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
- **環境変数設定** - 環境変数による設定オーバーライド
- **コマンドエイリアス** - 頻繁に使用するコマンドのショートカット定義
- **実行履歴管理** - タスク実行履歴の記録と検索

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

## 高度な使い方

### 動的タスク生成（再帰実行）

Sleepshipはタスク内で新しいタスクファイルを生成し、それを実行できます。これにより、段階的な開発フローを自動化できます。

#### 使用例1: 調査→計画→実装フロー

```markdown
## タスク1: 既存コードの調査

既存のユーザー認証機能を調査し、OAuth2.0対応の実装計画をtasks-oauth-impl.txtに出力してください。

調査すべき項目:
- 現在の認証フロー
- データベーススキーマ
- 必要な変更点

### 確認
- `test -f tasks-oauth-impl.txt`

## タスク2: 実装の実行

生成された実装計画に従って、OAuth2.0対応を実装してください。

### 確認
- `./bin/sleepship sync tasks-oauth-impl.txt`
- `go test ./auth/...`
```

#### 使用例2: 大規模機能の段階的実装

```markdown
## タスク1: 要件分析とタスク分解

以下の要件を実装可能な単位に分解し、tasks-feature-steps.txtを生成してください:

要件:
- ユーザープロフィール管理機能
- プロフィール画像アップロード
- プロフィール編集履歴
- プライバシー設定

出力形式:
各機能を独立したタスクに分解し、依存関係を明記する

### 確認
- `test -f tasks-feature-steps.txt`

## タスク2: 段階的実装

分解されたタスクを順次実行してください。

### 確認
- `./bin/sleepship sync tasks-feature-steps.txt`
- `go test ./...`
```

#### 使用例3: テスト駆動開発フロー

```markdown
## タスク1: テストケース設計

新規機能のテストケースを設計し、tasks-tdd.txtに出力してください。

テストケースに含めるべき項目:
- 正常系テスト
- 異常系テスト
- 境界値テスト

### 確認
- `test -f tasks-tdd.txt`

## タスク2: TDD実行

テストケースに基づいて実装を進めてください。

### 確認
- `./bin/sleepship sync tasks-tdd.txt`
```

### 再帰深度の制限

再帰実行は**最大3階層**まで制限されています:

```
tasks-main.txt          # 階層1
└── tasks-sub.txt       # 階層2
    └── tasks-subsub.txt # 階層3
        └── (エラー)     # 階層4は実行できない
```

### 動的タスク生成のベストプラクティス

1. **明確な命名規則**
   ```markdown
   - tasks-investigation.txt  # 調査フェーズ
   - tasks-design.txt         # 設計フェーズ
   - tasks-implementation.txt # 実装フェーズ
   ```

2. **検証の徹底**
   ```markdown
   ### 確認
   - `test -f tasks-next.txt`              # ファイル生成確認
   - `./bin/sleepship sync tasks-next.txt` # 実行確認
   - `go test ./...`                       # 最終検証
   ```

3. **段階的な進行**
   - 各タスクで適切な中間成果物を生成
   - 次のタスクに必要な情報を明示
   - 各段階で検証を実施

4. **適切な粒度**
   - 1つのタスクファイルは5-10タスク程度
   - 複雑すぎる場合は更に分解
   - シンプルな場合は直接実装

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

## 環境変数による設定

優先順位: **CLIフラグ > 環境変数 > デフォルト値**

### サポートされる環境変数

| 環境変数 | 説明 | デフォルト |
|---------|------|-----------|
| `SLEEPSHIP_PROJECT_DIR` | プロジェクトディレクトリ | カレントディレクトリ |
| `SLEEPSHIP_SYNC_MAX_RETRIES` | 最大リトライ回数 | 3 |
| `SLEEPSHIP_SYNC_LOG_DIR` | ログ出力ディレクトリ | logs |
| `SLEEPSHIP_SYNC_START_FROM` | 開始タスク番号 | 1 |
| `SLEEPSHIP_CLAUDE_FLAGS` | Claude Codeフラグ（カンマ区切り） | - |

### CI/CD環境での使用例

