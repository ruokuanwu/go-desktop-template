# Todo Desktop

使用 MyGo + Lobe MyGo 构建的本地原生桌面任务程序。界面由 Go 绘制，默认使用 CPU/内存渲染，无 WebView、JavaScript、HTTP 服务或前端构建步骤。

程序启动时，若未设置 `MYGO_GPU`，自动设为 `0`，以减少小程序的图形驱动内存开销；开发运行与打包产物均生效，从桌面图标启动也无需额外配置。显式设置的环境变量优先，例如 `MYGO_GPU=1 ./build/linux-amd64/todo-desktop` 可强制尝试 OpenGL 渲染。CPU 渲染可能增加滚动、动画和大窗口下的 CPU 开销。

## 功能

- 概览：全部、待完成和已完成任务统计。
- 任务：添加、标题搜索、状态筛选、切换完成状态、删除。
- SQLite 本地持久化；标题去除首尾空白后为 1–200 个 Unicode 字符。
- 浅色、深色、跟随系统，外观设置本地保存。
- 保存操作在后台执行；任务页通过固定宽度按钮显示保存状态，设置页就地更新保存提示，错误在各页固定高度区域显示，不插入临时状态行。

## 环境要求

- Go 1.27.3+，Make（可选）。不需要 Bun、Node.js 或 CGO。
- Linux 原生运行需要图形会话、GTK 3 和系统图形库；不是 WebKitGTK。各平台要求以 [MyGo 文档](https://github.com/egoist/mygo/tree/main/docs) 为准。
- VS Code 调试需要 Go 扩展与 Delve。

Linux 启动时会在创建窗口前追加进程内 fontconfig 规则，规避旧版 cairo 将灰度字形误渲染为子像素字形、导致文字忽略主题色的问题（如浅色模式文字发白、泛绿）。规则不修改系统配置，不影响 macOS 和 Windows。实现参考 [sshyard 的修复说明](https://github.com/Licoy/sshyard/commit/ddfb25d35ebfdda5ade2ff6eb587251837ee6f46)；若规则加载失败，启动日志会报告原因。

MyGo、Lobe MyGo 和原生 CLI 的版本固定在 [go.mod](./go.mod)。Lobe MyGo 当前需要 MyGo 的 0.8.4 预发布提交，不能单独降回原模板的 0.8.3。

## 快速开始

```bash
make install
make dev
```

也可以直接运行：

```bash
go run ./cmd
# 指定配置文件（文件必须存在）
go run ./cmd -p /path/to/config.json
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

- 配置路径由 Cobra 参数 `-p` / `--config-path` 指定，默认是上述目录下的 `config.json`。加载和保存始终使用同一路径；相对配置路径基于启动工作目录。
- 默认配置文件不存在时使用默认配置；显式指定的配置文件不存在时启动报错。不再读取旧的 `preferences.json`。
- `data_dir` 保留原字段名，但现在表示数据库文件路径：绝对路径直接使用；为空或非绝对路径时，统一使用 `os.UserConfigDir()/go-desktop-template/data.db`，不拼接相对路径。不读取 `TODO_DATA_DIR` 环境变量。
- 数据库父目录启动时自动创建。旧的 `todos.db` 不会自动迁移；如需保留任务，可在关闭程序后将其复制为 `data.db`，或将 `data_dir` 设置为旧数据库的绝对路径。旧配置中指向目录的 `data_dir` 也需要修改为数据库文件路径。

日志使用 Zap，输出 JSON 到日志目录的 `app.log`。启动层通过 `logger.New` 创建实例并使用 `zap.ReplaceGlobals` 注册为全局默认日志，其他包直接调用 `zap.L()`，无需逐层传递 logger。退出时先等待后台任务结束，再恢复原全局日志并统一刷新、关闭文件；业务代码不负责关闭日志。配置示例（空 `data_dir` 使用默认数据库路径；其余字段如下为默认值）：

```json
{
  "data_dir": "",
  "theme": "system",
  "log": {
    "dir": "logs",
    "max_age": 7,
    "level": "info"
  }
}
```

- `dir`：日志目录；相对路径基于数据库所在目录，也支持绝对路径。
- `max_age`：轮转日志保留天数，`0` 表示不按时间清理。单文件达到 100 MB 时轮转并压缩，过期备份在轮转时异步清理；当前日志不按时间删除。
- `level`：支持 `debug`、`info`、`warn`、`error`、`dpanic`、`panic`、`fatal`。设为 `debug` 可记录启动时的 `application starting` 日志；默认 `info` 会过滤该日志。
- 界面保存外观时保留日志配置；日志配置修改后重启生效。

关闭程序后复制数据目录即可备份；恢复时保持程序关闭。程序不会把数据库放进安装目录，也不需要网络连接。

## 目录结构

```text
cmd/main.go             Cobra 启动入口与配置路径参数
internal/app/           依赖组装、原生窗口创建及资源释放
internal/store/         任务和偏好设置的共享状态与动作
internal/pkg/asyncjobs/ 后台执行、结果发布及取消等待
internal/view/          根视图、主题应用及页面组合
internal/view/layouts/  侧栏与内容布局
internal/view/pages/    业务页面及输入、搜索、筛选等局部状态
internal/view/components/  任务行等展示组件
internal/service/       业务逻辑、数据访问接口
internal/repository/    参数化 SQLite CRUD
internal/model/         任务模型
internal/db/            SQLite 连接与自动建表
internal/config/        数据目录及应用配置加载、校验、保存
internal/pkg/logger/    Zap 文件日志、大小轮转与备份保留
resources/              桌面图标
.agents/skills/          原生开发与组件文档技能
.vscode/launch.json     原生 Go 调试配置
docker/                 预留构建环境目录，当前不依赖 Docker
deploy/test/            预留测试渠道分发目录
deploy/prod/            预留正式渠道分发目录
```

应用启动层创建并注入 Store 实例，业务状态不使用包级全局变量，也不设置 Controller 层；日志使用上述全局默认实例。业务页面直接使用构造时注入的 Store；根视图只负责主题、导航和页面组合，可复用组件通过参数和回调交互。任务 Store 调用 service，偏好设置 Store 调用配置保存，两者的忙碌状态和错误独立，保存外观不会查询任务。搜索、筛选和输入内容由任务页面持有，切换页面后保留。

后台任务使用可取消 context，结果通过启动层注入的 `Window.Update` 在 UI 线程发布；关闭窗口后忽略待发布结果，退出时等待后台任务结束后关闭数据库。
