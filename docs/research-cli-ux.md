# CLIツール設計パターン調査レポート

## 概要

本レポートでは、主要なCLIツール（GitHub CLI, Cargo, npm/yarn, Docker, kubectl）における設計パターンを調査し、リモートリソース指定、パス簡略化、設定ファイル管理、デフォルト値の扱い方について分析する。

---

## 1. GitHub CLI (gh)

### リモートリポジトリの指定方法

#### ショートハンド形式: `OWNER/REPO`

- **基本形式**: `OWNER/REPO`（例: `octo-org/octo-repo`）
- **URL形式**: 完全なGitHub URLも使用可能
- **OWNER省略**: 認証ユーザーのリポジトリの場合、`OWNER/`部分を省略可能（例: `myrepo`）

```bash
# 完全指定
gh repo view octo-org/octo-repo

# 自分のリポジトリの場合（OWNER省略可）
gh repo view myrepo

# URL指定
gh repo view https://github.com/octo-org/octo-repo
```

### デフォルトリポジトリの扱い

#### 自動検出

- カレントディレクトリのgit remoteから自動検出
- `gh repo view` 引数なしで現在のディレクトリのリポジトリを表示

#### 複数リモートの扱い

- **v2.21.0以降**: 複数リモートがある場合、デフォルトを明示的に設定する必要がある
- **設定コマンド**: `gh repo set-default`
- **確認コマンド**: `gh repo set-default --view`（v2.30.0以降）

```bash
# デフォルトリポジトリの設定
gh repo set-default

# デフォルトリポジトリの確認
gh repo set-default --view

# コマンド単位でのオーバーライド
gh pr list --repo OWNER/REPO
gh issue list -R OWNER/REPO
```

### 環境変数とコマンドラインフラグ

#### 優先順位（推定）

1. コマンドラインフラグ（`-R`, `--repo`）
2. 環境変数（`GH_REPO`）
3. デフォルト設定（`gh repo set-default`）
4. git remote自動検出

#### 主要な環境変数

```bash
# リポジトリ指定（HOST/OWNER/REPO形式）
export GH_REPO=github.com/owner/repo

# エディタ指定（優先順位順）
export GH_EDITOR=vim        # 最優先
export GIT_EDITOR=emacs     # 次点
export VISUAL=nano          # 次点
export EDITOR=vi            # 最後

# ページャ指定（優先順位順）
export GH_PAGER=less        # 最優先
export PAGER=more           # 次点
```

### 設定ファイル

- ホームディレクトリの `.config/gh/` に設定を保存
- 認証情報、デフォルト設定などを管理

### ベストプラクティス

1. **ローカルリポジトリでの作業**: 引数省略で現在のディレクトリのリポジトリを自動検出
2. **複数リモート環境**: `gh repo set-default` で明示的にデフォルトを設定
3. **スクリプト化**: `GH_REPO` 環境変数で動的にリポジトリを切り替え

---

## 2. Cargo (Rust)

### プロジェクトディレクトリの検出

#### 自動検出メカニズム

- カレントディレクトリから親ディレクトリへ再帰的に `Cargo.toml` を探索
- 最初に見つかった `Cargo.toml` をプロジェクトルートとして使用

#### マニフェストパスの明示指定

```bash
# --manifest-path でCargo.tomlを明示指定
cargo build --manifest-path /path/to/Cargo.toml

# ディレクトリパス指定も可能（Cargo.tomlが自動補完される）
cargo build --manifest-path /path/to/project/

# 作業ディレクトリの変更
cargo build -C /path/to/project/
```

### Cargo.tomlベースの設定

#### マニフェストの構成要素

```toml
[package]
name = "my-project"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = "1.0"

[workspace]
members = ["crate-a", "crate-b"]
```

### ワークスペース管理

#### ワークスペースの特徴

- 複数のパッケージ（ワークスペースメンバー）を一括管理
- **共有リソース**:
  - `Cargo.lock`: ワークスペースルートで一元管理
  - `target/`: ビルド成果物の共有ディレクトリ

#### ワークスペースの種類

