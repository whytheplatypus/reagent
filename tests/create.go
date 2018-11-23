package tests

import (
	"io"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
)

func Decode(r io.Reader) (tests, error) {
	var ts tests
	md, err := toml.DecodeReader(r, &ts)
	if err != nil {
		return nil, err
	}
	log.Println("[DEBUG]", md)
	log.Println("[DEBUG]", ts)
	return ts, nil
}

type tests map[string]*test

type test struct {
	Name    string
	URI     string
	Assert  map[string]map[string]interface{}
	Headers http.Header
	Method  string
}

func (t *test) Assertions() map[string]map[string]interface{} {
	return t.Assert
}

func (t *test) Do() (*http.Response, error) {
	req, err := http.NewRequest(t.Method, t.URI, nil)
	if err != nil {
		return nil, err
	}
	req.Header = t.Headers
	return http.DefaultClient.Do(req)
}
