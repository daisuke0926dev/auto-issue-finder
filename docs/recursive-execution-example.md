# Sleepship 再帰実行の実践例

## 概要

このドキュメントは、Sleepshipの再帰実行機能（動的タスク生成）の実際の動作例を記録したものです。`tasks-user-friendly.txt` を使用して、調査→改善提案→実装タスク生成→実行という段階的な開発フローを実現しました。

## 実行日時

- **実行開始**: 2025-12-08 21:47:10
- **実行コマンド**: `./bin/sleepship sync tasks-user-friendly.txt`

## タスク構成

### Phase 1: tasks-user-friendly.txt

Phase 1では以下の3つのタスクを実行しました:

1. **タスク1: 類似OSSツールの調査 - CLI設計パターン**
   - GitHub CLI, Cargo, npm/yarn, Docker, kubectlの調査
   - 出力: `docs/research-cli-ux.md`
   - 実行時間: 約5分

2. **タスク2: sleepshipの改善点の洗い出し**
   - タスク1の調査結果を基に改善提案を作成
   - 出力: `docs/improvement-proposals.md`
   - P0/P1/P2の優先度付き提案リスト
   - 実行時間: 約3分

3. **タスク3: 次フェーズのタスクファイル生成と実行**
   - タスク2の提案を基にPhase 2のタスクファイルを自動生成
   - 出力: `tasks-user-friendly-phase2.txt`
   - **再帰実行**: 生成したタスクファイルをsleepshipで自動実行
   - 実行時間: Phase 2完了まで継続

### Phase 2: tasks-user-friendly-phase2.txt（動的生成）

Phase 1のタスク3で自動生成され、再帰実行されたタスク群:

1. **タスク1: 環境変数による設定オーバーライド機能の実装**
   - `internal/config/env.go` の実装
   - `SLEEPSHIP_*` プレフィックスの環境変数サポート

2. **タスク2: エイリアス機能の実装**
   - `.sleepship.toml` にエイリアス設定追加
   - `sleepship alias list` コマンド実装

3. **タスク3: タスク実行履歴の記録機能の実装**
   - `.sleepship/history.json` に履歴記録
   - `sleepship history` コマンド実装

4. **タスク4: ドキュメントの更新**
   - CLAUDE.mdの更新
   - 新機能の使用例追加

5. **タスク5: 統合テストとビルド確認**
   - 全機能の統合テスト
   - ビルド確認

## 実行フロー

```
tasks-user-friendly.txt
  │
  ├─ タスク1: CLI設計パターン調査
  │   └─ 出力: docs/research-cli-ux.md
  │
  ├─ タスク2: 改善提案の洗い出し
  │   └─ 出力: docs/improvement-proposals.md
  │
  └─ タスク3: Phase2タスク生成と実行
      ├─ 出力: tasks-user-friendly-phase2.txt
      └─ 再帰実行: ./bin/sleepship sync tasks-user-friendly-phase2.txt
          │
          ├─ タスク1: 環境変数オーバーライド実装
          ├─ タスク2: エイリアス機能実装
          ├─ タスク3: 実行履歴記録実装
          ├─ タスク4: ドキュメント更新
          └─ タスク5: 統合テスト
```

## 再帰実行のメカニズム

### 環境変数による深度管理

Sleepshipは `SLEEPSHIP_DEPTH` 環境変数で再帰深度を管理します:

- **階層1**: メインタスクファイル (`SLEEPSHIP_DEPTH=1`)
- **階層2**: サブタスクファイル (`SLEEPSHIP_DEPTH=2`)
- **階層3**: サブサブタスクファイル (`SLEEPSHIP_DEPTH=3`)
- **階層4以降**: エラーで停止（無限再帰を防止）

本実行例では:
- Phase 1: `SLEEPSHIP_DEPTH=1`
- Phase 2: `SLEEPSHIP_DEPTH=2`

### プロセス分離

