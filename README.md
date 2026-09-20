# Summary Your Usage

最小用户 CRUD，参考 [go-clean-template](https://github.com/evrone/go-clean-template) 的 `controller → usecase → repository` 分层。

- 后端直接依赖只有 chi、goose、SQLite 和 MySQL 驱动；配置、日志、JSON、校验、SQL 使用 Go 标准库。
- 前端使用 Vue、Element Plus 和 Vite，源码在 `ui/`，使用 JavaScript。
- 页面只有用户列表、新增、编辑、删除。姓名与邮箱必填，邮箱唯一，删除前确认。

## 运行

需要 Go 1.26+、Node.js 20.19+（20.x）或 22.12+。

```sh
make ui-install
make build
./bin/app
```

访问 <http://127.0.0.1:8080>。服务读取当前目录的 `config.json`，自动创建 SQLite 数据库并运行 goose 迁移。

UI 构建产物通过 `go:embed` 嵌入应用二进制。部署只需复制 `bin/app` 和配置文件，页面无需额外目录。修改前端后执行 `make build` 重新打包。

开发时分别运行 `make run` 和 `make ui-dev`，访问 <http://127.0.0.1:5173>。前端通过 Vite 代理访问后端；若修改后端端口，可使用 `API_PROXY_TARGET=http://127.0.0.1:9090 make ui-dev` 指定开发代理地址。

## 配置与迁移

配置文件 `config.json`：

```json
{
  "http": {
    "host": "127.0.0.1",
    "port": 8080
  },
  "database": {
    "driver": "sqlite",
    "dsn": "file:app.db?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
  }
}
```

| 配置项 | 说明 |
| --- | --- |
| `http.host` | 监听地址，`127.0.0.1` 为本机，`0.0.0.0` 为所有 IPv4 网卡，也支持 `::1` 等 IPv6 地址 |
| `http.port` | 监听端口，范围为 1–65535 |
| `database.driver` | `sqlite` 或 `mysql` |
| `database.dsn` | 数据库连接串；SQLite 相对路径以启动工作目录为基准 |

这些配置项均为必填，文件缺失、配置错误时启动会报错。后端仅从 JSON 文件读取配置，UI 固定从二进制内的构建产物提供，没有 `UIDir` 配置项。

应用和迁移命令支持使用同一个指定配置文件：

```sh
./bin/app -config /path/to/config.json
./bin/migrate -config /path/to/config.json status
# Makefile 使用 CONFIG 指定路径
make run CONFIG=/path/to/config.json
make migrate-status CONFIG=/path/to/config.json
```

数据库迁移 SQL 同样嵌入 Go 可执行文件，按 `migrations/sqlite/` 和 `migrations/mysql/` 分开存放。默认读取根目录配置：

```sh
make migrate-status
make migrate-up
make migrate-down
```

`down` 回滚最近一条迁移，当前初始迁移会删除用户表及数据；下一次启动会重新执行待运行迁移。

使用 MySQL 8+ 时，先创建 `summary_your_usage` 数据库，再将配置中的 `database` 改为：

```json
{
  "driver": "mysql",
  "dsn": "root:password@tcp(127.0.0.1:3306)/summary_your_usage?charset=utf8mb4"
}
```

SQLite 与 MySQL 共用参数化 SQL，连接层处理驱动差异。切换数据库不会自动搬迁数据。

## 目录与 API

```text
cmd/                  应用与迁移入口
config.json           监听地址、端口、数据库配置
config/               JSON 配置读取
internal/app/         依赖组装、服务生命周期
internal/controller/  chi HTTP 接口
internal/entity/      用户实体
internal/usecase/     基础校验、仓储接口
internal/repo/        SQL 仓储
pkg/database/         数据库连接
migrations/           goose SQL
ui/embed.go           将 UI 构建产物嵌入 Go 二进制
ui/src/               页面、API 请求和样式
```

| 方法 | 路径 | 行为 |
| --- | --- | --- |
| GET | `/api/v1/users` | 返回全部用户数组 `[]`，按 ID 倒序 |
| GET | `/api/v1/users/{id}` | 用户详情 |
| POST | `/api/v1/users` | 新增，返回 201 |
| PUT | `/api/v1/users/{id}` | 更新，返回 200 |
| DELETE | `/api/v1/users/{id}` | 删除，返回 204 |

新增、更新使用 `Content-Type: application/json`，请求体为 `{"name":"张三","email":"zhangsan@example.com"}`。

保存时去除首尾空白、邮箱转小写。无效输入返回 400，用户不存在返回 404，重复邮箱返回 409；错误格式为 `{"error":{"code":"email_exists","message":"该邮箱已被使用"}}`。

## 验证

```sh
make test
make vet
npm --prefix ui run build
```

`make run`、`make build`、`make test`、`make vet` 会先构建 UI，需要先运行一次 `make ui-install`。直接使用 `go run`、`go build`、`go test` 时，也需要先执行 `make ui-build`，供 `go:embed` 打包。

测试使用临时 SQLite，覆盖 CRUD、校验、邮箱冲突、重启持久化和 goose 迁移回滚。MySQL 已提供驱动和迁移，尚未实库联调。
