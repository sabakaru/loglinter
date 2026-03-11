package analyzer

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "loglinter",
	Doc:      "Checks log messages for stylistic and security rules",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	ins.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !isLoggerFunc(pass, call) || len(call.Args) == 0 {
			return
		}

		checkSensitiveData(pass, call)
		checkLogMessage(pass, call.Args[0])
	})

	return nil, nil
}

func isLoggerFunc(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.Uses[sel.Sel]
	if obj == nil {
		return false
	}

	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}

	path := pkg.Path()
	return path == "log/slog" || path == "go.uber.org/zap"
}

func checkLogMessage(pass *analysis.Pass, arg ast.Expr) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}

	val, err := strconv.Unquote(lit.Value)
	if err != nil || len(val) == 0 {
		return
	}

	firstRune, size := utf8.DecodeRuneInString(val)
	if unicode.IsUpper(firstRune) {
		lowerRune := unicode.ToLower(firstRune)

		fix := analysis.SuggestedFix{
			Message: "Change first letter to lowercase",
			TextEdits: []analysis.TextEdit{
				{
					Pos:     lit.Pos() + 1,
					End:     lit.Pos() + 1 + token.Pos(size),
					NewText: []byte(string(lowerRune)),
				},
			},
		}

		pass.Report(analysis.Diagnostic{
			Pos:            lit.Pos(),
			End:            lit.End(),
			Message:        "log message must start with a lowercase letter",
			SuggestedFixes: []analysis.SuggestedFix{fix},
		})
	}

	for _, r := range val {
		if unicode.Is(unicode.Cyrillic, r) {
			pass.Reportf(lit.Pos(), "log message must be in English only")
			break
		}
	}

	if strings.HasSuffix(val, "!") || strings.HasSuffix(val, "...") || strings.Contains(val, "!!!") {
		pass.Reportf(lit.Pos(), "log message must not contain special characters")
	} else {
		for _, r := range val {
			if r > unicode.MaxASCII && !unicode.IsLetter(r) && !unicode.IsSpace(r) && !unicode.IsPunct(r) {
				pass.Reportf(lit.Pos(), "log message must not contain emojis")
				break
			}
		}
	}
}

func checkSensitiveData(pass *analysis.Pass, call *ast.CallExpr) {
	found := false

	var walk func(ast.Node)
	walk = func(n ast.Node) {
		if found || n == nil {
			return
		}

		switch x := n.(type) {
		case *ast.Ident:
			name := strings.ToLower(x.Name)
			if containsSensitive(name) {
				found = true
			}
		}
	}

	for _, arg := range call.Args {
		ast.Inspect(arg, func(n ast.Node) bool {
			walk(n)
			return !found
		})
	}

	if found {
		pass.Reportf(call.Pos(), "log message must not contain sensitive data")
	}
}

func containsSensitive(s string) bool {
	return strings.Contains(s, "password") ||
		strings.Contains(s, "api_key") ||
		strings.Contains(s, "apikey") ||
		strings.Contains(s, "token")
}