1. **ルートパッケージワークスペース**: `[package]` と `[workspace]` の両方を持つ
2. **仮想マニフェスト**: `[workspace]` のみを持つ（パッケージなし）

```toml
# ワークスペースルートのCargo.toml
[workspace]
members = [
    "crate-a",
    "crate-b",
    "tools/cli"
]

[patch.crates-io]
# 依存関係のパッチ（ワークスペースルートのみ有効）

[profile.release]
# プロファイル設定（ワークスペースルートのみ有効）
```

#### ワークスペースメンバーの自動検出

- ファイルシステム階層で最初に見つかった `[workspace]` セクションを持つ `Cargo.toml` がワークスペースルート
- メンバークレートは `package.workspace` キーを省略可能

### 設定ファイルの階層: `config.toml`

#### 検索順序（優先度の高い順）

1. カレントディレクトリの `.cargo/config.toml`
2. 親ディレクトリの `.cargo/config.toml`（再帰的に探索）
3. `$CARGO_HOME/config.toml`（グローバル設定）

```bash
# 例: /projects/foo/bar/baz でcargoを実行した場合
# 1. /projects/foo/bar/baz/.cargo/config.toml
# 2. /projects/foo/bar/.cargo/config.toml
# 3. /projects/foo/.cargo/config.toml
# 4. /projects/.cargo/config.toml
# 5. $CARGO_HOME/config.toml
```

#### 設定の統合

- 複数の階層からの設定を統合（マージ）
- より具体的な（深い）階層の設定が優先

### 環境変数

#### CARGO_HOME

- デフォルト: `$HOME/.cargo` (Windows: `%USERPROFILE%\.cargo`)
- レジストリインデックスとgitチェックアウトのキャッシュ保存場所
- 認証情報: `$CARGO_HOME/credentials.toml`

#### 環境変数による設定オーバーライド

- 形式: `CARGO_FOO_BAR` で `foo.bar` 設定をオーバーライド
- 変換規則:
  - 大文字に変換
  - `.` と `-` を `_` に変換

```bash
# config.tomlの build.jobs 設定をオーバーライド
export CARGO_BUILD_JOBS=4

# config.tomlの target.x86_64-unknown-linux-gnu.linker 設定をオーバーライド
export CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER=clang
```

#### 優先順位

1. 環境変数（最優先）
2. `--config KEY=VALUE` コマンドラインオプション
3. `.cargo/config.toml`（階層的）
4. デフォルト値

### ベストプラクティス

1. **プロジェクト特有の設定**: `.cargo/config.toml` をプロジェクトルートに配置
2. **グローバル設定**: `$CARGO_HOME/config.toml` でユーザー全体のデフォルト設定
3. **ワークスペース構成**: 関連クレートを統合管理し、ビルド効率を向上
4. **明示的パス指定**: CI/CD環境では `--manifest-path` で確実性を担保

---

## 3. npm / yarn

### package.jsonの自動検出

#### npm/yarnの検出メカニズム

- カレントディレクトリから親ディレクトリへ再帰的に `package.json` を探索
- 最初に見つかった `package.json` をプロジェクトルートとして使用

```bash
# ネストされたディレクトリからでも自動検出
cd /project/src/components
npm install  # /project/package.json を使用
```

### グローバル vs ローカル実行

#### インストールの種類

```bash
# ローカルインストール（デフォルト）
npm install lodash
# → ./node_modules/ にインストール
# → package.json の dependencies に追加

# グローバルインストール
npm install -g typescript
# → システム全体で使用可能
```

#### 実行可能ファイルの配置

- **ローカルインストール**: `node_modules/.bin/` にシンボリックリンク
- **グローバルインストール**: システムのグローバルbinディレクトリ

#### ローカルパッケージの実行方法

1. **npx（推奨）**:
   ```bash
   npx eslint .
   # $PATH → ./node_modules/.bin/ の順で探索して実行
   ```

2. **npm scripts**:
   ```json
   {
     "scripts": {
       "lint": "eslint .",
       "test": "jest"
     }
   }
   ```
   ```bash
   npm run lint
   # node_modules/.bin/ を自動的にPATHに追加
   ```

