# タスク検証機能の改善設計

## 現状の問題点

### 1. 確認コマンドの終了コードチェックの問題
- **現状**: `runCommand` 関数（cmd/sync.go:498-534）では、`cmd.CombinedOutput()` の返り値でエラーチェックを行っている
- **問題点**:
  - 終了コード0以外を明示的にチェックしていない
  - エラーメッセージがあっても警告だけの場合との区別が不明確
  - sleepship再帰呼び出し時の深度超過を「エラーではない」として扱うが、他のコマンドでは同じ挙動ができない

### 2. タスク失敗時の処理の問題
- **現状**: リトライ機能（cmd/sync.go:313-371）は実装されているが、検証失敗時にClaudeに修正を依頼している
- **問題点**:
  - Claudeが検証コマンドの出力を直接見られない場合がある
  - 検証コマンドの失敗原因が明確でない場合、修正が困難
  - リトライ回数の制限はあるが、各リトライでどう改善されているか追跡できない

### 3. Claudeによる成功判定の問題
- **現状**: タスク実行プロンプト（cmd/sync.go:457-475）は汎用的で、検証に関する明示的な指示がない
- **問題点**:
  - 「動作確認してください」という指示があるが、具体的な検証基準が不明確
  - Claudeが自己判断で「成功」と判断する余地がある
  - 検証コマンドが指定されていないタスクの成功基準が曖昧

### 4. タスク間依存関係の問題
- **現状**: タスクは順次実行され、各タスクは独立して扱われる
- **問題点**:
  - タスクNがタスクN-1の成果物に依存している場合、明示的なチェックがない
  - `--start-from` オプションで途中から実行する場合、前提条件が満たされているか不明
  - 複数のタスクファイルを組み合わせる場合の依存関係が管理できない

### 5. エラー情報の詳細度の問題
- **現状**: 履歴記録（internal/history/history.go）にはエラーメッセージを保存する
- **問題点**:
  - どの検証コマンドが失敗したかの詳細が記録されない
  - 失敗時の出力（stdout/stderr）が記録されない
  - リトライ履歴の詳細が記録されない

## 改善案

### 案1: 確認コマンドの終了コードチェック強化

#### 実装方法
1. `runCommand` 関数で終了コードを明示的に取得・検証
2. 終了コード0以外を失敗として扱う
3. sleepship再帰呼び出しの深度超過は特別扱い（警告のみ）
4. 検証コマンドの標準出力・標準エラー出力を分離して記録

```go
func runCommand(command string, logFile *os.File) error {
    cmd := exec.Command("bash", "-c", command)
    // ...

    output, err := cmd.CombinedOutput()
    logFile.Write(output)

    if err != nil {
        // 終了コードを取得
        exitCode := getExitCode(err)
        return &CommandError{
            Command:  command,
            ExitCode: exitCode,
            Output:   string(output),
            Err:      err,
        }
    }

    return nil
}
```

#### メリット
- 検証の厳密性が向上
- エラー原因の特定が容易になる
- デバッグ情報が充実する

#### デメリット
- 警告を出すだけのコマンドが失敗扱いになる可能性
- 実装が複雑になる
- 既存のタスクファイルとの互換性に注意が必要

### 案2: Claudeによる明示的な成功判定

#### 実装方法
1. タスク実行プロンプトに検証基準を明示的に含める
2. Claudeに「検証コマンド実行前に必ず成功基準を確認する」よう指示
3. 検証コマンドが存在する場合、Claudeに「検証コマンドが成功するまで修正を続ける」よう指示
4. 検証コマンドが存在しない場合、Claudeに「実装内容を詳細に報告する」よう指示

```go
func executeTask(task Task, logFile *os.File) error {
    verificationInstruction := ""
    if task.Command != "" {
        verificationInstruction = fmt.Sprintf(`
# 検証基準
以下のコマンドが成功する（終了コード0を返す）ことを確認してください:
%s

検証が失敗した場合は、エラーを修正して再度検証してください。
検証が成功するまで実装を繰り返してください。`, task.Command)
    } else {
        verificationInstruction = `
