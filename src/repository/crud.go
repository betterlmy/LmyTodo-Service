package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"todo-service/global"
)

// CategoryRepository 分类数据访问层
type CategoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository 创建分类仓库实例
func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{db: global.Db}
}

// CreateCategory 创建分类
func (r *CategoryRepository) CreateCategory(category *Category) error {
	query := `
		INSERT INTO categories (user_id, name, color, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRow(query, category.UserID, category.Name, category.Color,
		category.Icon, time.Now(), time.Now()).Scan(&category.ID)

	return err
}

// GetCategoriesByUserID 根据用户ID获取分类列表
func (r *CategoryRepository) GetCategoriesByUserID(userID int) ([]Category, error) {
	query := `
		SELECT id, user_id, name, color, icon, created_at, updated_at, is_deleted
		FROM categories 
		WHERE user_id = $1 AND is_deleted = FALSE
		ORDER BY created_at ASC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.Color,
			&category.Icon, &category.CreatedAt, &category.UpdatedAt, &category.IsDeleted)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// UpdateCategory 更新分类
func (r *CategoryRepository) UpdateCategory(category *Category) error {
	query := `
		UPDATE categories 
		SET name = $1, color = $2, icon = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6`

	_, err := r.db.Exec(query, category.Name, category.Color, category.Icon,
		time.Now(), category.ID, category.UserID)
	return err
}

// DeleteCategory 删除分类（软删除）
func (r *CategoryRepository) DeleteCategory(id, userID int) error {
	query := `
		UPDATE categories 
		SET is_deleted = TRUE, updated_at = $1
		WHERE id = $2 AND user_id = $3`

	_, err := r.db.Exec(query, time.Now(), id, userID)
	return err
}

// UserSettingsRepository 用户设置数据访问层
type UserSettingsRepository struct {
	db *sql.DB
}

// NewUserSettingsRepository 创建用户设置仓库实例
func NewUserSettingsRepository() *UserSettingsRepository {
	return &UserSettingsRepository{db: global.Db}
}

// GetUserSettings 获取用户设置
func (r *UserSettingsRepository) GetUserSettings(userID int) (*UserSettings, error) {
	query := `
		SELECT user_id, theme, notification_time, language, timezone, created_at, updated_at
		FROM user_settings 
		WHERE user_id = $1`

	var settings UserSettings
	err := r.db.QueryRow(query, userID).Scan(
		&settings.UserID, &settings.Theme, &settings.NotificationTime,
		&settings.Language, &settings.TimeZone, &settings.CreatedAt, &settings.UpdatedAt)

	if err == sql.ErrNoRows {
		// 如果没有设置，创建默认设置
		return r.CreateDefaultUserSettings(userID)
	}

	return &settings, err
}