3. **直接パス指定**:
   ```bash
   ./node_modules/.bin/eslint .
   ```

### スクリプトエイリアス

#### package.json の scripts セクション

```json
{
  "scripts": {
    "start": "node server.js",
    "dev": "nodemon server.js",
    "build": "webpack --mode production",
    "test": "jest",
    "lint": "eslint . --fix",
    "deploy": "npm run build && npm run upload"
  }
}
```

#### 実行方法

```bash
# npm
npm run dev
npm start  # 'run' を省略可能
npm test   # 'run' を省略可能

# yarn
yarn dev
yarn start
yarn test
```

### ワークスペース（モノレポ）

#### npmワークスペース（v7+）

```json
{
  "name": "my-monorepo",
  "workspaces": [
    "packages/*",
    "tools/cli"
  ]
}
```

#### ワークスペースメンバーの命名

- ワークスペースはフォルダ名ではなく、各 `package.json` の `name` フィールドで識別される

```
my-monorepo/
├── package.json  # "workspaces": ["packages/*"]
└── packages/
    ├── package-a/
    │   └── package.json  # "name": "@myorg/package-a"
    └── package-b/
        └── package.json  # "name": "@myorg/package-b"
```

#### ワークスペースでのスクリプト実行

```bash
# 特定のワークスペースでスクリプト実行
npm run test --workspace=@myorg/package-a
npm run test -w @myorg/package-a  # 短縮形

# 複数のワークスペースで実行
npm run test -w @myorg/package-a -w @myorg/package-b

# すべてのワークスペースで実行
npm run test --workspaces

# スクリプトがないワークスペースをスキップ
npm run test --workspaces --if-present
```

#### yarnワークスペースの特徴

- yarnは最初にワークスペースをサポートしたパッケージマネージャー
- `workspace:` プロトコルでワークスペース内依存を明示

```json
{
  "dependencies": {
    "@myorg/package-a": "workspace:*",
    "@myorg/package-b": "workspace:^1.0.0"
  }
}
```

### 依存関係の解決

#### yarn resolutions

```json
{
  "resolutions": {
    "lodash": "4.17.21",
    "**/lodash": "4.17.21"
  }
}
```

#### npm overrides（v8.3+）

```json
{
  "overrides": {
    "lodash": "4.17.21",
    "package-a": {
      "lodash": "4.17.20"
    }
  }
}
```

### 設定ファイル

#### .npmrc の階層

1. プロジェクトローカル: `/path/to/project/.npmrc`
2. ユーザーグローバル: `~/.npmrc`
3. グローバル: `$PREFIX/etc/npmrc`
4. npm組み込みデフォルト

```bash
# レジストリの変更
registry=https://registry.npmjs.org/

# スコープ別レジストリ
@myorg:registry=https://npm.myorg.com/
```

### ベストプラクティス

1. **ローカルインストール優先**: グローバルインストールは最小限に
2. **npx活用**: ローカルパッケージの実行にnpxを使用
3. **npm scriptsでタスク管理**: ビルド、テスト、デプロイなどをscriptsに集約
4. **ワークスペースでモノレポ管理**: 関連パッケージを統合管理
5. **依存関係の固定**: package-lock.jsonやyarn.lockをコミット

---

## 4. Docker

### コンテナ/イメージの指定方法

#### イメージリファレンス形式

完全な形式: `[registry/][namespace/]repository[:tag][@digest]`

```bash
# 完全指定
docker.io/library/nginx:1.21.0

# registry省略（docker.ioがデフォルト）
nginx:1.21.0

# tag省略（latestがデフォルト）
nginx

# namespace省略（libraryがデフォルト for Official Images）
nginx  # = docker.io/library/nginx:latest
```

### タグとIDの扱い

#### タグの命名規則

- **latest**: デフォルトタグ（明示しない場合）
- **セマンティックバージョニング**: `1.21.0`, `1.21`, `1`
- **機能別タグ**: `nginx:alpine`, `node:18-slim`