# 検証基準
検証コマンドが指定されていないため、以下を確認してください:
1. 実装が完了していること
2. 構文エラーがないこと
3. 実装内容が要求仕様を満たしていること
実装完了後、詳細な実装内容を報告してください。`
    }

    prompt := fmt.Sprintf(`あなたは自律的にソフトウェア開発を行うエンジニアです。

# タスク
%s

%s

%s

# 指示
1. このタスクを完全に実装してください
2. 必要なファイルを作成・編集してください
3. 実装後、上記の検証基準を満たすことを確認してください
4. エラーがあれば修正してください

プロジェクトディレクトリ: %s

実装を開始してください。`, task.Title, task.Description, verificationInstruction, projectDir)

    return executeClaude(prompt, logFile)
}
```

#### メリット
- Claudeが検証基準を明確に理解できる
- 検証コマンドがないタスクの扱いが明確になる
- プロンプト改善だけで実装できる（コード変更が少ない）

#### デメリット
- Claudeの判断に依存する部分が残る
- プロンプトが長くなる
- Claudeがリトライを繰り返しすぎる可能性

### 案3: タスク間依存関係チェック

#### 実装方法
1. タスクファイルフォーマットに依存関係を追加
2. 各タスクに前提条件（prerequisite）を定義可能にする
3. タスク実行前に前提条件をチェック
4. 前提条件が満たされていない場合はエラー

```markdown
## タスク2: APIエンドポイントのテスト追加

APIエンドポイントのテストを追加する

### 前提条件
- `test -f internal/api/handler.go`
- `grep -q "func HandleAPI" internal/api/handler.go`

### 確認
- `go test ./internal/api/...`
```

```go
type Task struct {
    Title         string
    Description   string
    Command       string
    Prerequisites []string // 前提条件コマンドのリスト
}

func validatePrerequisites(task Task, logFile *os.File) error {
    if len(task.Prerequisites) == 0 {
        return nil
    }

    for _, prereq := range task.Prerequisites {
        if err := runCommand(prereq, logFile); err != nil {
            return fmt.Errorf("prerequisite check failed: %s: %w", prereq, err)
        }
    }

    return nil
}
```

#### メリット
- タスク間の依存関係が明示的になる
- `--start-from` 使用時の安全性が向上
- タスクファイルの可読性が向上

#### デメリット
- タスクファイルフォーマットの変更が必要
- 既存のタスクファイルとの後方互換性が必要
- 前提条件の記述が面倒になる可能性

### 案4: 検証結果の詳細記録

#### 実装方法
1. 履歴エントリに検証詳細を追加
2. 各タスクの検証結果を記録
3. リトライ履歴を記録

```go
type VerificationResult struct {
    TaskNumber   int           `json:"task_number"`
    TaskTitle    string        `json:"task_title"`
    Command      string        `json:"command"`
    Success      bool          `json:"success"`
    ExitCode     int           `json:"exit_code,omitempty"`
    Output       string        `json:"output,omitempty"`
    RetryCount   int           `json:"retry_count"`
    Duration     time.Duration `json:"duration"`
}

type Entry struct {
    // 既存フィールド
    TaskFile     string        `json:"task_file"`
    ExecutedAt   time.Time     `json:"executed_at"`
    Success      bool          `json:"success"`
    // ...

    // 新規フィールド
    VerificationResults []VerificationResult `json:"verification_results,omitempty"`
}
```

#### メリット
- トラブルシューティングが容易になる
- 過去の実行結果を詳細に分析できる
- どの検証コマンドが頻繁に失敗するか特定できる

#### デメリット
- 履歴ファイルのサイズが大きくなる
- 大量の出力がある場合、記録が肥大化する
- パフォーマンスへの影響

## 推奨実装

以下の組み合わせで段階的に実装することを推奨します:

### フェーズ1: 基礎強化（優先度: 高）
1. **案1の一部**: 終了コードの明示的チェック
   - `runCommand` 関数で終了コードを明確に処理
   - エラー情報を構造化して返す

2. **案2**: Claudeプロンプトの改善
   - 検証基準を明示的に指示
   - 検証コマンドの有無で処理を分岐

