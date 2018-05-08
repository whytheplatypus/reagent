package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/whytheplatypus/errgroup"
	"github.com/whytheplatypus/tester/assertions"
)

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Enable for verbose logging")
	flag.Parse()
	tf := flag.Args()
	if verbose {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	for _, f := range tf {
		var tests Tests
		md, err := toml.DecodeFile(f, &tests)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("[DEBUG]", md)
		log.Println("[DEBUG]", tests)
		// render toml files
		// parse results
		// run each test
		ctx := context.Background()
		var g errgroup.Group
		for n, t := range tests {
			g.Go(func(t Test, n string) func() error {
				c := context.WithValue(ctx, &t, n)
				return func() error {
					return assertions.Parallell(c, &t)
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

type Tests map[string]Test

type Test struct {
	Name    string
	URI     string
	Assert  map[string]map[string]interface{}
	Headers http.Header
	Method  string
}

func (t *Test) Assertions() map[string]map[string]interface{} {
	return t.Assert
}

func (t *Test) Do() (*http.Response, error) {
	req, err := http.NewRequest(t.Method, t.URI, nil)
	if err != nil {
		return nil, err
	}
	req.Header = t.Headers
	// TODO Customize method
	return http.DefaultClient.Do(req)
}
