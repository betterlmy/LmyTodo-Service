# 数据同步和版本控制机制实现文档

## 概述

本文档描述了为跨平台TODO应用实现的数据同步和版本控制机制。该机制支持增量同步、批量同步和冲突解决，确保多设备间的数据一致性。

## 实现的功能

### 1. 同步版本字段 (sync_version)

为所有数据模型添加了 `sync_version` 字段：

- **todos表**: 已有 `sync_version` 字段
- **categories表**: 新增 `sync_version` 字段  
- **user_settings表**: 新增 `sync_version` 字段

#### 数据库触发器

创建了自动更新同步版本的触发器：

```sql
CREATE OR REPLACE FUNCTION update_sync_version()
RETURNS TRIGGER AS $
BEGIN
    NEW.sync_version = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;
    RETURN NEW;
END;
$ language 'plpgsql';

-- 为所有表添加触发器
CREATE TRIGGER update_todos_sync_version BEFORE UPDATE ON todos
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

CREATE TRIGGER update_categories_sync_version BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

CREATE TRIGGER update_user_settings_sync_version BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();
```

### 2. 增量同步API

#### 获取同步版本
- **接口**: `POST /api/v2/sync/version`
- **功能**: 获取服务器当前最大同步版本号
- **响应**: 
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "version": 1640995200000
  }
}
```

#### 增量同步数据
- **接口**: `POST /api/v2/sync/todos`
- **功能**: 基于时间戳获取增量更新的数据
- **请求参数**:
```json
{
  "since": 1640995200000
}
```
- **响应**: 包含所有在指定时间戳之后更新的数据
```json
{
  "code": 0,
  "message": "成功", 
  "data": {
    "todos": [...],
    "categories": [...],
    "settings": {...},
    "server_version": 1640995300000
  }
}
```

### 3. 批量同步API

#### 批量数据同步
- **接口**: `POST /api/v2/sync/batch`
- **功能**: 批量上传客户端数据并处理冲突
- **请求参数**:
```json
{
  "todos": [
    {
      "id": 0,  // 0表示新建，>0表示更新
      "title": "新任务",
      "description": "任务描述",
      "completed": false,
      "priority": 1,
      "tags": ["标签1", "标签2"],
      "is_deleted": false,
      "sync_version": 1640995200000,
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ],
  "categories": [...],
  "settings": {...}
}
```
- **响应**: 分类返回成功、冲突和错误的项目
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "success": [
      {
        "type": "todo",
        "local_id": 0,
        "server_id": 123,
        "action": "created",
        "message": "创建成功",
        "sync_version": 1640995300000
      }
    ],
    "conflicts": [...],
    "errors": [...]
  }
}
```

### 4. 冲突解决机制

#### 冲突检测规则
1. **时间戳比较**: 比较客户端 `updated_at` 和服务器 `updated_at`
2. **版本号比较**: 比较客户端 `sync_version` 和服务器 `sync_version`
3. **冲突条件**: 服务器数据更新时间晚于客户端且版本号更大

#### 冲突解决策略
- **服务器优先**: 默认策略，服务器数据优先
- **冲突标记**: 将冲突项目标记在响应中，由客户端处理
- **版本追踪**: 使用毫秒级时间戳确保版本唯一性

## 数据库架构更新

### 新增字段

```sql
-- categories表新增字段
ALTER TABLE categories ADD COLUMN sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;

-- user_settings表新增字段  
ALTER TABLE user_settings ADD COLUMN sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;
```

### 新增索引

```sql
-- 同步版本索引
CREATE INDEX IF NOT EXISTS idx_todos_sync_version ON todos(sync_version);
CREATE INDEX IF NOT EXISTS idx_categories_sync_version ON categories(sync_version);
CREATE INDEX IF NOT EXISTS idx_user_settings_sync_version ON user_settings(sync_version);
```

## 代码架构

### 数据模型更新

所有数据模型都包含了 `SyncVersion` 字段：

```go
type Todo struct {
    // ... 其他字段
    SyncVersion int64 `json:"sync_version"`
}

type Category struct {
    // ... 其他字段  
    SyncVersion int64 `json:"sync_version"`
}

type UserSettings struct {
    // ... 其他字段
    SyncVersion int64 `json:"sync_version"`
}
```

### 同步相关类型

```go
// 同步项目类型
type TodoSyncItem struct {
    ID          int      `json:"id,omitempty"`
    Title       string   `json:"title"`
    // ... 其他字段
    SyncVersion int64    `json:"sync_version"`
    UpdatedAt   string   `json:"updated_at"`
}

// 同步结果类型
type SyncResult struct {
    Type        string `json:"type"`
    LocalID     int    `json:"local_id,omitempty"`
    ServerID    int    `json:"server_id,omitempty"`
    Action      string `json:"action"`
    Message     string `json:"message,omitempty"`
    SyncVersion int64  `json:"sync_version,omitempty"`
}
```

### Repository层方法

新增了同步相关的数据访问方法：

```go
// 增量同步方法
func (r *ExtendedTodoRepository) GetTodosSince(userID int, since int64) ([]Todo, error)
func (r *CategoryRepository) GetCategoriesSince(userID int, since int64) ([]Category, error)
func (r *UserSettingsRepository) GetUserSettingsSince(userID int, since int64) (*UserSettings, error)

// 批量同步方法
func (r *ExtendedTodoRepository) BatchCreateOrUpdateTodos(userID int, todos []TodoSyncItem) ([]SyncResult, error)
func (r *CategoryRepository) BatchCreateOrUpdateCategories(userID int, categories []CategorySyncItem) ([]SyncResult, error)
func (r *UserSettingsRepository) BatchUpdateUserSettings(userID int, settings *UserSettingsSyncItem) (*SyncResult, error)

// 版本管理方法
func GetCurrentSyncVersion(db *sql.DB, userID int) (int64, error)
```

## 使用示例

### 客户端增量同步流程

1. **获取本地最后同步版本**
```javascript
const lastSyncVersion = localStorage.getItem('lastSyncVersion') || 0;
```

2. **请求增量数据**
```javascript
const response = await fetch('/api/v2/sync/todos', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ since: lastSyncVersion })
});
```

3. **处理增量数据**
```javascript
const syncData = await response.json();
if (syncData.code === 0) {
  // 更新本地数据
  updateLocalData(syncData.data);
  // 保存新的同步版本
  localStorage.setItem('lastSyncVersion', syncData.data.server_version);
}
```

### 客户端批量同步流程

1. **收集本地变更**
```javascript
const localChanges = {
  todos: getLocalTodoChanges(),
  categories: getLocalCategoryChanges(),
  settings: getLocalSettingsChanges()
};
```

2. **批量上传**
```javascript
const response = await fetch('/api/v2/sync/batch', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(localChanges)
});
```

3. **处理同步结果**
```javascript
const result = await response.json();
if (result.code === 0) {
  // 处理成功项目
  handleSuccessItems(result.data.success);
  // 处理冲突项目
  handleConflictItems(result.data.conflicts);
  // 处理错误项目
  handleErrorItems(result.data.errors);
}
```

## 性能优化

### 数据库优化
- 为 `sync_version` 字段创建索引
- 使用 `LIMIT` 和 `OFFSET` 进行分页查询
- 批量操作使用事务确保数据一致性

### 网络优化
- 增量同步减少数据传输量
- 批量操作减少网络请求次数
- 使用压缩减少传输大小

### 内存优化
- 流式处理大量数据
- 及时释放不需要的对象
- 使用对象池复用对象

## 测试

### 单元测试
- 测试同步版本更新逻辑
- 测试冲突检测算法
- 测试批量操作的事务性

### 集成测试
- 测试完整的同步流程
- 测试网络异常情况
- 测试并发同步场景

### 性能测试
- 测试大量数据的同步性能
- 测试高并发场景下的表现
- 测试内存使用情况

## 部署注意事项

### 数据库迁移
1. 执行 `db/migration_add_sync_version.sql` 脚本
2. 验证所有表都有 `sync_version` 字段
3. 确认触发器正常工作

### 配置更新
- 确保数据库连接池大小足够
- 配置适当的超时时间
- 启用数据库连接监控

### 监控指标
- 同步请求频率和响应时间
- 冲突发生率和解决时间
- 数据库查询性能指标

## 总结

本实现提供了完整的数据同步和版本控制机制，支持：

1. ✅ **同步版本字段**: 所有数据模型都有 `sync_version` 字段
2. ✅ **增量同步**: 基于时间戳的高效增量数据同步
3. ✅ **批量同步**: 支持离线数据的批量上传和下载
4. ✅ **冲突解决**: 基于时间戳和版本号的冲突检测和解决

该机制确保了多设备间的数据一致性，提供了良好的用户体验，并具备良好的性能和可扩展性。