#!/bin/bash
#
# Claude Code 自律開発システム（コミット機能付き）
# 使い方: ./auto-dev-with-commits.sh [tonight.txt]
#

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

# 色の定義
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Claude Code 自律開発システム${NC}"
echo -e "${BLUE}  （自動コミット機能付き）${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 入力ファイル
INPUT_FILE="${1:-tonight.txt}"

if [ ! -f "$INPUT_FILE" ]; then
    echo -e "${RED}エラー: $INPUT_FILE が見つかりません${NC}"
    exit 1
fi

# Gitリポジトリチェック
if [ ! -d .git ]; then
    echo -e "${YELLOW}警告: Gitリポジトリが初期化されていません${NC}"
    echo -e "${YELLOW}git init を実行しますか？ (y/n)${NC}"
    read -r answer
    if [ "$answer" = "y" ]; then
        git init
        echo -e "${GREEN}✓ Gitリポジトリを初期化しました${NC}"
    else
        echo -e "${RED}エラー: Gitリポジトリが必要です${NC}"
        exit 1
    fi
fi

# 初期コミット（変更がある場合）
if ! git rev-parse HEAD >/dev/null 2>&1; then
    echo -e "${YELLOW}初期コミットを作成します...${NC}"
    git add .
    git commit -m "chore: 初期コミット - 自律開発開始前の状態" || true
fi

# 実行前のコミットハッシュを保存
BEFORE_COMMIT=$(git rev-parse HEAD)
echo -e "${BLUE}実行前のコミット: $BEFORE_COMMIT${NC}"
echo ""

# タスク内容を表示
echo -e "${BLUE}タスク内容:${NC}"
cat "$INPUT_FILE"
echo ""

# 実行確認
if [ -t 0 ]; then
    echo -e "${YELLOW}Claude Code で自律開発を開始しますか？ (y/n)${NC}"
    echo -e "${YELLOW}※ タスク完了後に自動でgit commitします${NC}"
    read -r answer
    if [ "$answer" != "y" ]; then
        echo "キャンセルしました"
        exit 0
    fi
    echo ""
else
    echo -e "${GREEN}自動開始モード（バックグラウンド実行）${NC}"
    echo ""
fi

# ログディレクトリを作成
mkdir -p logs
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="logs/auto-dev-$TIMESTAMP.log"

echo -e "${BLUE}開始時刻: $(date)${NC}" | tee "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# タスクをClaudeに送る（コミット指示を追加）
TASK_CONTENT=$(cat "$INPUT_FILE")

PROMPT="あなたは自律的にソフトウェア開発を行うエンジニアです。

# 指示
以下のタスクを完全に実装してください。

# タスク
$TASK_CONTENT

# 重要なルール
1. すべてのファイルを実際に作成・編集すること
2. 必要なコマンドを実際に実行すること
3. テストを書いて実行すること
4. 動作確認を行うこと

# コミットルール
**実装が一区切りついたら、必ず git commit を実行してください。**

コミットメッセージの形式:
- feat: 新機能追加
- fix: バグ修正
- docs: ドキュメント更新
- test: テスト追加
- refactor: リファクタリング

例:
- git add .
- git commit -m \"feat: GitHub APIクライアント実装\"

プロジェクトディレクトリ: $PROJECT_DIR

実装を開始してください。"

echo -e "${YELLOW}Claude Code に送信中...${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Claude Codeを実行
echo "$PROMPT" | claude -p 2>&1 | tee -a "$LOG_FILE"

echo "" | tee -a "$LOG_FILE"
echo -e "${BLUE}完了時刻: $(date)${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# 実行後のコミット確認
AFTER_COMMIT=$(git rev-parse HEAD)

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  完了！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# コミット履歴を表示
if [ "$BEFORE_COMMIT" != "$AFTER_COMMIT" ]; then
    echo -e "${GREEN}✓ 新しいコミットが作成されました${NC}"
    echo ""
    echo -e "${BLUE}コミット履歴:${NC}"
    git log --oneline "$BEFORE_COMMIT".."$AFTER_COMMIT" --decorate --graph
    echo ""
    echo -e "${BLUE}変更されたファイル:${NC}"
    git diff --name-status "$BEFORE_COMMIT".."$AFTER_COMMIT"
else
    echo -e "${YELLOW}⚠ 新しいコミットは作成されませんでした${NC}"
    echo -e "${YELLOW}変更があった場合は手動でコミットしてください:${NC}"
    echo ""
    git status --short
fi

echo ""
echo -e "ログファイル: ${YELLOW}$LOG_FILE${NC}"
echo ""
