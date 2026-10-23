# codex-profiles

Codex CLIをアカウントごとに独立した `CODEX_HOME` で起動する、小さなランチャーです。

認証情報・設定・SQLite・daemonはプロファイルごとに分離し、`sessions/` と `archived_sessions/` を共有します。認証とトークン更新はCodex本体に任せます。

## インストール

前提条件：

- Go 1.27.3以降
- `codex` コマンドがPATHにあること
- シンボリックリンクを作成できること

macOS・Codex CLI 0.164.0で動作確認しています。他のOS・バージョンとの互換性は未検証です。

```bash
go install github.com/kazuki-kanaya/codex-profiles@latest
```

インストール先の `$(go env GOPATH)/bin` をPATHに追加してください。`GOBIN` を設定している場合は、そのディレクトリを使います。

短い名前で使いたい場合は、シェルの設定にエイリアスを追加できます。

```bash
alias cdx='codex-profiles'
```

## 使い方

```bash
# プロファイルを作成し、Codexでログイン
codex-profiles add main
codex-profiles add sub

# プロファイルと利用枠を表示
codex-profiles list

# 指定したプロファイルで起動
codex-profiles main

# 追加引数はCodexにそのまま転送
codex-profiles sub resume <SESSION_ID>
codex-profiles main --no-daemon
codex-profiles main app-server daemon restart

# 確認後に名前変更・削除
codex-profiles rename main work
codex-profiles remove sub

codex-profiles help
```

プロファイル名は小文字の英字で始め、以降は小文字の英字・数字・`-`・`_` を使えます。管理コマンドと同じ名前は使えません。

`list` はメールアドレス、プラン、5時間・週間の残り利用枠、リセット時刻、リセットチケット枚数を表示します。リセット時刻はローカル時刻です。取得できない値は `-` と表示します。Codex app-serverへの問い合わせは最大4件並列で、1件の失敗後も他のプロファイルを表示します。

`add` がログインを含む途中の処理で失敗した場合、作成したプロファイルを削除します。同じ名前で再試行できます。

## 保存先

```text
~/.codex/
├── sessions/                   # 共有する履歴の実体
└── archived_sessions/          # 共有するアーカイブの実体

~/.codex-profiles/
├── main/
│   ├── auth.json
│   ├── config.toml
│   ├── sessions/ -> ~/.codex/sessions/
│   └── archived_sessions/ -> ~/.codex/archived_sessions/
└── sub/
    └── ...
```

各プロファイルの `config.toml` には `cli_auth_credentials_store = "file"` を設定します。Desktopの認証情報や設定、既存の `~/.codex-cli/` は変更しません。共有履歴のディレクトリがなければ作成します。

保存先は環境変数で変更できます。

| 環境変数 | デフォルト |
| --- | --- |
| `CODEX_PROFILES_HOME` | `~/.codex-profiles` |
| `CODEX_PROFILES_SESSIONS_DIR` | `~/.codex/sessions` |
| `CODEX_PROFILES_ARCHIVED_SESSIONS_DIR` | `~/.codex/archived_sessions` |

3つの保存先はそれぞれ独立しています。環境変数を変更しても既存データやシンボリックリンクは移行しません。プロファイル保存先と共有履歴の保存先は、互いの配下に置かないでください。

子プロセスには `CODEX_HOME` を上書きして渡し、それ以外の環境変数は引き継ぎます。SQLiteの保存先をプロファイルごとに分離するため、`CODEX_SQLITE_HOME` は設定しないでください。

## 名前変更・削除

対象プロファイルのCodexを終了してから実行してください。確認で `y` を入力するとdaemonを自動停止します。停止に失敗した場合は、移行・削除を中断します。

- `remove` は対象プロファイルを削除します。共有履歴へのシンボリックリンクは削除しますが、リンク先の実体は削除しません。
- `rename` は新しいプロファイルを作成し、`auth.json` と `config.toml` をコピーして共有履歴へのリンクを作り直します。準備が完了したら旧プロファイルを削除します。
- 名前変更ではSQLite・ログ・daemonの状態などは引き継ぎません。認証情報と設定以外のプロファイル固有ファイルも引き継がないため、必要なものは事前に退避してください。
- 旧プロファイルの `archived_sessions/` が共有先へのリンクではない場合、名前変更を拒否します。内容を確認し、共有先へ移行してから実行してください。

## 履歴共有の制約

- 同じセッションを複数プロセスから同時に更新しないでください。
- SQLiteの索引は共有しないため、DesktopとCLIでセッション一覧が一致するとは限りません。ID指定の `resume` を基本とします。
- `/fork`・`/side`、Paginated History、圧縮済みrolloutの互換性はCodexのバージョンに依存します。共有した履歴を別アカウントで扱う動作をすべて保証するものではありません。

## 開発

```bash
go run . help
go fmt ./...
go vet ./...
go test -race ./...
go build .
```

remove・renameのテストは一時ディレクトリと仮の `codex` コマンドを使います。実際の認証情報やdaemonには触れません。

## ライセンス

[MIT](LICENSE)
