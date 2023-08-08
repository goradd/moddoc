package in

// notExported is a function that is not exported.
func notExported() {
	_ = 1
}

// Exported is a function that is exported.
func Exported() {
	_ = 2
}
