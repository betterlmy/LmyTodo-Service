# TODO Backend API

这是一个使用Go语言开发的TODO应用后端API服务。

## 功能特性

- 用户注册和登录
- JWT认证
- TODO的增删改查
- SQLite数据库存储

## 安装和运行

1. 进入backend目录：
```bash
cd todo-backend
```

2. 初始化Go模块并安装依赖：
```bash
go mod init todo-backend
go mod tidy
```

3. 运行服务：
```bash
go run main.go
```

服务将在 `http://localhost:8080` 启动。

## API接口

所有接口统一使用POST请求，返回HTTP状态码200，具体的业务状态通过响应体中的code字段判断。

### 统一响应格式
```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

### 错误码说明
- 0: 成功
- 10001: 参数错误
- 10002: 用户已存在
- 10003: 账号密码错误
- 10004: Token错误
- 10005: 资源不存在
- 10006: 内部错误
- 10007: 未授权

### 认证接口

- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录

### TODO接口（需要JWT认证）

- `POST /api/todos/list` - 获取当前用户的所有TODO
- `POST /api/todos/create` - 创建新的TODO
- `POST /api/todos/update` - 更新TODO
- `POST /api/todos/delete` - 删除TODO
- `POST /api/profile` - 获取用户信息

## 数据模型

### 用户注册
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

### 用户登录
```json
{
  "username": "testuser",
  "password": "password123"
}
```

### 创建TODO
```json
{
  "title": "完成项目",
  "description": "完成TODO应用的开发"
}
```

### 更新TODO
```json
{
  "id": 1,
  "title": "新标题",
  "description": "新描述",
  "completed": true
}
```

### 删除TODO
```json
{
  "id": 1
}
```
