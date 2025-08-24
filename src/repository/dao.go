package repository

import (
	"log"
	"todo-service/global"
)

// CreateTables 创建PostgreSQL数据库表
func CreateTables() {
	createPostgreSQLTables()
}

// createPostgreSQLTables 创建PostgreSQL表
func createPostgreSQLTables() {
	log.Println("Creating PostgreSQL tables...")

	// 用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// 分类表
	categoryTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(100) NOT NULL,
		color VARCHAR(7) DEFAULT '#2196F3',
		icon VARCHAR(50) DEFAULT 'folder',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN DEFAULT FALSE,
		UNIQUE(user_id, name)
	);`

	// 用户设置表
	userSettingsTable := `
	CREATE TABLE IF NOT EXISTS user_settings (
		user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		theme VARCHAR(10) DEFAULT 'light' CHECK (theme IN ('light', 'dark', 'auto')),
		notification_time TIME DEFAULT '09:00:00',
		language VARCHAR(10) DEFAULT 'zh-CN',
		timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	// TODO表（扩展版）
	todoTable := `
	CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(200) NOT NULL,
		description TEXT,
		completed BOOLEAN DEFAULT FALSE,
		priority INTEGER DEFAULT 0 CHECK (priority >= 0 AND priority <= 3),
		due_date TIMESTAMP WITH TIME ZONE,
		tags JSONB DEFAULT '[]'::jsonb,
		category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		reminder TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		is_deleted BOOLEAN DEFAULT FALSE,
		sync_version BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000
	);`

	tables := []string{userTable, categoryTable, userSettingsTable, todoTable}

	for _, table := range tables {
		if _, err := global.Db.Exec(table); err != nil {
			log.Fatal("Failed to create PostgreSQL table:", err)
		}
	}

	// 创建索引
	createPostgreSQLIndexes()

	log.Println("PostgreSQL tables created successfully")
}

// createPostgreSQLIndexes 创建PostgreSQL索引
func createPostgreSQLIndexes() {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_categories_user_id_name ON categories(user_id, name)",
		"CREATE INDEX IF NOT EXISTS idx_todos_user_id ON todos(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_todos_user_id_completed ON todos(user_id, completed)",
		"CREATE INDEX IF NOT EXISTS idx_todos_user_id_created_at ON todos(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority)",
		"CREATE INDEX IF NOT EXISTS idx_todos_category_id ON todos(category_id) WHERE category_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_todos_sync_version ON todos(sync_version)",
		"CREATE INDEX IF NOT EXISTS idx_todos_tags ON todos USING GIN(tags)",
		"CREATE INDEX IF NOT EXISTS idx_todos_reminder ON todos(reminder) WHERE reminder IS NOT NULL",
	}

	for _, index := range indexes {
		if _, err := global.Db.Exec(index); err != nil {
			log.Printf("Warning: Failed to create index: %v", err)
		}
	}
}
