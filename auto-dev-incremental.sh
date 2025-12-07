#!/bin/bash
#
# Claude Code 自律開発システム（タスク分割・個別コミット）
# 使い方: ./auto-dev-incremental.sh [tonight.txt]
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
echo -e "${BLUE}  （タスク分割・個別コミット版）${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 入力ファイル
INPUT_FILE="${1:-tonight.txt}"

if [ ! -f "$INPUT_FILE" ]; then
    echo -e "${RED}エラー: $INPUT_FILE が見つかりません${NC}"
    exit 1
fi

# Gitチェック
if [ ! -d .git ]; then
    echo -e "${YELLOW}git init を実行...${NC}"
    git init
fi

# タスクファイルを解析（#で始まる行をタスクとして認識）
mapfile -t TASKS < <(grep -E "^#+ (タスク|Task)" "$INPUT_FILE" || echo "")

if [ ${#TASKS[@]} -eq 0 ]; then
    echo -e "${YELLOW}警告: タスク分割マーカーが見つかりません${NC}"
    echo -e "${YELLOW}通常モードで実行します${NC}"
    exec ./auto-dev-with-commits.sh "$INPUT_FILE"
fi

echo -e "${BLUE}検出されたタスク: ${#TASKS[@]}件${NC}"
for i in "${!TASKS[@]}"; do
    echo -e "${YELLOW}  $((i+1)). ${TASKS[$i]}${NC}"
done
echo ""

# 確認
if [ -t 0 ]; then
    echo -e "${YELLOW}各タスクを順次実行して個別にコミットしますか？ (y/n)${NC}"
    read -r answer
    if [ "$answer" != "y" ]; then
        echo "キャンセルしました"
        exit 0
    fi
    echo ""
fi

# ログディレクトリ
mkdir -p logs
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
MAIN_LOG="logs/incremental-$TIMESTAMP.log"

echo -e "${BLUE}開始時刻: $(date)${NC}" | tee "$MAIN_LOG"
echo "" | tee -a "$MAIN_LOG"

# タスクファイルの内容を分割
TEMP_DIR="logs/tasks-$TIMESTAMP"
mkdir -p "$TEMP_DIR"

# タスクごとに分割（簡易版）
csplit -s -f "$TEMP_DIR/task-" "$INPUT_FILE" "/^#+ (タスク|Task)/" "{*}" 2>/dev/null || {
    echo -e "${YELLOW}タスク分割に失敗しました。通常モードで実行します${NC}"
    exec ./auto-dev-with-commits.sh "$INPUT_FILE"
}

# 各タスクを実行
TASK_NUM=0
for task_file in "$TEMP_DIR"/task-*; do
    if [ ! -s "$task_file" ]; then
        continue
    fi

    TASK_NUM=$((TASK_NUM + 1))

    echo -e "${BLUE}========================================${NC}" | tee -a "$MAIN_LOG"
    echo -e "${BLUE}  タスク $TASK_NUM 実行中${NC}" | tee -a "$MAIN_LOG"
    echo -e "${BLUE}========================================${NC}" | tee -a "$MAIN_LOG"
    echo "" | tee -a "$MAIN_LOG"

    # タスク内容を表示
    head -10 "$task_file" | tee -a "$MAIN_LOG"
    echo "" | tee -a "$MAIN_LOG"

    # 実行前のコミット
    BEFORE_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "none")

    # タスクを実行
    TASK_CONTENT=$(cat "$task_file")

    PROMPT="以下のタスクを実装してください:

$TASK_CONTENT

実装が完了したら、必ず git add と git commit を実行してください。
コミットメッセージは実装内容に応じて適切に設定してください。"

    TASK_LOG="logs/task-$TASK_NUM-$TIMESTAMP.log"
    echo "$PROMPT" | claude -p 2>&1 | tee "$TASK_LOG" | tee -a "$MAIN_LOG"

    # コミット確認
    AFTER_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "none")

    echo "" | tee -a "$MAIN_LOG"
    if [ "$BEFORE_COMMIT" != "$AFTER_COMMIT" ]; then
        echo -e "${GREEN}✓ タスク $TASK_NUM 完了（コミット作成済み）${NC}" | tee -a "$MAIN_LOG"
        git log -1 --oneline | tee -a "$MAIN_LOG"
    else
        echo -e "${YELLOW}⚠ タスク $TASK_NUM 完了（コミット未作成）${NC}" | tee -a "$MAIN_LOG"
    fi
    echo "" | tee -a "$MAIN_LOG"

    # 次のタスクまで少し待機
    sleep 2
done

echo -e "${BLUE}完了時刻: $(date)${NC}" | tee -a "$MAIN_LOG"
echo "" | tee -a "$MAIN_LOG"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  全タスク完了！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 全体のコミット履歴を表示
echo -e "${BLUE}コミット履歴:${NC}"
git log --oneline --graph --decorate -10

echo ""
echo -e "メインログ: ${YELLOW}$MAIN_LOG${NC}"
echo ""

# クリーンアップ
rm -rf "$TEMP_DIR"
