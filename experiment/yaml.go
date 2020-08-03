package experiment

import (
	"gopkg.in/yaml.v2"
)

func decodeYAML(b []byte) (steps, []string, error) {
	var s steps
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, nil, err
	}

	var ts yaml.MapSlice
	if err := yaml.Unmarshal(b, &ts); err != nil {
		return nil, nil, err
	}

	keys := make([]string, len(ts))
	for i, item := range ts {
		keys[i] = item.Key.(string)
	}

	return s, keys, nil
}
