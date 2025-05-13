package conv

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
	"io"
)

func YamlToJson(dst io.Writer, src io.Reader) error {
	d := map[string]any{}

	err := yaml.NewDecoder(src).Decode(d)
	if err != nil {
		return err
	}

	return json.NewEncoder(dst).Encode(d)
}
