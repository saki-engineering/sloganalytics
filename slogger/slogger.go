package slogger

import (
	"go/ast"
	"go/importer"
	"go/types"
	"log"
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

		// slog.Handlerを取得
		_, err := importer.Default().Import("log/slog")
		if err != nil {
			log.Printf("failed to import log/slog: %v", err)
			return
		}
		
		// まず、slog.Handlerを実装しようとしているかをチェック
		// （埋め込みフィールドがある）
		hasEmbeddedHandler := false
		
		// 構造体のフィールドをチェック
		if structType, ok := named.Underlying().(*types.Struct); ok {
			log.Printf("Checking %d fields in struct %s", structType.NumFields(), typeName)
			for i := 0; i < structType.NumFields(); i++ {
				field := structType.Field(i)
				log.Printf("  Field %d: name=%s, embedded=%v, type=%v", i, field.Name(), field.Embedded(), field.Type())
				if field.Embedded() {
					if namedType, ok := field.Type().(*types.Named); ok {
						log.Printf("    Named type: pkg=%v, name=%s", namedType.Obj().Pkg(), namedType.Obj().Name())
						if namedType.Obj().Pkg() != nil && 
						   namedType.Obj().Pkg().Path() == "log/slog" && 
						   namedType.Obj().Name() == "Handler" {
							hasEmbeddedHandler = true
							break
						}
					}
				}
			}
		}
		
		// slog.Handlerを実装しようとしている場合、HandlerInfoに追加
		if hasEmbeddedHandler {
			log.Printf("%s has embedded slog.Handler", typeName)
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
				log.Println("slice")
				continue
			}
			elem, ok := t.Elem().(*types.Named)
			if !ok {
				log.Println("elem")
				continue
			}
			if elem.Obj().Pkg() == nil {
				log.Println("pkg is nil")
				continue
			}
			if elem.Obj().Pkg().Path() != "log/slog" {
				log.Printf("pkg path is %s, not log/slog", elem.Obj().Pkg().Path())
				continue
			}
			if elem.Obj().Name() != "Attr" {
				log.Printf("elem name is %s, not Attr", elem.Obj().Name())
				continue
			}

			// 戻り値
			rslType := sig.Results().At(0).Type().(*types.Named)
			if rslType.Obj().Pkg().Path() != "log/slog" {
				log.Printf("pkg path is %s, not log/slog", rslType.Obj().Pkg().Path())
				continue
			}
			if rslType.Obj().Name() != "Handler" {
				log.Printf("elem name is %s, not Handler", rslType.Obj().Name())
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
