package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

var errNoContent = errors.New("нет подходящих условий для генерации кода")

func generateSrcCodeForPkg(pkg *packages.Package, sb *strings.Builder) error {
	var hasContent = false
	fmt.Fprintf(sb, "package %s\n\n", pkg.Name)
	var genErr error = nil

	for _, fileAST := range pkg.Syntax {
		if genErr != nil {
			break
		}
		fileObj := pkg.Fset.File(fileAST.Pos())
		fileName := fileObj.Name()
		fmt.Printf("Сканируем файл: %s \n", fileName)
		err := generateSrcCodeForFile(fileAST, pkg.TypesInfo, sb)
		if err != nil {
			if !errors.Is(err, errNoContent) {
				genErr = fmt.Errorf("ошибка генерации кода для пакета %s: %v", pkg.Name, err)
			}
		} else {
			hasContent = true
		}
	}
	if genErr != nil {
		return genErr
	}
	if hasContent {
		return nil
	} else {
		return errNoContent
	}
}

func generateSrcCodeForFile(fileAST *ast.File, typeInfo *types.Info, sb *strings.Builder) error {
	hasContent := false
	var genErr error = nil
	ast.Inspect(fileAST, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if genErr != nil {
			return true
		}
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}
		if genDecl.Doc == nil || !strings.Contains(genDecl.Doc.Text(), "generate:reset") {
			return true
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				return true
			}
			structDecl, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			typeObj := typeInfo.Defs[typeSpec.Name]
			rn := determineReceiverName(typeObj.Type())

			fmt.Printf("===+++ генерация метода Reset() для структуры %s +++=== \n", typeSpec.Name.Name)
			if err := generateSrcCodeForStruct(structDecl, typeInfo, typeSpec.Name.Name, rn, sb); err != nil {
				if !errors.Is(err, errNoContent) {
					genErr = fmt.Errorf("ошибка генерации кода для cnhernehs %s: %v", typeSpec.Name.Name, err)
				}
			} else {
				hasContent = true
			}
		}
		return true
	})
	if genErr != nil {
		return genErr
	}
	if hasContent {
		return nil
	} else {
		return errNoContent
	}
}

func generateSrcCodeForStruct(
	st *ast.StructType,
	typesInfo *types.Info,
	structName string,
	rn string,
	sb *strings.Builder,
) error {
	var stSb strings.Builder
	fmt.Fprintf(&stSb, "func (%s *%s) Reset(){\n", rn, structName)
	hasContent := false

	for _, field := range st.Fields.List {
		for _, ident := range field.Names {
			typeObj := typesInfo.Defs[ident]
			if typeObj == nil {
				return fmt.Errorf("для поля структуры не определен тип")
			}
			fieldName := typeObj.Name()
			fieldType := typeObj.Type()

			if ptr, ok := fieldType.(*types.Pointer); ok {
				if generateSrcCodeForPointers(ptr, rn+"."+fieldName, &stSb) {
					hasContent = true
				}
			} else {
				if generateSrcCodeForNonPointer(fieldType, rn+"."+fieldName, &stSb) {
					hasContent = true
				}
			}
		}
	}
	if hasContent {
		sb.WriteString(stSb.String())
		sb.WriteString("}\n")
		return nil
	} else {
		return errNoContent
	}
}

func generateSrcCodeForPointers(ptr *types.Pointer, accessPoint string, sb *strings.Builder) bool {
	ptType := ptr.Elem()
	var psb strings.Builder
	hasContent := false

	if _, ok := ptType.Underlying().(*types.Struct); ok {
		if hasResetMethod(ptType) {
			fmt.Fprintf(&psb, "\tif %s != nil{\n", accessPoint)
			hasContent = generateSrcCodeForNonPointer(ptType, accessPoint, &psb)
			if hasContent {
				fmt.Fprintf(&psb, "}\n")
			}
		} else {
			return false
		}
	} else {
		fmt.Fprintf(sb, "\tif %s != nil{\n", accessPoint)
		hasContent = generateSrcCodeForNonPointer(ptType, "*"+accessPoint, sb)
		if hasContent {
			sb.WriteString("}\n")
		}
	}
	if hasContent {
		sb.WriteString(psb.String())
		return true
	} else {
		return false
	}
}

