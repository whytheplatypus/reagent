package assertions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/whytheplatypus/errgroup"
)

var asserters map[string]Assertable = map[string]Assertable{}

type Assertable interface {
	Assert(*http.Response, map[string]interface{}) error
}

type Testable interface {
	Do() (*http.Response, error)
	Assertions() map[string]map[string]interface{}
}

func Parallell(ctx context.Context, t Testable) error {
	res, err := t.Do()
	if err != nil {
		return err
	}
	var g errgroup.Group
	for asrt, args := range t.Assertions() {
		g.Go(func(asrt string, args map[string]interface{}) func() error {
			asserter := asserters[asrt]
			return func() error {
				if asserter == nil {
					return fmt.Errorf("%s : asserter is not registered.", ctx.Value(t))
				}
				if err := asserter.Assert(res, args); err != nil {
					resp, _ := httputil.DumpResponse(res, true)
					req, _ := httputil.DumpRequest(res.Request, true)
					return fmt.Errorf("%v : %s : %s : %s", ctx.Value(t), err, string(req), string(resp))
				}
				log.Printf("[DEBUG] assertion passed %s %s \n", ctx.Value(t), asrt)
				return nil
			}
		}(asrt, args))
	}
	return g.Wait()
}
