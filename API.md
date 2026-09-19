# 校园失物招领系统 API
## 1. 基本约定
本地开发地址：
```text
http://localhost:8080
```
接口前缀：
```text
/api/v1
```
成功响应统一格式：
```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```
失败响应统一格式：
```json
{
  "code": "10001",
  "msg": "参数错误",
  "data": null
}
```
约定：
- `code = 0` 表示成功。
- `data` 没有内容时返回 `null`。
- 分页从 `page = 1` 开始。
- 默认每页 `page_size = 10`，最大值为 `100`。
- 需要登录的接口必须携带请求头：
```text
Authorization: Bearer <token>
```
## 2. 用户与认证
### 2.1 用户注册
```text
POST /api/v1/auth/register
```
请求体：
```json
{
  "username": "zhangsan",
  "password": "123456"
}
```
username要求在3到10个字符，允许中文、数字、字母，允许重名
password要求6到20个字符
成功响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "username": "zhangsan",
    "uid": 10001,
    "role": "user"
  }
}
```
uid由后端自动分配，从10001开始
前端注册角色role固定为普通user
### 2.2 用户登录
```text
POST /api/v1/auth/login
```
请求体：
```json
{
  "uid": 10001,
  "password": "123456"
}
```
成功响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "jwt-token",
    "user": {
      "username": "zhangsan",
      "uid": 10001,
      "role": "user"
    }
  }
}
```
### 2.3 获取当前用户
```text
GET /api/v1/auth/me
```
需要登录，请求头：
```text
Authorization: Bearer <token>
```
成功响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "uid": 10001,
    "username": "zhangsan",
    "role": "user",
    "status": "active",
    "created_at": "2026-09-17T10:00:00+08:00"
  }
}
```
未登录或 Token 无效：
```json
{
  "code": 10002,
  "msg": "未登录或令牌无效",
  "data": null
}
```
## 3. 失物和拾物信息
字段约定：
- `type`: `lost` 表示失物，`found` 表示拾物。
- `status`: `pending` 待审核，`approved` 已发布，`claimed` 已认领，`closed` 已关闭。
### 3.1 获取信息列表
```text
GET /api/v1/items?page=1&page_size=10&type=lost&keyword=钱包&location=图书馆
```
参数都是可选的：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `page` | number | 页码 |
| `page_size` | number | 每页数量 |
| `type` | string | `lost` 或 `found` |
| `keyword` | string | 搜索标题和描述 |
| `location` | string | 丢失或拾取地点 |
| `status` | string | 信息状态 |
| `sort` | string | `latest` 表示最新发布 |
成功响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "type": "lost",
        "title": "黑色钱包",
        "description": "内有学生卡和银行卡",
        "location": "图书馆三楼",
        "lost_at": "2026-09-10T14:30:00+08:00",
        "contact": "13800000000",
        "images": [
          "http://localhost:8080/uploads/wallet.jpg"
        ],
        "status": "approved",
        "user": {
          "id": 10001,
          "username": "zhangsan"
        },
        "created_at": "2026-09-10T15:00:00+08:00"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```