```bash
# ビルド時にタグ指定
docker build -t myapp:1.0.0 .

# 複数タグ指定
docker build -t myapp:1.0.0 -t myapp:latest .

# ビルド後にタグ追加
docker tag myapp:1.0.0 myapp:latest
docker tag myapp:1.0.0 myregistry.com/myapp:1.0.0
```

#### イメージID

- **短縮形**: 12文字のハッシュ（例: `a1b2c3d4e5f6`）
- **完全形**: 64文字のSHA256ハッシュ

```bash
# IDで指定
docker run a1b2c3d4e5f6
docker rmi sha256:a1b2c3d4e5f6...
```

#### レジストリホスト名の判定

- `.` （DNSセパレータ）を含む
- `:` （ポートセパレータ）を含む
- `localhost` で始まる

上記のいずれかに該当する最初の `/` までの部分をレジストリホスト名として扱う。該当しない場合は `docker.io` をデフォルトレジストリとして使用。

```bash
# カスタムレジストリ
myregistry.com:5000/myapp:1.0.0
localhost:5000/myapp:latest

# Docker Hub（デフォルト）
nginx  # = docker.io/library/nginx:latest
myuser/myapp  # = docker.io/myuser/myapp:latest
```

### 設定ファイル

#### Dockerfile

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 3000
CMD ["node", "server.js"]
```

- **デフォルトファイル名**: `Dockerfile`（拡張子なし）
- **明示指定**: `docker build -f Dockerfile.prod .`

#### docker-compose.yml

```yaml
version: '3.8'
services:
  web:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
    depends_on:
      - db
  db:
    image: postgres:14
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

#### 環境変数の補間

```yaml
services:
  web:
    image: "webapp:${TAG:-latest}"
    environment:
      - DATABASE_URL=${DATABASE_URL}
```

```bash
# .envファイルまたはシェル環境変数から読み込み
export TAG=1.0.0
docker-compose up
```

##### 制約事項

- **サービス名には変数補間不可**: サービス名自体を変数にすることはできない
- **利用可能な変数**: `COMPOSE_PROJECT_NAME` などのプロジェクトレベル変数

### LABELによるメタデータ管理

```dockerfile
LABEL version="1.0.0"
LABEL maintainer="team@example.com"
LABEL org.opencontainers.image.version="1.0.0"
LABEL org.opencontainers.image.authors="team@example.com"
```

### ベストプラクティス

1. **タグ戦略**:
   - `latest` は最新の安定版を指す
   - セマンティックバージョニングで複数タグ付与（`1.21.0`, `1.21`, `1`）
   - 環境別タグ（`myapp:dev`, `myapp:staging`, `myapp:prod`）

2. **イメージ命名**:
   - 組織名/プロジェクト名を明示（`myorg/myapp`）
   - レジストリを明示的に指定（本番環境）

3. **Dockerfile**:
   - マルチステージビルドで最終イメージサイズを削減
   - `.dockerignore` で不要なファイルを除外

4. **docker-compose**:
   - 環境変数で設定を外部化
   - ボリュームでデータ永続化

---

## 5. kubectl

### リソース指定の簡略化

#### リソースタイプの短縮名

kubectlは多くのリソースタイプに対して短縮名（エイリアス）を提供。`kubectl api-resources` で一覧確認可能。

```bash
# リソース短縮名の一覧表示
kubectl api-resources
```

##### 主要な短縮名

| リソースタイプ | 短縮名 | 例 |
|---------------|--------|-----|
| pods | po | `kubectl get po` |
| services | svc | `kubectl get svc` |
| deployments | deploy | `kubectl get deploy` |
| replicasets | rs | `kubectl get rs` |
| statefulsets | sts | `kubectl get sts` |
| daemonsets | ds | `kubectl get ds` |
| namespaces | ns | `kubectl get ns` |
| nodes | no | `kubectl get no` |
| persistentvolumes | pv | `kubectl get pv` |
| persistentvolumeclaims | pvc | `kubectl get pvc` |
| configmaps | cm | `kubectl get cm` |
| secrets | - | `kubectl get secrets` |
| ingresses | ing | `kubectl get ing` |
| horizontalpodautoscalers | hpa | `kubectl get hpa` |
| cronjobs | cj | `kubectl get cj` |
| customresourcedefinitions | crd, crds | `kubectl get crd` |

