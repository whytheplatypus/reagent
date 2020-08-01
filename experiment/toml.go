package experiment

import (
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

func decodeTOML(r io.Reader) (steps, []string, error) {
	var ts steps
	md, err := toml.DecodeReader(r, &ts)
	if err != nil {
		return nil, nil, err
	}
	return ts, orderSteps(md.Keys()), nil
}

func orderSteps(keys []toml.Key) []string {
	orderedKeys := []string{}
	for _, k := range keys {
		if len(strings.Split(k.String(), ".")) < 2 {
			orderedKeys = append(orderedKeys, k.String())
		}
	}
	return orderedKeys
}
