<div align="center">

<h1>Codex Profiles</h1>

<p><a href="README.md">English</a> · <a href="README.ja.md">日本語</a> · 简体中文</p>

<p>
用多个 ChatGPT 账号运行 Codex CLI。<br>
同时使用不同账号，并集中查看各账号的剩余额度。<br>
会话历史自动与 Codex Desktop 共享，也可以换一个账号继续之前的会话。
</p>

<p>
  <a href="https://github.com/kazuki-kanaya/codex-profiles/actions/workflows/ci.yaml"><img src="https://github.com/kazuki-kanaya/codex-profiles/actions/workflows/ci.yaml/badge.svg?branch=main" alt="CI"></a>
  <a href="https://github.com/kazuki-kanaya/codex-profiles/releases"><img src="https://img.shields.io/github/v/release/kazuki-kanaya/codex-profiles?label=release" alt="最新版本"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT 许可证"></a>
</p>

<p>
  <a href="#安装">安装</a> ·
  <a href="#使用方法">使用方法</a> ·
  <a href="#工作原理与存储位置">工作原理与存储位置</a>
</p>

</div>

![查看四个账号的剩余额度，并使用 main 账号启动 Codex](assets/usage.png)

## 安装

请先安装 Codex CLI，并确保可以在终端中运行 `codex`。
提供适用于 macOS 和 Linux 的 arm64／amd64 二进制文件。

```sh
curl -fsSL https://raw.githubusercontent.com/kazuki-kanaya/codex-profiles/main/install.sh | sh
```

默认安装到 `~/.local/bin`。如果该目录尚未加入 PATH，安装脚本会为 zsh 写入 `~/.zshrc`，为 bash 写入 `~/.bashrc`。安装后，执行提示中的 `source` 命令，或打开一个新终端。

更新时再次运行同一条安装命令即可。也可以通过 `INSTALL_DIR` 指定安装位置。

<details>
<summary>通过 Go 或 Homebrew 安装</summary>

Go 1.27.3 及以上版本：

```sh
go install github.com/kazuki-kanaya/codex-profiles@latest
```

请将 `GOBIN` 加入 PATH；如果未设置，则使用 `$(go env GOPATH)/bin`。

Homebrew：

```sh
brew install --cask kazuki-kanaya/tap/codex-profiles
```

</details>

> macOS 版本尚未经过 Apple 公证。如果系统阻止启动，请确认下载来源可信，然后前往“系统设置 → 隐私与安全性 → 仍要打开”允许运行。

## 使用方法

为每个账号创建一个具名配置。执行 `add` 后会启动 Codex 的登录流程。

```sh
codex-profiles add main
codex-profiles add work
codex-profiles list
```

指定配置名称即可使用对应账号启动 Codex。在不同终端中运行，就能同时使用多个账号。

```sh
codex-profiles main    # 终端 1
codex-profiles work    # 终端 2
```

配置名称后面的参数会原样传给 Codex。

```sh
codex-profiles work resume <SESSION_ID>
```

如果想用更短的命令，可以在 shell 配置中添加别名：

```sh
alias cdx='codex-profiles'
```

### 命令一览

| 命令 | 说明 |
| --- | --- |
| `add NAME` | 创建配置并登录 |
| `list` | 查看账号及剩余额度 |
| `NAME [args...]` | 使用指定账号运行 Codex |
| `rename OLD NEW` | 重命名配置 |
| `remove NAME` | 删除配置 |
| `version` / `--version` | 显示版本 |
| `help` | 显示帮助 |

`list` 中的额度重置时间使用本地时区。无法获取的数值显示为 `-`。

配置名称必须以小写英文字母开头，后面可以使用小写英文字母、数字、`-` 和 `_`。不能与命令名称相同。

## 工作原理与存储位置

添加账号时，工具会创建独立的 `CODEX_HOME`，并建立指向共享历史目录的符号链接。无需额外设置即可共享会话历史。认证和令牌刷新由 Codex 负责，不会修改 Desktop 的认证信息或设置。

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

只有 `sessions/` 和 `archived_sessions/` 会共享。SQLite 数据库和后台守护进程的状态按账号分别保存。

不要让多个 Codex 进程同时更新同一个会话。各客户端的会话列表可能不完全一致，请通过会话 ID 继续会话。`/fork`、`/side` 及历史格式的兼容性取决于 Codex 版本。

<details>
<summary>自定义存储位置</summary>

通常无需配置。如需更改，可以设置以下环境变量。

| 环境变量 | 默认值 |
| --- | --- |
| `CODEX_PROFILES_HOME` | `~/.codex-profiles` |
| `CODEX_PROFILES_SESSIONS_DIR` | `~/.codex/sessions` |
| `CODEX_PROFILES_ARCHIVED_SESSIONS_DIR` | `~/.codex/archived_sessions` |

修改这些设置不会迁移已有数据或更新链接。配置目录与共享历史目录不能互相嵌套。请勿设置 `CODEX_SQLITE_HOME`，以保持各账号的 SQLite 数据库相互独立。

</details>

## 重命名与删除

操作前，请先关闭使用该账号的 Codex 进程。确认后，工具会自动停止对应的后台守护进程；如果停止失败，则中止操作。共享的会话历史文件不会被删除。

`rename` 只会保留 `auth.json` 和 `config.toml`。SQLite 数据库、日志等运行时文件会在需要时重新生成。其他保存在配置目录内的文件请提前备份。如果 `archived_sessions/` 是普通目录而非符号链接，请先将其中的历史迁移到共享目录。

## 许可证

[MIT](LICENSE)
