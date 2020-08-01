package main_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/whytheplatypus/reagent/experiment"
)

var (
	things = []map[string]interface{}{}
	ts     *httptest.Server
)

func TestCRUD(t *testing.T) {
	vars := map[string]string{
		"host": ts.URL,
	}
	trial, err := experiment.NewTrial("crud", vars, "examples/crud.toml")
	if err != nil {
		t.Fatal(err)
	}

	if err := trial.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	ts = configureServer()
	defer ts.Close()
	// call flag.Parse() here if TestMain uses flags
	live := flag.Bool("live", false, "keep the test server alive")
	flag.Parse()
	if *live {
		fmt.Println(ts.URL)
		waitFor(syscall.SIGINT, syscall.SIGTERM)
	}
	os.Exit(m.Run())
}

func waitFor(calls ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, calls...)
	<-sigs
}

func handleThings(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		d := json.NewDecoder(r.Body)
		defer r.Body.Close()
		var t map[string]interface{}
		if err := d.Decode(&t); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		things = append(things, t)
		rw.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
		fmt.Fprintf(rw, "{\"id\": %d}", len(things)-1)
	default:
		http.Error(rw, "unsupported", http.StatusMethodNotAllowed)
	}
}

func handleThing(rw http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.Split(r.URL.Path, "/")[2])
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(things) < id+1 {
		http.NotFound(rw, r)
		return
	}

	t := things[id]
	switch r.Method {
	case http.MethodGet:
		res, err := json.Marshal(t)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(rw, "%s", string(res))
	case http.MethodDelete:
		things = append(things[:id], things[id+1:]...)
		rw.WriteHeader(http.StatusCreated)
	default:
		http.Error(rw, "unsupported", http.StatusMethodNotAllowed)
	}
}

func configureServer() *httptest.Server {
	r := http.NewServeMux()
	r.HandleFunc("/things/", func(rw http.ResponseWriter, r *http.Request) {
		switch idPart := strings.Split(r.URL.Path, "/")[2]; idPart {
		case "":
			handleThings(rw, r)
		default:
			handleThing(rw, r)
		}
	})
	ts := httptest.NewServer(r)
	return ts
}
