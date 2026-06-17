<div align="center">

# GinMiniServer

基于 [Gin](https://github.com/gin-gonic/gin) 的轻量级静态文件服务器，配备科技感 TUI 界面。

一个二进制文件，零配置，即刻启动。

[![Release](https://img.shields.io/github/v/tag/Moxin1044/GinMiniServer?label=发布&color=blue)](https://github.com/Moxin1044/GinMiniServer/releases)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/许可证-MIT-green)](LICENSE)

</div>

---

## 特性

- **单文件部署** — 无运行时依赖，下载即用
- **跨平台** — Windows / Linux / macOS，支持 amd64 / 386 / arm64 / arm
- **TUI 界面** — 基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) 的科技感终端界面
- **目录浏览** — 浏览器中直接浏览和下载文件
- **灵活配置** — 通过命令行参数指定端口、主机和目录
- **自动发布** — 推送标签即可自动构建并发布 Release

## 安装

从 [Releases](../../releases) 页面下载对应平台的二进制文件。

<details>
<summary>可用平台列表</summary>

| 系统 | 架构 | 文件名 |
|------|------|--------|
| Windows | amd64 | `GinMiniServer-windows-amd64.exe` |
| Windows | 386 | `GinMiniServer-windows-386.exe` |
| Windows | arm64 | `GinMiniServer-windows-arm64.exe` |
| Linux | amd64 | `GinMiniServer-linux-amd64` |
| Linux | 386 | `GinMiniServer-linux-386` |
| Linux | arm | `GinMiniServer-linux-arm7` |
| Linux | arm64 | `GinMiniServer-linux-arm64` |
| macOS | amd64 | `GinMiniServer-darwin-amd64` |
| macOS | arm64 | `GinMiniServer-darwin-arm64` |

</details>

## 使用方法

```bash
# 在当前目录启动服务，默认端口 8080
GinMiniServer

# 自定义端口和目录
GinMiniServer -port 3000 -dir /path/to/files

# 仅本机访问
GinMiniServer -host 127.0.0.1 -port 9000

# 设置访问密码（用户名固定为 admin）
GinMiniServer -password mysecret

# 启用日志记录到指定目录
GinMiniServer -log ./logs

# 组合使用
GinMiniServer -host 0.0.0.0 -port 8080 -dir /var/www -password mysecret -log /var/log/ginmini

# 查看版本
GinMiniServer -version
```

### TUI 界面预览

启动后会显示一个带圆角边框的科技感界面，包含：

- ASCII 艺术 Logo（紫色渐变）
- 服务器信息面板（版本、平台、时间、运行时长、监听地址、目录）
- 旋转动画状态指示器
- 实时刷新的运行时间


界面会实时刷新时间和运行时长，按 `Ctrl+C` 或 `q` 优雅退出。

### 网页文件浏览器

浏览器访问服务地址后，会看到一个现代化的深色主题文件浏览器：

- **深色主题** — GitHub 风格的暗色界面，护眼舒适
- **文件图标** — 根据文件类型自动显示对应 emoji 图标（🐹 Go、🐍 Python、📦 压缩包等）
- **面包屑导航** — 顶部显示当前路径，可点击快速跳转
- **实时搜索** — 右上角搜索框，输入即过滤文件列表
- **文件信息** — 显示文件大小和修改时间，目录优先排序
- **响应式布局** — 自动适配手机和桌面屏幕
- **密码保护** — 设置 `-password` 后需输入 Basic Auth 认证（用户名 `admin`）

## 命令行参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-port` | `8080` | HTTP 服务端口 |
| `-dir` | `.` | 要服务的文件目录 |
| `-host` | `0.0.0.0` | 绑定的主机地址 |
| `-password` | `（空）` | 访问密码，留空则无认证（用户名固定为 admin） |
| `-log` | `（空）` | 日志输出目录，留空则不记录日志 |
| `-version` | `false` | 显示版本信息并退出 |

### 日志格式

启用 `-log` 后，每天一个日志文件（如 `2026-06-18.log`），格式为：

```
2026-06-18 15:30:00 | 200 |      2.8966ms |             ::1 | GET     |       7925 | /
2026-06-18 15:30:01 | 200 |    142.5613ms |             ::1 | GET     |      20282 | /main.go
2026-06-18 15:30:02 | 404 |            0s |             ::1 | GET     |         53 | /nonexistent
```

字段依次为：时间 | 状态码 | 耗时 | 客户端IP | 方法 | 响应大小 | 路径

## 从源码构建

```bash
git clone https://github.com/Moxin1044/GinMiniServer.git
cd GinMiniServer
go mod tidy
go build -ldflags="-s -w" -o GinMiniServer .
```

### 交叉编译

```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o GinMiniServer .
```

## 自动发布

本项目使用 GitHub Actions 实现 CI/CD：

- **推送到 main** — 触发全平台构建验证
- **推送 `v*` 标签** — 构建所有平台并自动发布 GitHub Release

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 技术栈

| 组件 | 说明 |
|------|------|
| [Gin](https://github.com/gin-gonic/gin) | HTTP Web 框架 |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | 终端 TUI 框架 |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | 终端样式渲染 |

## 许可证

MIT
