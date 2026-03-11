package analyzer

import (
	"go/ast"
	"go/token"
	"regexp"
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
	CheckLowercase *bool    `mapstructure:"check_lowercase"`
	CheckEnglish   *bool    `mapstructure:"check_english"`
	CheckSpecials  *bool    `mapstructure:"check_specials"`
	CheckSensitive *bool    `mapstructure:"check_sensitive"`
}

var (
	reAllowedChars = regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)

	logMethods = map[string]bool{"Debug": true, "Info": true, "Warn": true, "Error": true}

	sensitiveWords = []string{"password", "api_key", "apikey", "token"}

	checkLowercase bool
	checkEnglish   bool
	checkSpecials  bool
	checkSensitive bool
	sensitiveFlag  string
)

func SetSensitiveWords(words []string) {
	if len(words) > 0 {
		sensitiveWords = words
	}
}

func ApplyConfig(cfg Config) {
	if len(cfg.SensitiveWords) > 0 {
		sensitiveWords = cfg.SensitiveWords
	}
	if cfg.CheckLowercase != nil {
		checkLowercase = *cfg.CheckLowercase
	}
	if cfg.CheckEnglish != nil {
		checkEnglish = *cfg.CheckEnglish
	}
	if cfg.CheckSpecials != nil {
		checkSpecials = *cfg.CheckSpecials
	}
	if cfg.CheckSensitive != nil {
		checkSensitive = *cfg.CheckSensitive
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

func init() {
	Analyzer.Flags.BoolVar(&checkLowercase, "check-lowercase", true, "check if log message starts with a lowercase letter")
	Analyzer.Flags.BoolVar(&checkEnglish, "check-english", true, "check if log message is in English only")
	Analyzer.Flags.BoolVar(&checkSpecials, "check-specials", true, "check if log message contains forbidden symbols or emojis")
	Analyzer.Flags.BoolVar(&checkSensitive, "check-sensitive", true, "check for sensitive data in logs")
	Analyzer.Flags.StringVar(&sensitiveFlag, "sensitive-words", "", "comma-separated list of sensitive words (overrides defaults)")
}

func run(pass *analysis.Pass) (any, error) {
	if sensitiveFlag != "" {
		words := strings.Split(sensitiveFlag, ",")
		for i := range words {
			words[i] = strings.TrimSpace(words[i])
		}
		sensitiveWords = words
	}

	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	ins.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !isLoggerFunc(pass, call) || len(call.Args) == 0 {
			return
		}

		if checkSensitive {
			checkSensitiveData(pass, call)
		}
		checkLogMessage(pass, call.Args[0])
	})
	return nil, nil
}

func isLoggerFunc(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if !logMethods[sel.Sel.Name] {
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

	if checkLowercase {
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
	}

	if checkEnglish {
		for _, r := range val {
			if unicode.Is(unicode.Cyrillic, r) {
				pass.Reportf(lit.Pos(), "log message must be in English only")
				return
			}
		}
	}

	if checkSpecials && !reAllowedChars.MatchString(val) {
		pass.Reportf(lit.Pos(), "log message contains forbidden symbols or emojis")
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
