// Package main is the entry point for the moddoc application.
//
// moddoc outputs static html documentation for the given source directory.
//
// See the [moddoc/mod.Module] structure for the structure that is passed to the main template for processing.
package main

import (
	"flag"
	"fmt"
	"log"
	"moddoc/mod"
	"moddoc/tmpl"
	"os"
	"path/filepath"
	"text/template"
)

var packageTemplateFlag = flag.String("pTmpl", "", "The path to a custom package page template.")
var indexTemplateFlag = flag.String("iTmpl", "", "The path to a custom index page template.")
var sourcePathFlag = flag.String("i", "", "The path to the module directory. Should have a go.mod file. Will use current working directory by default.")
var outPathFlag = flag.String("o", "", "The output directory. Will use current working directory by default.")
var outputTemplatesFlag = flag.Bool("t", false, "Will write out the default index.tmpl file and package.tmpl. Will output to the directory specified in the -o flag.")

func main() {
	flag.Parse()

	srcDir := *sourcePathFlag
	if srcDir == "" {
		var err error
		if srcDir, err = os.Getwd(); err != nil {
			log.Fatal(err)
		}
	}

	outDir := *outPathFlag
	if outDir == "" {
		var err error
		if outDir, err = os.Getwd(); err != nil {
			log.Fatal(err)
		}
	}

	if *outputTemplatesFlag {
		err := outputTemplates(outDir)
		if err != nil {
			log.Fatal(err)
		}

	}

	var err error
	packageTemplate := template.New("packageTemplate")
	if *packageTemplateFlag != "" {
		packageTemplate, err = packageTemplate.ParseFiles(*packageTemplateFlag)
		if err != nil {
			log.Fatalf("error opening package template %s", *packageTemplateFlag)
			return
		}
	} else {
		packageTemplate, err = packageTemplate.Parse(tmpl.PackageTemplate)
		if err != nil {
			log.Fatalf("error parsing default package template")
			return
		}
	}

	indexTemplate := template.New("indexTemplate")
	if *indexTemplateFlag != "" {
		indexTemplate, err = indexTemplate.ParseFiles(*indexTemplateFlag)
		if err != nil {
			log.Fatalf("error opening index template %s", *indexTemplateFlag)
			return
		}
	} else {
		indexTemplate, err = indexTemplate.Parse(tmpl.IndexTemplate)
		if err != nil {
			log.Fatalf("error parsing default index template")
			return
		}
	}

	m := mod.NewModule(srcDir)

	if err := createDirectoryIfNotExists(outDir); err != nil {
		log.Fatalf("error creating output directory: %s", err)
		return
	}

	for k := range m.Packages {
		execPackageTemplate(packageTemplate, m.Packages, k, outDir)
	}

	execModuleTemplate(indexTemplate, m, outDir)
}

func createDirectoryIfNotExists(directoryPath string) error {
	// Check if the directory already exists
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		// Directory does not exist, so create it
		err := os.MkdirAll(directoryPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	return nil
}

func execPackageTemplate(t *template.Template, pkgs map[string]*mod.Package, pkgKey string, outDir string) {
	pkg := pkgs[pkgKey]
	filePath := filepath.Join(outDir, pkg.FileName)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatalf("error opening file %s", filePath)
	}
	defer file.Close()
	err = t.Execute(file, pkg)
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

func outputTemplates(outDir string) error {
	filePath := filepath.Join(outDir, "index.tmpl")
	if err := writeFile(tmpl.IndexTemplate, filePath); err != nil {
		return err
	}

	filePath = filepath.Join(outDir, "package.tmpl")
	if err := writeFile(tmpl.PackageTemplate, filePath); err != nil {
		return err
	}
	return nil
}

func writeFile(inContent, outFile string) error {
	return os.WriteFile(outFile, []byte(inContent), 0644)
}