func generateSrcCodeForNonPointer(t types.Type, accessPoint string, sb *strings.Builder) bool {
	switch ft := t.Underlying().(type) {
	case *types.Map:
		fmt.Fprintf(sb, "\tclear(%s)\n", accessPoint)
		return true
	case *types.Slice:
		fmt.Fprintf(sb, "\t%s = %s[:0]\n", accessPoint, accessPoint)
		return true
	case *types.Struct:
		if hasResetMethod(t) {
			fmt.Fprintf(sb, "\t%s.Reset()\n", accessPoint)
			return true
		} else {
			return false
		}
	case *types.Basic:
		if ft.Kind() == types.String {
			fmt.Fprintf(sb, "\t%s = \"\"\n", accessPoint)
			return true
		} else if ft.Info()&types.IsNumeric != 0 {
			fmt.Fprintf(sb, "\t%s = 0\n", accessPoint)
			return true
		} else if ft.Kind() == types.Bool {
			fmt.Fprintf(sb, "\t%s = false\n", accessPoint)
			return true
		} else {
			return false
		}
	default:
		return false
	}
}

func determineReceiverName(t types.Type) string {
	named, ok := t.(*types.Named)
	if !ok {
		return "s"
	}

	if named.NumMethods() > 0 {
		firstMethod := named.Method(0)

		if sig, ok := firstMethod.Type().(*types.Signature); ok && sig.Recv() != nil {
			rn := sig.Recv().Name()
			if rn != "" {
				return rn
			}
		}
	}

	return "s"
}

func hasResetMethod(t types.Type) bool {
	obj, _, _ := types.LookupFieldOrMethod(t, true, nil, "Reset")
	if obj == nil {
		return false
	}
	if sig, ok := obj.Type().(*types.Signature); ok {
		return sig.Params().Len() == 0
	}
	return false
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

func writeSrcFile(sb *strings.Builder, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("не удалось создать файл %s: %w", filename, err)
	}

	defer file.Close()

	_, err = file.WriteString(sb.String())
	if err != nil {
		return fmt.Errorf("ошибка при записи в файл: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("ошибка синхронизации диска: %w", err)
	}

	fmt.Printf("Файл %s успешно сгенерирован и записан!\n", filename)
	return nil
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		log.Fatalf("не удалось определить корень проекта: %v", err)
	}
	cfg := &packages.Config{
		Dir:   root,
		Mode:  packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests: false,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			if filename == "reset.gen.go" {
				return nil, nil
			}
			const mode = parser.AllErrors | parser.ParseComments
			return parser.ParseFile(fset, filename, src, mode)
		},
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		log.Fatalf("Ошибка загрузки пакетов: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		log.Fatal("Обнаружены ошибки в коде проекта")
	}
	var sb strings.Builder

	for _, pkg := range pkgs {

		fmt.Printf("Сканируем пакет: %s \n", pkg.PkgPath)

		fmt.Printf("Удаляем файл reset.gen.go, если он существует: %s \n", pkg.PkgPath)
		err := os.Remove(path.Join(pkg.Dir, "reset.gen.go"))

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println("Файла не существовало, удалять нечего.")
			} else {
				log.Fatalf("Не удалось удалить файл: %v\n", err)
			}
		}

		sb.Reset()

		if err := generateSrcCodeForPkg(pkg, &sb); err != nil {
			if !errors.Is(err, errNoContent) {
				log.Fatalf("ошибка генерации кода: %v", err)
			}
		} else {
			writeSrcFile(&sb, path.Join(pkg.Dir, "reset.gen.go"))
		}
	}
}
