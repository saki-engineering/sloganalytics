# 自作静的解析

## まずはフレームワークを
```bash
# tenntennさんが作ったやつ
$ go install github.com/gostaticanalysis/skeleton/v2@latest
$ skeleton slogger
$ tree
.
├── README.md # これは最初からあった
└── slogger
    ├── cmd
    │   └── slogger
    │       └── main.go
    ├── go.mod
    ├── slogger.go
    ├── slogger_test.go
    └── testdata
        └── src
            └── a
                ├── a.go
                └── go.mod
```

## 実行
そのまま実行
```bash
$ go run cmd/slogger/main.go 
main is a tool for static analysis of Go programs.

Usage of main:
        main unit.cfg   # execute analysis specified by config file
        main help       # general help, including listing analyzers and flags
        main help name  # help on specific analyzer and its flags
exit status 1
```
何も指定しないと、unitchecker のヘルプやエラーが表示されます。

```bash
$ go run cmd/slogger/main.go ./testdata/src/a
main: invoking "go tool vet" directly is unsupported; use "go vet"
exit status 1
```
これもダメ。

多分go testで動作確認するのが一番手っ取り早い。
```bash
$ go test ./...
--- FAIL: TestAnalyzer (0.79s)
    analysistest.go:632: a/a.go:6: diagnostic "identifier is gopher" does not match pattern `pattern`
    analysistest.go:689: a/a.go:6: no diagnostic was reported matching `pattern`
FAIL
FAIL    slogger 1.101s
?       slogger/cmd/slogger     [no test files]
```

gopherという識別子があったら「"identifier is gopher"」と出すというのがデフォルト状態なので、`a.go`を以下のように変えればPASSします。
```diff:go
package a

func f() {
	// The pattern can be written in regular expression.
-	var gopher int // want "pattern"
+	var gopher int // want "identifier is gopher"
	print(gopher)  // want "identifier is gopher"
}
```
```bash
$ go test ./...
ok      slogger 0.671s
?       slogger/cmd/slogger     [no test files]
```

## まずは`With.Attr`メソッドがないかどうか
```go
func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		// ここはデフォルトとかえた
		// (*ast.Ident)(nil)
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		// n ast.Nodeをprint
		// &{<nil> TraceHandler <nil> 0 0xc006aa9d10 <nil>}

		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}

		// // 型名
		typeName := ts.Name.Name

		// 型オブジェクトを取得
		// objをprint
		// type a.TraceHandler struct{log/slog.Handler}
		obj := pass.TypesInfo.Defs[ts.Name]
		if obj == nil {
			return
		}
		typ := obj.Type()
		named, ok := typ.(*types.Named)
		if !ok {
			return
		}

		// WithAttrsメソッドを持っているか
		hasWithAttrs := false
		for i := 0; i < named.NumMethods(); i++ {
			log.Println(named.Method(i).Name())
			if named.Method(i).Name() == "WithAttrs" {
				hasWithAttrs = true
				break
			}
		}
		if !hasWithAttrs {
			pass.Reportf(ts.Pos(), "%s implements slog.Handler but does not implement WithAttrs method", typeName)
		}
	})

	return nil, nil
}
```

## 参考文献
https://budougumi0617.github.io/2019/03/24/go-create-type-check-handson/
