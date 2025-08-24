# 数据库架构升级指南

## 概述

本文档描述了 TODO 应用从 SQLite 迁移到 PostgreSQL 的数据库架构升级过程，包括新增的数据模型、索引优化和迁移步骤。

## Docker安装PostgreSQL
```bash
docker run \
--name my-postgres \
-e POSTGRES_PASSWORD=admin123 \
-p 5432:5432 \
-v pg-data:/var/lib/postgresql/data \
-d postgres
```
## 新增功能

### 1. 扩展的数据模型

#### Priority 枚举类型

- `PriorityLow` (0): 低优先级
- `PriorityMedium` (1): 中优先级
- `PriorityHigh` (2): 高优先级
- `PriorityUrgent` (3): 紧急

#### Category 分类模型

```go
type Category struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Name      string    `json:"name"`
    Color     string    `json:"color"`
    Icon      string    `json:"icon"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    IsDeleted bool      `json:"is_deleted"`
}
```

#### UserSettings 用户设置模型

```go
type UserSettings struct {
    UserID           int       `json:"user_id"`
    Theme            string    `json:"theme"`            // light/dark/auto
    NotificationTime string    `json:"notification_time"`
    Language         string    `json:"language"`
    TimeZone         string    `json:"timezone"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

#### 扩展的 Todo 模型

新增字段：

- `Priority`: 任务优先级
- `DueDate`: 截止日期
- `Tags`: 标签数组
- `CategoryID`: 分类 ID
- `Reminder`: 提醒时间
- `SyncVersion`: 同步版本号

### 2. 数据库支持

#### 支持的数据库

- **SQLite**: 开发和测试环境（向后兼容）
- **PostgreSQL**: 生产环境（推荐）

#### 环境变量配置

```bash
# PostgreSQL配置
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=todo_app
DB_SSLMODE=disable

# SQLite配置（默认）
DB_DRIVER=sqlite3
DB_FILEPATH=db/todo.db
```

### 3. 性能优化

#### 索引策略

- **用户表**: username, email, created_at
- **分类表**: user_id, (user_id, name), created_at
- **TODO 表**:
  - user_id, (user_id, completed), (user_id, created_at)
  - due_date, priority, category_id, sync_version
  - tags (GIN 索引，仅 PostgreSQL), reminder

#### 查询优化

- 分页查询支持
- 增量同步查询
- 全文搜索（PostgreSQL ILIKE，SQLite LIKE）

## 迁移步骤

### 1. 准备工作

1. **备份现有数据**

```bash
sqlite3 db/todo.db ".dump" > sqlite_backup.sql
```

2. **安装 PostgreSQL 依赖**

```bash
go get github.com/lib/pq
```

### 2. 数据库初始化

#### PostgreSQL 环境

```bash
# 1. 创建数据库
createdb todo_app

# 2. 执行DDL脚本
psql -d todo_app -f ddl.sql

# 3. 设置环境变量
export DB_DRIVER=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=todo_app
```

#### SQLite 环境（默认）

```bash
# 无需额外配置，使用默认设置
mkdir -p db
```

### 3. 运行应用

```bash
# 启动应用
go run main.go

# 应用会自动：
# 1. 检测数据库类型
# 2. 创建相应的表结构
# 3. 创建优化索引
```

### 4. 数据迁移（可选）

如果需要从现有 SQLite 迁移到 PostgreSQL：

```bash
# 1. 导出SQLite数据
sqlite3 db/todo.db ".mode insert" ".output sqlite_data.sql" "SELECT * FROM users;" "SELECT * FROM todos;"

# 2. 手动调整SQL格式适配PostgreSQL

# 3. 执行迁移脚本
psql -d todo_app -f migration.sql
```

## API 变更

### 新增接口

#### 分类管理

```bash
# 获取分类列表
GET /api/v2/categories

# 创建分类
POST /api/v2/categories
{
  "name": "工作",
  "color": "#FF5722",
  "icon": "work"
}

# 更新分类
PUT /api/v2/categories/:id

# 删除分类
DELETE /api/v2/categories/:id
```

#### 用户设置

```bash
# 获取用户设置
GET /api/v2/settings

# 更新用户设置
PUT /api/v2/settings
{
  "theme": "dark",
  "notification_time": "09:00:00",
  "language": "zh-CN",
  "timezone": "Asia/Shanghai"
}
```

#### 同步接口

```bash
# 增量同步
GET /api/v2/sync/todos?since=1640995200000

# 批量同步
POST /api/v2/sync/todos/batch

# 获取同步版本
GET /api/v2/sync/version
```

### 扩展的 TODO 接口

现有 TODO 接口支持新字段：

```json
{
  "id": 1,
  "user_id": 1,
  "title": "学习Go语言",
  "description": "学习Go语言基础语法",
  "completed": false,
  "priority": 2,
  "due_date": "2023-12-31T23:59:59Z",
  "tags": ["工作", "重要"],
  "category_id": 1,
  "reminder": "2023-12-30T09:00:00Z",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "sync_version": 1640995200000
}
```

## 测试验证

### 1. 功能测试

```bash
# 测试数据库连接
curl -X POST http://localhost:8080/api/test

# 测试用户注册
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password"}'

# 测试TODO创建（扩展字段）
curl -X POST http://localhost:8080/api/todos/create \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试任务",
    "description": "测试扩展字段",
    "priority": 2,
    "due_date": "2023-12-31T23:59:59Z",
    "tags": ["测试", "重要"],
    "category_id": 1
  }'
```

### 2. 性能测试

```bash
# 测试索引效果
EXPLAIN ANALYZE SELECT * FROM todos WHERE user_id = 1 ORDER BY created_at DESC LIMIT 20;

# 测试搜索性能
EXPLAIN ANALYZE SELECT * FROM todos WHERE user_id = 1 AND title ILIKE '%关键词%';
```

## 故障排除

### 常见问题

1. **PostgreSQL 连接失败**

   - 检查 PostgreSQL 服务是否启动
   - 验证连接参数和权限
   - 检查防火墙设置

2. **SQLite 文件权限问题**

   - 确保 db 目录存在且可写
   - 检查文件权限设置

3. **迁移数据丢失**
   - 确保备份文件完整
   - 检查字符编码问题
   - 验证外键约束

### 日志调试

应用启动时会输出数据库连接信息：

```
Successfully connected to postgres database
PostgreSQL tables created successfully
Server starting on port 8080 with postgres database
```

## 后续优化

1. **连接池配置**: 配置数据库连接池参数
2. **读写分离**: 支持主从数据库配置
3. **缓存策略**: 集成 Redis 缓存热点数据
4. **监控告警**: 添加数据库性能监控
5. **备份策略**: 自动化数据备份方案
