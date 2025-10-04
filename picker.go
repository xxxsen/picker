package picker

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/thoas/go-funk"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var (
	defaultSysImport = []string{
		"host",
		"context",
		"fmt",
	}
	defaultFnNameValidator = regexp.MustCompile("^[a-zA-Z0-9_]+$")
	defaultFnCodeValidator = regexp.MustCompile(`^\s*func\s*\(.*\).*`) //简单判断即可, 必须为 func(...) {} 这种类型, 不能带function name
)

type IPicker[T any] interface {
	Get(name string) (T, error)
	List() []string
}

type pickerImpl[T any] struct {
	i   *interp.Interpreter
	m   map[string]T
	lst []string
	c   *config
}

func checkIsTemplateTypeFunction[T any]() bool {
	var v T
	t := reflect.TypeOf(v)
	return t.Kind() == reflect.Func
}

func Load[T any](ps *Plugins, opts ...Option) (IPicker[T], error) {
	if !checkIsTemplateTypeFunction[T]() {
		return nil, fmt.Errorf("template type should be function")
	}
	c := applyOpts(opts...)
	pk := &pickerImpl[T]{
		i: interp.New(interp.Options{}),
		m: make(map[string]T),
		c: c,
	}
	if err := pk.i.Use(stdlib.Symbols); err != nil {
		return nil, fmt.Errorf("use stdlib failed, err:%w", err)
	}
	ct, err := buildInjectObjs(c.objs)
	if err != nil {
		return nil, fmt.Errorf("build inject object failed, err:%w", err)
	}
	if err := pk.i.Use(ct); err != nil {
		return nil, fmt.Errorf("use custom object failed, err:%w", err)
	}
	if err := pk.init(ps); err != nil {
		return nil, err
	}
	return pk, nil
}

func buildInjectObjs(objs map[string]interface{}) (map[string]map[string]reflect.Value, error) {
	ct := make(map[string]map[string]reflect.Value)
	ct["host/host"] = make(map[string]reflect.Value)
	ct["host/host"]["IContainer"] = reflect.ValueOf((*IContainer)(nil))
	for name, obj := range objs {
		if !strings.Contains(name, "/") { //没有pkg的, 自动注册到 host 中
			ct["host/host"][name] = reflect.Indirect(reflect.ValueOf(obj))
			continue
		}
		//如果存在pkg, 那么它的格式必定为 `$pkg_name/$object_name`, 仅能有一个`/`
		parts := strings.Split(name, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("custom object path should contains at most one slash, data:%s", name)
		}
		pkg := parts[0] + "/" + parts[0] //yaegi中, 任意一个pkg都得写成 `$pkg/$pkg` 的形式
		if _, ok := ct[pkg]; !ok {
			ct[pkg] = make(map[string]reflect.Value)
		}
		name = parts[1]
		ct[pkg][name] = reflect.Indirect(reflect.ValueOf(obj))
	}
	return ct, nil
}

func ParseData[T any](data []byte, dec DecoderFunc, opts ...Option) (IPicker[T], error) {
	ps := &Plugins{}
	if err := dec(data, ps); err != nil {
		return nil, fmt.Errorf("decode data failed, err:%w", err)
	}
	return Load[T](ps, opts...)
}

func ParseYamlFile[T any](f string, opts ...Option) (IPicker[T], error) {
	raw, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return ParseData[T](raw, YamlDecoder, opts...)
}

func ParseJsonFile[T any](f string, opts ...Option) (IPicker[T], error) {
	raw, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return ParseData[T](raw, JsonDecoder, opts...)
}

func ParseTomlFile[T any](f string, opts ...Option) (IPicker[T], error) {
	raw, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return ParseData[T](raw, TomlDecoder, opts...)
}

