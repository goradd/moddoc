package tmpl

import (
	_ "embed"
)

// PackageTemplate is the content of the per-package template.
//
//go:embed package.tmpl
var PackageTemplate string

// ModuleTemplate is the content of the module template.
//
//go:embed mod.tmpl
var ModuleTemplate string
