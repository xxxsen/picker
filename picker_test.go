package picker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testYamlFile = `
plugins:
  - name: testplugin
    import: |
      fmt
    define: |
      var a = 1
      var b = 2
    function: |
      func(ctx context.Context, args interface{}) error {
        fmt.Printf("hello world, a:%d, b:%d\n", a, b)
        return nil
      } 		
`
)

func TestParseData(t *testing.T) {
	pk, err := ParseData[func(ctx context.Context, args interface{}) error]([]byte(testYamlFile), YamlDecoder)
	assert.NoError(t, err)
	assert.NotNil(t, pk)
	fn, err := pk.Get("testplugin")
	assert.NoError(t, err)
	err = fn(context.Background(), nil)
	assert.NoError(t, err)
	fns := pk.List()
	t.Logf("fns:%v", fns)
	assert.Equal(t, 1, len(fns))
}

func TestInvalidTemplateType(t *testing.T) {
	_, err := Load[struct{}](&Plugins{})
	assert.Error(t, err)
}

func TestExecPlugin(t *testing.T) {
	pgs := &Plugins{
		Plugins: []*PluginConfig{
			{
				Name:     "p1",
				Function: `func(ctx context.Context) {fmt.Printf("hello world\n")}`,
			},
			{
				Name:     "p2",
				Function: `func(ctx context.Context) {fmt.Printf("test\n")}`,
			},
			{
				Name:     "p3",
				Function: `func(ctx context.Context) {fmt.Printf("r u ok?\n")}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context)](pgs)
	assert.NoError(t, err)
	for _, name := range pk.List() {
		fn, err := pk.Get(name)
		assert.NoError(t, err)
		fn(context.Background())
	}
	assert.Equal(t, 3, len(pk.List()))
}

func TestJson(t *testing.T) {
	code := `{
		"plugins": [
			{
				"name": "testjson",
				"import": "fmt",
				"define": "var a = 1\nvar b = 2",
				"function": "func(ctx context.Context, args interface{}) error {fmt.Printf(\"hello json\")\n\treturn nil\n}"
			}
		]
	}`
	pk, err := ParseData[func(ctx context.Context, args interface{}) error]([]byte(code), JsonDecoder)
	assert.NoError(t, err)
	assert.NotNil(t, pk)
	fn, err := pk.Get("testjson")
	assert.NoError(t, err)
	err = fn(context.Background(), nil)
	assert.NoError(t, err)
	fns := pk.List()
	t.Logf("fns:%v", fns)
	assert.Equal(t, 1, len(fns))
}
