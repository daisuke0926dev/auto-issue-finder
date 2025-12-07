# インストールガイド

このガイドでは、Auto Issue Finderの2つのツールをインストールする方法を説明します。

## 目次

- [Claude Code自律開発システム](#claude-code自律開発システム)
- [GitHub Issue Analyzer](#github-issue-analyzer)
- [前提条件](#前提条件)
- [トラブルシューティング](#トラブルシューティング)

---

## Claude Code自律開発システム

### 他のプロジェクトへのインストール

最も簡単な方法は、インストールスクリプトを使用することです:

```bash
curl -sSL https://raw.githubusercontent.com/isiidaisuke0926/auto-issue-finder/main/install-auto-dev.sh | bash
```

### インストールされるファイル

インストールスクリプトは以下のファイルをプロジェクトにコピーします:

| ファイル | 説明 |
|---------|------|
| `auto-dev.sh` | 基本実行スクリプト |
| `auto-dev-with-commits.sh` | 完了後に1つのコミットを作成 |
| `auto-dev-incremental.sh` | タスクごとに個別コミット |
| `run-overnight.sh` | バックグラウンド実行ラッパー |
| `.claude/settings.local.json` | 自動承認設定 |
| `tonight-with-tasks.txt.example` | サンプルタスクファイル |

### 手動インストール

スクリプトを使わずに手動でインストールする場合:

```bash
# リポジトリをクローン
git clone https://github.com/isiidaisuke0926/auto-issue-finder.git
cd auto-issue-finder

# 必要なファイルをコピー
cp auto-dev*.sh /path/to/your/project/
cp run-overnight.sh /path/to/your/project/
cp -r .claude /path/to/your/project/
cp tonight-with-tasks.txt.example /path/to/your/project/

# 実行権限を付与
cd /path/to/your/project
chmod +x auto-dev*.sh run-overnight.sh
```

### .gitignoreの設定

タスクファイルをgitで管理しない場合は、`.gitignore`に追加:

```bash
echo "tonight.txt" >> .gitignore
echo "test-*.txt" >> .gitignore
```

---

## GitHub Issue Analyzer

### 前提条件

- **Go 1.21以上** - [公式サイト](https://go.dev/)からインストール
- **GitHub Personal Access Token** - [作成方法](#github-tokenの作成)

### ソースからビルド

```bash
# 1. リポジトリをクローン
git clone https://github.com/isiidaisuke0926/auto-issue-finder.git
cd auto-issue-finder

# 2. 依存関係をダウンロード
go mod download

# 3. ビルド
go build -o auto-issue-finder cmd/analyze/main.go

# 4. （オプション）グローバルにインストール
go install cmd/analyze/main.go
```

### バイナリのインストール先

グローバルインストールした場合、バイナリは以下の場所に配置されます:

```bash
# Goのbinディレクトリを確認
go env GOPATH

# 通常は以下のパス
# macOS/Linux: $HOME/go/bin/main
# Windows: %USERPROFILE%\go\bin\main.exe
```

PATHに追加:

```bash
# .bashrc または .zshrc に追加
export PATH="$PATH:$(go env GOPATH)/bin"
```

### GitHub Tokenの作成

1. GitHubにログイン
2. **Settings** → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
3. **Generate new token** をクリック
4. スコープを選択:
   - `public_repo` - パブリックリポジトリ用
   - `repo` - プライベートリポジトリ用
5. トークンをコピー（**一度しか表示されません**）

### GitHub Tokenの設定

#### 方法1: 環境変数

```bash
# 一時的に設定
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx

# 永続的に設定（.bashrc / .zshrc に追加）
echo 'export GITHUB_TOKEN=ghp_xxxxxxxxxxxx' >> ~/.bashrc
source ~/.bashrc
```

#### 方法2: .envファイル

```bash
# プロジェクトルートに.envファイルを作成
cat > .env << EOF
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
EOF

# .gitignoreに追加（重要！）
echo ".env" >> .gitignore
```

#### 方法3: コマンドラインフラグ

```bash
./auto-issue-finder analyze owner/repo --token=ghp_xxxxxxxxxxxx
```

---

## トラブルシューティング

### Claude Code自律開発システム

#### "permission denied" エラー

```bash
chmod +x auto-dev.sh auto-dev-incremental.sh run-overnight.sh
```

#### スクリプトが見つからない

```bash
# 現在のディレクトリを確認
pwd

# ファイルが存在するか確認
ls -la auto-dev.sh
```

### GitHub Issue Analyzer

#### "invalid token" エラー

- トークンが正しいか確認
- トークンの有効期限を確認
- 適切なスコープ（`repo` または `public_repo`）が設定されているか確認

#### "go: command not found"

Goがインストールされていません:

```bash
# macOS (Homebrew)
brew install go

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# バージョン確認
go version
```

#### ビルドエラー

```bash
# 依存関係をクリーンアップして再取得
go clean -modcache
go mod download
go build -o auto-issue-finder cmd/analyze/main.go
```

#### レート制限エラー

GitHub APIのレート制限に達した場合:

```bash
# レート制限状況を確認
./auto-issue-finder analyze owner/repo --verbose

# 認証済み: 5,000 requests/hour
# 未認証: 60 requests/hour
```

対策:
1. GitHub Tokenを使用する（レート制限が大幅に緩和）
2. `--limit`フラグで取得件数を制限
3. しばらく待ってから再実行

---

## 次のステップ

- [使用方法](USAGE.md) - コマンドリファレンスと詳細な使い方
- [自律開発システム詳細](AUTO_DEV.md) - タスクファイルの書き方
- [テストガイド](TESTING.md) - テストの実行方法
