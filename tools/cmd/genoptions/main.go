package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/lestrrat-go/codegen"
	"github.com/lestrrat-go/xstrings"
)

var objectsFile = flag.String(`objects`, `objects.yaml`, `specify file containing object definitions`)

func main() {
	flag.Parse()

	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func writeComment(o *codegen.Output, comment string) bool {
	comment = strings.TrimSpace(comment)
	if comment == "" {
		return false
	}
	for i, line := range strings.Split(comment, "\n") {
		if i == 0 {
			o.LL(`// %s`, line)
		} else {
			o.L(`// %s`, line)
		}
	}
	return true
}

type Objects struct {
	Output      string
	PackageName string   `yaml:"package_name"`
	Imports     []string `yaml:"imports"`
	Interfaces  []*struct {
		Name         string
		Comment      string
		ConcreteType string `yaml:"concrete_type"`
		Methods      []string
		Embeds       []string
	} `yaml:"interfaces"`
	Options []*struct {
		Ident         string
		OptionName    string `yaml:"option_name"` // usually "With" + $Ident
		SkipOption    bool   `yaml:"skip_option"`
		Interface     string
		ConcreteType  string
		Comment       string
		ArgumentType  string `yaml:"argument_type"`
		ConstantValue string `yaml:"constant_value"`
	} `yaml:"options"`
}

func _main() error {
	var objects Objects

	{
		buf, err := os.ReadFile(*objectsFile)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(buf, &objects); err != nil {
			return err
		}
	}

	for _, iface := range objects.Interfaces {
		if iface.ConcreteType == "" {
			iface.ConcreteType = xstrings.LcFirst(iface.Name)
		}
		if len(iface.Methods) == 0 {
			iface.Methods = append(iface.Methods, iface.ConcreteType)
		}
	}

	for _, option := range objects.Options {
		if option.OptionName == "" {
			option.OptionName = `With` + option.Ident
		}
		if option.ConcreteType == "" {
			option.ConcreteType = xstrings.LcFirst(option.Interface)
		}
	}

	sort.Slice(objects.Interfaces, func(i, j int) bool {
		return objects.Interfaces[i].Name < objects.Interfaces[j].Name
	})
	sort.Slice(objects.Options, func(i, j int) bool {
		return objects.Options[i].Ident < objects.Options[j].Ident
	})

	if err := genOptions(&objects); err != nil {
		return fmt.Errorf(`failed to generate %q`, objects.Output)
	}

	if err := genOptionTests(&objects); err != nil {
		return fmt.Errorf(`failed to generate tests for %q`, objects.Output)
	}

	return nil
}

func genOptions(objects *Objects) error {
	var buf bytes.Buffer

	o := codegen.NewOutput(&buf)

	o.L("// This file is auto-generated by tools/cmd/genoptions/main.go. DO NOT EDIT")

	o.LL(`package %s`, objects.PackageName)

	imports := append(objects.Imports, []string{
		`github.com/lestrrat-go/option`,
	}...)
	o.WriteImports(imports...)

	o.LL(`type Option = option.Interface`)

	for _, iface := range objects.Interfaces {
		if writeComment(o, iface.Comment) {
			o.L(`type %s interface {`, iface.Name)
		} else {
			o.LL(`type %s interface {`, iface.Name)
		}
		if len(iface.Embeds) < 1 {
			o.L(`Option`)
		} else {
			for _, embed := range iface.Embeds {
				o.L(embed)
			}
		}

		for _, method := range iface.Methods {
			o.L(`%s()`, method)
		}
		o.L(`}`)

		o.LL(`type %s struct {`, iface.ConcreteType)
		o.L(`Option`)
		o.L(`}`)

		for _, method := range iface.Methods {
			o.LL(`func (*%s) %s() {}`, iface.ConcreteType, method)
		}
	}

	o.L(``)

	{
		seen := make(map[string]struct{})
		for _, option := range objects.Options {
			_, ok := seen[option.Ident]
			if !ok {
				o.L(`type ident%s struct{}`, option.Ident)
				seen[option.Ident] = struct{}{}
			}
		}
	}

	{
		seen := make(map[string]struct{})
		for _, option := range objects.Options {
			_, ok := seen[option.Ident]
			if !ok {
				o.LL(`func (ident%s) String() string {`, option.Ident)
				o.L(`return %q`, option.OptionName)
				o.L(`}`)
				seen[option.Ident] = struct{}{}
			}
		}
	}

	for _, option := range objects.Options {
		if option.SkipOption {
			continue
		}

		if writeComment(o, option.Comment) {
			o.L(`func %s(`, option.OptionName)
		} else {
			o.LL(`func %s(`, option.OptionName)
		}
		if argType := option.ArgumentType; argType != "" {
			o.R(`v %s`, argType)
		}
		o.R(`) %s {`, option.Interface)

		value := `v`
		if cv := option.ConstantValue; cv != "" {
			value = cv
		}

		o.L(`return &%s{option.New(ident%s{}, %s)}`, option.ConcreteType, option.Ident, value)
		o.L(`}`)
	}

	if err := o.WriteFile(objects.Output, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return fmt.Errorf(`failed to write to headers_gen.go: %w`, err)
	}
	return nil
}

func genOptionTests(objects *Objects) error {
	var buf bytes.Buffer

	o := codegen.NewOutput(&buf)

	o.L("// This file is auto-generated by tools/cmd/genoptions/main.go. DO NOT EDIT")

	o.LL(`package %s`, objects.PackageName)

	o.LL(`func TestOptionIdent(t *testing.T) {`)
	seen := make(map[string]struct{})
	for _, option := range objects.Options {
		_, ok := seen[option.Ident]
		if ok {
			continue
		}

		o.L(`require.Equal(t, %q, ident%s{}.String())`, option.OptionName, option.Ident)
		seen[option.Ident] = struct{}{}
	}

	o.L(`}`)

	filename := strings.Replace(objects.Output, `.go`, `_test.go`, -1)
	if err := o.WriteFile(filename, codegen.WithFormatCode(true)); err != nil {
		if cfe, ok := err.(codegen.CodeFormatError); ok {
			fmt.Fprint(os.Stderr, cfe.Source())
		}
		return fmt.Errorf(`failed to write to headers_gen.go: %w`, err)
	}

	return nil
}
