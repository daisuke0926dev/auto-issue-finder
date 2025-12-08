# Sleepship

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> Claude Codeによる自律開発システム

タスクファイルを作成するだけで、Claude Codeがバックグラウンドで自律的に開発を進めるツールです。

---

## クイックスタート

```bash
# ビルド
go build -o bin/sleepship

# タスクファイル作成
./bin/sleepship init tasks.txt

# 実行（バックグラウンドで自動実行）
./bin/sleepship sync tasks.txt

# ログをリアルタイム監視
tail -f logs/sync-*.log
```

---

## 主要機能

- **バックグラウンド実行** - コマンド実行後すぐに制御が戻る
- **同期処理** - タスクを順次実行、各タスク完了後に次へ進む
- **自動検証** - 各タスク後に確認コマンド実行
- **自動エラー修正** - 検証失敗時にClaude Codeが自動で修正を試みる
- **自動リトライ** - タスク実行と検証の失敗時に自動リトライ（デフォルト3回）
- **詳細ログ** - 全実行内容をログファイルに記録

---

## 使用例

### 基本的な使い方

```bash
# カレントディレクトリで実行
./bin/sleepship sync tasks.txt

# 別プロジェクトで実行
./bin/sleepship sync tasks.txt --dir=/path/to/project

# タスク6から再開
./bin/sleepship sync tasks.txt --start-from=6

# リトライ回数を変更
./bin/sleepship sync tasks.txt --max-retries=5
```

### タスクファイル例

```markdown
## タスク1: ユーザーモデルの実装
models/user.go にUserモデルを実装してください。

### 確認
- `go build`

## タスク2: テストの追加
models/user_test.go にユニットテストを追加してください。

### 確認
- `go test ./models`
```

### 実行の流れ

1. `sync`コマンドを実行すると、バックグラウンドプロセスが起動
2. すぐに制御が戻り、PIDとログファイルパスが表示される
3. バックグラウンドで各タスクが順次実行される
4. 各タスク後に確認コマンドが実行され、失敗時は自動修正を試みる
5. 全タスク完了までログファイルに進捗を記録

---

## タスクファイルの書き方

### テンプレートから始める

```bash
./bin/sleepship init my-tasks.txt
```

### 基本フォーマット

```markdown
## タスク[番号]: タスクのタイトル

タスクの詳細な説明

### 確認
- `実行する確認コマンド`
```

### 確認コマンド例

```markdown
### 確認
- `go build`
- `go test ./...`
- `npm run build`
- `make test`
```

### タスク分割のコツ

**良い例** - 適切に分割:
```markdown
## タスク1: データベースモデル作成
### 確認
- `go build`

## タスク2: CRUD操作実装
### 確認
- `go test ./repositories`
```

**悪い例** - 大きすぎる:
```markdown
## タスク1: ユーザー管理機能を全部作る
モデル、リポジトリ、API、テスト全部実装して
```

**ポイント**: 1タスク = 1つの明確な実装単位

---

## オプション

### --start-from

エラーで停止したタスクから再開できます。

```bash
./bin/sleepship sync tasks.txt --start-from=6
```

### --max-retries

自動リトライ回数を制御できます（デフォルト: 3回）。

```bash
./bin/sleepship sync tasks.txt --max-retries=5
```

### --dir

別プロジェクトで実行できます。

```bash
./bin/sleepship sync tasks.txt --dir=/path/to/project
```

---

## プロジェクト構造

```
sleepship/
├── cmd/
│   ├── sync.go          # 同期処理型自律開発コマンド
│   ├── sync_test.go     # syncコマンドのテスト
│   ├── init.go          # タスクファイル初期化コマンド
│   └── root.go          # CLIルート
├── bin/
│   └── sleepship        # 実行ファイル
├── logs/                # 実行ログ
├── main.go              # エントリーポイント
├── main_test.go         # メインテスト
├── tasks.txt.example    # サンプルタスクファイル
└── go.mod               # Goモジュール定義
```

---

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) をご覧ください。
