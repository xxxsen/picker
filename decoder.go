package picker

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type DecoderFunc func(data []byte, v interface{}) error

var (
	JsonDecoder DecoderFunc = json.Unmarshal
	YamlDecoder DecoderFunc = yaml.Unmarshal
)
