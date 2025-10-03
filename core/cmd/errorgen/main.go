package main

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sync"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage:", os.Args[0], `[options] [packages...]

gen searches Go packages for source files with struct literals that have
undeclared exported types from the same package and attempts to generate the
struct declaration. Declarations are put in output files named gen_types.go and
gen_types_test.go in the same package, configurable via --name.

Packages are either absolute or relative (beginning with a . or .. element)
paths to package directories, or names of packages that will be searched for in
GOPATH. If no packages are provided, then the package in the current directory
is processed. If a package contains one or more "..." wildcards, then it will
emulate the behavior of the go application (see 'go help packages').

options:`)
	flag.PrintDefaults()
}

var (
	namep = flag.String("name", "gen_types", "base name of output files")
	vp    = flag.Bool("v", false, "verbose output")
	base  = flag.String("base", "", "project root")

	pwd string
)

func main() {
	flag.Usage = usage
	flag.Parse()
	paths := flag.Args()
	if len(paths) == 0 {
		// Default to current directory.
		paths = []string{"."}
	}

	var err error
	pwd, err = os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot get current working directory:", err)
		os.Exit(1)
	}
	pkgs, ok := resolve(*base, paths)
	if !ok {
		// Errors have already been logged by resolve.
		os.Exit(1)
	}
	if len(pkgs) == 0 {
		// Wildcards matched no packages: this is not necessarily an
		// error so just return.
		return
	}

	// Simply start a parse-walk-generate goroutine for each package: there
	// will not be so many source files that we need to set up a worker
	// pool with a pipeline.
	var wg sync.WaitGroup
	var lock sync.Mutex // Used to synchronize access to ok.
	for _, pkg := range pkgs {
		if *vp {
			fmt.Println("checking", pkg.ImportPath)
		}

		wg.Add(1)
		go func(pkg *build.Package) {
			pok := pwg(pkg)
			lock.Lock()
			ok = ok && pok
			lock.Unlock()
			wg.Done()
		}(pkg)
	}
	wg.Wait()
	if !ok {
		// Errors have already been logged by pwg.
		os.Exit(1)
	}
}

// pwg parses the package AST, walks it looking for struct literals that need
// generation, and generates the missing type declarations.
func pwg(pkg *build.Package) (ok bool) {
	fset, parsed, ok := parse(pkg)
	if !ok {
		return
	}
	literals, ok := walk(pkg.ImportPath, fset, parsed)
	if !ok {
		return
	}
	if len(literals) > 0 {
		ok = generate(pkg, literals)
	}
	return
}

// rel attempts to convert the path to a relative path from the current working
// directory.
func rel(path string) string {
	if r, err := filepath.Rel(pwd, path); err == nil {
		return r
	}
	return path
}
