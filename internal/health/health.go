package health

type Meta struct {
	Name   string
	Value  string
	Passed bool
	Reason string
}

type Check struct {
	Passed bool
	Reason string
	Meta   []Meta
}

type Checker struct {
	Name  string
	Check func() Check
}
