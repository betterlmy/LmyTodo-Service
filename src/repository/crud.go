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
		INSERT INTO categories (user_id, name, color, icon, created_at, updated_at, sync_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	now := time.Now()
	syncVersion := now.UnixMilli()
	err := r.db.QueryRow(query, category.UserID, category.Name, category.Color,
		category.Icon, now, now, syncVersion).Scan(&category.ID)

	if err == nil {
		category.CreatedAt = now
		category.UpdatedAt = now
		category.SyncVersion = syncVersion
	}

	return err
}

// GetCategoriesByUserID 根据用户ID获取分类列表
func (r *CategoryRepository) GetCategoriesByUserID(userID int) ([]Category, error) {
	query := `
		SELECT id, user_id, name, color, icon, created_at, updated_at, is_deleted, sync_version
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
			&category.Icon, &category.CreatedAt, &category.UpdatedAt, &category.IsDeleted, &category.SyncVersion)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// UpdateCategory 更新分类
func (r *CategoryRepository) UpdateCategory(category *Category) error {
	query := `
		UPDATE categories 
		SET name = $1, color = $2, icon = $3, updated_at = $4, sync_version = $5
		WHERE id = $6 AND user_id = $7`

	now := time.Now()
	syncVersion := now.UnixMilli()
	result, err := r.db.Exec(query, category.Name, category.Color, category.Icon,
		now, syncVersion, category.ID, category.UserID)

	if err == nil {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("category not found or not owned by user")
		}
		category.UpdatedAt = now
		category.SyncVersion = syncVersion
	}

	return err
}

// DeleteCategory 删除分类（软删除）
func (r *CategoryRepository) DeleteCategory(id, userID int) error {
	query := `
		UPDATE categories 
		SET is_deleted = TRUE, updated_at = $1, sync_version = $2
		WHERE id = $3 AND user_id = $4`

	now := time.Now()
	syncVersion := now.UnixMilli()
	result, err := r.db.Exec(query, now, syncVersion, id, userID)
	if err == nil {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("category not found or not owned by user")
		}
	}
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
		SELECT user_id, theme, notification_time, language, timezone, created_at, updated_at, sync_version
		FROM user_settings 
		WHERE user_id = $1`

	var settings UserSettings
	err := r.db.QueryRow(query, userID).Scan(
		&settings.UserID, &settings.Theme, &settings.NotificationTime,
		&settings.Language, &settings.TimeZone, &settings.CreatedAt, &settings.UpdatedAt, &settings.SyncVersion)

	if err == sql.ErrNoRows {
		// 如果没有设置，创建默认设置
		return r.CreateDefaultUserSettings(userID)
	}

	return &settings, err
}

