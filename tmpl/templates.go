// Package tmpl imports the default templates into the application.
package tmpl

import (
	_ "embed"
)

// PackageTemplate is the content of the per-package template.
//
//go:embed package.tmpl
var PackageTemplate string

// IndexTemplate is the content of the module template that will become the index.html file.
//
//go:embed index.tmpl
var IndexTemplate string
