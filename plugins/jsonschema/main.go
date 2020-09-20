package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/whytheplatypus/reagent/extend"

	"github.com/hashicorp/go-plugin"
	"github.com/xeipuuv/gojsonschema"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func main() {
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"jsonschema": &extend.AssertablePlugin{Impl: assertJSONSchema},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
	})
}

func assertJSONSchema(r *http.Response, args map[string]interface{}) error {
	log.Println("hello world")
	ref, ok := args["ref"].(string)
	if !ok {
		return fmt.Errorf("ref must be a string %T", ref)
	}
	schema, err := ioutil.ReadFile(ref)
	if err != nil {
		return err
	}
	schemaLoader := gojsonschema.NewStringLoader(string(schema))

	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(content))

	documentLoader := gojsonschema.NewStringLoader(string(content))

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var s string
		for _, desc := range result.Errors() {
			s = fmt.Sprintf("%s\n%s", s, desc)
		}
		return fmt.Errorf(s)
	}
	log.Println("finished check")
	return nil
}
