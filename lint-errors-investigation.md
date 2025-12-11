# Lintエラー調査結果

## エラー1: errorlint
- ファイル: cmd/sync.go
- 行番号: 544
- 問題: fmt.Errorf で %s を使っている（%w を使うべき）
- 修正方法: %s を %w に変更
- 詳細: `fmt.Errorf("verification failed: %s\nOutput: %s", err, string(output))` のエラーラッピング部分

## エラー2: gosimple S1039 (1)
- ファイル: cmd/sync.go
- 行番号: 277
- 問題: 不要な fmt.Sprintf の使用
- 修正方法: 直接文字列を使用
- 詳細: `detailedError := fmt.Sprintf("【失敗の詳細】\n")` は直接文字列リテラルで良い

## エラー3: gosimple S1039 (2)
- ファイル: cmd/sync.go
- 行番号: 282
- 問題: 不要な fmt.Sprintf の使用
- 修正方法: 直接文字列を使用
- 詳細: `detailedError += fmt.Sprintf("\n実行を停止します。リトライ不可。\n")` は直接文字列リテラルで良い

## エラー4: gosimple S1039 (3)
- ファイル: cmd/sync.go
- 行番号: 303
- 問題: 不要な fmt.Sprintf の使用
- 修正方法: 直接文字列を使用
- 詳細: `retryDetail := fmt.Sprintf("【リトライ詳細】\n")` は直接文字列リテラルで良い

## その他の検出エラー（参考情報）

### gosec (セキュリティ警告) - 12件
これらは主にセキュリティベストプラクティスに関する警告で、今回のタスクで修正対象となっている errorlint と gosimple とは異なります：

- G301: ディレクトリパーミッションが 0755（0750以下を推奨）
  - cmd/sync.go:178
  - internal/config/alias_test.go:179
  - internal/history/history.go:63

- G304: 変数を使ったファイル操作（潜在的なファイルインクルージョン）
  - cmd/sync.go:185
  - internal/history/history.go:44
  - task/parser.go:23

- G204: 変数を使ったサブプロセス起動
  - cmd/sync.go:538, 674, 706

- G306: ファイルパーミッションが 0644（0600以下を推奨）
  - internal/config/alias_test.go:140, 173, 185

### revive (コード品質) - 1件
- unused-parameter: 未使用パラメータ
  - task/parser.go:174: `checkCircularDependencies` 関数の `tasks` パラメータが未使用

## 修正優先度

1. **高優先度**（タスク対象）
   - errorlint: cmd/sync.go:544 (エラーラッピング)
   - gosimple S1039: cmd/sync.go:277, 282, 303 (不要なfmt.Sprintf)

2. **中優先度**（将来的に修正推奨）
   - revive: task/parser.go:174 (未使用パラメータ)

3. **低優先度**（セキュリティ強化の観点から検討）
   - gosec: パーミッション設定、ファイル操作（現状の実装でも問題ない可能性が高い）
