// Package in contains test data.
package in

// TestConstantString is a string typed test constant
const TestConstantString = "test string" // a string

// TestConstantInt is an int typed test constant
const TestConstantInt = 2 // an int

// A group of test constants
const (
	// TestOne is the first item in the test group
	TestOne = iota + 1
	TestTwo // This is a comment on the same line
)