// CreateDefaultUserSettings 创建默认用户设置
func (r *UserSettingsRepository) CreateDefaultUserSettings(userID int) (*UserSettings, error) {
	now := time.Now()
	syncVersion := now.UnixMilli()
	settings := &UserSettings{
		UserID:           userID,
		Theme:            "light",
		NotificationTime: "09:00:00",
		Language:         "zh-CN",
		TimeZone:         "Asia/Shanghai",
		CreatedAt:        now,
		UpdatedAt:        now,
		SyncVersion:      syncVersion,
	}

	query := `
		INSERT INTO user_settings (user_id, theme, notification_time, language, timezone, created_at, updated_at, sync_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query, settings.UserID, settings.Theme, settings.NotificationTime,
		settings.Language, settings.TimeZone, settings.CreatedAt, settings.UpdatedAt, settings.SyncVersion)

	return settings, err
}

// UpdateUserSettings 更新用户设置
func (r *UserSettingsRepository) UpdateUserSettings(settings *UserSettings) error {
	query := `
		UPDATE user_settings 
		SET theme = $1, notification_time = $2, language = $3, timezone = $4, updated_at = $5, sync_version = $6
		WHERE user_id = $7`

	now := time.Now()
	syncVersion := now.UnixMilli()
	settings.UpdatedAt = now
	settings.SyncVersion = syncVersion
	_, err := r.db.Exec(query, settings.Theme, settings.NotificationTime, settings.Language,
		settings.TimeZone, settings.UpdatedAt, settings.SyncVersion, settings.UserID)
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

	now := time.Now()
	syncVersion := now.UnixMilli()

	err = r.db.QueryRow(query, todo.UserID, todo.Title, todo.Description, todo.Completed,
		todo.Priority, todo.DueDate, tagsJSON, todo.CategoryID, todo.Reminder,
		now, now, todo.IsDeleted, syncVersion).Scan(&todo.ID)

	if err == nil {
		todo.CreatedAt = now
		todo.UpdatedAt = now
		todo.SyncVersion = syncVersion
	}

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
		var tagsJSON []byte

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
			&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)
		if err != nil {
			return nil, err
		}

		// 反序列化标签
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &todo.Tags); err != nil {
				todo.Tags = StringSlice{} // 如果解析失败，设置为空切片
			}
		} else {
			todo.Tags = StringSlice{}
		}

		todos = append(todos, todo)
	}

	return todos, rows.Err()
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

	now := time.Now()
	syncVersion := now.UnixMilli()

	result, err := r.db.Exec(query, todo.Title, todo.Description, todo.Completed, todo.Priority,
		todo.DueDate, tagsJSON, todo.CategoryID, todo.Reminder,
		now, syncVersion, todo.ID, todo.UserID)

	if err == nil {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("todo not found or not owned by user")
		}
		todo.UpdatedAt = now
		todo.SyncVersion = syncVersion
	}

	return err
}

// SearchTodos 搜索TODO
func (r *ExtendedTodoRepository) SearchTodos(userID int, keyword string, limit, offset int) ([]Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, priority, due_date, tags,
			category_id, reminder, created_at, updated_at, is_deleted, sync_version
		FROM todos 
		WHERE user_id = $1 AND is_deleted = FALSE 
			AND (title ILIKE $2 OR description ILIKE $3 OR tags::text ILIKE $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6`

	searchPattern := "%" + keyword + "%"
	rows, err := r.db.Query(query, userID, searchPattern, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		var tagsJSON []byte

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
			&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)
		if err != nil {
			return nil, err
		}

		// 反序列化标签
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &todo.Tags); err != nil {
				todo.Tags = StringSlice{}
			}
		} else {
			todo.Tags = StringSlice{}
		}

		todos = append(todos, todo)
	}

	return todos, rows.Err()
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
		var tagsJSON []byte

		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
			&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
			&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)
		if err != nil {
			return nil, err
		}

		// 反序列化标签
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &todo.Tags); err != nil {
				todo.Tags = StringSlice{}
			}
		} else {
			todo.Tags = StringSlice{}
		}

		todos = append(todos, todo)
	}

	return todos, rows.Err()
}

