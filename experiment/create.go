package experiment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/whytheplatypus/reagent/hypothesis"
)

type steps map[string]*step

type step struct {
	Name    string
	URL     string
	Assert  map[string]map[string]interface{}
	Headers http.Header
	Method  string
	Body    string
}

func (s *step) Assertions() map[string]map[string]interface{} {
	return s.Assert
}

func (s *step) Do() (*http.Response, error) {
	req, err := http.NewRequest(s.Method, s.URL, bytes.NewBuffer([]byte(s.Body)))
	if err != nil {
		return nil, err
	}
	req.Header = s.Headers
	return http.DefaultClient.Do(req)
}

var funcMap = template.FuncMap{
	// The name "title" is what the function will be called in the template text.
	"json": parseJSON,
}

func parseJSON(b []byte, key string) interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		return nil
	}
	return result[key]
}

// NewTrial creates a new trial from the toml file and variables provided.
func NewTrial(name string, vars map[string]string, path string) (*Trial, error) {
	t := &Trial{
		Variables: vars,
		Name:      name,
	}
	if err := t.ParseFile(path); err != nil {
		return t, err
	}
	return t, nil
}

// Trial describes a set of behaviors an API is expected to exibit
// and the results of testing those behaviors.
type Trial struct {
	Variables map[string]string
	Name      string
	steps     map[string]*step
	keys      []string
	results   map[string]interface{}
}

// ParseFile parses a TOML file with this trials variables.
func (t *Trial) ParseFile(path string) error {
	tmp, err := template.ParseFiles(path)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer([]byte{})
	if err := tmp.Execute(b, t.Variables); err != nil {
		return err
	}
	steps, keys, err := decodeYAML(b.Bytes())
	if err != nil {
		return err
	}
	t.steps = steps
	t.keys = keys
	return nil
}

// Run executes the trial, recording the results and returning an error
// if the hypothesis in the trial are disproven.
func (t *Trial) Run() error {
	t.results = map[string]interface{}{}
	for _, name := range t.keys {
		step := t.steps[name]

		//template URL, Body

		b := bytes.NewBuffer([]byte{})
		if err := template.Must(template.New(name).Delims("${", "}").Funcs(funcMap).Parse(step.URL)).Execute(b, t.results); err != nil {
			return err
		}
		step.URL = b.String()

		result, err := hypothesis.Check(step)
		if err != nil {
			return err
		}
		t.results[name] = result
	}
	return nil
}
