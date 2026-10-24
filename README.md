<div align="center">

<h1>codex-profiles</h1>

<p>
  <a href="https://github.com/kazuki-kanaya/codex-profiles/actions/workflows/ci.yaml"><img src="https://github.com/kazuki-kanaya/codex-profiles/actions/workflows/ci.yaml/badge.svg?branch=main" alt="CI"></a>
  <a href="https://github.com/kazuki-kanaya/codex-profiles/releases"><img src="https://img.shields.io/github/v/release/kazuki-kanaya/codex-profiles?label=release" alt="Latest release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License"></a>
</p>

<p>
  <a href="#インストール">インストール</a> ·
  <a href="#使い方">使い方</a> ·
  <a href="#履歴の共有">履歴の共有</a> ·
  <a href="#設定">設定</a>
</p>

</div>

アカウントごとに名前をつけて登録し、その名前でCodexを起動します。認証情報・設定・daemonを個別に持つので、別々のターミナルで並行して使えます。

```sh
codex-profiles main    # ターミナル1
codex-profiles work    # ターミナル2
```

利用枠は `list` でまとめて確認できます。履歴はCodex Desktopと共有し、別のアカウントから続きを再開することもできます。

## インストール

先にCodex CLIをインストールし、`codex` コマンドが使える状態にしてください。
macOS・Linuxのarm64／amd64向けに配布しています。

```sh
curl -fsSL https://raw.githubusercontent.com/kazuki-kanaya/codex-profiles/main/install.sh | sh
```

インストール先は `~/.local/bin` です。PATHが通っていなければ、シェルの設定ファイルに追加してください。

```sh
export PATH="$HOME/.local/bin:$PATH"
```

更新も同じインストールコマンドで行えます。別の場所に入れる場合は、`INSTALL_DIR` で指定できます。

<details>
<summary>Go・Homebrewからインストールする</summary>

Go 1.27.3以降：

```sh
go install github.com/kazuki-kanaya/codex-profiles@latest
```

`GOBIN`、未設定なら `$(go env GOPATH)/bin` にPATHを通してください。

Homebrew：

```sh
brew install --cask kazuki-kanaya/tap/codex-profiles
```

</details>

> Mac版はAppleの公証を受けていません。起動をブロックされた場合は、入手元を確認したうえで「システム設定 → プライバシーとセキュリティ → このまま開く」から許可してください。

## 使い方

### アカウントを登録する

`add` を実行するとCodexのログインが始まります。使うアカウントごとに登録してください。

```sh
codex-profiles add main
codex-profiles add work
codex-profiles list
```

`list` にはメールアドレス、プラン、5時間・週間の残量、リセット時刻、リセットチケットの枚数を表示します。時刻はローカル時刻で、取得できない値は `-` になります。

### Codexを起動する

登録した名前を指定します。名前の後ろには、いつものCodexの引数を渡せます。

```sh
codex-profiles main
codex-profiles work resume <SESSION_ID>
```

短い名前で呼びたい場合は、エイリアスを設定してください。

```sh
alias cdx='codex-profiles'
```

### コマンド一覧

| コマンド | 説明 |
| --- | --- |
| `add NAME` | 登録してログイン |
| `list` | アカウントと利用枠を表示 |
| `NAME [args...]` | 指定したアカウントでCodexを起動 |
| `rename OLD NEW` | 名前を変更 |
| `remove NAME` | 削除 |
| `version` / `--version` | バージョンを表示 |
| `help` | ヘルプを表示 |

名前は小文字の英字で始め、以降は小文字の英字・数字・`-`・`_` が使えます。コマンド名は使えません。

## 履歴の共有

アカウントごとに `CODEX_HOME` を分け、認証とトークン更新はCodexに任せています。Desktopの認証情報や設定は変更しません。

```text
~/.codex-profiles/
├── main/
│   ├── auth.json
│   ├── config.toml
│   ├── sessions/ -> ~/.codex/sessions/
│   └── archived_sessions/ -> ~/.codex/archived_sessions/
└── work/
    └── ...
```

共有するのは `sessions/` と `archived_sessions/` だけです。SQLiteやdaemonの状態はアカウントごとに保存します。

同じセッションを複数のCodexから同時に更新しないでください。履歴一覧は完全には同期されないため、再開時はセッションIDを指定してください。`/fork`・`/side` や履歴形式の互換性は、Codexのバージョンに依存します。

## 設定

通常は設定不要です。保存先を変えたい場合は、以下の環境変数で指定できます。

| 環境変数 | デフォルト |
| --- | --- |
| `CODEX_PROFILES_HOME` | `~/.codex-profiles` |
| `CODEX_PROFILES_SESSIONS_DIR` | `~/.codex/sessions` |
| `CODEX_PROFILES_ARCHIVED_SESSIONS_DIR` | `~/.codex/archived_sessions` |

設定を変えても既存データやリンクは移動しません。プロファイルと共有履歴の保存先は、互いの配下に置かないでください。SQLiteを分離するため、`CODEX_SQLITE_HOME` は設定せずに使ってください。

## 名前変更・削除について

対象のアカウントで動かしているCodexを終了してから実行してください。確認後にdaemonを自動停止し、停止できなければ処理を中断します。共有履歴のリンク先は削除しません。

`rename` で引き継ぐのは `auth.json` と `config.toml` だけです。SQLiteやログなどは作り直します。その他の個別ファイルは事前に退避してください。`archived_sessions/` が通常のディレクトリになっている場合は、先に共有先への移行が必要です。

## ライセンス

[MIT](LICENSE)
