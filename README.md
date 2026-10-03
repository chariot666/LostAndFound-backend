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

服务启动时会自动执行数据库迁移，创建 `users`、`items`、`claims`、`announcements`、`favorites`、`notifications` 和 `reports` 表。用户 UID 使用数据库自增主键，初始值为 `10001`。普通用户注册时必须填写联系方式，发布信息时会默认带出该联系方式，也可以按单条信息修改。

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

当前后端已补齐前端页面使用的用户认证、物品发布与查询、图片上传、认领申请、收藏、站内通知、举报、公告、管理员审核、用户管理和统计接口。图片上传接口返回 `/uploads/...` 相对路径，静态资源由后端 `/uploads` 路由提供。

## 生产部署

生产环境使用以下服务：

```text
Nginx（前端静态文件、/api 和 /uploads 反向代理）
    └── Go 后端
            └── MySQL
```

### 本地打包

在 Windows PowerShell 中执行：

```powershell
tar --exclude=.git --exclude=.cache --exclude=.pnpm-store --exclude=uploads --exclude=frontend/node_modules --exclude=frontend/dist --exclude=bin --exclude=tmp --exclude=.docker-config -czf "$HOME\Desktop\lost-found-deploy.tar.gz" -C "D:\试用期作业" .
scp "$HOME\Desktop\lost-found-deploy.tar.gz" root@116.62.157.240:/root/
```

### 服务器启动

登录服务器后执行：

```bash
mkdir -p /opt/lost-found
tar -xzf /root/lost-found-deploy.tar.gz -C /opt/lost-found
cd /opt/lost-found
cp .env.production.example .env.production
APP_PASSWORD=$(openssl rand -hex 16)
ROOT_PASSWORD=$(openssl rand -hex 24)
JWT_SECRET=$(openssl rand -hex 32)
sed -i \
  -e "s/replace-with-a-strong-app-password/$APP_PASSWORD/g" \
  -e "s/replace-with-a-different-strong-root-password/$ROOT_PASSWORD/g" \
  -e "s/replace-with-a-long-random-secret-at-least-32-characters/$JWT_SECRET/g" \
  .env.production
chmod 600 .env.production
docker compose -p lost-found -f compose.production.yaml up -d --build
docker compose -p lost-found -f compose.production.yaml ps
```

### 创建系统管理员

管理员初始化命令只在服务器终端执行，不是公开接口。这样普通用户注册时仍然只能获得 `user` 角色。

```bash
read -r -p "管理员用户名（3-10位）: " ADMIN_USERNAME
read -r -s -p "管理员密码（6-20位）: " ADMIN_PASSWORD
echo
docker compose -p lost-found -f compose.production.yaml exec \
  -e ADMIN_USERNAME="$ADMIN_USERNAME" \
  -e ADMIN_PASSWORD="$ADMIN_PASSWORD" \
  backend /app/admininit
unset ADMIN_USERNAME ADMIN_PASSWORD
```

### 验证

```bash
curl http://127.0.0.1/health
```

浏览器访问：

```text
http://116.62.157.240
```

阿里云轻量服务器防火墙需要放行 TCP `80`；配置 HTTPS 后再放行 TCP `443`。不要对公网开放 `3306` 或 `8080`。
