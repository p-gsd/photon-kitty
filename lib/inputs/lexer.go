package inputs

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var (
	eof = rune(0)
)

type itemType int

const (
	itemError itemType = iota
	itemEOF
	itemURL
	itemCmd
	itemComment
)

type item struct {
	typ itemType
	val string
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

//stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	input              *bufio.Reader   // the data being scanned.
	buf                strings.Builder //the data already scanned
	line, pos, prevpos int
	items              chan item // channel of scanned items.
}

func lex(input io.Reader) *lexer {
	l := &lexer{
		input: bufio.NewReader(input),
		items: make(chan item, 5),
	}
	go l.run() // Concurrently run state machine.
	return l
}

//run lexes the input by executing state functions until the state is nil.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.buf.String()}
	l.buf.Reset()
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (l *lexer) read() rune {
	ch, _, err := l.input.ReadRune()
	if ch == '\n' {
		l.line++
		l.prevpos = l.pos
		l.pos = 0
	} else {
		l.pos++
		if err != nil {
			return eof
		}
	}
	l.buf.WriteRune(ch)
	return ch
}

// unread places the previously read rune back on the reader.
func (l *lexer) unread() {
	l.input.UnreadRune()
	if l.pos == 0 {
		l.pos = l.prevpos
		if l.line == 0 {
			panic("Cannot unread! No runes readed")
		}
		l.line--
	} else {
		l.pos--
	}
	buf := l.buf.String()
	l.buf.Reset()
	if len(buf) > 0 {
		l.buf.WriteString(buf[:len(buf)-1])
	}
}

//peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.read()
	l.unread()
	return r
}

//acceptRun consumes a run of runes from the valid set.
func (l *lexer) accept(valid string) {
	for strings.ContainsRune(valid, l.read()) {
	}
	l.unread()
}

func (l *lexer) acceptWhitespace() {
	l.accept(" \t\n\r")
	l.buf.Reset()
}

//acceptToLineBreak reads entire string to line break
func (l *lexer) acceptToLineBreak() {
	for {
		if ch := l.read(); ch == eof {
			break
		} else if ch == '\n' {
			l.unread()
			break
		}
	}
}

//readLetters reads all runes that are letters
func (l *lexer) readLetters() string {
	var buf strings.Builder
	for {
		ch := l.read()
		if ch == eof {
			break
		} else if !unicode.IsLetter(ch) {
			l.unread()
			break
		}
		buf.WriteRune(ch)
	}
	return buf.String()
}

func lexStart(l *lexer) stateFn {
	t := l.readLetters()
	switch t {
	case "cmd":
		return lexCommand
	case "http", "https":
		return lexURL
	default:
		r := l.peek()
		switch r {
		case '#':
			return lexComment
		case eof:
			l.emit(itemEOF)
			return nil
		case '\n', '\r':
			l.acceptWhitespace()
			l.buf.Reset()
			return lexStart
		}
		return l.errorf("unexpected characters (%s)", t)
	}
}

func lexCommand(l *lexer) stateFn {
	if r := l.read(); r != ':' {
		return l.errorf("unexpected character after cmd (%r) expected colon (:)", r)
	}
	if r := l.read(); r != '/' {
		return l.errorf("unexpected character after cmd: (%r) expected slash (/)", r)
	}
	if r := l.read(); r != '/' {
		return l.errorf("unexpected character after cmd:/ (%r) expected slash (/)", r)
	}
	l.acceptToLineBreak()
	l.emit(itemCmd)
	return lexStart
}

func lexURL(l *lexer) stateFn {
	if r := l.read(); r != ':' {
		return l.errorf("unexpected character after %s (%r) expected colon (:)", l.buf.String(), r)
	}
	if r := l.read(); r != '/' {
		return l.errorf("unexpected character after %s (%r) expected slash (/)", l.buf.String(), r)
	}
	if r := l.read(); r != '/' {
		return l.errorf("unexpected character after %s (%r) expected slash (/)", l.buf.String(), r)
	}
	l.acceptToLineBreak()
	l.emit(itemURL)
	return lexStart
}

func lexComment(l *lexer) stateFn {
	l.buf.Reset()
	l.acceptToLineBreak()
	l.emit(itemComment)
	return lexStart
}

//errorf returns an error token and terminates the scan
//by passing back a nil pointer that will be the next
//state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf("%d:%d:"+format, append([]interface{}{l.line, l.pos}, args...)...),
	}
	return nil
}
