# 【Claude Code × Go】タスクファイルを置くだけでAIが自律開発してくれるツール作った

## はじめに

「AIにコーディングを任せる」という発想は、もはや珍しいものではなくなりました。
しかし、「タスクファイルを作っておけば、**バックグラウンドで勝手に開発が進む**」というのはどうでしょうか？

本記事では、Claude Codeというローカル実行型のAIコーディングエージェントを活用した、**完全自律型の開発ツール「Auto Issue Finder」**を紹介します。

### この記事でわかること

- Auto Issue Finderの使い方（インストールから実行まで）
- 実際の使用例とユースケース
- タスクファイルの書き方のコツ
- AIを活用した開発の可能性

### 対象読者

- Claude Codeを使っている、または使ってみたい方
- 開発の自動化・効率化を考えているエンジニア
- プログラミング学習中で、AIの力を借りたい方

### 動作環境

```plaintext:開発環境
OS: macOS (Linux/Windowsでも動作)
言語: Go 1.21+
使用ツール: Claude Code
```

---

## Claude Codeとは？

**Claude Code**は、Anthropic社が提供するローカル実行型のAIコーディングエージェントです。
従来のAIアシスタントと異なり、以下のような**実際の開発作業を自律的に実行**できます:

- ファイルの検索・読み込み・編集
- コードのビルドやテスト実行
- Git操作（コミット、ブランチ作成、プルリクエスト作成）
- エラーが発生した場合の自動修正
- 複雑なタスクを複数のステップに分解して実行

つまり、**「人間のエンジニアが行う作業をAIが代行する」**ことが可能になります。

公式ドキュメント: https://docs.claude.com/claude-code

---

## 作ったもの:「Auto Issue Finder」

### コンセプト

```plaintext
タスクファイルに「やってほしいこと」を書く
    ↓
コマンド一発で実行
    ↓
バックグラウンドでClaude Codeが自律開発
    ↓
エラーが出ても自動修正
    ↓
完成！
```

**「開発を開発に任せる」** - これが本ツールのコンセプトです。

### 主要機能

- **バックグラウンド実行** - コマンド実行後すぐに制御が戻り、裏でタスクが進む
- **真の同期処理** - タスクを順次実行、各タスク完了後に次へ進む
- **自動検証** - 各タスク後に確認コマンド実行（`go build`, `go test`等）
- **自動エラー修正** - 検証失敗時にClaude Codeが自動で修正を試みる
- **汎用性** - `--dir`オプションで任意のプロジェクトで使用可能
- **詳細ログ** - 全実行内容をログファイルに記録

### 具体的な使い方

まず、タスクファイルを作成します:

```markdown:tasks.txt
## タスク1: HTTPサーバー実装

基本的なHTTPサーバーを main.go に実装してください。

### 実装
- ポート8080でHTTPサーバーを起動
- "/" エンドポイントで "Hello, World!" を返す
- "/health" エンドポイントでヘルスチェック

### 確認
- `go build`
- `go run main.go &`

---

## タスク2: テスト追加

HTTPハンドラーのユニットテストを追加してください。

### 確認
- `go test ./...`
```

次に、コマンド一発で実行:

```bash:実行コマンド
./bin/auto-issue-finder sync tasks.txt
```

出力:

```plaintext:実行結果
✅ Started background execution (PID: 12345)
📝 Log file: /path/to/logs/sync-20231208-115735.log
💡 Monitor: tail -f /path/to/logs/sync-20231208-115735.log
```

**これだけで、バックグラウンドでClaude Codeが動き出します。**

実行される流れ:

1. タスク1を Claude Code が実装
2. `go build`を実行して確認
3. ✅ 成功 → タスク2へ
4. ❌ 失敗 → 自動修正 → 再確認 → 次へ
5. すべて完了するまで繰り返し

---

## インストール方法

### 前提条件

- Go 1.21以上がインストールされていること
- Claude Codeがインストール済みであること

### インストール手順

```bash:インストール
# 1. リポジトリをクローン
git clone https://github.com/isiidaisuke0926/auto-issue-finder.git
cd auto-issue-finder

# 2. ビルド
go build -o bin/auto-issue-finder

# 3. 動作確認
./bin/auto-issue-finder --help
```

これでインストール完了です！

---

## 使い方

### 基本的な流れ

**STEP 1: タスクファイルを作成**

タスクファイルの作成は、**Claude Codeに書かせるのがおすすめ**です:

