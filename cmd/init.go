package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [task-file]",
	Short: "Initialize a new task file with template",
	Long: "Create a new task file with a template to help you get started.\n\n" +
		"The template includes example tasks that demonstrate the proper format\n" +
		"for defining tasks, implementation instructions, and verification commands.\n\n" +
		"Example:\n" +
		"  sleepship init tasks.txt",
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(_ *cobra.Command, args []string) error {
	taskFile := args[0]

	// Check if file already exists
	if _, err := os.Stat(taskFile); err == nil {
		return fmt.Errorf("file already exists: %s", taskFile)
	}

	// Write template to file
	if err := os.WriteFile(taskFile, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create task file: %w", err)
	}

	fmt.Printf("✅ Task file created: %s\n", taskFile)
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("  1. Edit the task file: %s\n", taskFile)
	fmt.Printf("  2. Run: sleepship sync %s\n", taskFile)

	return nil
}

const template = `# タスクファイル

このファイルに実装したいタスクを記述します。
Claude Codeが各タスクを順次実行します。

---

## タスク1: プロジェクト初期化

Go言語プロジェクトを初期化してください。

### 実装
以下を実行してください：
- go mod init example-project でプロジェクトを初期化
- 基本的なディレクトリ構造を作成（cmd/, internal/, pkg/）
- .gitignore ファイルを作成

### 確認
- ` + "`go mod tidy`" + `
- ` + "`ls -la`" + `

---

## タスク2: HTTPサーバー実装

基本的なHTTPサーバーを main.go に実装してください。

### 実装
main.go に以下の機能を実装：
- ポート8080でHTTPサーバーを起動
- "/" エンドポイントで "Hello, World!" を返す
- "/health" エンドポイントでヘルスチェック（JSON形式）

### 確認
- ` + "`go build`" + `

---

## タスク3: テスト追加

HTTPハンドラーのユニットテストを追加してください。

### 実装
main_test.go を作成：
- "/" エンドポイントのテスト
- "/health" エンドポイントのテスト
- レスポンスコードとボディの検証

### 確認
- ` + "`go test ./...`" + `
- ` + "`go test -cover ./...`" + `

---

## タスク4: README作成

プロジェクトのREADME.mdを作成してください。

### 実装
README.md に以下を記載：
- プロジェクトの説明
- インストール方法
- 使い方（実行コマンド）
- エンドポイント一覧

### 確認
- ` + "`cat README.md`" + `
`
