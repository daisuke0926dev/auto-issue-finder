#!/bin/bash
#
# 一晩実行用スクリプト
# 使い方: ./run-overnight.sh [tonight.txt]
#

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

INPUT_FILE="${1:-tonight.txt}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_FILE="logs/overnight-${TIMESTAMP}.log"

# ディレクトリ作成
mkdir -p logs

echo "========================================"
echo "  一晩実行モード - Claude Code自律開発"
echo "  開始時刻: $(date)"
echo "========================================"
echo ""

# nohupで実行（ターミナルを閉じても継続）
nohup bash -c "
    cd '$PROJECT_DIR'
    echo '開始時刻: \$(date)' >> '$LOG_FILE'
    echo '' >> '$LOG_FILE'
    ./auto-dev.sh '$INPUT_FILE' >> '$LOG_FILE' 2>&1
    echo '' >> '$LOG_FILE'
    echo '完了時刻: \$(date)' >> '$LOG_FILE'
    rm logs/overnight.pid 2>/dev/null || true
" > /dev/null 2>&1 &

PID=$!

echo "✓ バックグラウンドで実行中 (PID: $PID)"
echo ""
echo "📝 ログファイル: $LOG_FILE"
echo ""
echo "📊 進捗確認:"
echo "  tail -f $LOG_FILE"
echo ""
echo "🔍 プロセス確認:"
echo "  ps -p $PID"
echo ""
echo "⏹️  停止方法:"
echo "  kill $PID"
echo ""
echo "☀️  翌朝の確認:"
echo "  cat $LOG_FILE"
echo ""

# PIDをファイルに保存
echo "$PID" > logs/overnight.pid
echo "PIDを保存しました: logs/overnight.pid"
echo ""
echo "おやすみなさい 💤"
