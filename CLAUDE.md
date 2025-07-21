# CLAUDE.md

このファイルは、このリポジトリのコードを扱う際にClaude Code (claude.ai/code) にガイダンスを提供します。

## プロジェクト概要

これは、slog.Handlerの実装を解析することに焦点を当てたGo静的解析ツールプロジェクトです。メインツール（`slogger`）は、`golang.org/x/tools/go/analysis`フレームワークを使って構築されたカスタム静的解析器で、`slog.Handler`を実装する型が必要な`WithAttrs`メソッドを正しいシグネチャで適切に実装しているかをチェックします。

## 開発コマンド

### テスト
```bash
# 全てのテストを実行（解析器の機能を検証する主な方法）
go test ./...

# sloggerディレクトリのテストを具体的に実行
cd slogger && go test
```

### 解析器のビルドと実行
```bash
# 解析器を直接実行（sloggerディレクトリから）
go run cmd/slogger/main.go

# 解析器をビルド
go build -o slogger cmd/slogger/main.go

# テストデータで解析器を実行
go run cmd/slogger/main.go ./testdata/src/a
```

注意：解析器は`unitchecker`フレームワークで動作するよう設計されているため、適切な設定なしに直接実行するとヘルプテキストやエラーが表示される場合があります。

## アーキテクチャ

### 核となるコンポーネント

- **slogger.go**: メインの解析器実装で、以下を行います：
  - AST検査を使用して型指定を見つける
  - 型が`slog.Handler`インターフェースを実装しているかをチェック
  - `WithAttrs`メソッドのシグネチャを検証（パラメータ：`[]slog.Attr`、戻り値：`slog.Handler`）
  - Handlerの実装が適切な`WithAttrs`メソッドを欠いている場合に違反を報告

- **cmd/slogger/main.go**: `unitchecker.Main()`を使用するCLIエントリーポイント

- **slogger_test.go**: `analysistest`フレームワークを使用するテストランナー

- **testdata/**: 解析器をテストするためのGoモジュールを含むテストケース

### 解析フロー

1. 解析器は`*ast.TypeSpec`（型定義）のASTノードをフィルタリング
2. 各型について、`types.Implements()`を使用してその型が`slog.Handler`を実装しているかをチェック
3. インターフェースを実装している場合、`WithAttrs`メソッドのシグネチャを検証：
   - パラメータは`[]slog.Attr`である必要がある
   - 戻り値の型は`slog.Handler`である必要がある
4. 準拠していない実装に対して診断メッセージを報告

### テストデータ構造

テストケースは`testdata/src/a/`にあり、以下を含む：
- `a.go`: `slog.Handler`を実装する`TraceHandler`型を含む
- `// want "message"`形式を使用した期待される診断コメント
- 独立したテストのための別の`go.mod`

## プロジェクトセットアップ

このプロジェクトは`gostaticanalysis/skeleton`ツールを使用して生成されました：
```bash
go install github.com/gostaticanalysis/skeleton/v2@latest
skeleton slogger
```

これにより、事前設定されたテストインフラストラクチャを持つGo静的解析ツールの標準的な構造が作成されます。

# Conversation Guidelines
常に日本語で会話する
要求されたことを行う；それ以上でも以下でもない。
目標達成に絶対に必要でない限り、ファイルを作成しない。
新しいファイルを作成するよりも既存のファイルを編集することを常に優先する。
ドキュメントファイル（*.md）やREADMEファイルを積極的に作成しない。ユーザーから明示的に要求された場合のみドキュメントファイルを作成する。