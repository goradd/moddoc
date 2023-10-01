package mod

// Constant represents a single line constant or a constant group declaration
type Constant struct {
	Code        string
	Names       []string
	CommentHtml string
	Flags       map[string]string
}

// Variable represents a variable declaration, or a group of variables declared together with the same type.
type Variable struct {
	Code        string
	Names       []string
	CommentHtml string
	Flags       map[string]string
}

// Function represents a simple top-level function that is not associated with a type.
type Function struct {
	Code        string
	Name        string
	CommentHtml string
	Flags       map[string]string
}

// Method represents a method associated with a type.
type Method struct {
	Code        string
	Name        string
	CommentHtml string

	// The type of the receiver
	Receiver string
	// EmbeddedType is the name of the type this method is associated with if its embedded in the main type.
	EmbeddedType string
	Level        int
	Flags        map[string]string
}

// Type represents a type definition.
// This could by a simple type, an interface, or a structure
type Type struct {
	Code        string
	Name        string
	CommentHtml string
	Flags       map[string]string
	Constants   []Constant
	Variables   []Variable
	Functions   []Function
	Methods     []Method
}