#### 使用方法

- **大文字小文字を区別しない**
- **単数形、複数形、短縮形のいずれも使用可能**

```bash
# すべて同じ結果
kubectl get pod mypod
kubectl get pods mypod
kubectl get po mypod
```

### コンテキストとネームスペース

#### コンテキストとは

コンテキスト = クラスター + ユーザー + ネームスペース の組み合わせ

```yaml
# kubeconfigの構造
apiVersion: v1
kind: Config
current-context: dev-cluster
clusters:
  - name: dev-cluster
    cluster:
      server: https://dev.k8s.example.com
      certificate-authority: /path/to/ca.crt
  - name: prod-cluster
    cluster:
      server: https://prod.k8s.example.com
      certificate-authority: /path/to/ca.crt
users:
  - name: dev-user
    user:
      client-certificate: /path/to/client.crt
      client-key: /path/to/client.key
  - name: prod-user
    user:
      client-certificate: /path/to/client.crt
      client-key: /path/to/client.key
contexts:
  - name: dev-context
    context:
      cluster: dev-cluster
      user: dev-user
      namespace: development
  - name: prod-context
    context:
      cluster: prod-cluster
      user: prod-user
      namespace: production
```

#### コンテキストの操作

```bash
# コンテキスト一覧表示
kubectl config get-contexts

# 現在のコンテキスト表示
kubectl config current-context

# コンテキスト切り替え
kubectl config use-context prod-context

# コンテキスト作成/更新
kubectl config set-context my-context \
  --cluster=my-cluster \
  --user=my-user \
  --namespace=my-namespace
```

#### ネームスペースの操作

```bash
# 現在のコンテキストのデフォルトネームスペースを変更
kubectl config set-context --current --namespace=my-namespace

# 現在のネームスペース確認
kubectl config view --minify | grep namespace:

# コマンド単位でネームスペース指定（デフォルトをオーバーライド）
kubectl get pods --namespace=kube-system
kubectl get pods -n kube-system
```

#### すべてのネームスペースを対象

```bash
# すべてのネームスペースのPodを一覧表示
kubectl get pods --all-namespaces
kubectl get pods -A  # 短縮形
```

### 設定ファイル: kubeconfig

#### デフォルト場所

- `~/.kube/config`
- 環境変数 `KUBECONFIG` で変更可能

```bash
# 複数のkubeconfigを統合
export KUBECONFIG=~/.kube/config:~/.kube/config-dev:~/.kube/config-prod
```

#### kubeconfigの主要セクション

1. **clusters**: 接続先クラスターの情報（サーバーアドレス、CA証明書）
2. **users**: 認証情報（証明書、トークン、認証プロバイダー）
3. **contexts**: クラスターとユーザーとネームスペースの組み合わせ
4. **current-context**: 現在アクティブなコンテキスト名

#### kubeconfig表示

```bash
# 現在のkubeconfigを表示
kubectl config view

# 現在のコンテキストのみ表示
kubectl config view --minify

# パスワードなどのセンシティブ情報も表示
kubectl config view --raw
```

### リソース指定のパターン

```bash
# リソースタイプのみ（すべてのリソース）
kubectl get pods

# リソース名指定
kubectl get pod mypod

# 複数リソース指定
kubectl get pod mypod1 mypod2

# ラベルセレクタ
kubectl get pods -l app=nginx
kubectl get pods --selector=app=nginx,env=prod

# フィールドセレクタ
kubectl get pods --field-selector=status.phase=Running

# 出力形式指定
kubectl get pods -o wide
kubectl get pods -o yaml
kubectl get pods -o json
kubectl get pods -o jsonpath='{.items[*].metadata.name}'
```

### ベストプラクティス

1. **短縮名の活用**: タイピング量を削減し、作業効率を向上
2. **コンテキスト管理**:
   - 本番とdev/staging環境で異なるコンテキストを使用
   - 誤操作防止のため、コンテキスト名を明確に
