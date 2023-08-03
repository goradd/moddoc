package mod

import (
	"go/doc"
	"go/doc/comment"
	"go/token"
	"log"
	"os"
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
	// Fset is the fileset of the files in package.
	Fset *token.FileSet
	// ModuleName is the name of the module take from the go.mod file, and is also the root of the import path.
	ModuleName  string
	Path        string
	Name        string
	ImportPath  string
	Synopsis    string
	CommentHtml string
	FileName    string
	Constants   []Constant
	Variables   []Variable
	Functions   []Function
	Types       []Type
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

func NewPackage(p *doc.Package, fset *token.FileSet, dirPath string, importRoot string) *Package {
	n := new(Package)
	n.DocPkg = p
	n.Fset = fset
	n.ModuleName = importRoot
	n.Name = p.Name
	n.Synopsis = p.Synopsis(p.Doc)
	n.ImportPath = p.ImportPath
	n.Path = dirPath + `/`
	n.FileName = makeFileName(importRoot, dirPath, p.Name)
	n.CommentHtml = n.parseHtmlComment(p.Doc)
	n.parseConstants()
	n.parseVars()
	n.parseFuncs()
	n.parseTypes()
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

func (p *Package) parseConstant(c *doc.Value) Constant {
	var c2 Constant
	c2.Names = c.Names
	c2.CommentHtml = p.parseHtmlComment(c.Doc)
	c2.Code = p.getCodeFragment(c.Decl.Pos(), c.Decl.End())
	return c2
}

func (p *Package) parseConstants() {
	for _, c := range p.DocPkg.Consts {
		p.Constants = append(p.Constants, p.parseConstant(c))
	}
}

func (p *Package) getCodeFragment(start token.Pos, end token.Pos) string {
	startPosition := p.Fset.Position(start)
	endPosition := p.Fset.Position(end)
	if startPosition.Filename != endPosition.Filename {
		panic("filenames don't match. This is a programming error.")
	}
	f, err := os.Open(startPosition.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	b := make([]byte, endPosition.Offset-startPosition.Offset)
	_, err = f.ReadAt(b, int64(startPosition.Offset))
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func (p *Package) parseVariable(v *doc.Value) Variable {
	var v2 Variable
	v2.Names = v.Names
	v2.CommentHtml = p.parseHtmlComment(v.Doc)
	v2.Code = p.getCodeFragment(v.Decl.Pos(), v.Decl.End())
	return v2
}

func (p *Package) parseVars() {
	for _, v := range p.DocPkg.Vars {
		p.Variables = append(p.Variables, p.parseVariable(v))
	}
}

func (p *Package) parseFunction(f *doc.Func) Function {
	var f2 Function
	f2.Name = f.Name
	f2.CommentHtml = p.parseHtmlComment(f.Doc)
	f2.Code = p.getCodeFragment(f.Decl.Pos(), f.Decl.End())
	return f2
}

func (p *Package) parseFuncs() {
	for _, f := range p.DocPkg.Funcs {
		p.Functions = append(p.Functions, p.parseFunction(f))
	}
}

func (p *Package) parseMethod(f *doc.Func) Method {
	var f2 Method
	f2.Name = f.Name
	f2.CommentHtml = p.parseHtmlComment(f.Doc)
	f2.Code = p.getCodeFragment(f.Decl.Pos(), f.Decl.End())

	f2.Receiver = f.Orig
	f2.EmbeddedType = f.Recv
	f2.Level = f.Level
	return f2
}

func (p *Package) parseTypes() {
	for _, t := range p.DocPkg.Types {
		var t2 Type
		t2.Name = t.Name
		t2.CommentHtml = p.parseHtmlComment(t.Doc)
		t2.Code = p.getCodeFragment(t.Decl.Pos(), t.Decl.End())

		for _, c := range t.Consts {
			t2.Constants = append(p.Constants, p.parseConstant(c))
		}
		for _, v := range t.Vars {
			t2.Variables = append(t2.Variables, p.parseVariable(v))
		}
		for _, f := range t.Funcs {
			t2.Functions = append(t2.Functions, p.parseFunction(f))
		}

		for _, f := range t.Methods {
			t2.Methods = append(t2.Methods, p.parseMethod(f))
		}

		p.Types = append(p.Types, t2)
	}
}
