package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type fieldType struct {
	tp        string
	nullValue string
}

func determineFields(fields *ast.FieldList) []fieldType {
	// for _, field := range fields.List{
	// 	switch field.Type{
	// 		case
	// 	}
	// }
	return nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("файл go.mod не найден в структуре проекта")
}

func findGoFiles(root string) ([]string, error) {

	goFiles := make([]string, 0, 10)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() != "." && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".go" && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Ошибка при сканировании: %v\n", err)
	}
	return goFiles, nil
}

func main() {
	root, _ := findProjectRoot()
	files, _ := findGoFiles(root)
	//fmt.Printf("GO файлы:", files)
	fset := token.NewFileSet()

	for _, file := range files {
		astFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("Ошибка в файле %s: %v\n", file, err)
			continue
		}

		info := types.Info{
			Defs:  make(map[*ast.Ident]types.Object),
			Types: make(map[ast.Expr]types.TypeAndValue),
		}

		config := types.Config{Importer: importer.Default()}

		_, err = config.Check("main", fset, []*ast.File{astFile}, &info)
		if err != nil {
			log.Fatalf("Ошибка проверки типов: %v", err)
		}

		ast.Inspect(astFile, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			genDecl, ok := n.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				return true
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Doc == nil || strings.Contains(typeSpec.Doc.Text(), "generate:reset") {
					continue
				}
				structDecl, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				determineFields(structDecl.Fields)

			}

			return true
		})
	}
}
