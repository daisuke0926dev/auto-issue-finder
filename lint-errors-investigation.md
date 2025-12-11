# Lintエラー調査結果

## 調査日時
2025-12-11

## 検出されたエラーの概要
- gosec: 12件 (セキュリティ関連)
- revive: 1件 (未使用パラメータ)
- staticcheck: 3件 (不要なfmt.Sprintf)

合計: 16件

---

## エラー詳細

### 1. staticcheck S1039: 不要なfmt.Sprintf (3件)

#### エラー1-1
- **ファイル**: `cmd/sync.go`
- **行番号**: 277
- **問題**: 不要な fmt.Sprintf の使用
- **現在のコード**: `detailedError := fmt.Sprintf("【失敗の詳細】\n")`
- **修正方法**: 直接文字列を使用 → `detailedError := "【失敗の詳細】\n"`

#### エラー1-2
- **ファイル**: `cmd/sync.go`
- **行番号**: 282
- **問題**: 不要な fmt.Sprintf の使用
- **現在のコード**: `detailedError += fmt.Sprintf("\n実行を停止します。リトライ不可。\n")`
- **修正方法**: 直接文字列を使用 → `detailedError += "\n実行を停止します。リトライ不可。\n"`

#### エラー1-3
- **ファイル**: `cmd/sync.go`
- **行番号**: 303
- **問題**: 不要な fmt.Sprintf の使用
- **現在のコード**: `retryDetail := fmt.Sprintf("【リトライ詳細】\n")`
- **修正方法**: 直接文字列を使用 → `retryDetail := "【リトライ詳細】\n"`

---

### 2. gosec G301: ディレクトリパーミッション (2件)

#### エラー2-1
- **ファイル**: `cmd/sync.go`
- **行番号**: 178
- **問題**: ディレクトリパーミッションが 0755 (0750以下を期待)
- **現在のコード**: `os.MkdirAll(absLogDir, 0755)`
- **修正方法**: `os.MkdirAll(absLogDir, 0750)` に変更

#### エラー2-2
- **ファイル**: `cmd/sync.go`
- **行番号**: 815
- **問題**: ディレクトリパーミッションが 0755 (0750以下を期待)
- **現在のコード**: `os.MkdirAll(absLogDir, 0755)`
- **修正方法**: `os.MkdirAll(absLogDir, 0750)` に変更

---

### 3. gosec G306: ファイルパーミッション (3件)

#### エラー3-1
- **ファイル**: `internal/config/alias_test.go`
- **行番号**: 140
- **問題**: ファイルパーミッションが 0644 (0600以下を期待)
- **現在のコード**: `os.WriteFile(configPath, []byte(configContent), 0644)`
- **修正方法**: `os.WriteFile(configPath, []byte(configContent), 0600)` に変更

#### エラー3-2
- **ファイル**: `internal/config/alias_test.go`
- **行番号**: 173
- **問題**: ファイルパーミッションが 0644 (0600以下を期待)
- **現在のコード**: `os.WriteFile(invalidConfigPath, []byte("invalid toml [[["), 0644)`
- **修正方法**: `os.WriteFile(invalidConfigPath, []byte("invalid toml [[["), 0600)` に変更

#### エラー3-3
- **ファイル**: `internal/config/alias_test.go`
- **行番号**: 185
- **問題**: ファイルパーミッションが 0644 (0600以下を期待)
- **現在のコード**: `os.WriteFile(".sleepship.toml", []byte("invalid toml [[["), 0644)`
- **修正方法**: `os.WriteFile(".sleepship.toml", []byte("invalid toml [[["), 0600)` に変更

---

### 4. gosec G304: 変数によるファイル読み込み (2件)

#### エラー4-1
- **ファイル**: `cmd/sync.go`
- **行番号**: 185
- **問題**: 変数を使ったファイル作成（潜在的なセキュリティリスク）
- **現在のコード**: `os.Create(logFilePath)`
- **修正方法**: パス検証を追加するか、gosecの警告を抑制 (コメント `// #nosec G304`)

