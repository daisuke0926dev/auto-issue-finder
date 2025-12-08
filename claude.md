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
- 不要なmd等のファイルは作成しないでください。