3. **ネームスペース分離**: アプリケーション、チーム、環境ごとにネームスペースを分ける
4. **エイリアス設定**:
   ```bash
   alias k=kubectl
   alias kgp='kubectl get pods'
   alias kgs='kubectl get svc'
   alias kgd='kubectl get deploy'
   ```
5. **kubectxとkubens**: コンテキストとネームスペースの切り替えを簡単にするツール

---

## 横断的な設計パターン比較

### 1. リモートリソース指定方法

| ツール | 形式 | 例 | 省略時のデフォルト |
|--------|------|-----|-------------------|
| gh | `OWNER/REPO` | `octo-org/octo-repo` | 認証ユーザーのリポジトリ、またはgit remote検出 |
| cargo | `--manifest-path` | `--manifest-path /path/to/Cargo.toml` | カレントディレクトリから親へ再帰探索 |
| npm/yarn | N/A（ローカル） | N/A | カレントディレクトリから親へ再帰探索 |
| docker | `[registry/][namespace/]repository[:tag]` | `docker.io/library/nginx:latest` | `docker.io/library/<name>:latest` |
| kubectl | リソースタイプ + 名前 | `pod/mypod`, `deploy/myapp` | 現在のコンテキスト・ネームスペース |

### 2. パス指定の簡略化手法

| ツール | 手法 | 例 |
|--------|------|-----|
| gh | ショートハンド（OWNER省略） | `myrepo` instead of `myuser/myrepo` |
| cargo | 自動検出（親ディレクトリ探索） | カレントディレクトリから `Cargo.toml` を探索 |
| npm/yarn | 自動検出（親ディレクトリ探索） | カレントディレクトリから `package.json` を探索 |
| docker | デフォルト値（registry, namespace, tag） | `nginx` → `docker.io/library/nginx:latest` |
| kubectl | 短縮名、エイリアス | `po` → `pods`, `deploy` → `deployments` |

### 3. 設定ファイルの場所と形式

| ツール | 設定ファイル | デフォルト場所 | 階層構造 |
|--------|-------------|---------------|----------|
| gh | config.yml | `~/.config/gh/config.yml` | なし |
| cargo | config.toml | `.cargo/config.toml`, `$CARGO_HOME/config.toml` | あり（プロジェクト → グローバル） |
| npm | .npmrc, package.json | `.npmrc`, `~/.npmrc`, package.json | あり（プロジェクト → ユーザー → グローバル） |
| yarn | .yarnrc.yml, package.json | `.yarnrc.yml`, package.json | あり（プロジェクト → ユーザー） |
| docker | Dockerfile, docker-compose.yml | `./Dockerfile`, `./docker-compose.yml` | なし |
| kubectl | kubeconfig | `~/.kube/config` | 複数ファイル統合可（`KUBECONFIG`環境変数） |

### 4. よく使うパターン

| ツール | パターン | 例 |
|--------|---------|-----|
| gh | リモートオーバーライド | `gh pr list -R owner/repo` |
| cargo | ワークスペースメンバー指定 | `cargo build -p crate-name` |
| npm | ワークスペーススクリプト実行 | `npm run test -w package-a` |
| yarn | ワークスペースプロトコル | `"dependencies": { "pkg": "workspace:*" }` |
| docker | マルチタグビルド | `docker build -t app:1.0 -t app:latest .` |
| kubectl | ラベルセレクタ | `kubectl get pods -l app=nginx` |

### 5. デフォルト値の扱い方

#### 共通の優先順位パターン

ほとんどのツールで以下の優先順位が採用されている:

1. **コマンドラインフラグ**（最優先）
2. **環境変数**
3. **プロジェクトローカル設定ファイル**
4. **ユーザーグローバル設定ファイル**
5. **システムグローバル設定**
6. **組み込みデフォルト値**（最低優先）

#### 具体例

**Cargo:**
1. 環境変数（`CARGO_BUILD_JOBS`）
2. `--config KEY=VALUE` フラグ
3. `.cargo/config.toml`（プロジェクト）
4. `$CARGO_HOME/config.toml`（グローバル）
5. デフォルト値

