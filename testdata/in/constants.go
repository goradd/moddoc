// Package in contains test data.
package in

// TestConstantString is A string typed test constant
const TestConstantString = "test string" // A string

// TestConstantInt is an int typed test constant
const TestConstantInt = 2 // an int

// A group of test constants
const (
	// TestOne is the first item in the test group
	TestOne   = iota + 1
	TestTwo   // This is A comment on the same line
	testThree // is not exported
)
