package mod

type Constant struct {
	Name    string
	Value   string
	Comment string
}

type ConstantGroup struct {
	Comment string
	Consts  []Constant
}
