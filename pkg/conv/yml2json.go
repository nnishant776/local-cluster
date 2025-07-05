package conv

import (
	"encoding/json"

	"io"

	errstk "github.com/nnishant776/errstack"
	"gopkg.in/yaml.v3"
)

func YamlToJson(dst io.Writer, src io.Reader) error {
	d := map[string]any{}

	err := yaml.NewDecoder(src).Decode(d)
	if err != nil {
		return errstk.New(err, errstk.WithStack())
	}

	err = json.NewEncoder(dst).Encode(d)
	if err != nil {
		err = errstk.New(err, errstk.WithStack())
	}

	return err
}
