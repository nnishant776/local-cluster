package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	errstk "github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/pkg/conv"
	"github.com/nnishant776/local-cluster/pkg/model"
)

func Parse(filename string) (*model.Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errstk.New(err, errstk.WithStack())
	}

	return ParseStream(f)
}

func ParseStream(inStream io.Reader) (*model.Config, error) {
	outStream := bytes.NewBuffer(nil)
	err := conv.YamlToJson(outStream, inStream)
	if err != nil {
		return nil, errstk.NewChainString(
			"parse: unable to convert yaml to json", errstk.WithStack(),
		).Chain(err)
	}

	cfg := &model.Config{}
	err = json.NewDecoder(outStream).Decode(cfg)
	if err != nil {
		return nil, errstk.NewChainString(
			"parse: unable to convert yaml to json", errstk.WithStack(),
		).Chain(err)
	}

	return cfg, nil
}