// GetTodoByID 根据ID获取单个TODO
func (r *ExtendedTodoRepository) GetTodoByID(todoID, userID int) (*Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, priority, due_date, tags,
			category_id, reminder, created_at, updated_at, is_deleted, sync_version
		FROM todos 
		WHERE id = $1 AND user_id = $2 AND is_deleted = FALSE`

	var todo Todo
	var tagsJSON []byte

	err := r.db.QueryRow(query, todoID, userID).Scan(
		&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed,
		&todo.Priority, &todo.DueDate, &tagsJSON, &todo.CategoryID, &todo.Reminder,
		&todo.CreatedAt, &todo.UpdatedAt, &todo.IsDeleted, &todo.SyncVersion)

	if err != nil {
		return nil, err
	}

	// 反序列化标签
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &todo.Tags); err != nil {
			todo.Tags = StringSlice{}
		}
	} else {
		todo.Tags = StringSlice{}
	}

	return &todo, nil
}

// ===== 数据同步相关方法 =====

// GetCategoriesSince 获取指定时间戳之后的分类（用于增量同步）
func (r *CategoryRepository) GetCategoriesSince(userID int, since int64) ([]Category, error) {
	query := `
		SELECT id, user_id, name, color, icon, created_at, updated_at, is_deleted, sync_version
		FROM categories 
		WHERE user_id = $1 AND sync_version > $2
		ORDER BY sync_version ASC`

	rows, err := r.db.Query(query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.UserID, &category.Name, &category.Color,
			&category.Icon, &category.CreatedAt, &category.UpdatedAt, &category.IsDeleted, &category.SyncVersion)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// GetUserSettingsSince 获取指定时间戳之后的用户设置（用于增量同步）
func (r *UserSettingsRepository) GetUserSettingsSince(userID int, since int64) (*UserSettings, error) {
	query := `
		SELECT user_id, theme, notification_time, language, timezone, created_at, updated_at, sync_version
		FROM user_settings 
		WHERE user_id = $1 AND sync_version > $2`

	var settings UserSettings
	err := r.db.QueryRow(query, userID, since).Scan(
		&settings.UserID, &settings.Theme, &settings.NotificationTime,
		&settings.Language, &settings.TimeZone, &settings.CreatedAt, &settings.UpdatedAt, &settings.SyncVersion)

	if err == sql.ErrNoRows {
		return nil, nil // 没有更新的设置
	}

	return &settings, err
}

// BatchCreateOrUpdateTodos 批量创建或更新TODO
func (r *ExtendedTodoRepository) BatchCreateOrUpdateTodos(userID int, todos []TodoSyncItem) ([]SyncResult, error) {
	var results []SyncResult

	for _, todoItem := range todos {
		result := SyncResult{
			Type:    "todo",
			LocalID: todoItem.ID,
		}

		// 解析时间字段
		var dueDate, reminder *time.Time
		if todoItem.DueDate != nil {
			if parsed, err := time.Parse(time.RFC3339, *todoItem.DueDate); err == nil {
				dueDate = &parsed
			}
		}
		if todoItem.Reminder != nil {
			if parsed, err := time.Parse(time.RFC3339, *todoItem.Reminder); err == nil {
				reminder = &parsed
			}
		}

		if todoItem.ID == 0 {
			// 创建新TODO
			todo := &Todo{
				UserID:      userID,
				Title:       todoItem.Title,
				Description: todoItem.Description,
				Completed:   todoItem.Completed,
				Priority:    Priority(todoItem.Priority),
				DueDate:     dueDate,
				Tags:        StringSlice(todoItem.Tags),
				CategoryID:  todoItem.CategoryID,
				Reminder:    reminder,
				IsDeleted:   todoItem.IsDeleted,
			}

			if err := r.CreateTodoExtended(todo); err != nil {
				result.Action = "error"
				result.Message = err.Error()
			} else {
				result.Action = "created"
				result.ServerID = todo.ID
				result.SyncVersion = todo.SyncVersion
				result.Message = "创建成功"
			}
		} else {
			// 更新现有TODO
			existingTodo, err := r.GetTodoByID(todoItem.ID, userID)
			if err != nil {
				result.Action = "error"
				result.Message = "TODO不存在"
			} else {
				// 检查冲突
				clientUpdatedAt, _ := time.Parse(time.RFC3339, todoItem.UpdatedAt)
				if existingTodo.UpdatedAt.After(clientUpdatedAt) && existingTodo.SyncVersion > todoItem.SyncVersion {
					result.Action = "conflict"
					result.Message = "存在冲突，服务器版本更新"
				} else {
					// 更新TODO
					existingTodo.Title = todoItem.Title
					existingTodo.Description = todoItem.Description
					existingTodo.Completed = todoItem.Completed
					existingTodo.Priority = Priority(todoItem.Priority)
					existingTodo.DueDate = dueDate
					existingTodo.Tags = StringSlice(todoItem.Tags)
					existingTodo.CategoryID = todoItem.CategoryID
					existingTodo.Reminder = reminder
					existingTodo.IsDeleted = todoItem.IsDeleted

					if err := r.UpdateTodoExtended(existingTodo); err != nil {
						result.Action = "error"
						result.Message = err.Error()
					} else {
						result.Action = "updated"
						result.ServerID = existingTodo.ID
						result.SyncVersion = existingTodo.SyncVersion
						result.Message = "更新成功"
					}
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// BatchCreateOrUpdateCategories 批量创建或更新分类
func (r *CategoryRepository) BatchCreateOrUpdateCategories(userID int, categories []CategorySyncItem) ([]SyncResult, error) {
	var results []SyncResult

	for _, categoryItem := range categories {
		result := SyncResult{
			Type:    "category",
			LocalID: categoryItem.ID,
		}

		if categoryItem.ID == 0 {
			// 创建新分类
			category := &Category{
				UserID:    userID,
				Name:      categoryItem.Name,
				Color:     categoryItem.Color,
				Icon:      categoryItem.Icon,
				IsDeleted: categoryItem.IsDeleted,
			}

			if err := r.CreateCategory(category); err != nil {
				result.Action = "error"
				result.Message = err.Error()
			} else {
				result.Action = "created"
				result.ServerID = category.ID
				result.SyncVersion = category.SyncVersion
				result.Message = "创建成功"
			}
		} else {
			// 更新现有分类
			existingCategory, err := r.GetCategoryByID(categoryItem.ID, userID)
			if err != nil {
				result.Action = "error"
				result.Message = "分类不存在"
			} else {
				// 检查冲突
				clientUpdatedAt, _ := time.Parse(time.RFC3339, categoryItem.UpdatedAt)
				if existingCategory.UpdatedAt.After(clientUpdatedAt) && existingCategory.SyncVersion > categoryItem.SyncVersion {
					result.Action = "conflict"
					result.Message = "存在冲突，服务器版本更新"
				} else {
					// 更新分类
					existingCategory.Name = categoryItem.Name
					existingCategory.Color = categoryItem.Color
					existingCategory.Icon = categoryItem.Icon
					existingCategory.IsDeleted = categoryItem.IsDeleted

					if categoryItem.IsDeleted {
						if err := r.DeleteCategory(categoryItem.ID, userID); err != nil {
							result.Action = "error"
							result.Message = err.Error()
						} else {
							result.Action = "deleted"
							result.ServerID = existingCategory.ID
							result.Message = "删除成功"
						}
					} else {
						if err := r.UpdateCategory(existingCategory); err != nil {
							result.Action = "error"
							result.Message = err.Error()
						} else {
							result.Action = "updated"
							result.ServerID = existingCategory.ID
							result.SyncVersion = existingCategory.SyncVersion
							result.Message = "更新成功"
						}
					}
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// GetCategoryByID 根据ID获取单个分类
func (r *CategoryRepository) GetCategoryByID(categoryID, userID int) (*Category, error) {
	query := `
		SELECT id, user_id, name, color, icon, created_at, updated_at, is_deleted, sync_version
		FROM categories 
		WHERE id = $1 AND user_id = $2`

	var category Category
	err := r.db.QueryRow(query, categoryID, userID).Scan(
		&category.ID, &category.UserID, &category.Name, &category.Color,
		&category.Icon, &category.CreatedAt, &category.UpdatedAt, &category.IsDeleted, &category.SyncVersion)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

// BatchUpdateUserSettings 批量更新用户设置
func (r *UserSettingsRepository) BatchUpdateUserSettings(userID int, settingsItem *UserSettingsSyncItem) (*SyncResult, error) {
	result := &SyncResult{
		Type: "settings",
	}

	if settingsItem == nil {
		result.Action = "error"
		result.Message = "设置数据为空"
		return result, nil
	}

	existingSettings, err := r.GetUserSettings(userID)
	if err != nil {
		result.Action = "error"
		result.Message = "获取用户设置失败"
		return result, err
	}

	// 检查冲突
	clientUpdatedAt, _ := time.Parse(time.RFC3339, settingsItem.UpdatedAt)
	if existingSettings.UpdatedAt.After(clientUpdatedAt) && existingSettings.SyncVersion > settingsItem.SyncVersion {
		result.Action = "conflict"
		result.Message = "存在冲突，服务器版本更新"
		return result, nil
	}

	// 更新设置
	existingSettings.Theme = settingsItem.Theme
	existingSettings.NotificationTime = settingsItem.NotificationTime
	existingSettings.Language = settingsItem.Language
	existingSettings.TimeZone = settingsItem.TimeZone

	if err := r.UpdateUserSettings(existingSettings); err != nil {
		result.Action = "error"
		result.Message = err.Error()
		return result, err
	}

	result.Action = "updated"
	result.SyncVersion = existingSettings.SyncVersion
	result.Message = "更新成功"
	return result, nil
}

// GetCurrentSyncVersion 获取当前最大同步版本号
func GetCurrentSyncVersion(db *sql.DB, userID int) (int64, error) {
	query := `
		SELECT GREATEST(
			COALESCE((SELECT MAX(sync_version) FROM todos WHERE user_id = $1), 0),
			COALESCE((SELECT MAX(sync_version) FROM categories WHERE user_id = $1), 0),
			COALESCE((SELECT sync_version FROM user_settings WHERE user_id = $1), 0)
		) as max_version`

	var maxVersion int64
	err := db.QueryRow(query, userID).Scan(&maxVersion)
	return maxVersion, err
}
