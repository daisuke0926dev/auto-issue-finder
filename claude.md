# Sleepship - Claude Code 自律開発システム

## プロジェクト概要

Sleepship は Claude Code による自律開発システムです。タスクファイルから自動的にコードを実装し、検証まで行います。

### 重要: すべてのタスクは sleepship で実行すること

開発タスクを追加する場合は、以下の手順に従ってください:

1. `tasks-*.txt` ファイルを作成
2. `./bin/sleepship sync tasks-*.txt` で実行
3. `tasks-*.txt` ファイルは `.gitignore` で除外されているため、コミット不要

## アーキテクチャ

- **cmd/sync.go**: メインの同期実行ロジック
- **cmd/init.go**: タスクファイルの初期化
- **cmd/root.go**: CLI ルート

## タスクファイルフォーマット

タスクファイルはマークダウン形式で記述します:

```markdown
## タスク[番号]: タイトル

タスクの詳細説明

### 確認

検証コマンドを記載
```

### フォーマット詳細

- `## タスク[番号]: タイトル` で各タスクを開始
- `### 確認` セクションで検証コマンドを指定
- `### 前提条件` セクション（オプション）で依存タスクを指定
- タスク番号は連番で管理

### 確認ステップの重要性

`### 確認` セクションは、タスクの実装が正しく完了したことを自動的に検証するために**必須**です。

#### 確認ステップの役割

1. **実装の正確性保証**: コードが期待通りに動作することを自動検証
2. **リトライの判定基準**: 確認コマンドが失敗した場合、自動的にリトライ
3. **品質維持**: テストやビルドを強制することで、品質を担保

#### 確認コマンドの書き方

確認コマンドは複数行で記述でき、すべてのコマンドが成功する必要があります：

```markdown
### 確認
go test ./...
go build
test -f expected-file.txt
```

#### 良い確認コマンドの例

```markdown
### 確認
# テストの実行
go test ./internal/task/...

# ビルドの確認
go build -o bin/sleepship

# 生成ファイルの存在確認
test -f .sleepship/history.json

# 統合テスト
./bin/sleepship --help | grep "history"
```

#### 避けるべき確認コマンド

- **副作用のあるコマンド**: データベースの削除、ファイルの変更など
- **曖昧な成功条件**: 出力が不安定なコマンド
- **時間がかかりすぎるコマンド**: E2Eテストなど（別タスクに分離推奨）

### タスクの前提条件

タスクに依存関係がある場合、`### 前提条件` セクションで明示的に指定できます。

#### 基本構文

```markdown
## タスク3: データベースマイグレーション実行

データベースのマイグレーションを実行します。

### 前提条件
- タスク1
- タスク2

### 確認
./bin/sleepship db-status | grep "all migrations applied"
```

#### 前提条件の動作

- **実行順序の保証**: 前提条件のタスクが完了するまで実行されない
- **失敗時の挙動**: 前提タスクが失敗した場合、依存タスクはスキップ
- **循環依存の検出**: 循環依存がある場合、エラーで停止

#### 前提条件の実用例

```markdown
## タスク1: 環境構築

開発環境をセットアップします。

### 確認
go version
test -f go.mod

## タスク2: 依存パッケージのインストール

必要なパッケージをインストールします。

### 前提条件
- タスク1

### 確認
go mod verify

## タスク3: アプリケーションのビルド

アプリケーションをビルドします。

### 前提条件
- タスク1
- タスク2

### 確認
test -f bin/sleepship
./bin/sleepship --version
```

## リトライ機能

Sleepship には自動リトライ機能が組み込まれています:

- `--max-retries` オプション（デフォルト: 3回）
- タスク実行失敗時に自動リトライ
- 検証失敗時に自動修正+再検証

### 使用例

```bash
# デフォルト（3回まで）
./bin/sleepship sync tasks-feature.txt

# リトライ回数を指定
./bin/sleepship sync tasks-feature.txt --max-retries 5
```

### リトライの仕組み

Sleepship は以下のフローでタスクを実行し、失敗時に自動的にリトライします。

#### 実行フロー

```
1. タスク実行開始
   ↓
2. Claude Code に実装指示を送信
   ↓
3. 確認コマンドを実行
   ↓
4. 成功？
   ├─ Yes → 次のタスクへ
   └─ No  → リトライカウント確認
            ├─ リトライ可能 → ステップ2へ戻る
            └─ 上限到達   → タスク失敗として記録
```

#### リトライ時の動作

リトライ時には、以下の情報が Claude Code に渡されます：

1. **元のタスク内容**: 実装すべき機能の説明
2. **失敗した確認コマンド**: どのコマンドが失敗したか
3. **エラー出力**: コマンドの標準エラー出力
4. **リトライ回数**: 現在何回目のリトライか

