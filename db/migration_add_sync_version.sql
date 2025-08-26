-- 数据库迁移脚本：添加同步版本字段
-- 执行时间：2025-08-27

-- 为 categories 表添加 sync_version 字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'categories' AND column_name = 'sync_version') THEN
        ALTER TABLE categories ADD COLUMN sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;
    END IF;
END $$;

-- 为 user_settings 表添加 sync_version 字段（如果不存在）
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'user_settings' AND column_name = 'sync_version') THEN
        ALTER TABLE user_settings ADD COLUMN sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;
    END IF;
END $$;

-- 创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_categories_sync_version ON categories(sync_version);
CREATE INDEX IF NOT EXISTS idx_user_settings_sync_version ON user_settings(sync_version);

-- 创建同步版本更新触发器（如果不存在）
CREATE OR REPLACE FUNCTION update_sync_version()
RETURNS TRIGGER AS $
BEGIN
    NEW.sync_version = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000;
    RETURN NEW;
END;
$ language 'plpgsql';

-- 为 categories 表添加同步版本触发器
DROP TRIGGER IF EXISTS update_categories_sync_version ON categories;
CREATE TRIGGER update_categories_sync_version BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

-- 为 user_settings 表添加同步版本触发器
DROP TRIGGER IF EXISTS update_user_settings_sync_version ON user_settings;
CREATE TRIGGER update_user_settings_sync_version BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_sync_version();

-- 更新现有数据的 sync_version（设置为当前时间戳）
UPDATE categories SET sync_version = EXTRACT(EPOCH FROM updated_at) * 1000 WHERE sync_version IS NULL;
UPDATE user_settings SET sync_version = EXTRACT(EPOCH FROM updated_at) * 1000 WHERE sync_version IS NULL;

-- 验证迁移结果
SELECT 'categories' as table_name, COUNT(*) as total_rows, 
       COUNT(sync_version) as sync_version_count 
FROM categories
UNION ALL
SELECT 'user_settings' as table_name, COUNT(*) as total_rows, 
       COUNT(sync_version) as sync_version_count 
FROM user_settings;