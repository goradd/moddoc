package in

import "fmt"

// MyType is a test type
type MyType struct {
	// A is an exported type
	A int
	// b is an unexported type
	b int
	// C is also exported
	C string
}

// Init initializes the MyType object
func (m *MyType) Init() {
	m.A = 1
	m.b = Bstart
}

// private is a private method that should not generally appear in the doc unless asked for.
func (m *MyType) private() {
	m.A = 3
}

// ILikeMyType prints A message about MyType
//
// This is here to test the "type" doc command, that will associate this function with the [MyType] type.
//
// doc: type=MyType
func ILikeMyType() {
	fmt.Print("I like my types")
}

// Bstart is the initialized value of B
//
// doc: type=MyType
const Bstart = 4

// Dummy is A dummy value to test the hide command of moddoc
// doc: hide
const Dummy = 5

// MyiFace is an interface type
type MyiFace interface {
	// Here goes over here
	Here() // and this has an inline comment
	// There goes over there
	There()
	// where goes wherever
	where()
}
