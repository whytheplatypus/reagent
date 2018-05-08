package assertions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

func init() {
	asserters["jsonschema"] = &JSONSchema{}
}

type JSONSchema struct{}

func (jsa *JSONSchema) Assert(r *http.Response, args map[string]interface{}) error {
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

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. see errors :\n")
		var s string
		for _, desc := range result.Errors() {
			s = fmt.Sprintf("%s\n%s", s, desc)
		}
		return fmt.Errorf(s)
	}
	return nil
}
