package hypothesis

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func init() {
	asserters["response"] = assertResponse
}

var checks = map[string]func(*http.Response, interface{}) error{
	"code":       checkCode,
	"protomajor": checkProtoMajor,
	"protominor": checkProtoMinor,
	"headers":    checkHeader,
	"body":       checkBody,
}

func assertResponse(r *http.Response, args map[string]interface{}) error {
	for key, check := range checks {
		c, ok := args[key]
		if ok {
			if err := check(r, c); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkCode(r *http.Response, c interface{}) error {
	code, ok := c.(int)
	if !ok {
		return fmt.Errorf("Code must be an int %T", c)
	}
	if r.StatusCode != code {
		return fmt.Errorf("code not equal")
	}
	return nil
}

func checkProtoMajor(r *http.Response, c interface{}) error {
	proto, ok := c.(int64)
	if !ok {
		return fmt.Errorf("proto major must be an int %T", c)
	}
	if int64(r.ProtoMajor) != proto {
		return fmt.Errorf("proto major not equal")
	}
	return nil
}

func checkProtoMinor(r *http.Response, c interface{}) error {
	proto, ok := c.(int64)
	if !ok {
		return fmt.Errorf("proto minor must be an int %T", c)
	}
	if int64(r.ProtoMinor) != proto {
		return fmt.Errorf("proto minor not equal")
	}
	return nil
}

func checkHeader(r *http.Response, c interface{}) error {
	h, ok := c.(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("Headers must be of the form Header = [value] %T", c)
	}
	headers := http.Header{}
	for k, v := range h {
		for _, vv := range v.([]interface{}) {
			headers.Add(k.(string), vv.(string))
		}
	}
	for key := range headers {
		if headers.Get(key) != r.Header.Get(key) {
			return fmt.Errorf("%s header did not match; %s, %s", key, headers.Get(key), r.Header.Get(key))
		}
	}
	return nil
}

func checkBody(r *http.Response, c interface{}) error {
	b, ok := c.(string)
	if !ok {
		return fmt.Errorf("Expected body value could not be parsed %T", c)
	}
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(content))
	if string(content) != b {
		return fmt.Errorf("Response body did not match")
	}
	return nil
}
