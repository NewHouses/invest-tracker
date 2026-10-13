package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// formatArg indica, para cada función ou método de tradución, que argumento é
// o texto en galego.
var formatArg = map[string]int{"T": 0, "Sprintf": 0, "Errorf": 0, "Fprintf": 1, "Fprint": 1, "Fprintln": 1}

type usage struct {
	key string
	pos string
}

// TestCatalog percorre todo o código Go (agás os tests) e comproba que cada
// texto que se traduce ten a súa tradución inglesa rexistrada, e que as dúas
// versións usan os mesmos formatos de fmt na mesma orde.
func TestCatalog(t *testing.T) {
	root := filepath.Join("..", "..")
	fset := token.NewFileSet()
	catalog := map[string]string{}
	catalogPos := map[string]string{}
	usages := map[string][]usage{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".tools", "frontend", "node_modules", "dist-installers":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			// Infórmase pero séguese: así un ficheiro a medio editar non
			// impide comprobar o resto dos paquetes.
			t.Run("sintaxe", func(t *testing.T) { t.Errorf("%s: %v", path, err) })
			return nil
		}
		dir, _ := filepath.Rel(root, filepath.Dir(path))
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			recv, ok := sel.X.(*ast.Ident)
			if !ok {
				// Tamén langOf(r).T("...") no servidor web.
				if inner, isCall := sel.X.(*ast.CallExpr); isCall {
					if fn, isIdent := inner.Fun.(*ast.Ident); isIdent && fn.Name == "langOf" {
						recv = &ast.Ident{Name: "lang"}
					}
				}
			}
			if recv == nil || (recv.Name != "i18n" && recv.Name != "lang") {
				return true
			}
			if recv.Name == "i18n" && sel.Sel.Name == "Register" && len(call.Args) == 1 {
				lit, ok := call.Args[0].(*ast.CompositeLit)
				if !ok {
					return true
				}
				for _, elt := range lit.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					key, okK := stringValue(kv.Key)
					val, okV := stringValue(kv.Value)
					if !okK || !okV {
						t.Errorf("%s: as traducións deben ser textos literais", fset.Position(kv.Pos()))
						continue
					}
					if prev, dup := catalog[key]; dup && prev != val {
						t.Errorf("%s: tradución distinta para %q da de %s", fset.Position(kv.Pos()), key, catalogPos[key])
					}
					catalog[key] = val
					catalogPos[key] = fset.Position(kv.Pos()).String()
				}
				return true
			}
			idx, ok := formatArg[sel.Sel.Name]
			if !ok || len(call.Args) <= idx {
				return true
			}
			if key, ok := stringValue(call.Args[idx]); ok {
				usages[dir] = append(usages[dir], usage{key: key, pos: fset.Position(call.Args[idx].Pos()).String()})
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("percorrendo o código: %v", err)
	}

	dirs := make([]string, 0, len(usages))
	for dir := range usages {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		t.Run(filepath.ToSlash(dir), func(t *testing.T) {
			for _, u := range usages[dir] {
				if _, ok := catalog[u.key]; !ok {
					t.Errorf("%s: falta a tradución inglesa de %q", u.pos, u.key)
				}
			}
		})
	}

	t.Run("formatos", func(t *testing.T) {
		for gl, en := range catalog {
			if strings.TrimSpace(en) == "" {
				t.Errorf("%s: tradución baleira para %q", catalogPos[gl], gl)
			}
			if a, b := verbs(gl), verbs(en); !reflect.DeepEqual(a, b) {
				t.Errorf("%s: formatos distintos en %q %v e %q %v", catalogPos[gl], gl, a, en, b)
			}
		}
	})
}

var verbRe = regexp.MustCompile(`%(\[\d+\])?[-+# 0]*(\*|\d+)?(\.(\*|\d+))?[a-zA-Z%]`)

// verbs devolve os formatos de fmt dun texto, na orde na que aparecen.
func verbs(s string) []string {
	var out []string
	for _, m := range verbRe.FindAllString(s, -1) {
		if m != "%%" {
			out = append(out, m)
		}
	}
	return out
}

// stringValue devolve o valor dun texto literal ou dunha suma de textos
// literais.
func stringValue(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(v.Value)
		return s, err == nil
	case *ast.BinaryExpr:
		if v.Op != token.ADD {
			return "", false
		}
		l, okL := stringValue(v.X)
		r, okR := stringValue(v.Y)
		return l + r, okL && okR
	case *ast.ParenExpr:
		return stringValue(v.X)
	}
	return "", false
}
