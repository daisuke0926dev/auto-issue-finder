# タスク検証ロジックの調査結果

## 概要
このドキュメントは、`cmd/sync.go` の `executeTask` 関数における現在の確認ステップの実装を調査した結果をまとめたものです。

## 調査対象ファイル
- `cmd/sync.go`

## executeTask 関数の概要

### 関数シグネチャ
```go
func executeTask(task Task, logFile *os.File) error
```

**場所**: `cmd/sync.go:456-475`

### 実装内容
`executeTask` 関数は非常にシンプルで、以下の処理のみを行います:

1. タスク実行用のプロンプトを生成
2. `executeClaude` 関数を呼び出してClaude Codeでタスクを実行

```go
func executeTask(task Task, logFile *os.File) error {
	prompt := fmt.Sprintf(`あなたは自律的にソフトウェア開発を行うエンジニアです。

# タスク
%s

%s

# 指示
1. このタスクを完全に実装してください
2. 必要なファイルを作成・編集してください
3. 実装後、必ず動作確認してください
4. エラーがあれば修正してください

プロジェクトディレクトリ: %s

実装を開始してください。`, task.Title, task.Description, projectDir)

	return executeClaude(prompt, logFile)
}
```

### 重要な発見
**`executeTask` 関数自体は確認コマンドを実行しません。** 確認コマンドの実行は、`runSync` 関数内で別途行われています。

## 確認コマンドの実行方法

### 1. 確認コマンドの実行場所
確認コマンドは `runSync` 関数内で実行されます（`cmd/sync.go:306-372`）。

### 2. 実行フロー

```go
// タスク実行
if err := executeTask(task, f); err != nil {
	// エラーハンドリング（リトライロジック）
}

// 確認コマンド実行
if task.Command != "" {
	fmt.Printf("\n🔍 Running verification: %s\n", task.Command)

	retryCount := 0
	verificationPassed := false

	for retryCount <= maxRetries {
		if err := runCommand(task.Command, f); err != nil {
			// 検証失敗時の処理
		} else {
			// 検証成功
			verificationPassed = true
			break
		}
	}
}
```

### 3. 実行タイミング
1. **タスク実行後**: `executeTask` でClaude Codeがタスクを実装
2. **確認実行**: タスク実装が成功した後、`task.Command` が存在する場合のみ実行

## 確認コマンドの結果の扱い

### runCommand 関数の仕様
**場所**: `cmd/sync.go:498-534`

```go
func runCommand(command string, logFile *os.File) error {
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	_, _ = logFile.Write(output)

	if err != nil {
		return fmt.Errorf("%w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Output: %s\n", string(output))
	return nil
}
```

### 結果判定
- **成功**: コマンドの終了コードが0の場合
- **失敗**: コマンドの終了コードが0以外の場合（エラーを返す）

### 特殊ケース: sleepship 再帰実行
`runCommand` には、sleepship コマンドの再帰実行を検出する特別なロジックがあります:

```go
isSleepshipCommand := strings.Contains(command, "sleepship") || strings.Contains(command, "./bin/sleepship")

if isSleepshipCommand {
	currentDepth := getCurrentRecursionDepth()
	if currentDepth >= maxRecursionDepth {
		// 最大深度に達した場合はスキップ（エラーにしない）
		return nil
	}
	// 環境変数で深度を設定
	cmd.Env = append(cmd.Env, fmt.Sprintf("SLEEPSHIP_DEPTH=%d", currentDepth+1))
}
```

## タスクの成功/失敗の判定

### 判定基準
1. **タスク実行の成功**: `executeTask` がエラーを返さない
2. **確認コマンドの成功**: `task.Command` が空、または `runCommand` がエラーを返さない

### 失敗時の処理
両方のステップでリトライロジックが実装されています。

#### タスク実行失敗時（`cmd/sync.go:238-300`）
```go
for taskRetryCount <= maxRetries {
	if err := executeTask(task, f); err != nil {
		lastErr = err
		taskRetryCount++

		if taskRetryCount > maxRetries {
			// 最大リトライ回数超過: 実行停止
			return fmt.Errorf("task %d failed after %d attempts: %w", taskNum, maxRetries+1, err)
		}

		// リトライ: エラーコンテキスト付きでClaude Codeを再実行
		retryPrompt := fmt.Sprintf(`前回のタスク実行でエラーが発生しました (リトライ %d/%d):
エラー: %v

# タスク
%s

%s

# 指示
1. 前回のエラーを修正してください
...`, taskRetryCount, maxRetries, err, task.Title, task.Description)

		if err := executeClaude(retryPrompt, f); err != nil {
			continue
		}

		lastErr = nil
		break
	}
	break
}
```

#### 確認コマンド失敗時（`cmd/sync.go:313-367`）
```go
for retryCount <= maxRetries {
	if err := runCommand(task.Command, f); err != nil {
		retryCount++

		if retryCount > maxRetries {
			// 最大リトライ回数超過: 実行停止
			return fmt.Errorf("verification failed after %d attempts: %w", maxRetries+1, err)
		}

		// 修正を試みる: エラー情報付きでClaude Codeに修正を依頼
		fixPrompt := fmt.Sprintf(`検証コマンドが失敗しました（リトライ %d/%d 回目）:

