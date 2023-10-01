package in

import "moddoc/mod"

// notExported is A function that is not exported.
func notExported() {
	_ = 1
}

// Exported is A function that is exported.
//
// Here I am testing A package import link to [mod.Module]
// And here I am testing A link to A local, moved, function [ILikeMyType]
func Exported() {
	_ = 2
	_ = mod.Module{}
}
