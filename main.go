// Package main is the entry point for the docmod application.
//
// docmod outputs static html documentation for the given source directory.
//
// See the [mod.Module] structure for the structure that is passed to the main template for processing.
package main

import (
	"docmod/mod"
	"docmod/tmpl"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var packageTemplateFlag = flag.String("pTmpl", "", "The package page template.")
var indexTemplateFlag = flag.String("iTmpl", "", "The index page template.")
var sourcePathFlag = flag.String("i", "", "The path to the module directory. Should have a go.mod file.")
var outPathFlag = flag.String("o", "", "The output directory.")

func main() {
	flag.Parse()

	srcDir := *sourcePathFlag
	if srcDir == "" {
		var err error
		if srcDir, err = os.Getwd(); err != nil {
			panic(err)
		}
	}

	outDir := *outPathFlag
	if outDir == "" {
		var err error
		if outDir, err = os.Getwd(); err != nil {
			panic(err)
		}
	}

	var err error
	packageTemplate := template.New("packageTemplate")
	if *packageTemplateFlag != "" {
		packageTemplate, err = packageTemplate.ParseFiles(*packageTemplateFlag)
		if err != nil {
			fmt.Printf("error opening package template %s", *packageTemplateFlag)
		}
	} else {
		packageTemplate, err = packageTemplate.Parse(tmpl.PackageTemplate)
		if err != nil {
			fmt.Printf("error parsing default package template")
		}
	}

	indexTemplate := template.New("indexTemplate")
	if *indexTemplateFlag != "" {
		indexTemplate, err = indexTemplate.ParseFiles(*indexTemplateFlag)
		if err != nil {
			fmt.Printf("error opening index template %s", *indexTemplateFlag)
		}
	} else {
		indexTemplate, err = indexTemplate.Parse(tmpl.IndexTemplate)
		if err != nil {
			fmt.Printf("error parsing default index template")
		}
	}

	m := mod.NewModule(srcDir)

	for _, p := range m.Packages {
		execPackageTemplate(packageTemplate, p, outDir)
	}

	execModuleTemplate(indexTemplate, m, outDir)
}

func execPackageTemplate(t *template.Template, p *mod.Package, outDir string) {
	filePath := filepath.Join(outDir, p.FileName)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("error opening file %s", filePath)
	}
	defer file.Close()
	err = t.Execute(file, p)
}

func execModuleTemplate(t *template.Template, m *mod.Module, outDir string) {
	filePath := filepath.Join(outDir, "index.html")
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("error opening file %s", filePath)
	}
	defer file.Close()
	err = t.Execute(file, m)
}
