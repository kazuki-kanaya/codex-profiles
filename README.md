# codex-profiles

複数のChatGPTアカウントでCodex CLIを同時に動かしたくて作りました。

アカウントごとに名前をつけて登録しておくと、`codex-profiles main` や `codex-profiles work` で起動できます。それぞれ別の認証情報を使うので、別々のターミナルで同時に起動できます。

`list` で各アカウントの利用枠を確認できます。セッション履歴はDesktopと共有していて、別のアカウントから続きを再開することもできます。

## インストール

Go 1.27.3以降が必要です。Codex CLIも先にインストールして、`codex` コマンドが使える状態にしておいてください。

```sh
go install github.com/kazuki-kanaya/codex-profiles@latest
```

インストール先は `GOBIN`、未設定なら `$(go env GOPATH)/bin` です。PATHが通っていなければ追加してください。

コマンド名が長いので、好みでエイリアスをつけて使ってください。

```sh
alias cdx='codex-profiles'
```

## 使い方

まずアカウントを登録します。`add` を実行するとCodexのログイン画面が開きます。

```sh
codex-profiles add main
codex-profiles add work
codex-profiles list
```

登録した名前を指定すると、そのアカウントでCodexが起動します。以下を別々のターミナルで実行すれば、同時に使えます。

```sh
codex-profiles main    # ターミナル1
codex-profiles work    # ターミナル2
```

名前の後ろには、いつものCodexの引数をそのまま渡せます。

```sh
codex-profiles work resume <SESSION_ID>
```

| コマンド | 説明 |
| --- | --- |
| `add NAME` | プロファイルを作成してログイン |
| `list` | プロファイルと利用枠を表示 |
| `NAME [args...]` | 指定したプロファイルでCodexを実行 |
| `rename OLD NEW` | プロファイルの名前を変更 |
| `remove NAME` | プロファイルを削除 |

名前は小文字の英字で始め、以降は小文字の英字・数字・`-`・`_` が使えます。コマンド名は使えません。

## 仕組み

アカウントごとに `CODEX_HOME` を分けてCodexを起動しています。認証やトークンの更新はCodexに任せています。

```text
~/.codex-profiles/main/
├── auth.json
├── config.toml
├── sessions/ -> ~/.codex/sessions/
└── archived_sessions/ -> ~/.codex/archived_sessions/
```

`sessions` と `archived_sessions` はDesktop側のディレクトリへのシンボリックリンクです。認証情報・設定・SQLite・daemonはアカウントごとに持ちます。Desktopの認証情報や設定は変更しません。

<details>
<summary>保存先の変更</summary>

| 環境変数 | デフォルト |
| --- | --- |
| `CODEX_PROFILES_HOME` | `~/.codex-profiles` |
| `CODEX_PROFILES_SESSIONS_DIR` | `~/.codex/sessions` |
| `CODEX_PROFILES_ARCHIVED_SESSIONS_DIR` | `~/.codex/archived_sessions` |

環境変数を変えても、既存データやリンクは移動しません。プロファイルの中に共有履歴を置いたり、その逆の配置にしたりしないでください。`CODEX_SQLITE_HOME` を設定するとSQLiteの保存先も変わるため、設定せずに使ってください。

</details>

## 注意点

- 同じセッションを複数のCodexから同時に更新しないでください。一覧に履歴が出ないこともあるので、再開するときはセッションIDを指定してください。Codexのバージョンによっては履歴共有がうまく動かない場合があります。
- `rename`・`remove` の前には、そのアカウントで動かしているCodexを終了してください。daemonは確認後に自動停止します。共有している履歴は削除しません。
- `rename` では `auth.json` と `config.toml` だけを引き継ぎます。それ以外に保存したファイルがあれば、先に退避してください。`archived_sessions` が共有先へのリンクではなく通常のディレクトリになっている場合も、先に履歴の移行が必要です。

## ライセンス

[MIT](LICENSE)
