# テストガイド

このガイドでは、Auto Issue Finderのテスト実行方法とテスト戦略を説明します。

## 目次

- [テスト実行](#テスト実行)
- [カバレッジ](#カバレッジ)
- [テストの書き方](#テストの書き方)
- [CI/CD](#cicd)

---

## テスト実行

### 基本的なテスト実行

```bash
# 全テスト実行
go test ./...

# 詳細モード
go test ./... -v

# 特定パッケージのみ
go test ./internal/analyzer
go test ./internal/reporter
go test ./internal/github

# 並列実行（デフォルト）
go test ./... -parallel 4
```

### テスト出力の見方

```bash
$ go test ./...
ok      github.com/isiidaisuke0926/auto-issue-finder/internal/analyzer    0.123s
ok      github.com/isiidaisuke0926/auto-issue-finder/internal/reporter    0.089s
ok      github.com/isiidaisuke0926/auto-issue-finder/internal/github      0.067s
```

- `ok` - テスト成功
- `FAIL` - テスト失敗
- 数値 - 実行時間

### 失敗時のデバッグ

```bash
# 詳細な出力
go test ./... -v

# 特定のテスト関数のみ実行
go test ./internal/analyzer -run TestAnalyze

# テスト失敗時に即座に停止
go test ./... -failfast

# タイムアウト設定（デフォルト10分）
go test ./... -timeout 30s
```

---

## カバレッジ

### カバレッジ測定

```bash
# 全パッケージのカバレッジ
go test ./... -cover

# カバレッジファイル生成
go test ./... -coverprofile=coverage.out

# HTMLレポート生成
go tool cover -html=coverage.out -o coverage.html

# ブラウザで表示
go tool cover -html=coverage.out
```

### カバレッジ結果の見方

```bash
$ go test ./... -cover
ok      .../internal/analyzer    0.123s  coverage: 96.9% of statements
ok      .../internal/reporter    0.089s  coverage: 96.5% of statements
ok      .../internal/github      0.067s  coverage: 75.0% of statements
```

### パッケージ別の詳細カバレッジ

```bash
# analyzer パッケージ
go test ./internal/analyzer -coverprofile=coverage-analyzer.out
go tool cover -func=coverage-analyzer.out

# 出力例:
# github.com/isiidaisuke0926/auto-issue-finder/internal/analyzer/analyzer.go:15:  Analyze         100.0%
# github.com/isiidaisuke0926/auto-issue-finder/internal/analyzer/analyzer.go:45:  analyzeLabels   95.0%
# total:                                                                           (statements)    96.9%
```

### 目標カバレッジ

| パッケージ | 現在 | 目標 |
|-----------|------|------|
| `internal/analyzer` | 96.9% | 95%+ |
| `internal/reporter` | 96.5% | 95%+ |
| `internal/github` | 75.0% | 80%+ |
| **全体** | **>70%** | **80%+** |

---

## テストの書き方

### テーブル駆動テスト

Goの推奨パターン:

```go
func TestAnalyze(t *testing.T) {
    tests := []struct {
        name    string
        input   []Issue
        want    Analysis
        wantErr bool
    }{
        {
            name: "正常ケース - 複数Issue",
            input: []Issue{
                {ID: 1, State: "open", Labels: []string{"bug"}},
                {ID: 2, State: "closed", Labels: []string{"enhancement"}},
            },
            want: Analysis{
                TotalIssues: 2,
                OpenIssues:  1,
                ClosedIssues: 1,
            },
            wantErr: false,
        },
        {
            name:    "空のIssue",
            input:   []Issue{},
            want:    Analysis{},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Analyze(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Analyze() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Analyze() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### testifyを使ったテスト

より読みやすいアサーション:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestReporter(t *testing.T) {
    reporter := NewReporter()

    // 必須条件（失敗時に即座に終了）
    require.NotNil(t, reporter)

    // アサーション（失敗してもテスト続行）
    assert.Equal(t, "markdown", reporter.Format)
    assert.Greater(t, len(reporter.Buffer), 0)
}
```

### モックの使用

GitHub APIのモック:

```go
type MockGitHubClient struct {
    mock.Mock
}

func (m *MockGitHubClient) GetIssues(repo string) ([]Issue, error) {
    args := m.Called(repo)
    return args.Get(0).([]Issue), args.Error(1)
}

func TestWithMock(t *testing.T) {
    mockClient := new(MockGitHubClient)
    mockClient.On("GetIssues", "owner/repo").Return(
        []Issue{{ID: 1}}, nil,
    )

    // テスト実行
    issues, err := mockClient.GetIssues("owner/repo")

    assert.NoError(t, err)
    assert.Len(t, issues, 1)
    mockClient.AssertExpectations(t)
}
```

### サブテスト

関連するテストをグループ化:

```go
func TestUserOperations(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        user, err := CreateUser("test@example.com")
        assert.NoError(t, err)
        assert.NotEmpty(t, user.ID)
    })

    t.Run("Get", func(t *testing.T) {
        user, err := GetUser("123")
        assert.NoError(t, err)
        assert.Equal(t, "test@example.com", user.Email)
    })

    t.Run("Delete", func(t *testing.T) {
        err := DeleteUser("123")
        assert.NoError(t, err)
    })
}
```

### ヘルパー関数

テストコードの重複を減らす:

```go
func createTestIssues(t *testing.T, count int) []Issue {
    t.Helper()

    issues := make([]Issue, count)
    for i := 0; i < count; i++ {
        issues[i] = Issue{
            ID:    i + 1,
            State: "open",
            Labels: []string{"test"},
        }
    }
    return issues
}

func TestWithHelper(t *testing.T) {
    issues := createTestIssues(t, 5)
    assert.Len(t, issues, 5)
}
```

---

## ベンチマークテスト

### ベンチマークの書き方

```go
func BenchmarkAnalyze(b *testing.B) {
    issues := createTestIssues(nil, 100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Analyze(issues)
    }
}
```

### ベンチマーク実行

```bash
# 全ベンチマーク実行
go test ./... -bench=.

# 特定のベンチマーク
go test ./internal/analyzer -bench=BenchmarkAnalyze

# メモリ統計付き
go test ./... -bench=. -benchmem

# 実行回数指定
go test ./... -bench=. -benchtime=10s
```

### ベンチマーク結果の見方

```bash
$ go test ./internal/analyzer -bench=. -benchmem
BenchmarkAnalyze-8    1000000    1234 ns/op    512 B/op    10 allocs/op
```

- `BenchmarkAnalyze-8` - ベンチマーク名-CPU数
- `1000000` - 実行回数
- `1234 ns/op` - 1回あたりの実行時間
- `512 B/op` - 1回あたりのメモリ使用量
- `10 allocs/op` - 1回あたりのメモリアロケーション回数

---

## CI/CD

### GitHub Actions設定例

`.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Download dependencies
      run: go mod download

    - name: Run tests
      run: go test ./... -v -cover

    - name: Generate coverage
      run: go test ./... -coverprofile=coverage.out

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### カバレッジバッジの追加

README.mdに追加:

```markdown
[![Coverage](https://codecov.io/gh/isiidaisuke0926/auto-issue-finder/branch/main/graph/badge.svg)](https://codecov.io/gh/isiidaisuke0926/auto-issue-finder)
```

### Pre-commit Hook

`.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "Running tests..."
go test ./... -cover

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Running linter..."
golangci-lint run

if [ $? -ne 0 ]; then
    echo "Linter failed. Commit aborted."
    exit 1
fi

echo "All checks passed!"
```

実行権限付与:

```bash
chmod +x .git/hooks/pre-commit
```

---

## トラブルシューティング

### テストが失敗する

```bash
# 詳細な出力で原因確認
go test ./... -v

# 特定のテストのみ実行
go test ./internal/analyzer -run TestAnalyze -v

# キャッシュをクリア
go clean -testcache
go test ./...
```

### カバレッジが低い

```bash
# カバーされていない箇所を確認
go test ./internal/analyzer -coverprofile=coverage.out
go tool cover -html=coverage.out

# ブラウザで赤い部分（カバーされていない）を確認
```

### ベンチマークが不安定

```bash
# 実行時間を長くして平均化
go test ./... -bench=. -benchtime=30s

# CPU数を固定
GOMAXPROCS=1 go test ./... -bench=.

# 複数回実行して比較
go test ./... -bench=. -count=5
```

---

## ベストプラクティス

### 1. テストの命名

```go
// ✅ Good
func TestAnalyze_WithEmptyIssues_ReturnsEmptyAnalysis(t *testing.T)
func TestGetUser_InvalidID_ReturnsError(t *testing.T)

// ❌ Bad
func TestFunc1(t *testing.T)
func Test2(t *testing.T)
```

### 2. テストの独立性

```go
// ✅ Good - 各テストが独立
func TestCreateUser(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    user, err := CreateUser(db, "test@example.com")
    assert.NoError(t, err)
}

// ❌ Bad - グローバル状態に依存
var globalDB *DB

func TestCreateUser(t *testing.T) {
    user, err := CreateUser(globalDB, "test@example.com")
    assert.NoError(t, err)
}
```

### 3. エラーケースのテスト

```go
func TestAnalyze(t *testing.T) {
    tests := []struct {
        name    string
        input   []Issue
        wantErr error
    }{
        {
            name:    "nil input",
            input:   nil,
            wantErr: ErrNilInput,
        },
        {
            name:    "invalid issue",
            input:   []Issue{{ID: -1}},
            wantErr: ErrInvalidIssue,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := Analyze(tt.input)
            assert.ErrorIs(t, err, tt.wantErr)
        })
    }
}
```

---

## まとめ

### テストチェックリスト

- [ ] 全テストが通る: `go test ./...`
- [ ] カバレッジ80%以上: `go test ./... -cover`
- [ ] リンターエラーなし: `golangci-lint run`
- [ ] ベンチマークで性能確認: `go test ./... -bench=.`
- [ ] エラーケースをテスト
- [ ] エッジケース（境界値）をテスト
- [ ] テストが独立している

### 次のステップ

- [使用方法](USAGE.md) - コマンドリファレンス
- [インストールガイド](INSTALL.md) - セットアップ方法
- [自律開発システム詳細](AUTO_DEV.md) - 高度な使い方