func (p *pickerImpl[T]) wrapFunc(name string, t T) T {
	if !p.c.safeWrap {
		return t
	}
	funcType := reflect.TypeOf(t)
	numOut := funcType.NumOut()
	errKind := reflect.TypeOf((*error)(nil)).Elem()
	errIndex := -1
	for i := 0; i < numOut; i++ {
		if funcType.Out(i).AssignableTo(errKind) {
			errIndex = i
			break
		}
	}
	if errIndex < 0 { //没有err返回值, 捕获了也没卵用
		return t
	}
	return reflect.MakeFunc(funcType, func(args []reflect.Value) (results []reflect.Value) {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic recover from plugin:%s, recover:%v, stack:%s", name, r, string(debug.Stack()))
				results = make([]reflect.Value, numOut)
				for i := 0; i < numOut; i++ {
					if funcType.Out(i).AssignableTo(errKind) {
						results[i] = reflect.ValueOf(err)
						continue
					}
					results[i] = reflect.Zero(funcType.Out(i))
				}
			}
		}()
		results = reflect.ValueOf(t).Call(args)
		return
	}).Interface().(T)
}

func (p *pickerImpl[T]) init(ps *Plugins) error {
	m := make(map[string]interface{}, len(ps.Plugins))
	lst := make([]string, 0, len(ps.Plugins))
	ct := asContainer(m)
	for idx, item := range ps.Plugins {
		if err := p.validateConfig(item); err != nil {
			return fmt.Errorf("validate plugin config failed, idx:%d, name:%s, err:%w", idx, item.Name, err)
		}
		if err := p.createPlugin(ps, ct, item); err != nil {
			return fmt.Errorf("create plugin failed, name:%s, err:%w", item.Name, err)
		}
		lst = append(lst, item.Name)
	}
	for k, v := range m {
		vt, ok := v.(T)
		if !ok {
			var t T
			return fmt.Errorf("plugin:%s function type not match, need:%s, current_type:%s", k, p.inspectFunc(t), p.inspectFunc(v))
		}
		p.m[k] = p.wrapFunc(k, vt)
	}
	p.lst = lst
	return nil
}

func (p *pickerImpl[T]) createPlugin(ps *Plugins, ct IContainer, plg *PluginConfig) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("create plugin panic: %v, stack:%s", r, string(debug.Stack()))
		}
	}()
	args, err := p.buildArgs(plg, ps.Import)
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
	v, err := p.i.Eval(fmt.Sprintf("%s.%s", args.Package, "Register_"))
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

func (p *pickerImpl[T]) inspectFunc(fn interface{}) string {
	funcType := reflect.TypeOf(fn)

	if funcType.Kind() != reflect.Func {
		return "signature: non-func type"
	}

	return fmt.Sprintf("signature: %s", funcType)
}

func (p *pickerImpl[T]) buildArgs(plg *PluginConfig, extraImport []string) (*pluginTpltArgs, error) {
	args := &pluginTpltArgs{
		Package:  fmt.Sprintf("picker_%s", plg.Name),
		Name:     plg.Name,
		Import:   p.withSysImport(p.removeEmpty(plg.Import), p.removeEmpty(extraImport)),
		Define:   plg.Define,
		Function: plg.Function,
	}
	return args, nil
}

func (p *pickerImpl[T]) withSysImport(items ...[]string) []string {
	rs := make([]string, 0, len(defaultSysImport)+8)
	rs = append(rs, defaultSysImport...)
	for _, imports := range items {
		rs = append(rs, imports...)
	}
	return funk.Uniq(rs).([]string)
}

func (p *pickerImpl[T]) validateConfig(pc *PluginConfig) error {
	if len(pc.Name) == 0 {
		return fmt.Errorf("no plugin name found")
	}
	if !defaultFnNameValidator.MatchString(pc.Name) {
		return fmt.Errorf("plugin name is invalid")
	}
	if len(pc.Function) == 0 {
		return fmt.Errorf("no plugin function found")
	}
	if !defaultFnCodeValidator.MatchString(pc.Function) {
		return fmt.Errorf("plugin function code is invalid, should be a func without func name")
	}
	return nil
}

func (p *pickerImpl[T]) removeEmpty(lines []string) []string {
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
