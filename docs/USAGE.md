# 使用方法

このガイドでは、Auto Issue Finderの2つのツールの詳細な使用方法を説明します。

## 目次

- [Claude Code自律開発システム](#claude-code自律開発システム)
- [GitHub Issue Analyzer](#github-issue-analyzer)

---

## Claude Code自律開発システム

### 基本的な使い方

```bash
# 1. タスクファイルを作成
cp tonight-with-tasks.txt.example tonight.txt
vim tonight.txt

# 2. 実行
./auto-dev.sh tonight.txt
```

### 実行モード

#### 1. 基本実行（対話的）

```bash
./auto-dev.sh tonight.txt
```

- 実行前に確認プロンプトあり
- 標準出力に進捗表示
- コミットは手動

#### 2. コミット付き実行

```bash
./auto-dev-with-commits.sh tonight.txt
```

- タスク完了後に1つのコミットを作成
- コミット前後のハッシュを表示

#### 3. インクリメンタルコミット

```bash
./auto-dev-incremental.sh tonight.txt
```

- タスクファイルを `# タスク` または `# Task` で分割
- 各タスクごとに個別コミット
- タスクファイル内にコミットメッセージを記述可能

#### 4. バックグラウンド実行

```bash
./run-overnight.sh tonight.txt
```

- nohupでバックグラウンド実行
- ログは `nohup.out` に出力
- 実行中でもログ監視可能: `tail -f nohup.out`

### タスクファイルの書き方

#### 基本フォーマット

```markdown
今夜のタスク: [タスク名]

## 要件
- 要件1
- 要件2

## 制約
- 制約1
- 制約2

## 参考
- 参考情報
```

#### インクリメンタルコミット用

タスクごとにセクションを分割:

```markdown
# タスク1: データベース設計

users テーブルを作成してください。

## カラム
- id (UUID)
- email (string, unique)
- created_at (timestamp)

完了したら以下のメッセージでコミット:
"feat: usersテーブル追加"

---

# タスク2: API実装

RESTful APIを実装してください。

## エンドポイント
- GET /users
- POST /users
- PUT /users/:id
- DELETE /users/:id

完了したらコミット:
"feat: User CRUD API実装"
```

### 実行例

#### 例1: Webアプリケーション開発

```bash
cat > tonight.txt << 'EOF'
今夜のタスク: ブログシステムのMVP実装

## 要件
- 記事のCRUD機能
- マークダウンエディタ
- 記事一覧ページ
- テストカバレッジ70%以上

## 技術スタック
- Go 1.21
- chi router
- SQLite
- テンプレートエンジン: html/template

## 制約
- 認証機能は不要
- デプロイ設定は不要
- シンプルで読みやすいコード
EOF

./run-overnight.sh tonight.txt
```

#### 例2: CLIツール開発

```bash
cat > tonight.txt << 'EOF'
# タスク1: CLIの基本構造

cobraを使ってCLIの基本構造を作成。

コマンド構成:
- mytool version
- mytool config

完了したらコミット: "feat: CLI基本構造実装"

---

# タスク2: 設定ファイル読み込み

YAML形式の設定ファイルを読み込む機能を実装。

~/.mytool/config.yaml から設定を読み込む。

完了したらコミット: "feat: 設定ファイル読み込み機能追加"

---

# タスク3: テスト追加

ユニットテストとテストカバレッジを80%以上に。

完了したらコミット: "test: ユニットテスト追加"
EOF

./auto-dev-incremental.sh tonight.txt
```

### ログ確認

#### リアルタイム監視

```bash
# バックグラウンド実行中
tail -f nohup.out

# 最後の50行を表示
tail -50 nohup.out

# エラーのみ確認
grep ERROR nohup.out
```

#### 実行結果の確認

```bash
# プロセス確認
ps aux | grep claude

# 完了確認
cat nohup.out | grep "完了"

# コミット確認
git log --oneline -5
```

---

## GitHub Issue Analyzer

### 基本コマンド

```bash
auto-issue-finder analyze [owner/repo] [flags]
```

### コマンドラインフラグ

| フラグ | 説明 | デフォルト | 例 |
|--------|------|------------|-----|
| `--token` | GitHub Personal Access Token | `$GITHUB_TOKEN` | `--token=ghp_xxx` |
| `--state` | Issueの状態でフィルタ | `all` | `--state=open` |
| `--labels` | ラベルでフィルタ（カンマ区切り） | `[]` | `--labels=bug,help-wanted` |
| `--format` | 出力形式 | `markdown` | `--format=json` |
| `--output` | 出力ファイルパス | `stdout` | `--output=report.md` |
| `--limit` | 最大取得Issue数（0=全て） | `0` | `--limit=100` |
| `--verbose` | 詳細ログを有効化 | `false` | `--verbose` |

### 出力形式

#### 1. Console形式

ターミナルで見やすい形式:

```bash
./auto-issue-finder analyze microsoft/vscode --format=console --limit=50
```

出力例:
```
🔍 Analyzing microsoft/vscode...
✓ Fetched 50 issues

📊 Issue Statistics
─────────────────────────────────
Total Issues:        50
Open:                42 (84%)
Closed:              8 (16%)
Avg Resolution Time: 12.3 days

📋 Label Distribution
─────────────────────────────────
bug                  23 (46%)
feature-request      15 (30%)
enhancement          8 (16%)
```

#### 2. Markdown形式

詳細レポート生成:

```bash
./auto-issue-finder analyze golang/go \
  --state=open \
  --format=markdown \
  --output=report.md
```

レポートに含まれる内容:
- 統計テーブル
- ラベル分布
- 月次トレンドチャート（ASCII）
- 長期未解決Issue一覧
- 高活動Issue一覧
- 優先度付き推奨事項

#### 3. JSON形式

自動化・プログラム処理用:

```bash
./auto-issue-finder analyze owner/repo \
  --format=json \
  --output=analysis.json

# jqで処理
cat analysis.json | jq '.Stats.TotalIssues'
cat analysis.json | jq '.Patterns.LongStandingIssues | length'
cat analysis.json | jq '.Recommendations[0]'
```

### フィルタリング

#### 状態でフィルタ

```bash
# オープンなIssueのみ
./auto-issue-finder analyze owner/repo --state=open

# クローズしたIssueのみ
./auto-issue-finder analyze owner/repo --state=closed

# 全て
./auto-issue-finder analyze owner/repo --state=all
```

#### ラベルでフィルタ

```bash
# bugラベルのみ
./auto-issue-finder analyze owner/repo --labels=bug

# 複数ラベル（OR条件）
./auto-issue-finder analyze owner/repo --labels=bug,enhancement

# 状態とラベルの組み合わせ
./auto-issue-finder analyze owner/repo \
  --state=open \
  --labels=bug,critical
```

#### 件数制限

```bash
# 最初の100件のみ
./auto-issue-finder analyze owner/repo --limit=100

# 最初の10件（テスト用）
./auto-issue-finder analyze owner/repo --limit=10
```

### 実用例

#### 例1: 週次レポート自動生成

```bash
#!/bin/bash
# weekly-report.sh

REPO="microsoft/vscode"
DATE=$(date +%Y-%m-%d)
OUTPUT="reports/weekly-report-${DATE}.md"

./auto-issue-finder analyze $REPO \
  --state=open \
  --format=markdown \
  --output=$OUTPUT

echo "レポート生成完了: $OUTPUT"
```

#### 例2: バグの優先度分析

```bash
# 高活動のバグを抽出
./auto-issue-finder analyze owner/repo \
  --state=open \
  --labels=bug \
  --format=json \
  --output=bugs.json

# 20コメント以上のバグをリスト化
cat bugs.json | jq '.Patterns.HotTopics[] | select(.Comments > 20)'
```

#### 例3: 複数リポジトリの一括分析

```bash
#!/bin/bash
# analyze-all.sh

REPOS=(
  "golang/go"
  "rust-lang/rust"
  "microsoft/TypeScript"
)

for repo in "${REPOS[@]}"; do
  echo "Analyzing $repo..."

  slug=$(echo $repo | tr '/' '-')
  ./auto-issue-finder analyze $repo \
    --format=markdown \
    --output="reports/${slug}.md"
done

echo "全リポジトリの分析完了"
```

### トラブルシューティング

#### レート制限

```bash
# 詳細ログで確認
./auto-issue-finder analyze owner/repo --verbose

# 件数を制限して実行
./auto-issue-finder analyze owner/repo --limit=50
```

#### トークンエラー

```bash
# トークンの確認
echo $GITHUB_TOKEN

# トークンを明示的に指定
./auto-issue-finder analyze owner/repo --token=ghp_xxxxx
```

#### Issue未取得

```bash
# 状態フィルタを確認
./auto-issue-finder analyze owner/repo --state=all

# 詳細ログで原因確認
./auto-issue-finder analyze owner/repo --verbose
```

---

## 次のステップ

- [インストールガイド](INSTALL.md) - セットアップ方法
- [自律開発システム詳細](AUTO_DEV.md) - より高度な使い方
- [テストガイド](TESTING.md) - テストの実行方法