#### エラー4-2
- **ファイル**: `internal/history/history.go`
- **行番号**: 44
- **問題**: 変数を使ったファイル読み込み（潜在的なセキュリティリスク）
- **現在のコード**: `os.ReadFile(historyPath)`
- **修正方法**: パス検証を追加するか、gosecの警告を抑制 (コメント `// #nosec G304`)

#### エラー4-3
- **ファイル**: `task/parser.go`
- **行番号**: 23
- **問題**: 変数を使ったファイル読み込み（潜在的なセキュリティリスク）
- **現在のコード**: `os.Open(filename)`
- **修正方法**: パス検証を追加するか、gosecの警告を抑制 (コメント `// #nosec G304`)

---

### 5. gosec G204: 変数を使ったサブプロセス実行 (3件)

#### エラー5-1
- **ファイル**: `cmd/sync.go`
- **行番号**: 538
- **問題**: 変数を使ったbashコマンド実行（潜在的なセキュリティリスク）
- **現在のコード**: `exec.Command("bash", "-c", verifyCmd)`
- **修正方法**: コマンドインジェクション対策を追加するか、警告を抑制 (コメント `// #nosec G204`)

#### エラー5-2
- **ファイル**: `cmd/sync.go`
- **行番号**: 674
- **問題**: 変数を使ったgitコマンド実行
- **現在のコード**: `exec.Command("git", "checkout", "-b", branchName)`
- **修正方法**: ブランチ名の検証を追加するか、警告を抑制 (コメント `// #nosec G204`)

#### エラー5-3
- **ファイル**: `cmd/sync.go`
- **行番号**: 706
- **問題**: 変数を使ったgitコマンド実行
- **現在のコード**: `exec.Command("git", "commit", "-m", commitMessage)`
- **修正方法**: コミットメッセージの検証を追加するか、警告を抑制 (コメント `// #nosec G204`)

---

### 6. gosec G301: ディレクトリパーミッション (追加1件)

#### エラー6-1
- **ファイル**: `internal/history/history.go`
- **行番号**: 63
- **問題**: ディレクトリパーミッションが 0755 (0750以下を期待)
- **現在のコード**: `os.MkdirAll(dir, 0755)`
- **修正方法**: `os.MkdirAll(dir, 0750)` に変更

---

### 7. revive: 未使用パラメータ (1件)

#### エラー7-1
- **ファイル**: `task/parser.go`
- **行番号**: 174
- **問題**: パラメータ 'tasks' が使用されていない
- **現在のコード**: `func checkCircularDependencies(tasks []Task) error`
- **修正方法**: パラメータを `_` にリネームするか、実装を追加

---

## 修正優先順位

### 優先度高 (機能に影響)
1. **staticcheck S1039** (3件) - 不要なfmt.Sprintf → すぐに修正可能
2. **revive 未使用パラメータ** (1件) - 関数の実装が不完全

### 優先度中 (セキュリティベストプラクティス)
1. **gosec G301/G306** (6件) - パーミッション設定 → 簡単に修正可能

### 優先度低 (既存機能で問題なし、警告抑制を検討)
1. **gosec G304** (3件) - ファイル読み込み → 信頼できる入力のため警告抑制可
2. **gosec G204** (3件) - サブプロセス実行 → 内部使用のため警告抑制可

---

## 次のステップ

1. staticcheck S1039エラーを修正 (3箇所) - cmd/sync.go
2. revive 未使用パラメータエラーを修正 (1箇所) - task/parser.go
3. gosec パーミッションエラーを修正 (6箇所)
   - cmd/sync.go: 2箇所
   - internal/config/alias_test.go: 3箇所
   - internal/history/history.go: 1箇所
4. gosec G304/G204は警告抑制コメントを追加 (6箇所) - 必要に応じて