**npm:**
1. コマンドラインフラグ（`--registry`）
2. 環境変数（`NPM_CONFIG_REGISTRY`）
3. `.npmrc`（プロジェクト）
4. `~/.npmrc`（ユーザー）
5. グローバル設定
6. デフォルト値

**kubectl:**
1. コマンドラインフラグ（`--namespace`, `--context`）
2. 環境変数（`KUBECONFIG`）
3. `~/.kube/config` の `current-context` と各コンテキストの `namespace`
4. デフォルトネームスペース（`default`）

---

## 設計パターンのベストプラクティス

### 1. 自動検出とデフォルト値の活用

- **カレントディレクトリからの探索**: Cargo, npm/yarnのように、カレントディレクトリから親ディレクトリへ再帰的に設定ファイルを探索
- **git remoteからの推論**: GitHub CLIのように、既存のgit設定から推論
- **合理的なデフォルト**: Dockerの `latest` タグ、kubectlの `default` ネームスペース

### 2. 階層的な設定管理

- **プロジェクト設定 > ユーザー設定 > システム設定**: より具体的な設定が優先
- **環境変数による一時的なオーバーライド**: スクリプトやCI/CD環境で有用
- **コマンドラインフラグによる単発オーバーライド**: 最も高い優先度

### 3. ショートハンドとエイリアス

- **短縮形の提供**: kubectlのリソース短縮名（`po`, `svc`, `deploy`）
- **OWNER省略**: GitHub CLIの自分のリポジトリに対するOWNER省略
- **デフォルト値の補完**: Dockerのregistry, namespace, tag補完

### 4. ワークスペース/モノレポ対応

- **明示的なメンバー指定**: Cargoの `members` 配列、npmの `workspaces` 配列
- **共有リソース**: Cargoの `Cargo.lock` と `target/` 共有
- **個別実行**: npmの `-w` フラグ、Cargoの `-p` フラグ

### 5. コンテキスト管理

- **名前付きコンテキスト**: kubectlのコンテキスト、GitHub CLIのデフォルトリポジトリ
- **明示的な切り替え**: `kubectl config use-context`, `gh repo set-default`
- **コマンド単位のオーバーライド**: どのツールでもコマンドラインフラグで一時的にオーバーライド可能

### 6. 環境変数の命名規則

- **プレフィックス統一**: `GH_*`, `CARGO_*`, `NPM_CONFIG_*`, `DOCKER_*`, `KUBECONFIG`
- **階層的な命名**: Cargoの `CARGO_BUILD_JOBS`（`build.jobs` に対応）
- **大文字+アンダースコア**: シェル環境変数の慣例に従う

---

## まとめ

主要CLIツールの設計パターンには以下の共通点がある:

1. **インテリジェントなデフォルト**: 明示的な指定がない場合、合理的なデフォルト値やカレントディレクトリからの自動検出を行う
2. **階層的な設定管理**: プロジェクト → ユーザー → システム の順で設定を統合し、より具体的な設定が優先される
3. **柔軟なオーバーライド**: 環境変数やコマンドラインフラグで設定を一時的にオーバーライド可能
4. **ショートハンドの提供**: 頻繁に使うパターンに対して短縮形やエイリアスを提供
5. **モノレポ/ワークスペース対応**: 複数のプロジェクトを統合管理する仕組み
6. **明確な優先順位**: コマンドラインフラグ > 環境変数 > 設定ファイル > デフォルト値

これらのパターンを採用することで、以下の利点が得られる:

- **使いやすさ**: 多くの場合、引数なしで動作し、必要に応じて詳細を指定できる
- **柔軟性**: ローカル開発、CI/CD、本番環境など、異なる環境で柔軟に動作
- **一貫性**: 類似のツールと同様のパターンを採用することで、学習コストを削減
- **拡張性**: プロジェクトの成長に応じて、ワークスペースやコンテキスト管理で対応可能

新しいCLIツールを設計する際は、これらのパターンを参考に、ユーザーフレンドリーで拡張可能な設計を目指すべきである。
