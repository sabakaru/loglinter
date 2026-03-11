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

type Config struct {
	SensitiveWords []string `mapstructure:"sensitive_words"`
}

var sensitiveWords = []string{"password", "api_key", "apikey", "token"}

func SetSensitiveWords(words []string) {
	if len(words) > 0 {
		sensitiveWords = words
	}
}

func containsSensitive(s string) bool {
	s = strings.ToLower(s)
	for _, word := range sensitiveWords {
		if word != "" && strings.Contains(s, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

var Analyzer = &analysis.Analyzer{
	Name:     "loglinter",
	Doc:      "Checks log messages for stylistic and security rules",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

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
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	path := obj.Pkg().Path()
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
		pass.Report(analysis.Diagnostic{
			Pos:     lit.Pos(),
			End:     lit.End(),
			Message: "log message must start with a lowercase letter",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "lowercase first letter",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     lit.Pos() + 1,
							End:     lit.Pos() + 1 + token.Pos(size),
							NewText: []byte(string(lowerRune)),
						},
					},
				},
			},
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
	for _, arg := range call.Args {
		ast.Inspect(arg, func(n ast.Node) bool {
			if found {
				return false
			}
			switch x := n.(type) {
			case *ast.Ident:
				if containsSensitive(x.Name) {
					found = true
				}
			case *ast.BasicLit:
				if x.Kind == token.STRING {
					v, _ := strconv.Unquote(x.Value)
					v = strings.ToLower(v)
					if containsSensitive(v) && (strings.Contains(v, "=") || strings.Contains(v, ":")) {
						found = true
					}
				}
			}
			return true
		})
	}

	if found {
		pass.Reportf(call.Pos(), "log message must not contain sensitive data")
	}
}
