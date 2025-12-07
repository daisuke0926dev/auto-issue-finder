#!/bin/bash
#
# Claude Code 連続実行スクリプト
# 使い方: ./auto-dev.sh [tonight.txt]
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
echo -e "${BLUE}========================================${NC}"
echo ""

# 入力ファイル
INPUT_FILE="${1:-tonight.txt}"

if [ ! -f "$INPUT_FILE" ]; then
    echo -e "${YELLOW}入力ファイル ${INPUT_FILE} が見つかりません${NC}"
    echo ""
    echo "サンプルファイルを作成しますか？ (y/n)"
    read -r answer
    if [ "$answer" = "y" ]; then
        cat > "$INPUT_FILE" <<'EOF'
# 今夜のタスク

GitHubのIssueを取得するGoのCLIツールを作る。

まずはプロジェクトの初期化から始める:
- go mod init で auto-issue-finder を初期化
- 基本的なディレクトリ構造を作成（cmd/, internal/, pkg/）

次にGitHub APIクライアントを実装:
- GitHub REST APIを使ってIssueを取得する機能
- 認証はPersonal Access Tokenを使う
- 取得したIssueをJSON形式で出力

その後、基本的なテストも書く:
- ユニットテストをいくつか追加
- go test で実行できるように

最後にREADMEも更新して、使い方を説明する。
EOF
        echo -e "${GREEN}✓ サンプルファイルを作成しました: ${INPUT_FILE}${NC}"
        echo ""
    else
        exit 1
    fi
fi

# タスク内容を表示
echo -e "${BLUE}タスク内容:${NC}"
cat "$INPUT_FILE"
echo ""

# 対話モードかどうかチェック
if [ -t 0 ]; then
    # 標準入力が端末（対話モード）
    echo -e "${YELLOW}Claude Code で自律開発を開始しますか？ (y/n)${NC}"
    echo -e "${YELLOW}※ すべてのツールが自動承認されます${NC}"
    read -r answer
    if [ "$answer" != "y" ]; then
        echo "キャンセルしました"
        exit 0
    fi
    echo ""
else
    # バックグラウンド実行（自動的に開始）
    echo -e "${GREEN}自動開始モード（バックグラウンド実行）${NC}"
    echo ""
fi

# ログディレクトリを作成
mkdir -p logs
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="logs/auto-dev-$TIMESTAMP.log"

echo -e "${BLUE}開始時刻: $(date)${NC}" | tee "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# タスクをClaudeに送る
TASK_CONTENT=$(cat "$INPUT_FILE")

PROMPT="あなたは自律的にソフトウェア開発を行うエンジニアです。

# 指示
以下のタスクを完全に実装してください。必要なファイルを作成・編集し、コマンドを実行して、動作確認まで行ってください。

# タスク
$TASK_CONTENT

# 重要なルール
1. すべてのファイルを実際に作成・編集すること
2. 必要なコマンドを実際に実行すること
3. テストを書いて実行すること
4. 動作確認を行うこと
5. 完了したら詳細なレポートを作成すること

プロジェクトディレクトリ: $PROJECT_DIR

実装を開始してください。"

echo -e "${YELLOW}Claude Code に送信中...${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

# Claude Codeを実行（全て自動承認）
echo "$PROMPT" | claude -p 2>&1 | tee -a "$LOG_FILE"

echo "" | tee -a "$LOG_FILE"
echo -e "${BLUE}完了時刻: $(date)${NC}" | tee -a "$LOG_FILE"
echo "" | tee -a "$LOG_FILE"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  完了！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "ログファイル: ${YELLOW}$LOG_FILE${NC}"
echo ""
