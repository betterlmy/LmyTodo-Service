package api

import (
	"todo-service/src/repository"

	"github.com/golang-jwt/jwt/v5"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin" swaggertype:"string" description:"用户名"`      // 用户名
	Password string `json:"password" binding:"required" example:"password123" swaggertype:"string" description:"密码"` // 密码
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required" example:"newuser" swaggertype:"string" description:"用户名"`      // 用户名
	Email    string `json:"email" binding:"required" example:"user@example.com" swaggertype:"string" description:"邮箱"` // 邮箱
	Password string `json:"password" binding:"required" example:"password123" swaggertype:"string" description:"密码"`   // 密码
}

// TodoRequest 创建TODO请求
type TodoRequest struct {
	Title       string `json:"title" binding:"required" example:"学习Go语言" swaggertype:"string" description:"任务标题"` // 任务标题
	Description string `json:"description" example:"学习Go语言基础语法和框架" swaggertype:"string" description:"任务描述"`       // 任务描述
}

// UpdateTodoRequest 更新TODO请求
type UpdateTodoRequest struct {
	ID          int     `json:"id" binding:"required" example:"1" swaggertype:"integer" description:"TODO ID"`      // TODO ID
	Title       *string `json:"title,omitempty" example:"更新后的标题" swaggertype:"string" description:"任务标题（可选）"`       // 任务标题（可选）
	Description *string `json:"description,omitempty" example:"更新后的描述" swaggertype:"string" description:"任务描述（可选）"` // 任务描述（可选）
	Completed   *bool   `json:"completed,omitempty" example:"true" swaggertype:"boolean" description:"是否完成（可选）"`    // 是否完成（可选）
}

// DeleteTodoRequest 删除TODO请求
type DeleteTodoRequest struct {
	ID int `json:"id" binding:"required" example:"1" swaggertype:"integer" description:"TODO ID"` // TODO ID
}

// ExtendedTodoRequest 扩展TODO创建请求
type ExtendedTodoRequest struct {
	Title       string   `json:"title" binding:"required" example:"学习Go语言" swaggertype:"string" description:"任务标题"`
	Description string   `json:"description" example:"学习Go语言基础语法和框架" swaggertype:"string" description:"任务描述"`
	Priority    int      `json:"priority" example:"1" swaggertype:"integer" description:"优先级(0-3)"`
	DueDate     *string  `json:"due_date,omitempty" example:"2023-12-31T23:59:59Z" swaggertype:"string" description:"截止日期"`
	Tags        []string `json:"tags" example:"[\"工作\",\"重要\"]" swaggertype:"array,string" description:"标签"`
	CategoryID  *int     `json:"category_id,omitempty" example:"1" swaggertype:"integer" description:"分类ID"`
	Reminder    *string  `json:"reminder,omitempty" example:"2023-12-30T09:00:00Z" swaggertype:"string" description:"提醒时间"`
}

// UpdateExtendedTodoRequest 扩展TODO更新请求
type UpdateExtendedTodoRequest struct {
	ID          int      `json:"id" binding:"required" example:"1" swaggertype:"integer" description:"TODO ID"`
	Title       *string  `json:"title,omitempty" example:"更新后的标题" swaggertype:"string" description:"任务标题（可选）"`
	Description *string  `json:"description,omitempty" example:"更新后的描述" swaggertype:"string" description:"任务描述（可选）"`
	Completed   *bool    `json:"completed,omitempty" example:"true" swaggertype:"boolean" description:"是否完成（可选）"`
	Priority    *int     `json:"priority,omitempty" example:"2" swaggertype:"integer" description:"优先级（可选）"`
	DueDate     *string  `json:"due_date,omitempty" example:"2023-12-31T23:59:59Z" swaggertype:"string" description:"截止日期（可选）"`
	Tags        []string `json:"tags,omitempty" example:"[\"工作\",\"重要\"]" swaggertype:"array,string" description:"标签（可选）"`
	CategoryID  *int     `json:"category_id,omitempty" example:"1" swaggertype:"integer" description:"分类ID（可选）"`
	Reminder    *string  `json:"reminder,omitempty" example:"2023-12-30T09:00:00Z" swaggertype:"string" description:"提醒时间（可选）"`
}

// CategoryRequest 分类创建/更新请求
type CategoryRequest struct {
	Name  string `json:"name" binding:"required" example:"工作" swaggertype:"string" description:"分类名称"`
	Color string `json:"color" example:"#FF5722" swaggertype:"string" description:"分类颜色"`
	Icon  string `json:"icon" example:"work" swaggertype:"string" description:"分类图标"`
}

// UpdateCategoryRequest 分类更新请求
type UpdateCategoryRequest struct {
	ID    int    `json:"id" binding:"required" example:"1" swaggertype:"integer" description:"分类ID"`
	Name  string `json:"name" binding:"required" example:"工作" swaggertype:"string" description:"分类名称"`
	Color string `json:"color" example:"#FF5722" swaggertype:"string" description:"分类颜色"`
	Icon  string `json:"icon" example:"work" swaggertype:"string" description:"分类图标"`
}

// DeleteCategoryRequest 分类删除请求
type DeleteCategoryRequest struct {
	ID int `json:"id" binding:"required" example:"1" swaggertype:"integer" description:"分类ID"`
}

// UserSettingsRequest 用户设置更新请求
type UserSettingsRequest struct {
	Theme            string `json:"theme" example:"light" swaggertype:"string" description:"主题设置"`
	NotificationTime string `json:"notification_time" example:"09:00" swaggertype:"string" description:"通知时间"`
	Language         string `json:"language" example:"zh-CN" swaggertype:"string" description:"语言设置"`
	TimeZone         string `json:"timezone" example:"Asia/Shanghai" swaggertype:"string" description:"时区设置"`
}

// SearchTodosRequest 搜索TODO请求
type SearchTodosRequest struct {
	Keyword string `json:"keyword" binding:"required" example:"学习" swaggertype:"string" description:"搜索关键词"`
	Limit   int    `json:"limit" example:"20" swaggertype:"integer" description:"返回数量限制"`
	Offset  int    `json:"offset" example:"0" swaggertype:"integer" description:"偏移量"`
}

// GetTodosRequest 获取TODO列表请求
type GetTodosRequest struct {
	Limit  int `json:"limit" example:"20" swaggertype:"integer" description:"返回数量限制"`
	Offset int `json:"offset" example:"0" swaggertype:"integer" description:"偏移量"`
}

// Claims JWT Claims
type Claims struct {
	UserID   int    `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
	jwt.RegisteredClaims
}

// ===== 数据同步相关请求 =====

// IncrementalSyncRequest 增量同步请求
type IncrementalSyncRequest struct {
	Since int64 `json:"since" example:"1640995200000" swaggertype:"integer" description:"同步起始时间戳（毫秒）"`
}

// BatchSyncRequest 批量同步请求
type BatchSyncRequest struct {
	Todos      []repository.TodoSyncItem        `json:"todos,omitempty" description:"待同步的TODO列表"`
	Categories []repository.CategorySyncItem    `json:"categories,omitempty" description:"待同步的分类列表"`
	Settings   *repository.UserSettingsSyncItem `json:"settings,omitempty" description:"待同步的用户设置"`
}
