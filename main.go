package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/whytheplatypus/errgroup"
	"github.com/whytheplatypus/tester/assertions"
	"github.com/whytheplatypus/tester/tests"
)

type StringMapVar map[string]string

func (av StringMapVar) String() string {
	return ""
}

func (av StringMapVar) Set(s string) error {
	args := strings.SplitN(s, "=", 2)
	if len(args) != 2 {
		return fmt.Errorf("Variables must be of the form key=value")
	}
	av[args[0]] = args[1]
	return nil
}

func main() {
	vars := StringMapVar{}
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Enable for verbose logging")
	flag.Var(&vars, "var", "Variables to use in the rendering of test files. Must be of the form key=value")
	flag.Parse()
	tf := flag.Args()
	if verbose {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	for _, f := range tf {
		t := template.Must(template.ParseFiles(f))
		b := bytes.NewBuffer([]byte{})
		if err := t.Execute(b, vars); err != nil {
			log.Fatal(err)
		}
		ts, err := tests.Decode(b)
		if err != nil {
			log.Fatal(err)
		}
		// render toml files
		// parse results
		// run each test
		ctx := context.Background()
		var g errgroup.Group
		for n, t := range ts {
			g.Go(func(t assertions.Testable, n string) func() error {
				c := context.WithValue(ctx, t, n)
				return func() error {
					return assertions.Parallell(c, t)
				}
			}(t, n))
		}
		if err := g.Wait(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	// assert response code
	// validate response json schema if present
	// collect errors
	// report errors
	// report statistics (request time)
}
