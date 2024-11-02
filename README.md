# Path Utility Library

This Go library provides a robust and expressive suite of utilities for handling file system paths with ease. Building on the standard path/filepath package, it introduces a set of intuitive, chainable methods that simplify file and directory operations such as path manipulation, pattern matching, directory traversal, and more. With this library, developers can effortlessly perform common file operations in a clean, readable style, making it ideal for projects that require efficient and flexible file system interactions.

## Installation

To install the library, use `go get`:

```sh
go get github.com/maa3x/path
```

## Usage

Here is a quick example of how to use the library:

```go
package main

import (
  "fmt"
  "log"

  "github.com/maa3x/path"
)

func main() {
  p := path.New("/Users/matrix/Code/go/projects")
  p = p.Join("path", "README.md")

  exists, err := p.Exists()
  if err != nil {
    log.Fatal(err)
  }

  p2 := p.NthParent(2).Join("copy")
  if err := p.Dir().Copy(p2); err != nil {
    log.Fatal(err)
  }
}
```