### 3.2 获取信息详情
```text
GET /api/v1/items/:id
```
例如：
```text
GET /api/v1/items/1
```
### 3.3 发布失物或拾物信息
```text
POST /api/v1/items
```
需要登录。
请求体：
```json
{
  "type": "lost",
  "title": "黑色钱包",
  "description": "内有学生卡和银行卡",
  "location": "图书馆三楼",
  "lost_at": "2026-09-10T14:30:00+08:00",
  "contact": "13800000000",
  "images": [
    "http://localhost:8080/uploads/wallet.jpg"
  ]
}
```
新发布的信息默认进入 `pending` 状态，等待管理员审核。
### 3.4 修改自己的信息
```text
PUT /api/v1/items/:id
```
需要登录。只能修改自己发布且还没有完成认领的信息。
请求体字段与发布接口相同，允许只提交需要修改的字段。
### 3.5 删除自己的信息
```text
DELETE /api/v1/items/:id
```
需要登录。只能删除自己的信息。
### 3.6 获取我发布的信息
```text
GET /api/v1/me/items?page=1&page_size=10&status=pending
```
需要登录。
## 4. 图片上传
### 4.1 上传图片
```text
POST /api/v1/upload
```
需要登录，使用 `multipart/form-data`。
表单字段：
```text
file: 图片文件
```
成功响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "url": "http://localhost:8080/uploads/20260910-wallet.jpg"
  }
}
```
建议限制：
- 只允许 `jpg`、`jpeg`、`png`、`webp`。
- 单张图片最大 `5 MB`。
- 后端保存文件名时使用随机名称，不能直接使用用户上传的原始文件名。
## 5. 认领申请
### 5.1 提交认领申请
```text
POST /api/v1/items/:id/claims
```
需要登录。
请求体：
```json
{
  "proof": "钱包内有我的学生卡，学号是 20260001",
  "contact": "13800000000"
}
```
### 5.2 获取我提交的申请
```text
GET /api/v1/me/claims?page=1&page_size=10
```
需要登录。
### 5.3 管理员获取待处理申请
```text
GET /api/v1/admin/claims?page=1&page_size=10&status=pending
```
需要失物招领管理员或系统管理员权限。
### 5.4 管理员审核认领申请
```text
PUT /api/v1/admin/claims/:id
```
需要失物招领管理员或系统管理员权限。
请求体：
```json
{
  "status": "approved",
  "remark": "已核对学生卡信息"
}
```
`status` 只能是 `approved` 或 `rejected`。
## 6. 管理员功能
角色约定：
- `user`: 普通用户。
- `item_admin`: 失物招领管理员。
- `system_admin`: 系统管理员。
### 6.1 审核失物或拾物信息
```text
GET /api/v1/admin/items?page=1&page_size=10&status=pending
PUT /api/v1/admin/items/:id
```
审核请求体：
```json
{
  "status": "approved",
  "remark": "内容审核通过"
}
```
### 6.2 获取用户列表
```text
GET /api/v1/admin/users?page=1&page_size=10&keyword=zhangsan
```
需要系统管理员权限。
### 6.3 修改用户角色或状态
```text
PUT /api/v1/admin/users/:id
```
请求体：
```json
{
  "role": "item_admin",
  "status": "active"
}
```
需要系统管理员权限。
## 7. 公告和统计
### 7.1 获取公告列表
```text
GET /api/v1/announcements?page=1&page_size=10
```
### 7.2 管理公告
```text
POST /api/v1/admin/announcements
PUT /api/v1/admin/announcements/:id
DELETE /api/v1/admin/announcements/:id
```
需要系统管理员权限。
新增公告请求体：
```json
{
  "title": "失物招领平台使用说明",
  "content": "请如实填写物品信息。",
  "published": true
}
```
### 7.3 获取统计数据
```text
GET /api/v1/admin/statistics
```
成功响应：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total_items": 100,
    "pending_items": 8,
    "claimed_items": 42,
    "total_users": 80,
    "total_claims": 50
  }
}
```
## 8. 错误码
| 错误码 | 含义 |
| --- | --- |
| `0` | 成功 |
| `10001` | 参数错误 |
| `10002` | 未登录或令牌无效 |
| `10003` | 没有权限 |
| `10004` | 资源不存在 |
| `10005` | 资源状态不允许当前操作 |
| `10006` | 注册信息不符合规范 |
| `10007` | uid或密码错误 |
| `10008` | 重复提交 |
| `20001` | 服务器内部错误 |
## 9. 接口实现顺序
建议按以下顺序开发和联调：
1. `GET /health`
2. `POST /auth/register`
3. `POST /auth/login`
4. `GET /auth/me`
5. `GET /items`
6. `GET /items/:id`
7. `POST /items`
8. `PUT /items/:id` 和 `DELETE /items/:id`
9. `POST /items/:id/claims`
10. 管理员审核接口
11. 公告和统计接口
12. 图片上传和生产环境部署
