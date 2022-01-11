.POSIX:
.SUFFIXES:
.SUFFIXES: .1 .5 .7 .1.scd .5.scd .7.scd

VPATH=doc
PREFIX?=/usr/local
BINDIR?=$(PREFIX)/bin
MANDIR?=$(PREFIX)/share/man
GO?=go
GOFLAGS?=

GOSRC:=$(shell find . -name '*.go')
GOSRC+=go.mod go.sum

photon: $(GOSRC)
	$(GO) build $(GOFLAGS) -o $@

DOCS := \
	photon.1 \
	photon.5 \
	photon-lua.7

.1.scd.1:
	scdoc < $< > $@

.5.scd.5:
	scdoc < $< > $@

.7.scd.7:
	scdoc < $< > $@

doc: $(DOCS)

all: photon doc

clean:
	rm -rf $(DOCS) photon

install: all
	mkdir -m755 -p $(DESTDIR)$(BINDIR) $(DESTDIR)$(MANDIR)/man1 $(DESTDIR)$(MANDIR)/man5 $(DESTDIR)$(MANDIR)/man7 \
		$(DESTDIR)$(SHAREDIR)
	install -m755 photon $(DESTDIR)$(BINDIR)/photon
	install -m644 photon.1 $(DESTDIR)$(MANDIR)/man1/photon.1
	install -m644 photon.5 $(DESTDIR)$(MANDIR)/man5/photon.5
	install -m644 photon-lua.7 $(DESTDIR)$(MANDIR)/man7/photon-lua.7

uninstall:
	rm -rf $(DESTDIR)$(BINDIR)/photon
	rm -rf $(DESTDIR)$(MANDIR)/man1/photon.1
	rm -rf $(DESTDIR)$(MANDIR)/man5/photon.5
	rm -rf $(DESTDIR)$(MANDIR)/man7/photon-lua.7
