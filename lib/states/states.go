package states

type Enum int

const (
	Normal Enum = iota
	Article
	Search
)

type Func func() Enum
