package picker

import "testing"

func Test_buildTemplate(t *testing.T) {
	args := &pluginTpltArgs{
		Package:  "testpkg",
		Name:     "TestPlugin",
		Import:   []string{"context", "fmt"},
		Define:   []string{"var a = 1", "var b = 2"},
		Function: "func(ctx context.Context, args interface{}) error {\n\t\tfmt.Println(\"hello world\")\n\t\treturn nil\n\t}",
	}
	got, err := buildPluginCode(args)
	if err != nil {
		t.Fatalf("buildPlugin() error = %v", err)
	}
	t.Logf("data:%s", got)
}
