package picker

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"gopkg.in/yaml.v3"
)

type IPicker[T any] interface {
	Get(name string) (T, error)
}

type pickerImpl[T any] struct {
	i  *interp.Interpreter
	ct IContainer
}

func ParseData[T any](data []byte) (IPicker[T], error) {
	p := &Plugins{}
	if err := yaml.Unmarshal(data, p); err != nil {
		return nil, err
	}
	pk := &pickerImpl[T]{
		i:  interp.New(interp.Options{}),
		ct: NewContainer(),
	}
	pk.i.Use(stdlib.Symbols)
	host := make(map[string]map[string]reflect.Value)
	host["host/host"] = make(map[string]reflect.Value)
	host["host/host"]["IContainer"] = reflect.ValueOf((*IContainer)(nil))
	pk.i.Use(host)
	if err := pk.init(p); err != nil {
		return nil, err
	}
	return pk, nil
}

func ParseFile[T any](f string) (IPicker[T], error) {
	raw, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return ParseData[T](raw)
}

func (p *pickerImpl[T]) init(ps *Plugins) error {
	argslist, err := p.buildPluginTemplateArgs(ps)
	if err != nil {
		return fmt.Errorf("build template args failed, err:%w", err)
	}
	for _, item := range argslist {
		if err := p.createPlugiuns(p.ct, item); err != nil {
			return fmt.Errorf("create plugin failed, err:%w", err)
		}
	}
	return nil
}

func (p *pickerImpl[T]) createPlugiuns(ct IContainer, args *pluginTpltArgs) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("create plugin panic: %v, stack:%s", r, string(debug.Stack()))
		}
	}()
	code, err := buildPluginCode(args)
	if err != nil {
		return fmt.Errorf("build code failed, name:%s, err:%w", args.Name, err)
	}
	_, err = p.i.Eval(code)
	if err != nil {
		return fmt.Errorf("eval plugin code failed, name:%s, err:%w", args.Name, err)
	}
	v, err := p.i.Eval(fmt.Sprintf("%s.%s", args.Package, "Register"))
	if err != nil {
		return fmt.Errorf("eval register func failed, name:%s, err:%w", args.Name, err)
	}
	fn, ok := v.Interface().(func(IContainer) error)
	if !ok {
		return fmt.Errorf("register func type not match signature, name:%s", args.Name)
	}
	if err := fn(ct); err != nil {
		return fmt.Errorf("register func failed, err:%w", err)
	}
	return nil
}

func (p *pickerImpl[T]) buildPluginTemplateArgs(ps *Plugins) ([]*pluginTpltArgs, error) {
	rs := make([]*pluginTpltArgs, 0, len(ps.Plugins))
	for _, plg := range ps.Plugins {
		args := &pluginTpltArgs{
			Package:  fmt.Sprintf("picker_%s_%d", plg.Name, time.Now().Unix()),
			Name:     plg.Name,
			Import:   p.noEmptyLine(strings.Split(plg.Import, "\n")),
			Define:   strings.Split(plg.Define, "\n"),
			Function: plg.Function,
		}
		rs = append(rs, args)
	}
	return rs, nil
}

func (p *pickerImpl[T]) noEmptyLine(in []string) []string {
	rs := make([]string, 0, len(in))
	for _, item := range in {
		if strings.TrimSpace(item) != "" {
			rs = append(rs, item)
		}
	}
	return rs
}

func (p *pickerImpl[T]) Get(name string) (T, error) {
	v, ok := p.ct.Get(name)
	if !ok {
		var zero T
		return zero, fmt.Errorf("plugin:%s not found", name)
	}
	vt, ok := v.(T)
	if !ok {
		return vt, fmt.Errorf("invalid plugin type, name:%s", name)
	}
	return vt, nil
}
