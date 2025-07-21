# /commit カスタムコマンド

このコマンドは、変更をgit commitします。コミットメッセージは引数として受け取ります。

## 手順

1. 現在のgit statusを確認
2. 変更をdiffで表示  
3. 最近のコミット履歴を確認してコミットメッセージのスタイルを把握
4. すべての変更をステージングエリアに追加
5. 引数で受け取ったメッセージでコミットを作成
6. コミット作成後の状態を確認

## 注意事項

- git repositoryでない場合はエラーメッセージを表示
- 変更がない場合は何もしない
- コミットメッセージの最後には必ず以下のフッターを追加：

```
🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## 実行

以下の手順でgit commitを実行してください：

1. `git status` でステータス確認
2. `git diff` で変更内容確認  
3. `git log --oneline -3` で最近のコミット確認
4. `git add .` で変更を追加
5. 引数のコミットメッセージを使って `git commit` 実行
6. `git status` で完了確認