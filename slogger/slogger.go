package slogger

import (
	"go/ast"
	"go/importer"
	"go/types"
	"log"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "slogger is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "slogger",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		// n ast.Nodeをprint
		// &{<nil> TraceHandler <nil> 0 0xc006aa9d10 <nil>}

		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}
		// _, ok = ts.Type.(*ast.StructType)
		// if !ok {
		// 	return
		// }

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
		
		// slog.Handlerを実装しようとしていない場合はスキップ
		if !hasEmbeddedHandler {
			return
		}
		
		log.Printf("%s has embedded slog.Handler", typeName)

		// WithAttrsメソッドを持っているか
		hasWithAttrs := false
		for i := 0; i < named.NumMethods(); i++ {
			// メソッド名
			if named.Method(i).Name() != "WithAttrs" {
				continue
			}

			// 第一引数
			sig := named.Method(i).Signature()
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

		if !hasWithAttrs {
			pass.Reportf(ts.Pos(), "%s implements slog.Handler but does not implement WithAttrs method", typeName)
		}
	})

	return nil, nil
}