### フェーズ2: 可視性向上（優先度: 中）
3. **案4の一部**: 検証結果の基本記録
   - 各タスクの検証コマンド実行結果を記録
   - 失敗時の出力を保存（最大文字数制限付き）

### フェーズ3: 高度な機能（優先度: 低）
4. **案3**: タスク間依存関係チェック
   - 新しいタスクファイルフォーマットとして追加
   - 既存フォーマットとの後方互換性を維持

### 実装しない機能
- 案1の「出力の分離記録」: 複雑性に対する効果が小さい
- 案4の「全出力記録」: ファイルサイズの問題

## 実装手順

### ステップ1: CommandErrorの定義
`cmd/sync.go` に構造化エラー型を追加

```go
type CommandError struct {
    Command  string
    ExitCode int
    Output   string
    Err      error
}

func (e *CommandError) Error() string {
    return fmt.Sprintf("command failed (exit code %d): %s\nOutput: %s",
        e.ExitCode, e.Command, e.Output)
}
```

### ステップ2: runCommand関数の改善
終了コードを明示的に取得・返却

```go
func runCommand(command string, logFile *os.File) error {
    // ... (既存のコード)

    output, err := cmd.CombinedOutput()
    logFile.Write(output)

    if err != nil {
        exitCode := 1 // デフォルト
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        }

        return &CommandError{
            Command:  command,
            ExitCode: exitCode,
            Output:   string(output),
            Err:      err,
        }
    }

    return nil
}
```

### ステップ3: executeTask関数のプロンプト改善
検証基準を明示的に含める

```go
func executeTask(task Task, logFile *os.File) error {
    verificationInstruction := ""
    if task.Command != "" {
        verificationInstruction = fmt.Sprintf(`
### 検証基準
以下のコマンドが成功する（終了コード0を返す）ことを確認してください:
- %s

実装後、このコマンドを実行して検証してください。
検証が失敗した場合は、エラーを修正して再度検証してください。`, task.Command)
    } else {
        verificationInstruction = `
### 検証基準
検証コマンドが指定されていません。以下を確認してください:
1. 実装が完了していること
2. 構文エラーがないこと
3. 要求仕様を満たしていること`
    }

    prompt := fmt.Sprintf(`あなたは自律的にソフトウェア開発を行うエンジニアです。

# タスク
%s

%s

%s

# 指示
1. このタスクを完全に実装してください
2. 必要なファイルを作成・編集してください
3. 実装後、必ず動作確認してください
4. エラーがあれば修正してください

プロジェクトディレクトリ: %s

実装を開始してください。`, task.Title, task.Description, verificationInstruction, projectDir)

    return executeClaude(prompt, logFile)
}
```

### ステップ4: VerificationResultの追加
`internal/history/history.go` に検証結果の記録機能を追加

```go
type VerificationResult struct {
    TaskNumber int           `json:"task_number"`
    TaskTitle  string        `json:"task_title"`
    Command    string        `json:"command"`
    Success    bool          `json:"success"`
    ExitCode   int           `json:"exit_code,omitempty"`
    Output     string        `json:"output,omitempty"` // 最大1000文字に制限
    RetryCount int           `json:"retry_count"`
    Duration   time.Duration `json:"duration"`
}

type Entry struct {
    TaskFile            string               `json:"task_file"`
    ExecutedAt          time.Time            `json:"executed_at"`
    Success             bool                 `json:"success"`
    Duration            time.Duration        `json:"duration"`
    TaskCount           int                  `json:"task_count"`
    ErrorMessage        string               `json:"error_message,omitempty"`
    StartFrom           int                  `json:"start_from,omitempty"`
    MaxRetries          int                  `json:"max_retries,omitempty"`
    BranchName          string               `json:"branch_name,omitempty"`
    VerificationResults []VerificationResult `json:"verification_results,omitempty"`
}
```

### ステップ5: sync.goの検証ロジック修正
検証結果を記録しながら実行