```bash:Claude Codeで作成
# Claude Codeに以下のように依頼
"HTTPサーバーとテストを実装するタスクファイル(my-tasks.txt)を作ってください"
```

もちろん、テンプレートや手動での作成も可能です:

```bash:テンプレート生成
./bin/auto-issue-finder init my-tasks.txt
```

タスクファイルの例:

```markdown:my-tasks.txt
## タスク1: HTTPサーバー実装

基本的なHTTPサーバーを main.go に実装してください。

### 実装
- ポート8080でHTTPサーバーを起動
- "/" エンドポイントで "Hello, World!" を返す
- "/health" エンドポイントでヘルスチェック

### 確認
- `go build`
- `go run main.go &`

---

## タスク2: テスト追加

HTTPハンドラーのユニットテストを追加してください。

### 確認
- `go test ./...`
```

**STEP 2: 実行**

```bash:実行
./bin/auto-issue-finder sync my-tasks.txt
```

出力:

```plaintext:実行結果
✅ Started background execution (PID: 12345)
📝 Log file: /path/to/logs/sync-20231208-115735.log
💡 Monitor: tail -f /path/to/logs/sync-20231208-115735.log
```

**STEP 3: ログで進捗確認**

```bash:ログ監視
tail -f logs/sync-*.log
```

これだけです！あとはClaude Codeが自動で開発を進めてくれます。

### 別プロジェクトで実行する場合

```bash:別プロジェクトでの実行
./bin/auto-issue-finder sync tasks.txt --dir=/path/to/your/project
```

---

## タスクファイルの書き方のコツ

### 1. タスクは小さく分割する

**良い例（適切に分割）:**

```markdown:good_example.txt
## タスク1: データベースモデル作成
models/user.go にUserモデルを作成

### 確認
- `go build`

## タスク2: CRUD操作実装
repositories/user.go にCRUD操作を実装

### 確認
- `go test ./repositories`
```

**悪い例（大きすぎる）:**

```markdown:bad_example.txt
## タスク1: ユーザー管理機能を全部作る
モデル、リポジトリ、API、テスト全部実装して

### 確認
- `go test ./...`
```

**ポイント**: 1タスク = 1つの明確な実装単位

### 2. 確認コマンドを明確にする

「何が成功なのか」を具体的に指定することで、AIが正しく実装できたか自動判定できます:

```markdown:確認コマンド例
### 確認
- `go build`                    # ビルドが通る
- `go test ./...`               # すべてのテストが通る
- `docker build -t app .`       # Dockerイメージがビルドできる
```

### 3. タスクの詳細は具体的に

AIに何を期待するかを明確に:

```markdown:良い例
## タスク1: ユーザー認証API

handlers/auth.go に以下のエンドポイントを実装してください:
- POST /auth/login - メールとパスワードでログイン
- POST /auth/logout - ログアウト
- JWTトークンを使った認証
- エラーハンドリング（400, 401, 500）

### 確認
- `go build`
- `go test ./handlers`
```

---

## 実際の使用例とユースケース

### ケース1: プロトタイプの高速作成

```markdown:prototype.txt
## タスク1: REST API基盤
Go言語でRESTful APIの基盤を作成

### 確認
- `go build`

## タスク2: ユーザー登録エンドポイント
POST /users でユーザー登録APIを実装

### 確認
- `go test ./handlers`

## タスク3: Docker対応
Dockerfileとdocker-compose.ymlを作成

### 確認
- `docker build -t app .`
```

**結果**: 夜実行しておけば、朝には動くプロトタイプができている

### ケース2: テストコードの自動生成

```markdown:test_gen.txt
## タスク1: handlers パッケージのテスト
handlers/ 配下のすべてのハンドラーにユニットテストを追加

### 確認
- `go test ./handlers -cover`

## タスク2: repositories パッケージのテスト
repositories/ 配下のすべてのリポジトリにテストを追加

### 確認
- `go test ./repositories -cover`
```

**結果**: 退屈なテストコード作成から解放される

### ケース3: 他のプロジェクトでの利用

```bash:別プロジェクトで実行
# Node.jsプロジェクトでも使える
./bin/auto-issue-finder sync tasks.txt --dir=/path/to/node-project
```

```markdown:node_tasks.txt
## タスク1: TypeScript設定
tsconfig.json を作成してTypeScript環境を構築

### 確認
- `npm run build`

## タスク2: Express APIサーバー
src/server.ts にExpressサーバーを実装

### 確認
- `npm test`
```

**結果**: 言語・フレームワークを問わず使える汎用ツール

---

## よくある質問（FAQ）

