package mod

import (
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"path"
	"path/filepath"
	"strings"
)

// StdLibSource is the URL for documentation of packages outside the module.
const StdLibSource = `https://pkg.go.dev/`

type HTMLer interface {
	toHTML(pkg *Package) string
}

// Package is a deconstruction of a package into its documentation parts for relatively easy consumption by a template.
//
// All strings listed are not escaped.
// Call HTML on an item to convert it to html.
type Package struct {
	DocPkg *doc.Package
	// ModuleName is the name of the module take from the go.mod file, and is also the root of the import path.
	ModuleName  string
	Path        string
	Name        string
	ImportPath  string
	Synopsis    string
	CommentHtml string
	FileName    string
	Consts      []Constant
	ConstGroups []ConstantGroup
}

// HTML should be called from within a template to convert the passed item to html.
func (p *Package) HTML(t any) string {
	switch v := t.(type) {
	case string:
		return string(p.DocPkg.HTML(v))
	case HTMLer:
		return v.toHTML(p)
	default:
		panic("cannot convert this type to HTML")
	}
}

func NewPackage(p *doc.Package, dirPath string, importRoot string) *Package {
	n := new(Package)
	n.DocPkg = p
	n.ModuleName = importRoot
	n.Name = p.Name
	n.Synopsis = p.Synopsis(p.Doc)
	n.ImportPath = p.ImportPath
	n.Path = dirPath + `/`
	n.FileName = makeFileName(importRoot, dirPath, p.Name)
	n.CommentHtml = n.parseHtmlComment(p.Doc)
	n.parseConstants()
	return n
}

func (p *Package) parseHtmlComment(text string) string {
	parser := p.DocPkg.Parser().Parse(text)
	printer := p.DocPkg.Printer()

	printer.DocLinkURL = func(link *comment.DocLink) (url string) {
		if strings.HasPrefix(link.ImportPath, p.ModuleName) {
			// A local package, so use the file name
			l := strings.TrimPrefix(link.ImportPath, p.ModuleName)
			l = strings.TrimPrefix(l, "/")
			url = makeFileName(p.ModuleName, l, p.Name)
		} else if strings.ContainsRune(link.ImportPath, '.') {
			// A path to a url
			url = "http://" + link.ImportPath
		} else {
			// assume this is a link to the standard library, so point to online doc
			url = StdLibSource + link.ImportPath
		}

		if link.Name != "" {
			if link.Recv != "" {
				url += "#" + link.Recv + "." + link.Name
			} else {
				url += "#" + link.Name
			}
		}
		return
	}
	c := printer.HTML(parser)
	return string(c)
}

func makeFileName(importRoot string, dirPath string, packageName string) string {
	var fileName string
	if dirPath == "." || dirPath == "" {
		fileName = path.Base(importRoot)
	} else {
		fileName = strings.ReplaceAll(dirPath, string(filepath.Separator), "_")
		fileName := strings.TrimPrefix(fileName, "._")
		if filepath.Base(dirPath) != packageName {
			// a package that is not named the same as its directory, so append the package name to keep the file name unique
			fileName += "_" + packageName
		}
	}
	return fileName + ".html"
}

func (p *Package) parseConstants() {
	for _, c := range p.DocPkg.Consts {
		if len(c.Names) == 1 {
			var c2 Constant
			c2.Name = c.Names[0]
			c2.Comment = p.parseHtmlComment(c.Doc)
			c2.Value = c.Decl.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
			p.Consts = append(p.Consts, c2)
		}
	}
}