```go
// Run verification command with retry logic
if task.Command != "" {
    fmt.Printf("\n🔍 Running verification: %s\n", task.Command)

    retryCount := 0
    verificationPassed := false
    startVerification := time.Now()
    var verificationResult VerificationResult

    for retryCount <= maxRetries {
        err := runCommand(task.Command, f)

        // 検証結果を記録
        if cmdErr, ok := err.(*CommandError); ok {
            verificationResult = VerificationResult{
                TaskNumber: taskNum,
                TaskTitle:  task.Title,
                Command:    task.Command,
                Success:    false,
                ExitCode:   cmdErr.ExitCode,
                Output:     truncateString(cmdErr.Output, 1000),
                RetryCount: retryCount,
                Duration:   time.Since(startVerification),
            }
        } else if err == nil {
            verificationResult = VerificationResult{
                TaskNumber: taskNum,
                TaskTitle:  task.Title,
                Command:    task.Command,
                Success:    true,
                RetryCount: retryCount,
                Duration:   time.Since(startVerification),
            }
            verificationPassed = true
        }

        // 検証結果をスライスに追加（後で履歴に保存）
        verificationResults = append(verificationResults, verificationResult)

        if err != nil {
            // ... (既存のリトライロジック)
        } else {
            break
        }
    }

    // ...
}
```

### ステップ6: history.Record関数の更新
検証結果を受け取って記録

```go
func Record(projectDir, taskFile, branchName string, success bool, duration time.Duration,
    taskCount, startFrom, maxRetries int, errorMsg string, verificationResults []VerificationResult) error {
    // ...
    entry := Entry{
        TaskFile:            taskFile,
        ExecutedAt:          time.Now(),
        Success:             success,
        Duration:            duration,
        TaskCount:           taskCount,
        ErrorMessage:        errorMsg,
        StartFrom:           startFrom,
        MaxRetries:          maxRetries,
        BranchName:          branchName,
        VerificationResults: verificationResults,
    }
    // ...
}
```

### ステップ7: テストとドキュメント更新
1. 既存のテストが通ることを確認
2. 新機能のテストを追加
3. CLAUDE.mdを更新して新機能を説明

## テスト方法

### 1. 終了コードチェックのテスト

#### テストケース1: 正常終了
```bash
cat > tasks-test-success.txt << 'EOF'
## タスク1: 正常終了テスト

echo コマンドでファイルを作成

### 確認
- `echo "test" > /tmp/sleepship-test.txt && test -f /tmp/sleepship-test.txt`
EOF

./bin/sleepship sync tasks-test-success.txt
```

期待結果: タスクが成功する

#### テストケース2: 異常終了
```bash
cat > tasks-test-failure.txt << 'EOF'
## タスク1: 異常終了テスト

存在しないファイルを確認

### 確認
- `test -f /tmp/nonexistent-file-12345.txt`
EOF

./bin/sleepship sync tasks-test-failure.txt
```

期待結果: 検証が失敗し、リトライ後にタスクが失敗する

### 2. プロンプト改善のテスト

#### テストケース3: 検証コマンドあり
```bash
cat > tasks-test-with-verification.txt << 'EOF'
## タスク1: ファイル作成

/tmp/sleepship-prompt-test.txt を作成し、内容を "Hello Sleepship" にする

### 確認
- `grep -q "Hello Sleepship" /tmp/sleepship-prompt-test.txt`
EOF

./bin/sleepship sync tasks-test-with-verification.txt
```

期待結果: Claudeが検証基準を理解し、ファイルを正しく作成する

#### テストケース4: 検証コマンドなし
```bash
cat > tasks-test-without-verification.txt << 'EOF'
## タスク1: ログファイル作成

/tmp/sleepship-log.txt にタイムスタンプ付きログを追加する
EOF

./bin/sleepship sync tasks-test-without-verification.txt
```

期待結果: Claudeが実装内容を報告し、タスクが完了する

### 3. 検証結果記録のテスト

#### テストケース5: 履歴確認
```bash
# 上記のテストを実行後
./bin/sleepship history --last 1
```

