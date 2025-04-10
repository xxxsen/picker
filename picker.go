package picker

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/thoas/go-funk"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"gopkg.in/yaml.v3"
)

var (
	defaultSysImport = []string{
		"host",
		"context",
		"fmt",
	}
)

type IPicker[T any] interface {
	Get(name string) (T, error)
	List() []string
}

type pickerImpl[T any] struct {
	i   *interp.Interpreter
	m   map[string]T
	lst []string
}

func ParseData[T any](data []byte) (IPicker[T], error) {
	p := &Plugins{}
	if err := yaml.Unmarshal(data, p); err != nil {
		return nil, err
	}
	pk := &pickerImpl[T]{
		i: interp.New(interp.Options{}),
		m: make(map[string]T),
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
	m := make(map[string]interface{}, len(ps.Plugins))
	lst := make([]string, 0, len(ps.Plugins))
	ct := asContainer(m)
	for idx, item := range ps.Plugins {
		if err := p.validateConfig(item); err != nil {
			return fmt.Errorf("validate plugin config failed, idx:%d, name:%s, err:%w", idx, item.Name, err)
		}
		if err := p.createPlugin(ct, item); err != nil {
			return fmt.Errorf("create plugin failed, name:%s, err:%w", item.Name, err)
		}
		lst = append(lst, item.Name)
	}
	for k, v := range m {
		vt, ok := v.(T)
		if !ok {
			return fmt.Errorf("plugin:%s function type not match", k)
		}
		p.m[k] = vt
	}
	p.lst = lst
	return nil
}

func (p *pickerImpl[T]) createPlugin(ct IContainer, plg *PluginConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("create plugin panic: %v, stack:%s", r, string(debug.Stack()))
		}
	}()
	args, err := p.buildArgs(plg)
	if err != nil {
		return fmt.Errorf("build template args failed, err:%w", err)
	}
	code, err := buildCode(args)
	if err != nil {
		return fmt.Errorf("build code failed, err:%w", err)
	}
	_, err = p.i.Eval(code)
	if err != nil {
		return fmt.Errorf("eval plugin code failed, err:%w", err)
	}
	v, err := p.i.Eval(fmt.Sprintf("%s.%s", args.Package, "Register"))
	if err != nil {
		return fmt.Errorf("eval register func failed, err:%w", err)
	}
	fn, ok := v.Interface().(func(IContainer) error)
	if !ok {
		return fmt.Errorf("register func type not match signature")
	}
	if err := fn(ct); err != nil {
		return fmt.Errorf("register func failed, err:%w", err)
	}
	return nil
}

func (p *pickerImpl[T]) buildArgs(plg *PluginConfig) (*pluginTpltArgs, error) {
	args := &pluginTpltArgs{
		Package:  fmt.Sprintf("picker_%s", plg.Name),
		Name:     plg.Name,
		Import:   p.withSysImport(p.asLine(plg.Import)),
		Define:   p.asLine(plg.Define),
		Function: plg.Function,
	}
	return args, nil
}

func (p *pickerImpl[T]) withSysImport(in []string) []string {
	rs := make([]string, 0, len(in)+len(defaultSysImport))
	rs = append(rs, defaultSysImport...)
	rs = append(rs, in...)
	return funk.Uniq(rs).([]string)
}

func (p *pickerImpl[T]) validateConfig(pc *PluginConfig) error {
	if len(pc.Name) == 0 {
		return fmt.Errorf("no plugin name found")
	}
	if len(pc.Function) == 0 {
		return fmt.Errorf("no plugin function found")
	}

	return nil
}

func (p *pickerImpl[T]) asLine(in string) []string {
	lines := strings.Split(strings.TrimSpace(in), "\n")
	rs := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		rs = append(rs, line)
	}
	return rs
}

func (p *pickerImpl[T]) Get(name string) (T, error) {
	fn, ok := p.m[name]
	if !ok {
		return fn, fmt.Errorf("plugin:%s not found", name)
	}
	return fn, nil
}

func (p *pickerImpl[T]) List() []string {
	return p.lst
}
