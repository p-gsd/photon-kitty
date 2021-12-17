package inputs

type Inputs []string

func (ins Inputs) Len() int {
	return len(ins)
}

func (ins Inputs) Get(i int) string {
	return ins[i]
}

func (ins Inputs) Set(i int, v string) {
	ins[i] = v
}