期待結果:
- 検証結果が記録されている
- 終了コード、リトライ回数、実行時間が記録されている
- 失敗時の出力が記録されている（最大1000文字）

### 4. リトライ機能のテスト

#### テストケース6: リトライ成功
```bash
cat > tasks-test-retry.txt << 'EOF'
## タスク1: リトライテスト

/tmp/sleepship-retry-test.txt を作成し、内容を "Success" にする

### 確認
- `test -f /tmp/sleepship-retry-test.txt`
- `grep -q "Success" /tmp/sleepship-retry-test.txt`
EOF

# 事前に誤ったファイルを作成（Claudeが修正する必要がある）
echo "Wrong" > /tmp/sleepship-retry-test.txt

./bin/sleepship sync tasks-test-retry.txt
```

期待結果: Claudeがリトライで修正し、検証が成功する

### 5. 統合テスト

#### テストケース7: 複数タスクの連続実行
```bash
cat > tasks-test-integration.txt << 'EOF'
## タスク1: 初期化

/tmp/sleepship-integration/ ディレクトリを作成

### 確認
- `test -d /tmp/sleepship-integration`

## タスク2: ファイル作成

/tmp/sleepship-integration/data.txt を作成

### 確認
- `test -f /tmp/sleepship-integration/data.txt`

## タスク3: 検証

data.txt の内容を確認

### 確認
- `test -s /tmp/sleepship-integration/data.txt`
EOF

./bin/sleepship sync tasks-test-integration.txt
```

期待結果:
- すべてのタスクが順次成功する
- 各タスクの検証結果が記録される
- 履歴に3つの検証結果が記録される

### 6. エラーハンドリングのテスト

#### テストケース8: 最大リトライ超過
```bash
cat > tasks-test-max-retries.txt << 'EOF'
## タスク1: 必ず失敗するテスト

このタスクは意図的に失敗します

### 確認
- `false`
EOF

./bin/sleepship sync tasks-test-max-retries.txt --max-retries 2
```

期待結果:
- 2回のリトライ後にタスクが失敗する
- 履歴に失敗が記録される
- エラーメッセージが適切に記録される

## 後方互換性

### 既存タスクファイルとの互換性
- 検証コマンドがないタスクも引き続き動作する
- 既存の `### 確認` セクションの形式を維持
- 新しいフィールド（前提条件など）はオプション扱い

### 履歴ファイルの互換性
- 新しいフィールド（VerificationResults）はオプショナル
- 既存の履歴ファイルは引き続き読み込み可能
- 新しい形式で保存しても古いバージョンでエラーにならない（新フィールドは無視される）

## 性能への影響

### メモリ使用量
- 検証結果の出力を1000文字に制限することで、メモリ使用量を抑制
- 大量のタスクがある場合でも、履歴ファイルサイズは妥当な範囲に収まる

### 実行時間
- 終了コードチェックの追加による遅延はほぼゼロ
- プロンプト改善による実行時間の増加もほぼゼロ
- 検証結果の記録による遅延はほぼゼロ

## セキュリティ考慮事項

### 出力の記録
- 検証コマンドの出力に秘密情報（パスワード、APIキーなど）が含まれる可能性
- 履歴ファイルのパーミッションは0600に設定済み（history.go:72）
- 必要に応じて、特定のパターンをマスクする機能を追加可能

### コマンドインジェクション
- `runCommand` は `bash -c` でコマンドを実行
- タスクファイルから読み込まれたコマンドをそのまま実行するため、タスクファイルの信頼性が重要
- タスクファイルは開発者が作成するため、通常は問題なし

## まとめ

この設計により、以下の改善が期待できます:

1. **検証の厳密性向上**: 終了コードを明示的にチェックすることで、検証の信頼性が向上
2. **Claudeの動作改善**: 検証基準を明示的に指示することで、Claudeの実装精度が向上
3. **トラブルシューティング改善**: 検証結果の詳細記録により、問題の特定が容易に
4. **段階的実装**: 3つのフェーズに分けることで、リスクを最小化しつつ改善を進められる

実装は、まずフェーズ1から開始し、動作を確認しながら次のフェーズに進むことを推奨します。
