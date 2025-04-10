package main

//Note the additional imports
import (
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// Note 'custom' lib was imported
const src = `
package pickertestplugin1744293217
import "custom"
func Bar(d *custom.Data) string { 
	return d.Message + "-Foo" 
}
`

type Data struct {
	Message string
}

func main() {
	d := Data{
		Message: "Kung",
	}

	i := interp.New(interp.Options{})

	//This allows use of standard libs
	i.Use(stdlib.Symbols)

	//This will make a 'custom' lib  available that can be imported and contains your Data struct
	custom := make(map[string]map[string]reflect.Value)
	custom["custom/custom"] = make(map[string]reflect.Value)
	custom["custom/custom"]["Data"] = reflect.ValueOf((*Data)(nil))
	i.Use(custom)

	_, err := i.Eval(src)
	if err != nil {
		panic(err)
	}

	v, err := i.Eval("foo.Bar")
	if err != nil {
		panic(err)
	}

	bar := v.Interface().(func(*Data) string)

	r := bar(&d)
	println(r)
}