コマンド: %s
エラー: %v

# 指示
1. 上記のエラーを修正してください
2. 修正後、検証が通ることを確認してください
...`, retryCount, maxRetries, task.Command, err)

		if err := executeClaude(fixPrompt, f); err != nil {
			continue
		}

		// 修正後、検証を再実行
		continue
	}

	// 検証成功
	verificationPassed = true
	break
}
```

## 次のタスクに進む条件

### 条件
1. **タスク実行が成功** している
2. **確認コマンドが成功** している（確認コマンドが存在する場合）

### 実装詳細
`runSync` 関数のメインループ（`cmd/sync.go:222-382`）:

```go
for i, task := range tasks {
	taskNum := i + 1

	// タスク実行（リトライ付き）
	// ...

	// 確認コマンド実行（リトライ付き）
	if task.Command != "" {
		// ...
		if !verificationPassed {
			return fmt.Errorf("verification failed after all retries")
		}
	}

	// タスク完了後の処理
	if err := commitTaskChanges(task, taskNum, f); err != nil {
		// コミット失敗は警告のみ（処理は継続）
	}

	fmt.Printf("\n✅ Task %d completed\n\n", taskNum)

	// 次のループイテレーションで次のタスクへ
}
```

**重要**: タスクまたは確認が失敗してリトライ上限に達した場合、`return` でエラーを返すため、ループが中断され、次のタスクには進みません。

## Task構造体の定義

```go
type Task struct {
	Title       string
	Description string
	Command     string // 確認コマンド（go build, go test等）
}
```

**場所**: `cmd/sync.go:34-38`

### フィールド説明
- `Title`: タスクのタイトル（`## タスク1: ...` から抽出）
- `Description`: タスクの説明文（本文）
- `Command`: 確認コマンド（`### 確認` セクションの `` - `command` `` 形式から抽出）

## 確認コマンドのパース方法

### parseTaskFile 関数
**場所**: `cmd/sync.go:401-454`

```go
// 確認コマンド (line starting with "- `")
if currentTask != nil && strings.HasPrefix(line, "- `") && strings.HasSuffix(line, "`") {
	// Extract command from "- `command`" format
	cmd := strings.TrimPrefix(line, "- `")
	cmd = strings.TrimSuffix(cmd, "`")
	currentTask.Command = cmd
	continue
}
```

### パース仕様
- **パターン**: `` - `コマンド` ``
- **抽出**: バッククォート内のコマンド文字列を取得
- **格納**: `Task.Command` に格納（最後に見つかったコマンドのみ）

### 複数の確認コマンド
現在の実装では、**複数の確認コマンドがある場合、最後の1つだけが保存されます**。これは潜在的な問題点です。

例:
```markdown
### 確認
- `go build`
- `go test ./...`
```

この場合、`Task.Command` には `"go test ./..."` のみが保存され、`"go build"` は失われます。

## まとめ

### 確認ステップの実装状況

1. **確認コマンドの実行**:
   - `runCommand` 関数で `bash -c` 経由で実行
   - コマンドの終了コードで成功/失敗を判定
   - sleepship 再帰実行の検出と深度管理あり

2. **確認コマンドの結果**:
   - **成功**: 終了コード0 → 次のタスクへ進む
   - **失敗**: 終了コード非0 → リトライロジックに入る

3. **タスクの成功/失敗判定**:
   - タスク実行と確認コマンド両方が成功した場合のみ成功
   - どちらかが失敗した場合、最大リトライ回数まで自動修正を試みる

4. **次のタスクに進む条件**:
   - タスク実行成功 + 確認コマンド成功（または確認コマンドなし）
   - 失敗時はリトライ上限まで達した時点で処理を停止

### 現状の課題

1. **複数の確認コマンドに未対応**:
   - 最後の1つだけが保存される
   - 複数のコマンドを実行したい場合、1つのコマンドに `&&` でつなげる必要がある

2. **確認コマンドの出力が限定的**:
   - コマンドの標準出力/標準エラーをログに記録
   - 画面にも出力を表示
   - しかし、出力の構造化された解析は行われていない

3. **エラーメッセージの伝達**:
   - エラー発生時、コマンドの出力全体をエラーメッセージに含める
   - Claude Codeへのリトライプロンプトにエラー内容を含める
   - 効果的だが、大量の出力がある場合に冗長になる可能性

### 改善の余地

1. **複数確認コマンドのサポート**:
   - `Task.Command` を文字列スライスに変更
   - すべてのコマンドを順次実行

2. **確認結果の詳細化**:
   - 確認コマンドの種類（テスト、ビルド、リントなど）を認識
   - 結果をより構造化して記録

3. **段階的な確認**:
   - タスク実行中にも中間確認を入れる
   - より早期にエラーを検出
