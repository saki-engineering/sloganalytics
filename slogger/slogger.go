package slogger

import (
	"go/ast"
	"go/importer"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "slogger checks slog.Handler implementations for missing required methods"

// HandlerInfo holds information about a struct that implements slog.Handler
type HandlerInfo struct {
	Name     string
	TypeSpec *ast.TypeSpec
	Named    *types.Named
}

// HandlerInfos represents a slice of HandlerInfo
type HandlerInfos []*HandlerInfo

// HandlerFinderAnalyzer finds structs that implement slog.Handler interface
var HandlerFinderAnalyzer = &analysis.Analyzer{
	Name: "handlerfinder",
	Doc:  "finds structs that implement slog.Handler interface",
	Run:  handlerFinderRun,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	ResultType: reflect.TypeOf(HandlerInfos{}),
}

// Analyzer checks if the found handlers implement required methods
var Analyzer = &analysis.Analyzer{
	Name: "slogger",
	Doc:  doc,
	Run:  methodValidatorRun,
	Requires: []*analysis.Analyzer{
		HandlerFinderAnalyzer,
	},
}

func handlerFinderRun(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	var handlers HandlerInfos

	// slog.Handlerインターフェースの型を取得
	// pass.TypesInfoから直接取得する方法を試す
	var handlerInterface *types.Interface
	
	// パッケージ内でslog.Handlerの使用を見つける
	for _, obj := range pass.TypesInfo.Uses {
		if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "log/slog" && obj.Name() == "Handler" {
			if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
				handlerInterface = iface
				break
			}
		}
	}
	
	// 見つからない場合はimporterで取得
	if handlerInterface == nil {
		slogPkg, err := importer.Default().Import("log/slog")
		if err != nil {
			return nil, err
		}
		handlerObj := slogPkg.Scope().Lookup("Handler")
		if handlerObj == nil {
			return nil, nil
		}
		handlerInterface = handlerObj.Type().Underlying().(*types.Interface)
	}

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}

		typeName := ts.Name.Name

		// 型オブジェクトを取得
		obj := pass.TypesInfo.Defs[ts.Name]
		if obj == nil {
			return
		}
		typ := obj.Type()
		named, ok := typ.(*types.Named)
		if !ok {
			return
		}

		// types.Implementsを使用してslog.Handlerインターフェースを実装しているかチェック
		// 値型とポインタ型の両方をチェック
		pointerType := types.NewPointer(named)
		if types.Implements(named, handlerInterface) || types.Implements(pointerType, handlerInterface) {
			handlers = append(handlers, &HandlerInfo{
				Name:     typeName,
				TypeSpec: ts,
				Named:    named,
			})
		}
	})

	return handlers, nil
}

func methodValidatorRun(pass *analysis.Pass) (any, error) {
	handlers := pass.ResultOf[HandlerFinderAnalyzer].(HandlerInfos)

	for _, handler := range handlers {
		// WithAttrsメソッドを持っているか
		hasWithAttrs := false
		for i := 0; i < handler.Named.NumMethods(); i++ {
			// メソッド名
			if handler.Named.Method(i).Name() != "WithAttrs" {
				continue
			}

			// 第一引数
			sig := handler.Named.Method(i).Signature()
			paramType := sig.Params().At(0).Type()
			t, ok := paramType.(*types.Slice)
			if !ok {
				continue
			}
			elem, ok := t.Elem().(*types.Named)
			if !ok {
				continue
			}
			if elem.Obj().Pkg() == nil {
				continue
			}
			if elem.Obj().Pkg().Path() != "log/slog" {
				continue
			}
			if elem.Obj().Name() != "Attr" {
				continue
			}

			// 戻り値
			rslType := sig.Results().At(0).Type().(*types.Named)
			if rslType.Obj().Pkg().Path() != "log/slog" {
				continue
			}
			if rslType.Obj().Name() != "Handler" {
				continue
			}

			hasWithAttrs = true
			break
		}

		// WithGroupメソッドを持っているか
		hasWithGroup := false
		for i := 0; i < handler.Named.NumMethods(); i++ {
			// メソッド名
			if handler.Named.Method(i).Name() != "WithGroup" {
				continue
			}

			// 第一引数（string）
			sig := handler.Named.Method(i).Signature()
			if sig.Params().Len() != 1 {
				continue
			}
			paramType := sig.Params().At(0).Type()
			basicType, ok := paramType.(*types.Basic)
			if !ok || basicType.Kind() != types.String {
				continue
			}

			// 戻り値（slog.Handler）
			if sig.Results().Len() != 1 {
				continue
			}
			rslType, ok := sig.Results().At(0).Type().(*types.Named)
			if !ok {
				continue
			}
			if rslType.Obj().Pkg() == nil || rslType.Obj().Pkg().Path() != "log/slog" {
				continue
			}
			if rslType.Obj().Name() != "Handler" {
				continue
			}

			hasWithGroup = true
			break
		}

		if !hasWithAttrs {
			pass.Reportf(handler.TypeSpec.Pos(), "%s implements slog.Handler but does not implement WithAttrs method", handler.Name)
		}
		if !hasWithGroup {
			pass.Reportf(handler.TypeSpec.Pos(), "%s implements slog.Handler but does not implement WithGroup method", handler.Name)
		}
	}

	return nil, nil
}