これにより、Claude Code は失敗原因を分析し、修正を試みます。

#### リトライ回数の選び方

| 状況 | 推奨リトライ回数 | 理由 |
|------|-----------------|------|
| 単純なタスク | 1-3回 | すぐに成功するはず。失敗なら設計見直し |
| 複雑なタスク | 3-5回 | 複数の修正が必要な可能性 |
| 実験的な実装 | 5-10回 | 試行錯誤が必要 |
| 本番デプロイ | 1回 | 失敗時は手動確認すべき |

### タスク失敗時の挙動

タスクが最大リトライ回数に達して失敗した場合、以下の処理が実行されます。

#### 失敗時の処理

1. **実行停止**: 残りのタスクは実行されない
2. **履歴への記録**: 失敗が `.sleepship/history.json` に記録される
   - 失敗したタスク番号
   - エラーメッセージ
   - リトライ回数
   - 実行時間
3. **終了コード**: プロセスが終了コード 1 で終了

#### 失敗後の対処方法

##### 方法1: 失敗したタスクから再開

```bash
# タスク3で失敗した場合、タスク3から再実行
./bin/sleepship sync tasks-feature.txt --start-from=3
```

##### 方法2: リトライ回数を増やして再実行

```bash
# より多くのリトライを許可
./bin/sleepship sync tasks-feature.txt --max-retries=10
```

##### 方法3: タスク内容を修正して再実行

失敗原因がタスク定義の曖昧さにある場合、タスクファイルを修正：

```markdown
## タスク3: API実装

# 修正前（曖昧）
APIを実装してください。

# 修正後（具体的）
以下の仕様でREST APIを実装してください：
- エンドポイント: POST /api/users
- リクエスト: JSON形式でname, emailを受け取る
- レスポンス: 作成されたユーザー情報を返す
- エラーハンドリング: バリデーションエラーは400を返す

### 確認
go test ./internal/api/...
curl -X POST http://localhost:8080/api/users -d '{"name":"test","email":"test@example.com"}'
```

##### 方法4: 手動で修正してタスクをスキップ

```bash
# 手動でタスク3を実装
vim internal/api/user.go

# タスク4から実行
./bin/sleepship sync tasks-feature.txt --start-from=4
```

#### 失敗の分析

履歴コマンドで失敗の詳細を確認：

```bash
# 失敗した実行のみ表示
./bin/sleepship history --failed

# 出力例:
# ❌ tasks-feature.txt (2025-12-11 10:30:00)
#    Status: Failed
#    Failed at: Task 3
#    Error: verification command failed: go test ./...
#    Retries: 3/3
#    Duration: 5m30s
```

#### よくある失敗パターンと対処法

| 失敗パターン | 原因 | 対処法 |
|-------------|------|--------|
| テスト失敗 | 実装の不具合 | タスク説明を具体化、リトライ増 |
| ビルド失敗 | 構文エラー | タスクを小さく分割 |
| ファイル不在 | 生成パス間違い | 期待パスを明示 |
| タイムアウト | 処理時間超過 | タスクを分割、または--max-retries増 |
| 依存エラー | 前提条件未完了 | 前提条件セクション追加 |

## 再帰実行機能（動的タスク生成）

Sleepship はタスクファイル内で `./bin/sleepship` コマンドを実行できます。これにより、実行時に動的にタスクを生成し、段階的な開発フローを実現できます。

### 特徴

- **動的タスク生成**: タスク実行中に新しいタスクファイルを作成して実行可能
- **再帰深度制限**: 環境変数 `SLEEPSHIP_DEPTH` で再帰深度を管理（最大3階層）
- **段階的開発**: 調査→計画→実装のフローを自動化

### 使用例

#### パターン1: 調査→計画→実装フロー

```markdown
## タスク1: APIエンドポイント調査

既存のAPIエンドポイントを調査し、新規エンドポイント実装計画をtasks-api-impl.txtに出力してください。

### 確認
- `test -f tasks-api-impl.txt`

## タスク2: 実装実行

調査結果を元に実装を実行してください。

### 確認
- `./bin/sleepship sync tasks-api-impl.txt`
- `go test ./...`
```

#### パターン2: サブタスク自動生成

```markdown
## タスク1: 機能分析とタスク分解

大きな機能要件を分析し、実装可能な単位にタスクを分解してtasks-subtasks.txtに出力してください。

### 確認
- `test -f tasks-subtasks.txt`

## タスク2: サブタスク実行

分解されたタスクを順次実行してください。

### 確認
- `./bin/sleepship sync tasks-subtasks.txt`
```

### 再帰深度の制限

再帰実行は最大3階層まで制限されています:

- **階層1**: メインタスクファイル (`SLEEPSHIP_DEPTH=1`)
- **階層2**: サブタスクファイル (`SLEEPSHIP_DEPTH=2`)
- **階層3**: サブサブタスクファイル (`SLEEPSHIP_DEPTH=3`)
- **階層4以降**: エラーで停止（無限再帰を防止）

深度超過時のエラーメッセージ:
```
最大再帰深度(3)に達しました。これ以上のsleepship実行はできません。
```

### ベストプラクティス

1. **タスクファイル命名規則**: 用途がわかる名前をつける
   - `tasks-investigation.txt` (調査)
   - `tasks-plan.txt` (計画)
   - `tasks-impl.txt` (実装)

2. **検証の徹底**: 生成したタスクファイルの存在確認を行う
   ```markdown
   ### 確認
   - `test -f tasks-impl.txt`
   ```

3. **明確な指示**: 生成するタスクファイルの内容を具体的に指示
   ```markdown
   以下の形式でtasks-impl.txtを作成してください:
   - タスク1: モデル実装
   - タスク2: テスト追加
   ```

4. **段階的実行**: 調査→計画→実装の順で段階を分ける
   - 各段階で適切な情報を次の段階に渡す
   - 中間成果物（タスクファイル）を検証する

## 環境変数による設定

Sleepship は環境変数による設定オーバーライドをサポートしています。優先順位は以下の通りです：

**設定の優先順位**: CLIフラグ > 環境変数 > デフォルト値

### サポートされる環境変数

| 環境変数 | 説明 | 例 |
|---------|------|-----|
| `SLEEPSHIP_PROJECT_DIR` | プロジェクトディレクトリ | `/path/to/project` |
| `SLEEPSHIP_SYNC_DEFAULT_TASK_FILE` | デフォルトタスクファイル | `tasks-default.txt` |
| `SLEEPSHIP_SYNC_MAX_RETRIES` | 最大リトライ回数 | `5` |
| `SLEEPSHIP_SYNC_LOG_DIR` | ログ出力ディレクトリ | `./logs` |
| `SLEEPSHIP_SYNC_START_FROM` | 開始タスク番号 | `3` |
| `SLEEPSHIP_CLAUDE_FLAGS` | Claude Codeフラグ（カンマ区切り） | `--flag1,--flag2` |

### 使用例

#### CI/CD環境での設定

```bash
# GitHub Actionsなどでの使用例
export SLEEPSHIP_PROJECT_DIR=/workspace
export SLEEPSHIP_SYNC_MAX_RETRIES=10
export SLEEPSHIP_SYNC_LOG_DIR=/logs

./bin/sleepship sync tasks-ci.txt
```

#### 開発環境での設定

```bash
# .bashrc や .zshrc に設定
export SLEEPSHIP_SYNC_MAX_RETRIES=5
export SLEEPSHIP_SYNC_LOG_DIR=~/sleepship-logs

# デフォルト設定で実行
./bin/sleepship sync tasks-dev.txt
```

#### 一時的な設定変更

```bash
# 特定の実行時のみ設定を変更
SLEEPSHIP_SYNC_MAX_RETRIES=10 ./bin/sleepship sync tasks-critical.txt
```

## エイリアス機能

頻繁に使用するコマンドをエイリアスとして定義できます。

### 設定方法

プロジェクトディレクトリまたはホームディレクトリに `.sleepship.toml` ファイルを作成：

```toml
[aliases]
dev = "sync tasks-dev.txt"
test = "sync tasks-test.txt --max-retries=5"
prod = "sync tasks-prod.txt --max-retries=10"
staging = "sync tasks-staging.txt --dir=/path/to/staging"
```

### 使用方法

```bash
# エイリアスを使用して実行
./bin/sleepship dev
./bin/sleepship test
./bin/sleepship prod

# エイリアス一覧を表示
./bin/sleepship alias list

# 特定のエイリアスの詳細を表示
./bin/sleepship alias get dev
```

### エイリアスの連鎖

エイリアスから別のエイリアスを参照できます：

```toml
[aliases]
base = "sync tasks-base.txt"
extended = "@base --max-retries=5"  # baseエイリアスを参照
```

### 実用例

```toml
[aliases]
# 開発フロー
quick = "sync tasks-quick.txt --max-retries=1"
normal = "sync tasks-dev.txt"
careful = "sync tasks-dev.txt --max-retries=10"

# 環境別
local = "sync tasks-local.txt"
staging = "sync tasks-staging.txt --dir=/path/to/staging"
production = "sync tasks-prod.txt --dir=/path/to/prod --max-retries=10"

# 特定機能
db-migrate = "sync tasks-db-migrate.txt"
api-deploy = "sync tasks-api-deploy.txt --max-retries=5"
```

