package hypothesis

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
)

type assertable func(*http.Response, map[string]interface{}) error

var asserters = map[string]assertable{}

type testable interface {
	Do() (*http.Response, error)
	Assertions() map[string]map[string]interface{}
}

var (
	// ErrUnregisteredAsserter is thrown if the type of assertion is not recognized.
	// currentlly supported assertions are `jsonschema` and `code`
	ErrUnregisteredAsserter = errors.New("asserter is not registered")
)

// Check a hypothesis. Returns the content of the body of the API response and, if the hypothesis is disproven, an error.
func Check(t testable) (result []byte, err error) {

	res, err := t.Do()
	if err != nil {
		return nil, err
	}

	for asrt, args := range t.Assertions() {
		asserter, ok := asserters[asrt]
		if !ok {
			return nil, ErrUnregisteredAsserter
		}
		if err := asserter(res, args); err != nil {
			resp, _ := httputil.DumpResponse(res, true)
			req, _ := httputil.DumpRequest(res.Request, true)
			return nil, fmt.Errorf("%s : %s : %s", err, string(req), string(resp))
		}
	}
	result, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return
}
