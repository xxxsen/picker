package picker

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testYamlFile = `
plugins:
  - name: testplugin
    import:
      - fmt
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
				Define:   "a := 3.14",
				Function: `func(ctx context.Context) {fmt.Printf("r u ok? value a:%f\n", a)}`,
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
				"import": ["fmt"],
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

type TestData struct {
	V string
}

func TestCutsomObject(t *testing.T) {
	pgs := &Plugins{
		Plugins: []*PluginConfig{
			{
				Name:     "p1",
				Function: `func(ctx context.Context, in *host.TestData) {in.V = "hello world"}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context, in *TestData)](pgs,
		WithCustomObject(
			map[string]interface{}{
				"TestData": &TestData{},
			},
		),
	)
	assert.NoError(t, err)
	fn, _ := pk.Get("p1")
	in := &TestData{
		V: "123",
	}
	fn(context.Background(), in)
	t.Logf("in.V:%s", in.V)
	assert.Equal(t, "hello world", in.V)
}

func TestCustomObjectWithPkg(t *testing.T) {
	pgs := &Plugins{
		Plugins: []*PluginConfig{
			{
				Name:     "p1",
				Import:   []string{"cstpkg"},
				Function: `func(ctx context.Context, in *cstpkg.TestData) {in.V = "hello world"}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context, in *TestData)](pgs,
		WithCustomObject(
			map[string]interface{}{
				"cstpkg/TestData": &TestData{},
			},
		),
	)
	assert.NoError(t, err)
	fn, _ := pk.Get("p1")
	in := &TestData{
		V: "123",
	}
	fn(context.Background(), in)
	t.Logf("in.V:%s", in.V)
	assert.Equal(t, "hello world", in.V)
}

func TestWithSafeWrap(t *testing.T) {
	pgs := &Plugins{
		Plugins: []*PluginConfig{
			{
				Name:     "normal",
				Function: `func(ctx context.Context) (string, error) {return "123", nil}`,
			},
			{
				Name:     "panic",
				Function: `func(ctx context.Context) (string, error) {panic(1)}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context) (string, error)](pgs,
		WithSafeFuncWrap(true),
	)
	assert.NoError(t, err)
	fn, _ := pk.Get("normal")
	data, err := fn(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "123", data)
	fn, _ = pk.Get("panic")
	data, err = fn(context.Background())
	assert.Error(t, err)
	t.Logf("panic err:%v", err)
	assert.Equal(t, "", data)
}

func TestGlobalImport(t *testing.T) {
	pgs := &Plugins{
		Import: []string{"io", "bytes"},
		Plugins: []*PluginConfig{
			{
				Name: "normal",
				Function: `func(ctx context.Context) io.Reader {
					 return bytes.NewReader([]byte("hello world"))
				}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context) io.Reader](pgs)
	assert.NoError(t, err)
	fn, _ := pk.Get("normal")
	rd := fn(context.Background())
	data, err := io.ReadAll(rd)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(data))
}

func TestClosure(t *testing.T) {
	pgs := &Plugins{
		Import: []string{"io", "bytes"},
		Plugins: []*PluginConfig{
			{
				Name: "normal",
				Define: `
					var str = "hello world"
					cb := func(ctx context.Context) []byte {
						return []byte(str)
					}
				`,
				Function: `func(ctx context.Context) io.Reader {
					 return bytes.NewReader(cb(ctx))
				}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context) io.Reader](pgs)
	assert.NoError(t, err)
	fn, _ := pk.Get("normal")
	rd := fn(context.Background())
	data, err := io.ReadAll(rd)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(data))
}

func BenchmarkFuncCallWithYaegi(b *testing.B) {
	//BenchmarkFuncCallWithYaegi-12            1000000              1728 ns/op             552 B/op         15 allocs/op
	pgs := &Plugins{
		Plugins: []*PluginConfig{
			{
				Name:     "normal",
				Function: `func(ctx context.Context) (int, error) {return 123, nil}`,
			},
		},
	}
	pk, err := Load[func(ctx context.Context) (int, error)](pgs)
	assert.NoError(b, err)
	fn, _ := pk.Get("normal")
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fn(ctx)
	}
}

func testNormalCall(ctx context.Context) (int, error) {
	return 123, nil
}

func BenchmarkFuncCallWithoutYaegi(b *testing.B) {
	//BenchmarkFuncCallWithoutYaegi-12        1000000000               0.1256 ns/op          0 B/op          0 allocs/op
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = testNormalCall(ctx)
	}
}