```bash
# GitHub Actionsでの設定例
export SLEEPSHIP_PROJECT_DIR=/workspace
export SLEEPSHIP_SYNC_MAX_RETRIES=10
export SLEEPSHIP_SYNC_LOG_DIR=/logs

./bin/sleepship sync tasks-ci.txt
```

### 開発環境での使用例

```bash
# シェル設定ファイル（.bashrc, .zshrc等）に追加
export SLEEPSHIP_SYNC_MAX_RETRIES=5
export SLEEPSHIP_SYNC_LOG_DIR=~/sleepship-logs

# 環境変数が自動的に適用される
./bin/sleepship sync tasks-dev.txt

# 一時的に設定を変更
SLEEPSHIP_SYNC_MAX_RETRIES=10 ./bin/sleepship sync tasks-critical.txt
```

---

## コマンドエイリアス

頻繁に使用するコマンドをエイリアスとして定義できます。

### 設定ファイルの作成

プロジェクトディレクトリまたはホームディレクトリに `.sleepship.toml` を作成：

```toml
[aliases]
dev = "sync tasks-dev.txt"
test = "sync tasks-test.txt --max-retries=5"
prod = "sync tasks-prod.txt --max-retries=10"
staging = "sync tasks-staging.txt --dir=/path/to/staging"
```

### エイリアスの使用

```bash
# 通常のコマンドの代わりにエイリアスを使用
./bin/sleepship dev
./bin/sleepship test
./bin/sleepship prod

# エイリアス一覧を表示
./bin/sleepship alias list

# 特定のエイリアスの内容を確認
./bin/sleepship alias get dev
```

### 実用的なエイリアス設定例

```toml
[aliases]
# 開発フロー別
quick = "sync tasks-quick.txt --max-retries=1"
normal = "sync tasks-dev.txt"
careful = "sync tasks-dev.txt --max-retries=10"

# 環境別
local = "sync tasks-local.txt"
staging = "sync tasks-staging.txt --dir=/path/to/staging"
production = "sync tasks-prod.txt --dir=/path/to/prod --max-retries=10"

# 機能別
db-migrate = "sync tasks-db-migrate.txt"
api-deploy = "sync tasks-api-deploy.txt --max-retries=5"
frontend-build = "sync tasks-frontend-build.txt"
```

### エイリアスの連鎖

エイリアスから別のエイリアスを参照できます：

```toml
[aliases]
base = "sync tasks-base.txt"
extended = "@base --max-retries=5"  # baseエイリアスを参照
```

---

## タスク実行履歴

すべてのタスク実行履歴を自動記録し、後から確認できます。

### 履歴の表示

```bash
# すべての実行履歴を表示
./bin/sleepship history

# 最新5件の履歴を表示
./bin/sleepship history --last 5

# 最新の実行結果のみ表示（直前の実行を確認）
./bin/sleepship history --last 1

# 失敗した実行のみ表示
./bin/sleepship history --failed
```

### 履歴に記録される情報

- ✅ タスクファイル名
- ✅ 実行日時
- ✅ 成功/失敗ステータス
- ✅ 実行時間
- ✅ タスク数
- ✅ リトライ回数
- ✅ ブランチ名
- ✅ エラーメッセージ（失敗時）

### トラブルシューティングでの活用

```bash
# 1. 最近失敗したタスクを確認
./bin/sleepship history --failed

# 2. エラーの原因を特定
# （履歴にエラーメッセージが表示される）

# 3. 失敗したタスクから再実行
./bin/sleepship sync tasks-feature.txt --start-from=3

# 4. 実行結果を確認
./bin/sleepship history --last 1
```

### パフォーマンス分析

```bash
# 実行時間の長いタスクを特定
./bin/sleepship history | grep "Duration"

# 成功率の確認（統計情報が表示される）
./bin/sleepship history
```

### 履歴ファイルの場所

履歴は `.sleepship/history.json` に保存されます（Git管理対象外）。

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

## 開発

### Lint

コードの品質チェック:
```bash
make lint
```

自動修正:
```bash
make lint-fix
```

---

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) をご覧ください。
