# 校园失物招领系统后端

校园失物招领系统的后端服务，采用 Go 和 Gin 开发，提供统一的 RESTful API，并使用 MySQL 持久化业务数据。

## 技术栈

- Go 1.26
- Gin 1.10
- GORM 1.25
- MySQL 8.4
- Docker Compose

## 项目结构

```text
.
├── cmd/server/             # 服务启动入口
├── internal/config/        # 应用配置
├── internal/database/      # 数据库连接与迁移
├── internal/handler/       # HTTP 接口处理器
├── internal/middleware/    # JWT 鉴权和权限中间件
├── internal/model/         # 数据模型
├── internal/response/      # 统一响应结构
├── API.md                  # API 接口文档
├── docker-compose.yml      # MySQL 开发环境
├── .env.example            # 环境变量示例
├── go.mod
└── go.sum
```

## 环境要求

- Go 1.26 或兼容版本
- Docker Desktop（用于运行 MySQL）

## 配置

服务通过环境变量读取配置，未设置时使用以下默认值：

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `APP_PORT` | `8080` | HTTP 服务端口 |
| `DB_HOST` | `127.0.0.1` | MySQL 地址 |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USER` | `app` | MySQL 用户名 |
| `DB_PASSWORD` | `app_password` | MySQL 密码 |
| `DB_NAME` | `lost_found` | 数据库名称 |
| `JWT_SECRET` | `local-development-secret` | JWT 签名密钥，生产环境必须修改 |
| `UPLOAD_DIR` | `uploads` | 图片上传保存目录 |

配置示例见 `.env.example`。

## 本地运行

启动 MySQL：

```powershell
docker compose -p lost-found up -d
```

安装依赖并启动后端：

```powershell
go mod download
go run .\cmd\server
```

服务启动时会自动执行数据库迁移，创建 `users`、`items`、`claims` 和 `announcements` 表。用户 UID 使用数据库自增主键，初始值为 `10001`。

## 验证服务

访问健康检查接口：

```http
GET http://localhost:8080/health
```

成功响应：

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "status": "ok"
  }
}
```

运行项目测试：

```powershell
go test ./...
```

## 接口约定

业务接口统一使用 `/api/v1` 前缀，响应结构包含 `code`、`msg` 和 `data`。完整接口定义、请求字段、响应示例及错误码见 [API.md](./API.md)。

当前后端已补齐前端页面使用的用户认证、物品发布与查询、图片上传、认领申请、公告、管理员审核、用户管理和统计接口。图片上传接口返回 `/uploads/...` 相对路径，静态资源由后端 `/uploads` 路由提供。
