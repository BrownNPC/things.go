package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"

	j "github.com/dave/jennifer/jen"
)

func main() {
	names, err := ExportedStructs()
	if err != nil {
		fmt.Println(err)
		return
	}
	f := j.NewFile("t")
	f.PackageComment("This file is auto generated. DO NOT EDIT")

	for _, name := range names {
		poolName := "_cache" + name
		f.Var().Id(poolName).Id("cache").Types(j.Id(name))
		f.Func().Params(
			j.Id("p*").Id(name),
		).Id("onDestroy").Params().Block(
			j.Id(poolName).Dot("Store").Params(j.Id("p")),
		)
	}
	f.Comment("New creates a new Thing. It tries reusing the object from a cache to avoid allocations.")
	f.Func().Id("New").Types(j.Id("T any")).Params(
		j.Id("v T"),
	).Id("*T").
		Block(
			j.Switch(
				j.Id("arg := any(v).(type)"),
			).Block(func() []j.Code {
				ret := make([]j.Code, 0)
				for _, name := range names {
					poolName := "_cache" + name
					ret = append(ret,
						j.Case(j.Id(name)),
						j.Return(
							j.Id(poolName).Dot("New").Call(j.Id("arg")).Assert(j.Id("*T")),
						),
					)
				}
				ret = append(ret,
					j.Default().Return(j.Id("&v")),
				)
				return ret
			}()...),
		)
	file, err := os.Create("thing_allocation_caches.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := f.Render(file); err != nil {
		fmt.Println(err)
	}

}

func ExportedStructs() ([]string, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, ".", func(fi fs.FileInfo) bool {
		switch fi.Name() {
		case "thing_allocation_caches.go":
			return false
		}
		return true
	}, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}

	var structs []string

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					ts := spec.(*ast.TypeSpec)

					if !ts.Name.IsExported() {
						continue
					}

					if _, ok := ts.Type.(*ast.StructType); !ok {
						continue
					}

					structs = append(structs, ts.Name.Name)
				}
			}
		}
	}

	return structs, nil
}
