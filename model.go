package picker

type PluginConfig struct {
	Name     string   `yaml:"name" json:"name" toml:"name"`
	Import   []string `yaml:"import" json:"import" toml:"import"`
	Define   string   `yaml:"define" json:"define" toml:"define"`
	Function string   `yaml:"function" json:"function" toml:"function"`
}

type Plugins struct {
	Plugins []*PluginConfig `yaml:"plugins" json:"plugins" toml:"plugins"`
	Import  []string        `yaml:"import" json:"import" toml:"import"`
}

type pluginTpltArgs struct {
	Package  string
	Name     string
	Import   []string
	Define   string
	Function string
}
