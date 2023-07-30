package tmpl

import (
	_ "embed"
)

//go:embed package.tmpl
var PackageTemplate string

//go:embed mod.tmpl
var ModuleTemplate string