### Q1. 本当にバックグラウンドで動きますか？

はい。コマンド実行後すぐに制御が戻るので、他の作業をしながらタスクが進みます。

### Q2. エラーが出たらどうなりますか？

Claude Codeが自動でエラーを検出し、修正を試みます。1回の修正で直らない場合はログに記録されて停止します。

### Q3. 任意のプログラミング言語で使えますか？

はい。`--dir`オプションで任意のプロジェクトを指定でき、確認コマンドも自由に設定できます（`npm test`, `python -m pytest`など）。

### Q4. タスクが失敗したらどうすればいいですか？

ログファイル（`logs/sync-*.log`）を確認して、どこで失敗したかを把握できます。タスクを小さく分割するか、より具体的な指示に修正してください。

### Q5. 複数のプロジェクトで同時に実行できますか？

はい。別々のプロジェクトディレクトリで実行すれば、並列で複数のタスクを進められます。

---

## 注意事項とTips

### 注意事項

- **完全放置は推奨しません**: たまにログを確認して進捗をチェックしましょう
- **大きすぎるタスクは避ける**: 1タスクは1機能に絞ることで成功率が上がります
- **確認コマンドは必須**: 確認コマンドがないと、正しく実装されたか判定できません

### Tips

- **夜間実行がおすすめ**: 寝る前に実行しておけば、朝には完成しています
- **テンプレート活用**: よく使うタスクはテンプレート化しておくと便利です
- **ログは宝の山**: ログを見るとAIの思考プロセスがわかり、学びになります

---

## まとめ

Auto Issue Finderは、タスクファイルを作成するだけでClaude Codeが自律的に開発を進めてくれるツールです。

### こんな人におすすめ

- Claude Codeを使っているが、もっと効率化したい
- アイデアはあるけど実装に時間がかかる
- テストコードやドキュメント作成を自動化したい
- プログラミング学習中で、AIの力を借りたい

### 始めるには

```bash:クイックスタート
# 1. インストール
git clone https://github.com/isiidaisuke0926/auto-issue-finder.git
cd auto-issue-finder
go build -o bin/auto-issue-finder

# 2. タスクファイル作成
./bin/auto-issue-finder init my-tasks.txt

# 3. 実行（これだけ！）
./bin/auto-issue-finder sync my-tasks.txt
```

### 開発の未来

AIエージェントの登場により、開発の本質は変わりつつあります:

- **従来**: コードを書くスキルが必要
- **これから**: タスクを分解し、適切に指示を出すスキルが必要

プログラミングは「書く」から「設計する」へ。
AIと一緒に開発する時代が、もう始まっています。

---

## おわりに

ぜひAuto Issue Finderを使って、あなたのアイデアを形にしてみてください。
きっと、**「こんなに簡単に開発できるんだ」**と驚くはずです。

バグ報告や機能提案は[GitHub Issues](https://github.com/isiidaisuke0926/auto-issue-finder/issues)でお待ちしています。
あなたの使用例やタスクファイルのベストプラクティスもぜひシェアしてください！

---

## 参考資料

### プロジェクト情報

- **GitHubリポジトリ**: [auto-issue-finder](https://github.com/isiidaisuke0926/auto-issue-finder)
- **使用言語**: Go 1.21+
- **ライセンス**: MIT License

### 関連リンク

- [Claude Code公式ドキュメント](https://docs.claude.com/claude-code)
- [Go言語公式サイト](https://go.dev/)

---

## 付録: タスクファイルテンプレート集

リポジトリには、以下のようなテンプレートを用意しています。

### Webアプリケーション開発

```markdown:webapp_template.txt
## タスク1: プロジェクト初期化
Go言語でWebアプリケーションのプロジェクト構造を作成

### 確認
- `go mod init example.com/webapp`
- `go mod tidy`

## タスク2: HTTPサーバー実装
Ginフレームワークを使ってRESTful APIサーバーを実装

### 確認
- `go build`
- `go test ./handlers`
```

### テスト駆動開発

```markdown:tdd_template.txt
## タスク1: テストファイル作成
既存のコードに対してユニットテストを作成

### 確認
- `go test ./... -v`

## タスク2: カバレッジ向上
テストカバレッジを80%以上に

### 確認
- `go test -cover ./... | grep "80"`
```

---

最後まで読んでいただき、ありがとうございました！

ぜひAuto Issue Finderを試してみて、感想や改善案をシェアしてください。
あなたの開発ライフが少しでも楽になれば嬉しいです。

**Happy Coding with AI! 🚀**