// CreateDefaultUserSettings 创建默认用户设置
func (r *UserSettingsRepository) CreateDefaultUserSettings(userID int) (*UserSettings, error) {
	settings := &UserSettings{
		UserID:           userID,
		Theme:            "light",
		NotificationTime: "09:00:00",
		Language:         "zh-CN",
		TimeZone:         "Asia/Shanghai",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	query := `
		INSERT INTO user_settings (user_id, theme, notification_time, language, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(query, settings.UserID, settings.Theme, settings.NotificationTime,
		settings.Language, settings.TimeZone, settings.CreatedAt, settings.UpdatedAt)

	return settings, err
}

// UpdateUserSettings 更新用户设置
func (r *UserSettingsRepository) UpdateUserSettings(settings *UserSettings) error {
	query := `
		UPDATE user_settings 
		SET theme = $1, notification_time = $2, language = $3, timezone = $4, updated_at = $5
		WHERE user_id = $6`

	settings.UpdatedAt = time.Now()
	_, err := r.db.Exec(query, settings.Theme, settings.NotificationTime, settings.Language,
		settings.TimeZone, settings.UpdatedAt, settings.UserID)
	return err
}

// ExtendedTodoRepository 扩展的TODO数据访问层
type ExtendedTodoRepository struct {
	db *sql.DB
}

// NewExtendedTodoRepository 创建扩展TODO仓库实例
func NewExtendedTodoRepository() *ExtendedTodoRepository {
	return &ExtendedTodoRepository{db: global.Db}
}

// CreateTodoExtended 创建扩展TODO
func (r *ExtendedTodoRepository) CreateTodoExtended(todo *Todo) error {
	// 序列化标签
	tagsJSON, err := json.Marshal(todo.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %v", err)
	}

	query := `
		INSERT INTO todos (user_id, title, description, completed, priority, due_date, tags, 
			category_id, reminder, created_at, updated_at, is_deleted, sync_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	err = r.db.QueryRow(query, todo.UserID, todo.Title, todo.Description, todo.Completed,
		todo.Priority, todo.DueDate, tagsJSON, todo.CategoryID, todo.Reminder,
		time.Now(), time.Now(), todo.IsDeleted, time.Now().UnixMilli()).Scan(&todo.ID)

	return err
}

// GetTodosByUserIDExtended 根据用户ID获取扩展TODO列表
func (r *ExtendedTodoRepository) GetTodosByUserIDExtended(userID int, limit, offset int) ([]Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, priority, due_date, tags,
			category_id, reminder, created_at, updated_at, is_deleted, sync_version
		FROM todos 
		WHERE user_id = $1 AND is_deleted = FALSE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var tagsJSON string

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
			&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)
		if err != nil {
			return nil, err
		}

		// 反序列化标签
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &todo.Tags); err != nil {
				todo.Tags = StringSlice{} // 如果解析失败，设置为空切片
			}
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// UpdateTodoExtended 更新扩展TODO
func (r *ExtendedTodoRepository) UpdateTodoExtended(todo *Todo) error {
	// 序列化标签
	tagsJSON, err := json.Marshal(todo.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %v", err)
	}

	query := `
		UPDATE todos 
		SET title = $1, description = $2, completed = $3, priority = $4, due_date = $5, tags = $6,
			category_id = $7, reminder = $8, updated_at = $9, sync_version = $10
		WHERE id = $11 AND user_id = $12`

	todo.UpdatedAt = time.Now()
	todo.SyncVersion = time.Now().UnixMilli()

	_, err = r.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.Priority,
		todo.DueDate, tagsJSON, todo.CategoryID, todo.Reminder,
		todo.UpdatedAt, todo.SyncVersion, todo.ID, todo.UserID)

	return err
}

// SearchTodos 搜索TODO
func (r *ExtendedTodoRepository) SearchTodos(userID int, keyword string, limit, offset int) ([]Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, priority, due_date, tags,
			category_id, reminder, created_at, updated_at, is_deleted, sync_version
		FROM todos 
		WHERE user_id = $1 AND is_deleted = FALSE 
			AND (title ILIKE $2 OR description ILIKE $3)
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`

	searchPattern := "%" + keyword + "%"
	rows, err := r.db.Query(query, userID, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var tagsJSON string

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
			&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)
		if err != nil {
			return nil, err
		}

		// 反序列化标签
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &todo.Tags); err != nil {
				todo.Tags = StringSlice{}
			}
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// GetTodosSince 获取指定时间戳之后的TODO（用于增量同步）
func (r *ExtendedTodoRepository) GetTodosSince(userID int, since int64) ([]Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, priority, due_date, tags,
			category_id, reminder, created_at, updated_at, is_deleted, sync_version
		FROM todos 
		WHERE user_id = $1 AND sync_version > $2
		ORDER BY sync_version ASC`

	rows, err := r.db.Query(query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var tagsJSON string

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
			&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)
		if err != nil {
			return nil, err
		}

		// 反序列化标签
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &todo.Tags); err != nil {
				todo.Tags = StringSlice{}
			}
		}

		todos = append(todos, todo)
	}

	return todos, nil
}
