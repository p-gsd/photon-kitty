package inputs

type Inputs []string

func (ins *Inputs) Len() int {
	return len(*ins)
}

func (ins *Inputs) Get(i int) string {
	return (*ins)[i]
}

func (ins *Inputs) Set(i int, v string) {
	(*ins)[i] = v
}

func (ins *Inputs) Add(i int, v string) {
	(*ins) = append((*ins)[:i-1], append(Inputs{v}, (*ins)[i-1:]...)...)
}

func (ins *Inputs) Append(v string) {
	(*ins) = append((*ins), Inputs{v}...)
}