各フェーズは独立したプロセスとして実行されます:

```bash
# Phase 1のプロセス
PID: 66614 - sleepship sync tasks-user-friendly.txt --worker

# Phase 2のプロセス（Phase 1から起動）
PID: 82605 - sleepship sync tasks-user-friendly-phase2.txt --worker
```

### ログファイル分離

各実行は独立したログファイルを生成します:

- Phase 1: `logs/sync-20251208-214710.log`
- Phase 2: `logs/sync-20251208-215708.log`（およびその他）

## 生成されたドキュメント

### 1. docs/research-cli-ux.md

主要CLIツールの設計パターン調査結果:

- **対象ツール**: GitHub CLI, Cargo, npm/yarn, Docker, kubectl
- **調査項目**:
  - リモートリソース指定方法
  - パス指定の簡略化手法
  - 設定ファイルの場所と形式
  - よく使うパターン（エイリアス、ショートハンド等）
  - デフォルト値の扱い方
- **分量**: 約940行

#### 主要な発見

1. **自動検出とデフォルト値の活用**
   - Cargo, npm/yarnはカレントディレクトリから親へ再帰探索
   - GitHub CLIはgit remoteから推論

2. **階層的な設定管理**
   - 優先順位: CLI > 環境変数 > プロジェクト設定 > ユーザー設定 > デフォルト

3. **ショートハンドとエイリアス**
   - kubectlのリソース短縮名（`po`, `svc`, `deploy`）
   - GitHub CLIの自分のリポジトリに対するOWNER省略

### 2. docs/improvement-proposals.md

Sleepshipに適用できる改善提案:

- **現在の課題**: タスクファイル指定の制約、プロジェクトディレクトリの扱い等
- **優先度付き提案**:
  - **P0（Critical）**: 設定ファイルサポート、プロジェクトルート自動検出等
  - **P1（High）**: 環境変数オーバーライド、エイリアス機能等
  - **P2（Medium）**: インタラクティブモード、プラグインシステム等
- **実装難易度と影響範囲**の分析
- **分量**: 約745行

#### 次フェーズで実装すべき機能（P0）

1. **設定ファイルサポート**（`.sleepship.toml`）
2. **プロジェクトルートの自動検出**
3. **デフォルトタスクファイルの自動検出**

### 3. tasks-user-friendly-phase2.txt

Phase 2の実装タスクファイル（自動生成）:

- **タスク数**: 5個
- **対象機能**: P1の高優先度機能を中心に選定
- **形式**: 標準的なSleepshipタスクファイル形式
- **分量**: 約170行

## 再帰実行の検証ポイント

### ✅ 成功したポイント

1. **動的タスク生成**: Phase 1のタスク3で、改善提案を読み取って具体的な実装タスクを自動生成
2. **再帰実行の開始**: 生成したタスクファイルが自動的に実行開始
3. **深度制限**: `SLEEPSHIP_DEPTH` が正しく機能し、無限再帰を防止
4. **プロセス分離**: 各フェーズが独立したプロセスとして動作
5. **ログ分離**: 各実行が独立したログファイルを生成
6. **ブランチ管理**: Phase 1で作成した `feature/user-friendly` ブランチがPhase 2でも継続使用

### 📊 実行統計

| 項目 | 値 |
|------|-----|
| Phase 1タスク数 | 3 |
| Phase 2タスク数 | 5 |
| 合計タスク数 | 8 |
| 生成されたドキュメント | 3ファイル |
| 生成されたコード行数 | 約1,685行（ドキュメント） |
| 再帰深度 | 2階層 |
| 実行プロセス数 | 2個（分離実行） |
| ログファイル数 | 複数（Phase毎に分離） |

### 🎯 実現したユースケース

1. **調査→計画→実装フロー**
   - Phase 1で調査と計画を実施
   - Phase 2で具体的な実装タスクを自動生成・実行

