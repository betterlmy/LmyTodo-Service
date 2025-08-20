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

### 认证接口

- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录

### TODO接口（需要JWT认证）

- `GET /api/todos` - 获取当前用户的所有TODO
- `POST /api/todos` - 创建新的TODO
- `PUT /api/todos/:id` - 更新TODO
- `DELETE /api/todos/:id` - 删除TODO
- `GET /api/profile` - 获取用户信息

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
  "title": "新标题",
  "description": "新描述",
  "completed": true
}
```
