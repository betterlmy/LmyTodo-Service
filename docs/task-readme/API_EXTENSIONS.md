# TODO应用API扩展功能实现

## 概述

本文档描述了为跨平台TODO应用实现的扩展API接口，包括分类管理、用户设置、扩展TODO功能和搜索功能。

## 新增API接口

### 1. 扩展TODO管理 API

#### 1.1 获取扩展TODO列表
- **接口**: `POST /api/v2/todos/list`
- **功能**: 获取用户的TODO列表，支持分页和扩展字段
- **请求体**:
```json
{
  "limit": 20,
  "offset": 0
}
```

#### 1.2 创建扩展TODO
- **接口**: `POST /api/v2/todos/create`
- **功能**: 创建新的TODO任务，支持优先级、标签等扩展字段
- **请求体**:
```json
{
  "title": "学习Go语言",
  "description": "学习Go语言基础语法和框架",
  "priority": 2,
  "due_date": "2023-12-31T23:59:59Z",
  "tags": ["学习", "编程", "Go"],
  "category_id": 1,
  "reminder": "2023-12-30T09:00:00Z"
}
```

#### 1.3 更新扩展TODO
- **接口**: `POST /api/v2/todos/update`
- **功能**: 更新TODO任务，支持扩展字段的部分更新
- **请求体**:
```json
{
  "id": 1,
  "title": "更新后的标题",
  "priority": 3,
  "tags": ["紧急", "重要"]
}
```

#### 1.4 搜索TODO
- **接口**: `POST /api/v2/todos/search`
- **功能**: 根据关键词搜索TODO任务，支持标题、描述和标签搜索
- **请求体**:
```json
{
  "keyword": "Go",
  "limit": 20,
  "offset": 0
}
```

### 2. 分类管理 API

#### 2.1 获取分类列表
- **接口**: `POST /api/v2/categories`
- **功能**: 获取用户的所有分类

#### 2.2 创建分类
- **接口**: `POST /api/v2/categories/create`
- **功能**: 创建新的分类
- **请求体**:
```json
{
  "name": "工作",
  "color": "#FF5722",
  "icon": "work"
}
```

#### 2.3 更新分类
- **接口**: `POST /api/v2/categories/update`
- **功能**: 更新分类信息
- **请求体**:
```json
{
  "id": 1,
  "name": "工作",
  "color": "#FF5722",
  "icon": "work"
}
```

#### 2.4 删除分类
- **接口**: `POST /api/v2/categories/delete`
- **功能**: 删除分类（软删除）
- **请求体**:
```json
{
  "id": 1
}
```

### 3. 用户设置 API

#### 3.1 获取用户设置
- **接口**: `POST /api/v2/settings`
- **功能**: 获取用户的个性化设置

#### 3.2 更新用户设置
- **接口**: `POST /api/v2/settings/update`
- **功能**: 更新用户的个性化设置
- **请求体**:
```json
{
  "theme": "dark",
  "notification_time": "08:00",
  "language": "en-US",
  "timezone": "America/New_York"
}
```

## 数据模型扩展

### 扩展的TODO模型
```go
type Todo struct {
    ID          int         `json:"id"`
    UserID      int         `json:"user_id"`
    Title       string      `json:"title"`
    Description string      `json:"description"`
    Completed   bool        `json:"completed"`
    Priority    Priority    `json:"priority"`        // 新增：优先级 (0-3)
    DueDate     *time.Time  `json:"due_date"`        // 新增：截止日期
    Tags        StringSlice `json:"tags"`            // 新增：标签数组
    CategoryID  *int        `json:"category_id"`     // 新增：分类ID
    Reminder    *time.Time  `json:"reminder"`        // 新增：提醒时间
    CreatedAt   time.Time   `json:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at"`
    IsDeleted   bool        `json:"is_deleted"`
    SyncVersion int64       `json:"sync_version"`    // 新增：同步版本号
}
```

### 分类模型
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

### 用户设置模型
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

## 技术特性

### 1. 数据库支持
- 从SQLite升级到PostgreSQL
- 支持JSONB字段存储标签数据
- 优化的索引设计提升查询性能
- 软删除机制保护数据完整性

### 2. 搜索功能
- 支持标题、描述和标签的全文搜索
- 使用PostgreSQL的ILIKE进行大小写不敏感搜索
- 支持分页查询

### 3. 数据同步
- 每个TODO都有sync_version字段用于增量同步
- 自动更新时间戳和同步版本号

### 4. 错误处理
- 统一的错误响应格式
- 详细的参数验证
- 数据库约束检查

## 向后兼容性

所有v1 API接口保持不变，新功能通过v2 API提供，确保现有客户端不受影响。

## 测试

使用提供的`test_api.sh`脚本可以测试所有新增的API接口功能。

## 部署要求

1. PostgreSQL数据库 (推荐版本 12+)
2. 执行`db/ddl.sql`脚本初始化数据库结构
3. 配置环境变量：
   - `DB_HOST`: 数据库主机
   - `DB_PORT`: 数据库端口
   - `DB_USER`: 数据库用户名
   - `DB_PASSWORD`: 数据库密码
   - `DB_NAME`: 数据库名称