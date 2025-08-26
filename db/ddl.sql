-- 跨平台TODO应用数据库架构 (PostgreSQL)
-- 版本: 2.0
-- 创建时间: 2025-08-25

-- 启用UUID扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 分类表
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7) DEFAULT '#2196F3', -- 默认蓝色
    icon VARCHAR(50) DEFAULT 'folder',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,
    sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, -- 毫秒时间戳
    UNIQUE(user_id, name) -- 同一用户下分类名称唯一
);

-- 用户设置表
CREATE TABLE IF NOT EXISTS user_settings (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(10) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),
    notification_time TIME DEFAULT '09:00:00',
    language VARCHAR(10) DEFAULT 'zh-CN',
    timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 -- 毫秒时间戳
);

-- TODO任务表（扩展版）
CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    priority INTEGER DEFAULT 0 CHECK (priority >= 0 AND priority <= 3), -- 0:低 1:中 2:高 3:紧急
    due_date TIMESTAMP WITH TIME ZONE,
    tags JSONB DEFAULT '[]'::jsonb, -- 存储标签数组
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    reminder TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,
    sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000 -- 毫秒时间戳
);

-- 创建索引优化查询性能
-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 分类表索引
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_categories_user_id_name ON categories(user_id, name);
CREATE INDEX IF NOT EXISTS idx_categories_created_at ON categories(created_at);
CREATE INDEX IF NOT EXISTS idx_categories_sync_version ON categories(sync_version);

-- TODO表索引
CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id);
CREATE INDEX IF NOT EXISTS idx_todos_user_id_completed ON todos(user_id, completed);
CREATE INDEX IF NOT EXISTS idx_todos_user_id_created_at ON todos(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
CREATE INDEX IF NOT EXISTS idx_todos_category_id ON todos(category_id) WHERE category_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_todos_sync_version ON todos(sync_version);
CREATE INDEX IF NOT EXISTS idx_todos_tags ON todos USING GIN(tags); -- GIN索引用于JSONB查询
CREATE INDEX IF NOT EXISTS idx_todos_reminder ON todos(reminder) WHERE reminder IS NOT NULL;

-- 用户设置表索引
CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_user_settings_sync_version ON user_settings(sync_version);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表添加更新时间触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_todos_updated_at BEFORE UPDATE ON todos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 创建同步版本更新触发器
CREATE OR REPLACE FUNCTION update_sync_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.sync_version = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_todos_sync_version BEFORE UPDATE ON todos
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

CREATE TRIGGER update_categories_sync_version BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

CREATE TRIGGER update_user_settings_sync_version BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

-- 插入默认分类数据
INSERT INTO categories (user_id, name, color, icon) VALUES 
(1, '工作', '#FF5722', 'work'),
(1, '个人', '#4CAF50', 'person'),
(1, '学习', '#2196F3', 'school'),
(1, '购物', '#FF9800', 'shopping_cart')
ON CONFLICT (user_id, name) DO NOTHING;

-- 插入默认用户设置
INSERT INTO user_settings (user_id) VALUES (1)
ON CONFLICT (user_id) DO NOTHING;

-- 创建视图：活跃任务（未删除且未完成）
CREATE OR REPLACE VIEW active_todos AS
SELECT 
    t.*,
    c.name as category_name,
    c.color as category_color,
    c.icon as category_icon
FROM todos t
LEFT JOIN categories c ON t.category_id = c.id AND c.is_deleted = FALSE
WHERE t.is_deleted = FALSE;

-- 创建视图：任务统计
CREATE OR REPLACE VIEW todo_stats AS
SELECT 
    user_id,
    COUNT(*) as total_todos,
    COUNT(CASE WHEN completed = TRUE THEN 1 END) as completed_todos,
    COUNT(CASE WHEN completed = FALSE THEN 1 END) as pending_todos,
    COUNT(CASE WHEN due_date IS NOT NULL AND due_date < CURRENT_TIMESTAMP AND completed = FALSE THEN 1 END) as overdue_todos,
    COUNT(CASE WHEN priority = 3 AND completed = FALSE THEN 1 END) as urgent_todos
FROM todos 
WHERE is_deleted = FALSE
GROUP BY user_id;

-- 添加注释
COMMENT ON TABLE users IS '用户表';
COMMENT ON TABLE categories IS '任务分类表';
COMMENT ON TABLE user_settings IS '用户个性化设置表';
COMMENT ON TABLE todos IS 'TODO任务表（扩展版）';

COMMENT ON COLUMN todos.priority IS '优先级：0-低，1-中，2-高，3-紧急';
COMMENT ON COLUMN todos.tags IS '任务标签，JSON数组格式';
COMMENT ON COLUMN todos.sync_version IS '同步版本号，用于增量同步';