## タスク実行履歴

Sleepship はすべてのタスク実行履歴を記録します。これにより、過去の実行結果を確認し、トラブルシューティングに活用できます。

### 履歴の表示

```bash
# すべての実行履歴を表示
./bin/sleepship history

# 最新5件の履歴を表示
./bin/sleepship history --last 5

# 最新の実行結果のみ表示
./bin/sleepship history --last 1

# 失敗した実行のみ表示
./bin/sleepship history --failed
```

### 履歴に記録される情報

- タスクファイル名
- 実行日時
- 成功/失敗ステータス
- 実行時間
- タスク数
- リトライ回数
- ブランチ名
- エラーメッセージ（失敗時）

### 履歴ファイルの場所

履歴は `.sleepship/history.json` に保存されます。このファイルは Git 管理対象外です。

### トラブルシューティングでの活用例

```bash
# 最近失敗したタスクを確認
./bin/sleepship history --failed

# 失敗したタスクから再実行
./bin/sleepship sync tasks-feature.txt --start-from=3

# 実行時間が長いタスクを特定
./bin/sleepship history | grep "Duration"
```

### 履歴の統計情報

履歴表示時には以下の統計情報が表示されます：

- 総実行回数
- 成功回数
- 失敗回数
- 合計実行時間

## 開発フロー

1. **タスクファイルを作成**
   ```bash
   # 例: 新機能開発用タスク
   cat > tasks-feature.txt << 'EOF'
   ## タスク1: 機能Aの実装

   機能Aの詳細説明

   ### 確認
   go test ./...
   EOF
   ```

2. **Sleepship を実行**
   ```bash
   ./bin/sleepship sync tasks-feature.txt
   ```

3. **ログを確認**
   ```bash
   # リアルタイムでログを監視
   tail -f logs/sync-*.log
   ```

4. **完了後の処理**
   - タスクファイルは削除してOK（`.gitignore` で除外済み）
   - 実装内容をコミット・プッシュ

## ディレクトリ構成

```
sleepship/
├── bin/
│   └── sleepship         # 実行ファイル
├── cmd/
│   ├── root.go          # CLI ルート
│   ├── sync.go          # 同期実行ロジック
│   └── init.go          # 初期化コマンド
├── logs/                # 実行ログ（.gitignore）
├── tasks-*.txt          # タスクファイル（.gitignore）
└── claude.md           # このファイル
```

## コマンド一覧

```bash
# タスクファイルのテンプレート作成
./bin/sleepship init

# タスクの実行
./bin/sleepship sync <タスクファイル>

# リトライ回数を指定して実行
./bin/sleepship sync <タスクファイル> --max-retries 5

# エイリアスの管理
./bin/sleepship alias list              # エイリアス一覧表示
./bin/sleepship alias get <name>        # 特定のエイリアス表示

# 実行履歴の確認
./bin/sleepship history                 # すべての履歴表示
./bin/sleepship history --last 5        # 最新5件表示
./bin/sleepship history --failed        # 失敗した実行のみ表示

# ヘルプ表示
./bin/sleepship --help
```

## 開発ガイドライン

### コード品質管理

#### golangci-lintの設定

プロジェクトでは `.golangci.yml` で以下のlinterを有効化しています：

- **misspell**: スペルミスの検出
- **gosec**: セキュリティ問題の検出
- **revive**: Goのコーディング規約チェック
- **errorlint**: エラーハンドリングの確認（%wを使ったエラーラッピング）
- **govet**: Go公式の静的解析ツール
- **errcheck**: エラーハンドリングの確認
- **staticcheck**: 高度な静的解析
- **unused**: 未使用コードの検出
- **ineffassign**: 非効率な代入の検出

#### CIでのlint実行

GitHub Actionsで自動的にlintが実行されます：

- **プッシュ時（mainブランチ）**: golangci-lint が自動実行
- **プルリクエスト時**: lint結果がチェックされる
- **失敗時**: マージがブロックされる

設定ファイル: `.github/workflows/lint.yml`

#### コミット前の確認事項

コミット前に以下を必ず実行してください：

1. **テスト実行**: `go test ./...` がすべて成功すること
2. **ビルド確認**: `go build` が成功すること

CIでlintが自動実行されるため、ローカルでのlint実行は必須ではありません。

## 注意事項

- タスクファイル (`tasks-*.txt`) は Git 管理対象外です
- ログファイルは `logs/` ディレクトリに保存されます（Git 管理対象外）
- すべての開発タスクは Sleepship 経由で実行してください
- タスクの検証コマンドは必ず記載してください
- 不要なmd等のファイルは作成しないでください