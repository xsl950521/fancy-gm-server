# 认证与授权系统文档

## 概述

系统已实现完整的JWT认证和RBAC权限管理，所有API接口都需要认证才能访问（除登录、注册接口外）。

## 认证流程

### 1. 用户登录

**接口**：`POST /api/auth/login`

**请求体**：
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "访问token",
    "refresh_token": "刷新token",
    "expires_at": "2024-01-02T12:00:00Z",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com",
      "role_id": 1,
      "role_name": "super_admin"
    }
  }
}
```

### 2. 使用Token访问API

在请求头中添加：
```
Authorization: Bearer <token>
```

### 3. 刷新Token

当访问token过期时，使用刷新token获取新的访问token。

**接口**：`POST /api/auth/refresh`

**请求体**：
```json
{
  "refresh_token": "刷新token"
}
```

## 角色和权限

### 角色定义

1. **超级管理员 (super_admin)**
   - 拥有所有权限
   - 可以管理用户、角色、配置

2. **GM管理员 (admin)**
   - 可以上传、处理、补发、查看历史
   - 不能管理用户和角色

3. **操作员 (operator)**
   - 可以上传、处理和查看数据
   - 不能删除和补发

4. **查看者 (viewer)**
   - 只能查看数据

### 权限列表

| 权限代码 | 说明 | 默认角色 |
|---------|------|---------|
| `file.upload` | 文件上传 | admin, operator |
| `file.download` | 文件下载 | 所有角色 |
| `file.delete` | 文件删除 | super_admin |
| `data.process` | 数据处理 | admin, operator |
| `data.view` | 数据查看 | 所有角色 |
| `data.export` | 数据导出 | admin, operator |
| `resend.mail` | 补发邮件（高危） | admin |
| `history.view` | 查看历史 | 所有角色 |
| `history.delete` | 删除历史 | admin |
| `config.view` | 查看配置 | admin |
| `config.modify` | 修改配置 | super_admin |
| `user.*` | 用户管理 | super_admin |
| `role.*` | 角色管理 | super_admin |
| `audit.view` | 查看审计日志 | super_admin |

## API权限要求

### 无需认证的接口
- `POST /api/auth/login` - 登录
- `POST /api/auth/register` - 注册
- `POST /api/auth/refresh` - 刷新token

### 需要认证的接口

所有其他API接口都需要在请求头中提供有效的JWT token。

### 权限要求示例

- **文件上传**：需要 `file.upload` 权限
- **数据处理**：需要 `data.process` 权限
- **补发邮件**：需要 `resend.mail` 权限（高危操作）
- **删除历史**：需要 `history.delete` 权限
- **查看数据**：需要 `data.view` 权限

## 配置

### JWT配置

在 `config.yaml` 中配置：

```yaml
auth:
  jwt:
    secret_key: "your-secret-key"  # 生产环境必须修改！
    expiration: "24h"               # Token过期时间
    refresh_exp: "168h"             # 刷新Token过期时间（7天）
    issuer: "gm-server"            # 签发者
```

### 密码策略

```yaml
auth:
  password:
    min_length: 8          # 密码最小长度
    require_upper: false   # 要求大写字母
    require_lower: true    # 要求小写字母
    require_number: false  # 要求数字
    require_special: false # 要求特殊字符
```

### 会话配置

```yaml
auth:
  session:
    max_login_attempts: 5    # 最大登录尝试次数
    lockout_duration: "30m"  # 锁定持续时间
```

## 审计日志

系统会自动记录以下操作的审计日志：
- 用户登录
- 文件上传
- 数据处理
- 补发操作
- 历史记录删除
- 配置修改
- 用户管理操作

审计日志包含：
- 用户ID
- 操作类型
- 资源类型和ID
- IP地址
- 用户代理
- 请求ID
- 操作状态（成功/失败）
- 时间戳

## 错误处理

### 认证错误

- **401 Unauthorized**：未提供token或token无效
- **401 Unauthorized**：token已过期
- **403 Forbidden**：权限不足

### 错误响应格式

```json
{
  "code": 1004,
  "message": "未授权",
  "error": "[1004] 未授权: token已过期"
}
```

## 安全建议

1. **修改默认密钥**：生产环境必须修改JWT密钥
2. **修改默认密码**：首次登录后立即修改默认管理员密码
3. **使用HTTPS**：生产环境必须使用HTTPS
4. **定期轮换密钥**：定期更换JWT密钥
5. **监控审计日志**：定期查看审计日志，发现异常操作
6. **最小权限原则**：为用户分配最小必要权限

## 默认账户

系统首次启动会自动创建：
- **用户名**：`admin`
- **密码**：`admin123`
- **角色**：超级管理员

**⚠️ 首次登录后请立即修改密码！**
