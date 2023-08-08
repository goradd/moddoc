# ModDoc
ModDoc outputs documentation for a module as static html using simple Go templates.

While go provides a built-in documentation server to serve your documentation through the source code,
this doesn't work for everyone. ModDoc is for situations where you would rather extract the documentation
and present it as html. Its great for situations where:

1. You want to present your documentation without having to host all the source code.
2. You need to extract the documentation and store it for regulatory purposes (FDA, MilSpec, etc.)
3. You would like to style your documentation differently.
4. You would like to change what items are output.
