# gm — Go 工具链管理器

类 [uv](https://github.com/astral-sh/uv) 体验的 Go 项目管理 CLI：多版本 SDK、模块依赖、全局工具。

## 安装

### 一键安装（推荐）

**macOS / Linux：**

```bash
curl -fsSL https://raw.githubusercontent.com/angei24/go-manager/main/scripts/install.sh | bash
```

或在仓库目录内：

```bash
./scripts/install.sh
```

**Windows（PowerShell）：**

```powershell
irm https://raw.githubusercontent.com/angei24/go-manager/main/scripts/install.ps1 | iex
```

或 CMD：

```cmd
scripts\install.bat
```

默认安装到 `~/.local/bin`（Windows 为 `%USERPROFILE%\.local\bin`）。若该目录不在 PATH 中，脚本会提示如何配置。

### 安装选项

| 选项 | 说明 |
|------|------|
| `--from-source` / `-FromSource` | 从源码编译（需 Go 1.21+ 与 git） |
| `--dir PATH` / `-InstallDir PATH` | 自定义安装目录 |
| `--repo OWNER/REPO` | GitHub 仓库（默认 `angei24/go-manager`） |
| `--version TAG` | 指定 Release 标签；无 Release 时自动回退到源码编译 |
| `-AddToPath`（仅 Windows） | 自动加入当前用户 PATH |

示例：

```bash
# 安装到 ~/bin
GM_INSTALL_DIR=$HOME/bin ./scripts/install.sh

# 强制从当前仓库源码编译
./scripts/install.sh --from-source
```

### 手动编译

```bash
git clone https://github.com/angei24/go-manager.git
cd go-manager
go build ./cmd/gm/.
```

> **说明：** 若 GitHub 尚未发布 Release，安装脚本会自动克隆仓库并用本地 Go 编译。发布 Release 后，脚本会优先下载预编译二进制（命名格式：`gm_<tag>_<os>_<arch>.tar.gz` / `.zip`）。

## 快速开始

```bash
# 安装并切换 Go 版本
gm go install 1.22.5
gm go use 1.22.5 --global

# 新建项目
gm init myapp --module example.com/myapp
cd myapp

# 依赖管理
gm add github.com/spf13/cobra@latest
gm sync
gm remove github.com/spf13/cobra

# 全局工具（类似 pipx / uv tool）
gm tool install golang.org/x/tools/gopls@latest
export PATH="$HOME/.local/share/gm/bin:$PATH"
gm tool list
```

## 命令

| 命令 | 说明 |
|------|------|
| `gm init [dir]` | 创建项目（git、go.mod、main.go、README、.gm-version） |
| `gm go list` | 已安装版本 + 当前支持安装的最新两个 stable minor（如 1.25.x / 1.26.x 的最新 patch） |
| `gm go install <ver>` | 下载安装 Go SDK（仅限最近两个 stable minor 内的版本，不含 rc/beta） |
| `gm go use <ver> [--global]` | 项目 `.gm-version` 或全局默认版本 |
| `gm go uninstall <ver>` | 卸载已安装版本 |
| `gm add <pkg>` | `go get` 添加依赖 |
| `gm remove <pkg>` | `go mod edit -droprequire` |
| `gm sync [--check]` | `go mod tidy` + `go mod download` |
| `gm tool list/install/uninstall` | 管理 `~/.local/share/gm/bin` 下的工具 |

## 配置与目录

| 路径 | 用途 |
|------|------|
| `~/.local/share/gm/versions/` | 已安装的 Go SDK（GOROOT） |
| `~/.local/share/gm/bin/` | 全局工具 GOBIN |
| `~/.config/gm/config.toml` | 全局默认 Go 版本 |
| `.gm-version` | 项目锁定 Go 版本 |

版本解析优先级：**`.gm-version` > `GM_GO_VERSION` > 全局 config > 系统 PATH 中的 `go`**。

## 环境变量

| 变量 | 说明 |
|------|------|
| `GM_GO_VERSION` | 覆盖当前 Go 版本 |
| `GM_DATA_DIR` | 数据目录（默认 `~/.local/share/gm`） |
| `GM_CONFIG_DIR` | 配置目录 |
| `GM_GO_DOWNLOAD_BASE` | 下载镜像前缀（如 `https://mirrors.example/go/`） |
| `GM_GO_DOWNLOAD_API` | 版本列表 API URL |

## 验证

```bash
./scripts/verify.sh
```

## 许可

MIT
