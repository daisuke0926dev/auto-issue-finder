# Auto Issue Finder

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

> 🤖 Claude Codeによる自律開発システム + 📊 GitHub Issue分析ツール

このリポジトリには3つの強力なツールが含まれています:

1. **同期処理型自律開発ツール (sync)** - タスクを順次実行し、エラーを自動修正する次世代ツール
2. **Claude Code自律開発システム (bash版)** - 寝ている間にClaude Codeが自律的に開発を進める従来のシステム
3. **GitHub Issue Analyzer** - リポジトリのIssueを自動分析し、パターン検出と推奨事項を提供

---

## 🚀 クイックスタート

### 同期処理型自律開発ツール（推奨）

```bash
# ビルド
go build -o bin/auto-issue-finder

# タスクファイルを作成
cat > tasks.txt << 'EOF'
## タスク1: ユーザーモデルの実装
models/user.go にUserモデルを実装してください。

### 確認
- `go build`

## タスク2: テストの追加
models/user_test.go にユニットテストを追加してください。

### 確認
- `go test ./models`
EOF

# 実行（タスクを順次実行、エラーは自動修正）
./bin/auto-issue-finder sync tasks.txt

# 別プロジェクトで実行
./bin/auto-issue-finder sync tasks.txt --dir=/path/to/project
```

### Claude Code自律開発システム（bash版）

```bash
# 任意のプロジェクトにワンライナーでインストール
curl -sSL https://raw.githubusercontent.com/isiidaisuke0926/auto-issue-finder/main/install-auto-dev.sh | bash

# タスクファイルを作成
cp tonight-with-tasks.txt.example tonight.txt
vim tonight.txt

# 実行（寝ている間に開発）
./run-overnight.sh tonight.txt
```

### GitHub Issue Analyzer

```bash
# ビルド
go build -o bin/auto-issue-finder

# 実行
export GITHUB_TOKEN=your_token_here
./bin/auto-issue-finder analyze microsoft/vscode --format=console
```

---

## ✨ 主要機能

### 同期処理型自律開発ツール (sync)

- 🔄 **真の同期処理** - タスクを順次実行、各タスク完了後に次へ進む
- ✅ **自動検証** - 各タスク後に確認コマンド実行（`go build`, `go test`等）
- 🔧 **自動エラー修正** - 検証失敗時にClaude Codeが自動で修正を試みる
- 🌐 **汎用性** - `--dir`オプションで任意のプロジェクトで使用可能
- 📝 **詳細ログ** - 全実行内容をログファイルに記録
- 🎯 **柔軟なタスク記述** - 日本語・英語両対応のマークダウン形式

### Claude Code自律開発システム (bash版)

- 🌙 **夜間自律実行** - バックグラウンドで数時間の開発を自動実行
- 📝 **タスクベース** - マークダウンでタスクを記述するだけ
- 🔄 **自動コミット** - タスクごと、または完了時に自動git commit
- 🎯 **柔軟な実行モード** - 対話的/バックグラウンド/インクリメンタル
- 🔧 **自動承認設定** - 全ツールの使用を自動承認して中断なし

### GitHub Issue Analyzer

- 📊 **包括的な分析** - 統計、ラベル分布、月次トレンド
- 🔍 **パターン検出** - 長期未解決Issue、高活動Issue、重複疑い
- 📈 **複数の出力形式** - Console/Markdown/JSON
- 💡 **実用的な推奨** - 分析結果に基づく具体的なアクション提案
- ⚡ **高速処理** - ページネーション対応、効率的なAPI利用

---

## 📦 インストール

### 同期処理型自律開発ツール (sync) & GitHub Issue Analyzer

```bash
git clone https://github.com/isiidaisuke0926/auto-issue-finder.git
cd auto-issue-finder
go mod download
go build -o bin/auto-issue-finder

# 使用可能なコマンド確認
./bin/auto-issue-finder --help
```

ビルドされるコマンド:
- `sync` - 同期処理型自律開発
- `analyze` - GitHub Issue分析

### Claude Code自律開発システム (bash版)

他のプロジェクトで使用する場合:

```bash
# ワンライナーインストール
curl -sSL https://raw.githubusercontent.com/isiidaisuke0926/auto-issue-finder/main/install-auto-dev.sh | bash
```

インストールされるファイル:
- `auto-dev.sh` - 基本実行
- `auto-dev-incremental.sh` - タスクごとにコミット
- `run-overnight.sh` - バックグラウンド実行
- `.claude/settings.local.json` - 自動承認設定

---

## 🎯 使用例

### 同期処理型自律開発ツール (sync)

**基本的な使い方:**

```bash
# タスクファイル作成
cat > my-tasks.txt << 'EOF'
## タスク1: データベースモデルの追加

models/product.go に Product モデルを実装してください。
- ID, Name, Price, CreatedAt フィールドを持つ
- GORM タグを適切に設定

### 確認
- `go build`

## タスク2: CRUD操作の実装

repositories/product.go に ProductRepository を実装してください。
- Create, Read, Update, Delete メソッド
- エラーハンドリングを適切に行う

### 確認
- `go build`
- `go test ./repositories`

## タスク3: API エンドポイントの追加

handlers/product.go に HTTP ハンドラーを実装してください。
- GET /products
- POST /products
- PUT /products/:id
- DELETE /products/:id

### 確認
- `go build`
- `go test ./handlers`
EOF

# 実行（カレントディレクトリ）
./bin/auto-issue-finder sync my-tasks.txt

# 別プロジェクトで実行
./bin/auto-issue-finder sync my-tasks.txt --dir=/path/to/your/project

# ログディレクトリを指定
./bin/auto-issue-finder sync my-tasks.txt --log-dir=./custom-logs
```

