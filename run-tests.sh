#!/bin/bash

# Test runner script
# 全テストを実行してカバレッジレポートを生成

set -e

# 色定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Test Runner${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# オプション解析
VERBOSE=false
COVERAGE=false
HTML=false
INTEGRATION=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        --html)
            HTML=true
            COVERAGE=true
            shift
            ;;
        -i|--integration)
            INTEGRATION=true
            shift
            ;;
        -h|--help)
            echo "使用方法: $0 [オプション]"
            echo
            echo "オプション:"
            echo "  -v, --verbose      詳細な出力"
            echo "  -c, --coverage     カバレッジ測定"
            echo "  --html             HTMLカバレッジレポート生成"
            echo "  -i, --integration  統合テストも実行"
            echo "  -h, --help         このヘルプを表示"
            echo
            echo "例:"
            echo "  $0                   # 基本的なテスト実行"
            echo "  $0 -v -c             # 詳細出力とカバレッジ"
            echo "  $0 --html            # HTMLカバレッジレポート"
            echo "  $0 -i                # 統合テストも含む"
            exit 0
            ;;
        *)
            echo -e "${RED}不明なオプション: $1${NC}"
            exit 1
            ;;
    esac
done

# プロジェクトルート
PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_ROOT"

# 1. ユニットテスト
echo -e "${BLUE}━━━ ユニットテスト ━━━${NC}"

TEST_CMD="go test ./internal/..."

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -coverprofile=coverage.out"
fi

echo -e "${YELLOW}実行中: $TEST_CMD${NC}"
echo

if eval "$TEST_CMD"; then
    echo -e "${GREEN}✓ ユニットテスト成功${NC}"
else
    echo -e "${RED}✗ ユニットテスト失敗${NC}"
    exit 1
fi
echo

# 2. カバレッジ詳細
if [ "$COVERAGE" = true ]; then
    echo -e "${BLUE}━━━ カバレッジ詳細 ━━━${NC}"

    # パッケージごとのカバレッジ
    echo -e "${YELLOW}パッケージ別カバレッジ:${NC}"
    go tool cover -func=coverage.out | grep -E "total:|analyzer|reporter|github"
    echo

    # 総カバレッジ
    total_coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    echo -e "${GREEN}総カバレッジ: $total_coverage${NC}"
    echo
fi

# 3. HTMLレポート
if [ "$HTML" = true ]; then
    echo -e "${BLUE}━━━ HTMLレポート生成 ━━━${NC}"

    go tool cover -html=coverage.out -o coverage.html
    echo -e "${GREEN}✓ coverage.html を生成しました${NC}"

    # ブラウザで開く（オプション）
    if command -v open &> /dev/null; then
        echo -e "${YELLOW}ブラウザでレポートを開きますか？ (y/n)${NC}"
        read -r answer
        if [[ "$answer" =~ ^[Yy]$ ]]; then
            open coverage.html
        fi
    fi
    echo
fi

# 4. 統合テスト
if [ "$INTEGRATION" = true ]; then
    echo -e "${BLUE}━━━ 統合テスト ━━━${NC}"

    INTEGRATION_CMD="go test ./test/..."

    if [ "$VERBOSE" = true ]; then
        INTEGRATION_CMD="$INTEGRATION_CMD -v"
    fi

    echo -e "${YELLOW}実行中: $INTEGRATION_CMD${NC}"
    echo

    if eval "$INTEGRATION_CMD"; then
        echo -e "${GREEN}✓ 統合テスト成功${NC}"
    else
        echo -e "${YELLOW}⚠️  統合テストに一部失敗がありました${NC}"
    fi
    echo
fi

# 5. サマリー
echo -e "${BLUE}━━━ サマリー ━━━${NC}"

# テストファイル数
test_files=$(find . -name "*_test.go" | wc -l | tr -d ' ')
echo "テストファイル数: $test_files"

# パッケージ数
packages=$(go list ./... 2>/dev/null | grep -v "/test" | grep -v "/cmd/executor" | grep -v "/cmd/parse-tasks" | grep -v "/cmd/reporter" | wc -l | tr -d ' ')
echo "テスト対象パッケージ: $packages"

if [ "$COVERAGE" = true ]; then
    echo "総カバレッジ: $total_coverage"
fi

echo

# 完了
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ テスト完了${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo

# 次のステップ
if [ "$COVERAGE" = false ]; then
    echo "💡 カバレッジを確認するには:"
    echo "   ${YELLOW}$0 --coverage${NC}"
    echo
fi

if [ "$HTML" = false ] && [ "$COVERAGE" = true ]; then
    echo "💡 HTMLレポートを生成するには:"
    echo "   ${YELLOW}$0 --html${NC}"
    echo
fi

if [ "$INTEGRATION" = false ]; then
    echo "💡 統合テストも実行するには:"
    echo "   ${YELLOW}$0 --integration${NC}"
    echo
fi
