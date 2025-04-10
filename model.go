package picker

type PluginConfig struct {
	Name     string `yaml:"name"`
	Import   string `yaml:"import"`
	Define   string `yaml:"define"`
	Function string `yaml:"function"`
}

type Plugins struct {
	Plugins []*PluginConfig `yaml:"plugins"`
}

type pluginTpltArgs struct {
	Package  string
	Name     string
	Import   []string
	Define   []string
	Function string
}
