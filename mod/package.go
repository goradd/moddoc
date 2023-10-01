package mod

import (
	"bytes"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// StdLibSource is the URL for documentation of packages outside the module.
const StdLibSource = `https://pkg.go.dev/`

const hideCommand = "hide"
const typeCommand = "type"

type HTMLer interface {
	toHTML(pkg *Package) string
}

// Package is a deconstruction of a package into its documentation parts for relatively easy consumption by a template.
//
// All strings listed are not escaped.
// Call [HTML] on an item to convert it to html.
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
	Types       []*Type
	types       map[string]*Type // to manipulate the type after its inserted
	//paths       map[string]struct{} // the set of valid paths in the package to know if we can link to them
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
	cmt, flags := parseCommentFlags(p.Doc)
	if _, ok := flags[hideCommand]; ok {
		// We are being told to hide the package documentation completely
		return nil
	}
	n.types = make(map[string]*Type)
	n.CommentHtml = n.parseHtmlComment(cmt)
	n.parseConstants()
	n.parseVars()
	n.parseFuncs()
	n.parseTypes()
	n.applyFlags()
	return n
}

const docPrefix = "doc:"

// parseCommentFlags will search through the text and look for "doc:" commands, parse the commands, and return
// the text with the commands removed.
func parseCommentFlags(text string) (newText string, flags map[string]string) {
	// Not the fastest implementation, but the most supportable perhaps
	lines := strings.SplitAfter(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, docPrefix) {
			if flags == nil {
				flags = make(map[string]string)
			}
			parts := strings.Split(line[len(docPrefix):], "=")
			if len(parts) == 2 {
				flags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			} else {
				flags[strings.TrimSpace(parts[0])] = "" // a boolean flag
			}
		} else {
			newText += line
		}
	}
	return
}

