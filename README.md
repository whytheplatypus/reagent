# reagent [![GoDoc](https://godoc.org/github.com/whytheplatypus/reagent?status.svg)](http://godoc.org/github.com/whytheplatypus/reagent) [![Report card](https://goreportcard.com/badge/github.com/whytheplatypus/reagent)](https://goreportcard.com/report/github.com/whytheplatypus/reagent)

A cli and library for declaring what how an API is expected to behave and checking that hypothesis.

## Installation
`go install github.com/whytheplatypus/reagent`

## Usage

Describe how you expect an API to behave (currently in YAML)

Variables can be set from the command-line with `-var` e.g. `-var host=<address>`.
This can be done multiple times `-var host=<address> -var token=<authToken>`
[examples/crud.yaml](/example/crud.yaml)

```
create_a_thing:
  name: "Create a thing"
  url: "{{ .host }}/things/"
  method: "POST"
  body: |
          {
                  "hello": "world",
                  "works": false
          }
  headers:
    Authorization: ["Bearer sample_bearer_token"]
```
State what you expect the API to do with this input.
```
  assert:
    response:
      code: 200
      body: '{"id": 0}'
      headers:
        Content-Type: ["application/json"]
```

Steps from the same file are run in order and results from previous steps can be used
e.g. `json .create_a_thing "id"` returns the value from the `id` key of the json response from the `create_a_thing` step.
```
read_a_thing:
  name: "Read a thing"
  url: "{{ .host }}/things/${ json .create_a_thing `id` }"
  method: "GET"
  headers:
    Authorization: ["Bearer sample_bearer_token"]
  assert:
    response:
      code: 200
    jsonschema:
      ref: "examples/thing.json"
```

For a full example run you can pull down the repository and use the server used for tests to experiment.
```
git clone github.com/whytheplatypus/reagent
cd reagent
```
Start the test server
```
go test -live -v .
```
This will output the port the test server is running on.

In another terminal run the example hypothesis.
```
go run . -v -var host=<address from go test> examples/crud.yaml
```
Try modifying [examples/crud.yaml](/examples/crud.yaml) to make the run fail.

