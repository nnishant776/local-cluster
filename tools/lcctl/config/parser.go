package config

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/nnishant776/local-cluster/pkg/conv"
	"github.com/nnishant776/local-cluster/pkg/model"
)

func Parse(filename string) (*model.Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ParseStream(f)
}

func ParseStream(inStream io.Reader) (*model.Config, error) {
	outStream := bytes.NewBuffer(nil)
	err := conv.YamlToJson(outStream, inStream)
	if err != nil {
		return nil, err
	}

	cfg := &model.Config{}
	err = json.NewDecoder(outStream).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
