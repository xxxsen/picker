package picker

type PluginConfig struct {
	Name     string   `yaml:"name" json:"name"`
	Import   []string `yaml:"import" json:"import"`
	Define   string   `yaml:"define" json:"define"`
	Function string   `yaml:"function" json:"function"`
}

type Plugins struct {
	Plugins []*PluginConfig `yaml:"plugins" json:"plugins"`
	Import  []string        `yaml:"import" json:"import"`
}

type pluginTpltArgs struct {
	Package  string
	Name     string
	Import   []string
	Define   []string
	Function string
}