**タスクファイルフォーマット:**

```markdown
## タスク[番号]: タスクタイトル

実装の詳細な説明をここに記述します。
マークダウン形式で自由に記述できます。

### 確認
- `実行したい確認コマンド`
- `go test ./...`
- `npm run build`

---

## Task2: 英語タイトルも可

English description is also supported.

### 確認
- `make test`
```

**実行の流れ:**

1. タスク1を Claude Code で実装
2. 確認コマンド（`go build`）を実行
3. ✅ 成功 → タスク2へ
4. ❌ 失敗 → 自動修正を試みる → 再確認 → 成功/失敗で停止
5. すべてのタスクが完了するまで繰り返し

### Claude Code自律開発システム (bash版)

**基本的な使い方:**

```bash
# 1. タスクファイル作成
cat > tonight.txt << 'EOF'
今夜のタスク: RESTful APIサーバーの実装

## 要件
- ユーザーCRUD機能
- JWT認証
- テストカバレッジ80%以上

## 技術スタック
- Go 1.21
- chi router
- PostgreSQL
EOF

# 2. 実行方法を選択

# 対話的実行
./auto-dev.sh tonight.txt

# タスクごとに個別コミット
./auto-dev-incremental.sh tonight.txt

# バックグラウンド実行（推奨）
./run-overnight.sh tonight.txt
tail -f nohup.out  # ログ監視
```

**インクリメンタルコミットの例:**

```markdown
# タスク1: データベーススキーマ設計
- users テーブル作成
- マイグレーションファイル作成

# タスク2: APIエンドポイント実装
- GET /users
- POST /users
- PUT /users/:id
- DELETE /users/:id

# タスク3: テスト追加
- ユニットテスト
- 統合テスト
```

### GitHub Issue Analyzer

**コンソール出力:**

```bash
./auto-issue-finder analyze golang/go --format=console --limit=100
```

**マークダウンレポート生成:**

```bash
./auto-issue-finder analyze microsoft/vscode \
  --state=open \
  --format=markdown \
  --output=report.md
```

**JSON出力（自動化向け）:**

```bash
./auto-issue-finder analyze owner/repo \
  --format=json \
  --output=analysis.json

# jqで処理
cat analysis.json | jq '.Stats.TotalIssues'
```

---

## 📚 ドキュメント

- [📖 詳細なインストールガイド](docs/INSTALL.md)
- [🔧 使用方法とコマンドリファレンス](docs/USAGE.md)
- [🤖 自律開発システム詳細](docs/AUTO_DEV.md)
- [🧪 テストとカバレッジ](docs/TESTING.md)
- [🤝 貢献ガイド](CONTRIBUTING.md)

---

## 🛠️ 開発

### 動作確認

```bash
# デモスクリプト実行（推奨）
./demo.sh

# 全テストとカバレッジ確認
./run-tests.sh --coverage

# HTMLカバレッジレポート生成
./run-tests.sh --html

# 統合テストも含めて実行
./run-tests.sh --integration
```

### テスト実行

```bash
# 全テスト実行
go test ./...

# ユニットテストのみ
go test ./internal/...

# 統合テスト
go test ./test/...

# カバレッジ付き
go test ./... -cover

# カバレッジレポート
go test ./internal/analyzer -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### テストカバレッジ

- `internal/analyzer`: 96.9%
- `internal/reporter`: 96.5%
- `internal/github`: 9.1% (モックなしAPI呼び出しのため低い)
- **全体**: 83.0%

### プロジェクト構造

```
auto-issue-finder/
├── cmd/
│   ├── sync.go               # 同期処理型自律開発コマンド
│   ├── analyze.go            # GitHub Issue分析コマンド
│   └── root.go               # CLI ルート
├── internal/
│   ├── github/               # GitHub API クライアント
│   ├── analyzer/             # Issue 分析ロジック (96.9% coverage)
│   └── reporter/             # レポート生成 (96.5% coverage)
├── bin/                      # ビルド成果物
│   └── auto-issue-finder     # 実行ファイル
├── auto-dev.sh               # 自律開発スクリプト (bash版)
├── auto-dev-incremental.sh   # インクリメンタルコミット版
├── run-overnight.sh          # バックグラウンド実行
└── install-auto-dev.sh       # インストーラー
```

---

## 🤝 貢献

貢献を歓迎します！以下の方法で参加できます:

1. 🐛 [Issueを報告](https://github.com/isiidaisuke0926/auto-issue-finder/issues)
2. 💡 機能提案
3. 🔧 プルリクエストの送信
4. 📖 ドキュメントの改善

詳細は [CONTRIBUTING.md](CONTRIBUTING.md) をご覧ください。

---

## 📄 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) をご覧ください。

---

## 🙏 謝辞

使用しているライブラリ:
- [go-github](https://github.com/google/go-github) - GitHub API client
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [godotenv](https://github.com/joho/godotenv) - Environment variables
- [testify](https://github.com/stretchr/testify) - Testing toolkit

---

**Made with ❤️ and Claude Code**
