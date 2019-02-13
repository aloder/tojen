jenjen
======

jenjen is a code generator that generates
[jennifer](http://www.github.com/dave/jennifer) code from a existing file.

## Why?

Well writing code that generates code is tedious. This tool removes some of the
tedium by setting up a base that can be changed and extended to suit your needs
for code generation.


## How?

The command line is all you need.

go install github.com/aloder/jenjen

jenjen gen [source file]

This just takes the sourcefile and outputs the code in the terminal.

jenjen gen [source file] [output file]

This takes the source file and outputs the code in the specified file