2. **段階的な開発**
   - 各段階で適切な情報を次の段階に渡す
   - 中間成果物（タスクファイル、ドキュメント）を検証

3. **動的なタスク分解**
   - 大きな要件を分析して実装可能な単位に自動分解
   - 依存関係を考慮した順序付け

## ベストプラクティス

### 1. タスクファイル命名規則

用途がわかる名前をつける:

```
tasks-investigation.txt  # 調査
tasks-plan.txt           # 計画
tasks-impl.txt           # 実装
tasks-<feature>-phase2.txt  # フェーズ2
```

### 2. 検証の徹底

生成したタスクファイルの存在確認:

```markdown
### 確認
- `test -f tasks-impl.txt`
- `cat tasks-impl.txt`
```

### 3. 明確な指示

生成するタスクファイルの内容を具体的に指示:

```markdown
`tasks-impl.txt` を作成:
- P0の機能から優先的にタスク化
- 各タスクに設計・実装・テスト・ドキュメント更新を含める
- タスクは適切な粒度に分割（1タスク = 1機能の実装完了）
```

### 4. 段階的実行

調査→計画→実装の順で段階を分ける:

```
Phase 1: 調査 + 計画 + Phase 2タスク生成
  ↓
Phase 2: 実装 + テスト + ドキュメント更新
```

## 制限事項と注意点

### 1. 再帰深度の制限

- 最大3階層まで（4階層目以降はエラー）
- 深いネストは避けて、フラットな構造を推奨

### 2. 実行時間

- 各タスクはClaude APIを呼び出すため、時間がかかる
- Phase 1の3タスクで約8分、Phase 2はさらに時間がかかる見込み

### 3. エラーハンドリング

- サブタスクのエラーは親タスクに伝播しない
- 各ログファイルを個別に確認する必要がある

### 4. リソース管理

- 複数のsleepshipプロセスが同時実行される
- メモリとCPUリソースに注意

## まとめ

Sleepshipの再帰実行機能により、以下のような高度な開発フローが実現できました:

1. **自動化された段階的開発**: 調査→計画→実装のフローを1つのタスクファイルで記述
2. **動的なタスク生成**: 実行時の状況に応じてタスクを自動生成
3. **柔軟なワークフロー**: 大きな要件を自動的に分解して段階的に実装

この機能により、Claudeによる自律開発がより高度になり、複雑な開発タスクを効率的に実行できるようになりました。

## 参考ログ

### Phase 1のログ抜粋

```
📋 Total tasks: 3
📁 Project directory: /Users/isiidaisuke0926/Documents/GitHub/sleepship

=== Creating Branch: feature/user-friendly
✅ Branch created: feature/user-friendly

Task 1/3: 1: 類似OSSツールの調査 - CLI設計パターン
🤖 Executing with Claude...
[タスク1実行...]
✅ Task 1 completed

Task 2/3: 2: sleepshipの改善点の洗い出し
🤖 Executing with Claude...
[タスク2実行...]
✅ Task 2 completed

Task 3/3: 3: 次フェーズのタスクファイル生成と実行
🤖 Executing with Claude...
[タスク3実行...]
✅ Verification passed
[Phase 2の再帰実行開始...]
```

### Phase 2のプロセス確認

```bash
$ ps aux | grep sleepship
isiidaisuke0926  66614  ... sleepship sync tasks-user-friendly.txt --worker
isiidaisuke0926  82605  ... sleepship sync tasks-user-friendly-phase2.txt --worker
```

### 生成されたファイル

```bash
$ ls -la docs/
docs/research-cli-ux.md          # 940行 - CLI設計パターン調査
docs/improvement-proposals.md     # 745行 - 改善提案
docs/recursive-execution-example.md  # このファイル

$ ls -la tasks-*
tasks-user-friendly.txt           # Phase 1のタスクファイル
tasks-user-friendly-phase2.txt    # Phase 2のタスクファイル（自動生成）
```
