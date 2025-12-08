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
- タスク番号は連番で管理

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

# ヘルプ表示
./bin/sleepship --help
```

## 注意事項

- タスクファイル (`tasks-*.txt`) は Git 管理対象外です
- ログファイルは `logs/` ディレクトリに保存されます（Git 管理対象外）
- すべての開発タスクは Sleepship 経由で実行してください
- タスクの検証コマンドは必ず記載してください
