# 校园失物招领系统后端

这是一个 Go + Gin 的前后端分离后端项目。

## 第 1 步：启动基础服务

设置本项目自己的 Go 缓存目录，避免 Windows 用户目录权限问题：

```powershell
$env:GOCACHE = (Join-Path (Get-Location) '.cache\go-build')
$env:GOPATH = (Join-Path (Get-Location) '.cache\go')
```

整理依赖：

```powershell
go mod tidy
```

启动服务：

```powershell
go run .\cmd\server
```

验证接口：

```powershell
Invoke-WebRequest -UseBasicParsing http://localhost:8080/health
```

成功时会返回：

```json
{"code":0,"msg":"success","data":{"status":"ok"}}
```

## 接下来要做的步骤

1. 添加配置读取和数据库连接。
2. 设计用户表，实现注册和登录。
3. 加入 JWT 鉴权和路由守卫对应的后端中间件。
4. 实现失物/招领信息发布、查询、详情、管理接口。
5. 实现认领申请和管理员审核流程。
6. 增加 RBAC 权限、公告管理、统计接口。
7. 编写接口文档，和前端联调。
