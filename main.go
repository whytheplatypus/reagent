package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/hashicorp/go-plugin"
	"github.com/whytheplatypus/reagent/experiment"
	"github.com/whytheplatypus/reagent/extend"
	"github.com/whytheplatypus/reagent/hypothesis"
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
		log.SetFlags(log.LstdFlags | log.Llongfile)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	loadPlugins()
	log.Println("plugins loaded")
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

func loadPlugins() {
	var handshakeConfig = plugin.HandshakeConfig{
		ProtocolVersion:  1,
		MagicCookieKey:   "BASIC_PLUGIN",
		MagicCookieValue: "hello",
	}
	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          extend.AssertablePlugins(),
		Cmd:              exec.Command("sh", "-c", "plugins/jsonschema/jsonschema"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
	})
	//defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("jsonschema")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	kv := raw.(extend.Assertable)
	hypothesis.Register("jsonschema", kv.Assert)
}
