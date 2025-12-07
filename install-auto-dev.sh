#!/bin/bash

# Claude Code 自律開発システム インストールスクリプト
# 任意のプロジェクトにこのシステムをインストールします

set -e

# 色定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# リポジトリURL（デフォルト）
REPO_URL="${REPO_URL:-https://raw.githubusercontent.com/isiidaisuke0926/auto-issue-finder/main}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Claude Code 自律開発システム インストール${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# カレントディレクトリ確認
echo -e "${YELLOW}インストール先:${NC} $(pwd)"
echo -e "${YELLOW}このディレクトリにインストールしますか？ (y/n)${NC}"
read -r answer
if [[ ! "$answer" =~ ^[Yy]$ ]]; then
    echo -e "${RED}インストールをキャンセルしました${NC}"
    exit 1
fi

echo

# 既存ファイルチェック
if [ -f "auto-dev.sh" ] || [ -f "auto-dev-incremental.sh" ]; then
    echo -e "${YELLOW}⚠️  既存のauto-dev関連ファイルが見つかりました${NC}"
    echo -e "${YELLOW}上書きしますか？ (y/n)${NC}"
    read -r overwrite
    if [[ ! "$overwrite" =~ ^[Yy]$ ]]; then
        echo -e "${RED}インストールをキャンセルしました${NC}"
        exit 1
    fi
fi

# ダウンロード関数
download_file() {
    local file=$1
    local url="${REPO_URL}/${file}"
    echo -e "${BLUE}📥 ${file} をダウンロード中...${NC}"

    if curl -sSfL "$url" -o "$file"; then
        echo -e "${GREEN}✓ ${file}${NC}"
    else
        echo -e "${RED}✗ ${file} のダウンロードに失敗しました${NC}"
        return 1
    fi
}

# メインスクリプトのダウンロード
echo -e "${BLUE}━━━ スクリプトファイルのダウンロード ━━━${NC}"
download_file "auto-dev.sh"
download_file "auto-dev-with-commits.sh"
download_file "auto-dev-incremental.sh"
download_file "run-overnight.sh"

# 実行権限付与
chmod +x auto-dev.sh auto-dev-with-commits.sh auto-dev-incremental.sh run-overnight.sh
echo

# .claudeディレクトリとsettings.local.jsonの設定
echo -e "${BLUE}━━━ Claude Code 設定ファイルの作成 ━━━${NC}"
mkdir -p .claude

if [ -f ".claude/settings.local.json" ]; then
    echo -e "${YELLOW}⚠️  .claude/settings.local.json が既に存在します${NC}"
    echo -e "${YELLOW}バックアップを作成して上書きしますか？ (y/n)${NC}"
    read -r backup_answer
    if [[ "$backup_answer" =~ ^[Yy]$ ]]; then
        cp .claude/settings.local.json ".claude/settings.local.json.backup.$(date +%Y%m%d_%H%M%S)"
        echo -e "${GREEN}✓ バックアップを作成しました${NC}"
    fi
fi

cat > .claude/settings.local.json << 'EOF'
{
  "permissions": {
    "allow": [
      "Bash",
      "Read",
      "Write",
      "Edit",
      "Glob",
      "Grep"
    ],
    "deny": [],
    "ask": []
  }
}
EOF
echo -e "${GREEN}✓ .claude/settings.local.json${NC}"
echo

# サンプルタスクファイルのダウンロード
echo -e "${BLUE}━━━ サンプルファイルのダウンロード ━━━${NC}"
download_file "tonight-with-tasks.txt.example"
echo

# .gitignoreに追加
if [ -f ".gitignore" ]; then
    if ! grep -q "tonight.txt" .gitignore; then
        echo -e "${BLUE}━━━ .gitignore の更新 ━━━${NC}"
        echo "" >> .gitignore
        echo "# Claude Code auto-dev system" >> .gitignore
        echo "tonight.txt" >> .gitignore
        echo "test-*.txt" >> .gitignore
        echo -e "${GREEN}✓ .gitignore に追加しました${NC}"
        echo
    fi
fi

# インストール完了
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ インストールが完了しました！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# 使い方の表示
echo -e "${BLUE}📚 使い方:${NC}"
echo
echo "1. タスクファイルを作成:"
echo -e "   ${YELLOW}cp tonight-with-tasks.txt.example tonight.txt${NC}"
echo -e "   ${YELLOW}vim tonight.txt  # タスク内容を編集${NC}"
echo
echo "2. 実行方法を選択:"
echo
echo "   【基本実行】"
echo -e "   ${YELLOW}./auto-dev.sh tonight.txt${NC}"
echo
echo "   【完了後に1つのコミット】"
echo -e "   ${YELLOW}./auto-dev-with-commits.sh tonight.txt${NC}"
echo
echo "   【タスクごとに個別コミット】"
echo -e "   ${YELLOW}./auto-dev-incremental.sh tonight.txt${NC}"
echo
echo "   【バックグラウンド実行（寝てる間に開発）】"
echo -e "   ${YELLOW}./run-overnight.sh tonight.txt${NC}"
echo -e "   ${YELLOW}tail -f nohup.out  # ログ確認${NC}"
echo
echo -e "${BLUE}💡 ヒント:${NC}"
echo "- タスクファイルはマークダウン形式で記述"
echo "- インクリメンタル実行では '# タスク' または '# Task' で分割"
echo "- .claude/settings.local.json で自動承認が有効化されています"
echo
echo -e "${GREEN}Happy Coding! 🚀${NC}"
