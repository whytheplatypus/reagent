package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/whytheplatypus/reagent/experiment"
)

var exitCode = 0

var errBadStringMapVar = errors.New("Variables must be of the form key=value")

type stringMapVar map[string]string

func (av stringMapVar) String() string {
	return ""
}

func (av stringMapVar) Set(s string) error {
	args := strings.SplitN(s, "=", 2)
	if len(args) != 2 {
		return errBadStringMapVar
	}
	av[args[0]] = args[1]
	return nil
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.Usage = func() {
		fmt.Fprintf(flags.Output(), "Usage of %s:\n %s [flags] [files]\nFlags:\n", os.Args[0], os.Args[0])
		flags.PrintDefaults()
	}
	vars := stringMapVar{}
	var verbose bool
	flags.BoolVar(&verbose, "v", false, "Enable for verbose logging")
	flags.Var(&vars, "var", "Variables to use in the rendering of test files. Must be of the form key=value")
	flags.Parse(os.Args[1:])
	tf := flags.Args()
	if verbose {
		log.SetFlags(log.LstdFlags)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	var wg sync.WaitGroup
	for _, f := range tf {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			t, err := experiment.NewTrial(f, vars, f)
			if err != nil {
				exitCode = 1
				fmt.Println(t.Name, err)
				return
			}
			if err := t.Run(); err != nil {
				exitCode = 1
				log.Println(t.Name, err)
				return
			}
			log.Println("PASS:", f)
		}(f)
	}
	wg.Wait()
	os.Exit(exitCode)
}
