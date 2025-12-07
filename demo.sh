#!/bin/bash

# Demo script for Auto Issue Finder
# このスクリプトでツールの動作を確認できます

set -e

# 色定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Auto Issue Finder デモ${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# プロジェクトルートの確認
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

echo -e "${GREEN}📁 プロジェクトディレクトリ:${NC} $PROJECT_ROOT"
echo

# 1. ビルド
echo -e "${BLUE}━━━ 1. ビルド ━━━${NC}"
echo -e "${YELLOW}CLIツールをビルドしています...${NC}"

if go build -o auto-issue-finder cmd/analyze/main.go; then
    echo -e "${GREEN}✓ ビルド成功${NC}"
else
    echo -e "${RED}✗ ビルド失敗${NC}"
    exit 1
fi
echo

# 2. ヘルプ表示
echo -e "${BLUE}━━━ 2. ヘルプ表示 ━━━${NC}"
./auto-issue-finder --help
echo

# 3. バージョン確認
echo -e "${BLUE}━━━ 3. バージョン ━━━${NC}"
./auto-issue-finder version 2>&1 || echo "（バージョンコマンドは未実装の可能性があります）"
echo

# 4. テスト実行
echo -e "${BLUE}━━━ 4. テスト実行 ━━━${NC}"
echo -e "${YELLOW}ユニットテストを実行しています...${NC}"

if go test ./internal/... -cover; then
    echo -e "${GREEN}✓ テスト成功${NC}"
else
    echo -e "${RED}✗ テスト失敗${NC}"
    exit 1
fi
echo

# 5. 統合テスト実行
echo -e "${BLUE}━━━ 5. 統合テスト ━━━${NC}"
echo -e "${YELLOW}統合テストを実行しています...${NC}"

if go test ./test/... -v; then
    echo -e "${GREEN}✓ 統合テスト成功${NC}"
else
    echo -e "${RED}✗ 統合テスト失敗${NC}"
fi
echo

# 6. 実際のリポジトリで分析（オプション）
echo -e "${BLUE}━━━ 6. 実際のリポジトリ分析（オプション） ━━━${NC}"

if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${YELLOW}⚠️  GITHUB_TOKEN が設定されていません${NC}"
    echo "環境変数 GITHUB_TOKEN を設定すると、実際のリポジトリを分析できます。"
    echo
    echo "設定方法:"
    echo "  export GITHUB_TOKEN=your_token_here"
    echo
else
    echo -e "${GREEN}✓ GITHUB_TOKEN が設定されています${NC}"
    echo -e "${YELLOW}小規模リポジトリで動作確認しています...${NC}"
    echo

    # 小さなリポジトリで簡易分析
    if ./auto-issue-finder analyze golang/go --format=console --limit=10; then
        echo -e "${GREEN}✓ 分析成功${NC}"
    else
        echo -e "${YELLOW}⚠️  分析に失敗しました（レート制限の可能性）${NC}"
    fi
    echo
fi

# 7. Auto-devスクリプトの確認
echo -e "${BLUE}━━━ 7. Auto-devスクリプト確認 ━━━${NC}"

scripts=(
    "auto-dev.sh"
    "auto-dev-incremental.sh"
    "auto-dev-with-commits.sh"
    "run-overnight.sh"
    "install-auto-dev.sh"
)

all_exist=true
for script in "${scripts[@]}"; do
    if [ -f "$script" ] && [ -x "$script" ]; then
        echo -e "${GREEN}✓${NC} $script"
    else
        echo -e "${RED}✗${NC} $script (見つからないか実行権限がありません)"
        all_exist=false
    fi
done

if [ "$all_exist" = true ]; then
    echo -e "${GREEN}✓ 全てのスクリプトが存在します${NC}"
fi
echo

# 8. ドキュメント確認
echo -e "${BLUE}━━━ 8. ドキュメント確認 ━━━${NC}"

docs=(
    "README.md"
    "CONTRIBUTING.md"
    "docs/INSTALL.md"
    "docs/USAGE.md"
    "docs/AUTO_DEV.md"
    "docs/TESTING.md"
)

all_docs_exist=true
for doc in "${docs[@]}"; do
    if [ -f "$doc" ]; then
        lines=$(wc -l < "$doc")
        echo -e "${GREEN}✓${NC} $doc ($lines 行)"
    else
        echo -e "${RED}✗${NC} $doc"
        all_docs_exist=false
    fi
done

if [ "$all_docs_exist" = true ]; then
    echo -e "${GREEN}✓ 全てのドキュメントが存在します${NC}"
fi
echo

# 完了
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ デモ完了！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

echo "次のステップ:"
echo "1. GitHub Issue Analyzer を試す:"
echo "   ${YELLOW}./auto-issue-finder analyze microsoft/vscode --format=console --limit=20${NC}"
echo
echo "2. Auto-dev システムを試す:"
echo "   ${YELLOW}cp tonight-with-tasks.txt.example tonight.txt${NC}"
echo "   ${YELLOW}vim tonight.txt  # タスクを編集${NC}"
echo "   ${YELLOW}./auto-dev.sh tonight.txt${NC}"
echo
echo "3. ドキュメントを読む:"
echo "   ${YELLOW}cat docs/USAGE.md${NC}"
echo

# クリーンアップ
if [ -f "auto-issue-finder" ]; then
    echo -e "${BLUE}ビルドしたバイナリをクリーンアップしますか？ (y/n)${NC}"
    read -r answer
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        rm auto-issue-finder
        echo -e "${GREEN}✓ クリーンアップ完了${NC}"
    fi
fi