func (p *Package) parseHtmlComment(text string) (html string) {
	// Parse out the flags and adjust the text

	parser := p.DocPkg.Parser().Parse(text)
	printer := p.DocPkg.Printer()

	printer.DocLinkURL = func(link *comment.DocLink) (url string) {
		if strings.HasPrefix(link.ImportPath, p.ModuleName) {
			// A local package, so use the file name
			l := strings.TrimPrefix(link.ImportPath, p.ModuleName)
			l = strings.TrimPrefix(l, "/")
			url = makeFileName(p.ModuleName, l, p.Name)
		} else if strings.ContainsRune(link.ImportPath, '.') {
			// Indicates this is a path to a URL
			url = "https://" + link.ImportPath
		} else if link.ImportPath != "" {
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
	html = string(c)
	return
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
	cmt, flags := parseCommentFlags(c.Doc)
	c2.Flags = flags
	c2.CommentHtml = p.parseHtmlComment(cmt)
	c2.Code, _ = p.generateCode(c.Decl)
	return c2
}

func (p *Package) parseConstants() {
	for _, c := range p.DocPkg.Consts {
		newC := p.parseConstant(c)
		if _, ok := newC.Flags[hideCommand]; !ok {
			p.Constants = append(p.Constants, newC)
		}
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

func (p *Package) generateCode(decl ast.Decl) (string, error) {
	// Create an AST node slice containing only the target GenDecl.
	astFile := &ast.File{
		Name:  ast.NewIdent("main"),
		Decls: []ast.Decl{decl},
	}

	buf := bytes.Buffer{}
	err := format.Node(&buf, p.Fset, astFile)
	if err != nil {
		return "", err
	}
	s := buf.String()
	// remove the package preamble
	return s[14:], nil
}

func (p *Package) parseVariable(v *doc.Value) Variable {
	var v2 Variable
	v2.Names = v.Names
	cmt, flags := parseCommentFlags(v.Doc)
	v2.Flags = flags
	v2.CommentHtml = p.parseHtmlComment(cmt)
	v2.Code, _ = p.generateCode(v.Decl)
	return v2
}

func (p *Package) parseVars() {
	for _, v := range p.DocPkg.Vars {
		newV := p.parseVariable(v)
		if _, ok := newV.Flags[hideCommand]; !ok {
			p.Variables = append(p.Variables, newV)
		}
	}
}

func (p *Package) parseFunction(f *doc.Func) Function {
	var f2 Function
	f2.Name = f.Name
	cmt, flags := parseCommentFlags(f.Doc)
	f2.Flags = flags
	f2.CommentHtml = p.parseHtmlComment(cmt)
	f2.Code = p.getCodeFragment(f.Decl.Pos(), f.Decl.End())
	return f2
}

func (p *Package) parseFuncs() {
	for _, f := range p.DocPkg.Funcs {
		newF := p.parseFunction(f)
		if _, ok := newF.Flags[hideCommand]; !ok {
			p.Functions = append(p.Functions, newF)
		}
	}
}

func (p *Package) parseMethod(f *doc.Func) Method {
	var f2 Method
	f2.Name = f.Name
	cmt, flags := parseCommentFlags(f.Doc)
	f2.Flags = flags
	f2.CommentHtml = p.parseHtmlComment(cmt)
	f2.Code, _ = p.generateCode(f.Decl)

	f2.Receiver = f.Recv
	f2.EmbeddedType = f.Orig
	f2.Level = f.Level
	return f2
}

func (p *Package) parseTypes() {
	for _, t := range p.DocPkg.Types {
		var t2 Type
		t2.Name = t.Name
		cmt, flags := parseCommentFlags(t.Doc)
		if _, ok := flags[hideCommand]; ok {
			continue // skip
		}
		t2.Flags = flags
		t2.CommentHtml = p.parseHtmlComment(cmt)
		t2.Code, _ = p.generateCode(t.Decl)

		for _, c := range t.Consts {
			item := p.parseConstant(c)
			if _, ok := item.Flags[hideCommand]; !ok {
				t2.Constants = append(t2.Constants, item)
			}
		}
		for _, v := range t.Vars {
			item := p.parseVariable(v)
			if _, ok := item.Flags[hideCommand]; !ok {
				t2.Variables = append(t2.Variables, item)
			}
		}
		for _, f := range t.Funcs {
			item := p.parseFunction(f)
			if _, ok := item.Flags[hideCommand]; !ok {
				t2.Functions = append(t2.Functions, item)
			}
		}

		for _, f := range t.Methods {
			item := p.parseMethod(f)
			if _, ok := item.Flags[hideCommand]; !ok {
				t2.Methods = append(t2.Methods, item)
			}
		}

		pT := &t2
		p.Types = append(p.Types, pT)
		p.types[t2.Name] = pT // to get to types by name
	}
}

// applyFlags will apply the flag values that were parsed earlier, deleting or moving specific objects as needed.
func (p *Package) applyFlags() {
	var newConstants []Constant
	for _, item := range p.Constants {
		if t := item.Flags[typeCommand]; t != "" {
			if pT := p.types[t]; pT != nil {
				pT.Constants = append(pT.Constants, item)
			} else {
				log.Printf("Warning: type %s not found in comment for constant %s", t, item.Names[0])
				newConstants = append(newConstants, item)
			}
		} else {
			newConstants = append(newConstants, item)
		}
	}
	p.Constants = newConstants

	var newVars []Variable
	for _, item := range p.Variables {
		if t := item.Flags[typeCommand]; t != "" {
			if pT := p.types[t]; pT != nil {
				pT.Variables = append(pT.Variables, item)
			} else {
				log.Printf("Warning: type %s not found in comment for variable %s", t, item.Names[0])
				newVars = append(newVars, item)
			}
		} else {
			newVars = append(newVars, item)
		}
	}
	p.Variables = newVars

	var newFuncs []Function
	for _, item := range p.Functions {
		if t := item.Flags[typeCommand]; t != "" {
			if pT := p.types[t]; pT != nil {
				pT.Functions = append(pT.Functions, item)
			} else {
				log.Printf("Warning: type %s not found in comment for function %s", t, item.Name)
				newFuncs = append(newFuncs, item)
			}
		} else {
			newFuncs = append(newFuncs, item)
		}
	}
	p.Functions = newFuncs
}

/*

func (p *Package) collectPaths() {
	p.paths = make(map[string]struct{})

	for _, c := range p.Constants {
		for _, n := range c.Names {
			p.paths[n] = struct{}{}
		}
	}
	for _, v := range p.Variables {
		for _, n := range v.Names {
			p.paths[n] = struct{}{}
		}
	}
	for _, f := range p.Functions {
		p.paths[f.Name] = struct{}{}
	}
	for _, t := range p.Types {
		for _, c := range t.Constants {
			for _, n := range c.Names {
				p.paths[t.Name+"."+n] = struct{}{}
			}
		}
		for _, v := range t.Variables {
			for _, n := range v.Names {
				p.paths[t.Name+"."+n] = struct{}{}
			}
		}
		for _, f := range t.Functions {
			p.paths[t.Name+"."+f.Name] = struct{}{}
		}
		for _, m := range t.Methods {
			p.paths[t.Name+"."+m.Name] = struct{}{}
		}
	}
}
*/
