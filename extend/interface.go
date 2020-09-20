package extend

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"log"
	"net/http"
	"net/http/httputil"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/whytheplatypus/reagent/hypothesis"
)

func init() {
	gob.Register(map[string]interface{}{})
}

func AssertablePlugins() plugin.PluginSet {
	return map[string]plugin.Plugin{
		"jsonschema": &AssertablePlugin{},
	}
}

type Assertable interface {
	Assert(*http.Response, map[string]interface{}) error
}

type AssertableRPC struct{ client *rpc.Client }

func (g *AssertableRPC) Assert(resp *http.Response, args map[string]interface{}) error {
	log.Println("calling plugin")
	bresp, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	var r interface{}
	return g.client.Call("Plugin.Assert", map[string]interface{}{
		"resp": bresp,
		"args": args,
	}, &r)
}

// Here is the RPC server that AssertableRPC talks to, conforming to
// the requirements of net/rpc
type AssertableRPCServer struct {
	// This is the real implementation
	Impl hypothesis.Assertable
}

func (s *AssertableRPCServer) Assert(args map[string]interface{}, resp *interface{}) error {
	bresp := args["resp"].([]byte)
	r, err := http.ReadResponse(bufio.NewReader(bytes.NewBuffer(bresp)), nil)
	if err != nil {
		return err
	}
	return s.Impl(r, args["args"].(map[string]interface{}))
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a AssertableRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return AssertableRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type AssertablePlugin struct {
	// Impl Injection
	Impl hypothesis.Assertable
}

func (p *AssertablePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &AssertableRPCServer{Impl: p.Impl}, nil
}

func (AssertablePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &AssertableRPC{client: c}, nil
}
