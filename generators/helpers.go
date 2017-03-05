package generators

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
)

func AddRoute(method, path, handlerName string) error {
	routeDefinition := fmt.Sprintf(`app.%v("%v", %v)`, method, path, handlerName)
	return AddInsideAppBlock(routeDefinition)
}

func AddInsideAppBlock(expressions ...string) error {
	src, err := ioutil.ReadFile("actions/app.go")
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "actions/app.go", string(src), 0)
	if err != nil {
		return err
	}

	srcContent := string(src)
	fileLines := strings.Split(srcContent, "\n")

	end := findClosingRouteBlockEnd(f, fset, fileLines)
	if end < 0 {
		return errors.New("could not find desired block on the app.go file")
	}

	for i := 0; i < len(expressions); i++ {
		expressions[i] = fmt.Sprintf("\t\t%s", expressions[i])
	}
	fileLines = append(fileLines[:end], append(expressions, fileLines[end:]...)...)

	fileContent := strings.Join(fileLines, "\n")
	err = ioutil.WriteFile("actions/app.go", []byte(fileContent), 0755)
	return err
}

func findClosingRouteBlockEnd(f *ast.File, fset *token.FileSet, fileLines []string) int {
	var end = -1

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BlockStmt:
			start := fset.Position(x.Lbrace).Line
			blockDeclaration := fmt.Sprintf("%s\n", fileLines[start-1])

			if strings.Contains(blockDeclaration, "if app == nil {") {
				end = fset.Position(x.Rbrace).Line - 1
			}

		}
		return true
	})

	return end
}

func AddImport(path string, imports ...string) error {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, string(src), 0)
	if err != nil {
		return err
	}

	srcContent := string(src)
	fileLines := strings.Split(srcContent, "\n")

	end := findLastImport(f, fset, fileLines)

	x := make([]string, len(imports), len(imports)+2)
	for _, i := range imports {
		x = append(x, fmt.Sprintf("\t\"%s\"", i))

	}
	if end < 0 {
		x = append([]string{"import ("}, x...)
		x = append(x, ")")
	}

	fileLines = append(fileLines[:end], append(x, fileLines[end:]...)...)

	fileContent := strings.Join(fileLines, "\n")
	err = ioutil.WriteFile(path, []byte(fileContent), 0755)
	return err
}

func findLastImport(f *ast.File, fset *token.FileSet, fileLines []string) int {
	var end = -1

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ImportSpec:
			end = fset.Position(x.End()).Line
			return true
		}
		return true
	})

	return end
}
