// Package mod processes a directory with a go.mod file and extracts the documentation.
//
// This process complements the [go/doc] package in that it
//   - Specifically is designed to output HTML, and
//   - It creates a structure that is easily consumed by Go templates.
package mod

import (
	"go/doc"
	"go/parser"
	"go/token"
	"golang.org/x/mod/modfile"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Module represents the documentation for an entire module.
//
// Module is structured to be easily consumed by Go templates.
type Module struct {
	// Name is the module name extracted from the go.mod file.
	Name string
	// DirName is the name of the directory holding the module. This is not always the same, but often is.
	DirName string
	// Packages is the documentation for all the packages in the module.
	// The first package in the list represents the package in the same
	// directory as the go.mod file, if there is a package there.
	Packages map[string]*Package
}

// NewModule walks a module directory, returning a Module structure.
//
// The directory dirPath should contain a go.mod file.
func NewModule(modPath string) *Module {
	m := new(Module)
	importPath := getImportPath(modPath)
	m.Name = path.Base(importPath)
	m.DirName = filepath.Base(modPath)

	var dirPaths []string

	err := filepath.WalkDir(modPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if d.Name()[0] == '.' {
				return filepath.SkipDir
			}
			if d.Name() == "internal" {
				return filepath.SkipDir
			}
			dirPaths = append(dirPaths, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error examining directory: %s", err)
	}

	if len(dirPaths) == 0 {
		log.Fatalf("no directories found")
	}

	m.Packages = getPackages(dirPaths, modPath, m)
	return m
}

func getImportPath(modPath string) string {
	modPath = filepath.Join(modPath, "go.mod")
	if _, err := os.Stat(modPath); !os.IsNotExist(err) {
		b, err := os.ReadFile(modPath)
		if err != nil {
			log.Fatalf("could not open %s:%s", modPath, err)
		}
		f, err := modfile.Parse("go.mod", b, nil)
		if err != nil {
			log.Fatalf("could not parse %s:%s", modPath, err)
		}
		return f.Module.Mod.Path
	} else {
		log.Fatalf("could not find go.mod file")
	}
	return ""
}

func getPackages(dirPaths []string, modPath string, module *Module) (pkgs map[string]*Package) {
	pkgs = make(map[string]*Package)

	for _, dirPath := range dirPaths {
		fset := token.NewFileSet()
		parsedPackages, err := parser.ParseDir(fset, dirPath, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("parse error: %s, %s", dirPath, err)
		}

		// A directory may have multiple packages.
		// Often this is used for test packages.
		for _, pkg := range parsedPackages {
			// Do not deal with unit test files for now.
			if strings.HasSuffix(pkg.Name, "_test") {
				continue
			}

			relPath, _ := filepath.Rel(modPath, dirPath)
			pkgImportPath := path.Join(module.Name, relPath)

			docPkg := doc.New(pkg, pkgImportPath, 0)
			if err != nil {
				log.Fatalf("doc package error: %s, %s", pkg.Name, err)
			}

			p := NewPackage(docPkg, fset, relPath, module)
			if p != nil {
				pkgs[p.Path] = p
			}
		}
	}

	// Take another pass through the packages and build each the PathParts
	for path, pkg := range pkgs {
		parts := strings.Split(path, string(filepath.Separator))
		pathParts := []PathPart{
			{
				DirName: module.DirName,
				DocFile: "index.html",
			},
		}
		for i, part := range parts {
			link := ""
			pkgPath := filepath.Join(parts[:i+1]...)
			if p2, ok := pkgs[pkgPath]; ok {
				link = p2.FileName
			}
			pathParts = append(pathParts, PathPart{
				DirName: part,
				DocFile: link,
			})
		}
		pkg.PathParts = pathParts
	}

	return
}
