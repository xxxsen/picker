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
      context
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
	pk, err := ParseData[func(ctx context.Context, args interface{}) error]([]byte(testYamlFile))
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
