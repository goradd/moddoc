# ModDoc
Outputs static html documentation for the given Go module source directory.

ModDoc will read a directory that has a go.mod file and output html documentation
files for all the Go source code files that are in the module's source tree. By default,
it will output an index.html file and an HTML file for each package.

This process is controlled by standard Go templates. By default, it uses templates
embedded in the application to produce an approximation of what go doc displays,
but you can provide your own templates to display your documentation however you like.

ModDoc is useful for the following situations:
1) You want a different style or content from what `go doc` displays.
2) You want to serve your documentation as static html, rather than running a go doc server.
3) You need to save the documentation separate from the source code for archiving or regulatory purposes.

## Installation
`go install github.com/goradd/moddoc`

## Requirements
- Go 1.18 or greater

## Usage

```shell
moddoc [options]
```

options:
- o: The output directory. By default, output goes to the current working directory.
- i: The input directory. Must have a go.mod file in that directory. By default will use the current working directory.
- iTmpl: The path to the index template file. By default, it will use its internal index template file. 
- pTmpl: The path to the package template file. By default, it will use its internal package template file.
- t: Instead of writing out the html, will output the default template files. You can use these as starting points for your custom template files. 

## Tags
Add the following to the bottom of a comment to prevent documentation from being
generated for that item. This works with package comments too:
```
\\ doc: hide
```

Add a "type=" specifier to a comment to assign the documentation for that item
to a particular struct type. The type given must be a struct type in the same
package.
```
\\ doc: type=MyType
```
For example, the following function would normally be included in package level
documentation, but the doc: type= tag will put it in the Animal section where
it logically belongs.

```
var animals []Animal

// func CountAnimals returns the number of animals found.
// doc: type=Animal
func CountAnimals() int {
    return len(animals)
}
```
## Styles
The default template imports the "styles.css" file. To style your documentation, create a styles.css
file and put it in the directory with the documentation. See a sample [here](styles.css).


## Contributions
Please submit your suggestions for improvement.

Also, if you create some great documentation templates, please share those and we
can incorporate those into the application for others to enjoy.
