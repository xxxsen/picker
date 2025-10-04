package picker

import (
	"encoding/json"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type DecoderFunc func(data []byte, v interface{}) error

var (
	JsonDecoder DecoderFunc = json.Unmarshal
	YamlDecoder DecoderFunc = yaml.Unmarshal
	TomlDecoder DecoderFunc = toml.Unmarshal
)
