# Todo Desktop

使用 MyGo + Lobe MyGo 构建的本地原生桌面任务程序。界面由 Go 在 GPU 上绘制，无 WebView、JavaScript、HTTP 服务或前端构建步骤。

## 功能

- 概览：全部、待完成和已完成任务统计。
- 任务：添加、标题搜索、状态筛选、切换完成状态、删除。
- SQLite 本地持久化；标题去除首尾空白后为 1–200 个 Unicode 字符。
- 浅色、深色、跟随系统，外观设置本地保存。
- 保存操作在后台执行，错误在界面显示。

## 环境要求

- Go 1.27.3+，Make（可选）。不需要 Bun、Node.js 或 CGO。
- Linux 原生运行需要图形会话、GTK 3 和系统图形库；不是 WebKitGTK。各平台要求以 [MyGo 文档](https://github.com/egoist/mygo/tree/main/docs) 为准。
- VS Code 调试需要 Go 扩展与 Delve。

MyGo、Lobe MyGo 和原生 CLI 的版本固定在 [go.mod](./go.mod)。Lobe MyGo 当前需要 MyGo 的 0.8.4 预发布提交，不能单独降回原模板的 0.8.3。

## 快速开始

```bash
make install
make dev
```

也可以直接运行：

```bash
go run .
```

构建本机桌面程序：

```bash
make build
```

Linux 产物位于 `build/linux-amd64/`（架构随本机变化），包含可执行文件、桌面入口、Debian 包等。应用标识、版本和输出目录见 [mygo.json](./mygo.json)。

## 常用命令

| 命令 | 用途 |
|---|---|
| `make install` | 下载 Go 依赖 |
| `make dev` | MyGo 原生开发模式 |
| `make build` | 构建本机桌面应用与分发产物 |
| `make test` | 执行 Go 测试命令；当前未包含测试文件 |
| `make check` | Go vet、MyGo UI 生命周期检查、Go 编译验证 |
| `make clean` | 清理构建产物，不删除用户数据 |

可使用 `make GO=/path/to/go build` 指定工具路径。

## 数据与备份

默认目录为 `os.UserConfigDir()/go-desktop-template`：Linux 通常为 `~/.config/go-desktop-template`，macOS 为 `~/Library/Application Support/go-desktop-template`，Windows 为 `%AppData%/go-desktop-template`。

- `todos.db`：任务数据。
- `preferences.json`：外观偏好。
- 可通过 `TODO_DATA_DIR` 覆盖目录；启动时自动创建。

```bash
TODO_DATA_DIR=/tmp/todo-desktop-demo go run .
```

关闭程序后复制数据目录即可备份；恢复时保持程序关闭。程序不会把数据库放进安装目录，也不需要网络连接。

## 目录结构

```text
main.go                 启动入口
internal/app/           依赖组装、初始化及资源释放
internal/desktop/       原生窗口与关闭事件
internal/view/          Lobe MyGo 页面、界面状态及外观设置
internal/service/       业务逻辑、数据访问接口
internal/repository/    参数化 SQLite CRUD
internal/model/         任务模型
internal/db/            SQLite 连接与自动建表
internal/config/        数据目录配置
resources/              桌面图标
.agents/skills/          原生开发与组件文档技能
.vscode/launch.json     原生 Go 调试配置
docker/                 预留构建环境目录，当前不依赖 Docker
deploy/test/            预留测试渠道分发目录
deploy/prod/            预留正式渠道分发目录
```

界面直接调用 service，不包含 HTTP controller 或 IPC 绑定层。后台任务使用可取消 context，结果通过 `Window.Update` 在 UI 线程发布；退出时等待后台任务结束后关闭数据库。
