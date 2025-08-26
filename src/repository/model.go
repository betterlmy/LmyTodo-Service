package repository

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Priority 优先级枚举
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityMedium:
		return "medium"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return "low"
	}
}

// Value 实现 driver.Valuer 接口，用于数据库存储
func (p Priority) Value() (driver.Value, error) {
	return int64(p), nil
}

// Scan 实现 sql.Scanner 接口，用于数据库读取
func (p *Priority) Scan(value interface{}) error {
	if value == nil {
		*p = PriorityLow
		return nil
	}
	if v, ok := value.(int64); ok {
		*p = Priority(v)
		return nil
	}
	return fmt.Errorf("cannot scan %T into Priority", value)
}

// StringSlice 字符串切片类型，用于存储标签
type StringSlice []string

// Value 实现 driver.Valuer 接口
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}

	return json.Unmarshal(bytes, s)
}

// User 用户模型
type User struct {
	ID        int       `json:"id" example:"1" swaggertype:"integer" description:"用户ID"`                           // 用户ID
	Username  string    `json:"username" example:"admin" swaggertype:"string" description:"用户名"`                   // 用户名
	Email     string    `json:"email" example:"admin@example.com" swaggertype:"string" description:"邮箱"`           // 邮箱
	Password  string    `json:"-" swaggerignore:"true"`                                                            // 密码（不返回给客户端）
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"创建时间"` // 创建时间
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"更新时间"` // 更新时间
}

// Category 分类模型
type Category struct {
	ID          int       `json:"id" example:"1" swaggertype:"integer" description:"分类ID"`                           // 分类ID
	UserID      int       `json:"user_id" example:"1" swaggertype:"integer" description:"用户ID"`                      // 用户ID
	Name        string    `json:"name" example:"工作" swaggertype:"string" description:"分类名称"`                         // 分类名称
	Color       string    `json:"color" example:"#FF5722" swaggertype:"string" description:"分类颜色"`                   // 分类颜色
	Icon        string    `json:"icon" example:"work" swaggertype:"string" description:"分类图标"`                       // 分类图标
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"创建时间"` // 创建时间
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"更新时间"` // 更新时间
	IsDeleted   bool      `json:"is_deleted" example:"false" swaggertype:"boolean" description:"是否删除"`               // 是否删除
	SyncVersion int64     `json:"sync_version" example:"1640995200000" swaggertype:"integer" description:"同步版本号"`    // 同步版本号
}

// UserSettings 用户设置模型
type UserSettings struct {
	UserID           int       `json:"user_id" example:"1" swaggertype:"integer" description:"用户ID"`                      // 用户ID
	Theme            string    `json:"theme" example:"light" swaggertype:"string" description:"主题设置"`                     // 主题设置 (light/dark/auto)
	NotificationTime string    `json:"notification_time" example:"09:00" swaggertype:"string" description:"通知时间"`         // 通知时间
	Language         string    `json:"language" example:"zh-CN" swaggertype:"string" description:"语言设置"`                  // 语言设置
	TimeZone         string    `json:"timezone" example:"Asia/Shanghai" swaggertype:"string" description:"时区设置"`          // 时区设置
	CreatedAt        time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"创建时间"` // 创建时间
	UpdatedAt        time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"更新时间"` // 更新时间
	SyncVersion      int64     `json:"sync_version" example:"1640995200000" swaggertype:"integer" description:"同步版本号"`    // 同步版本号
}

// Todo TODO任务模型（扩展版）
type Todo struct {
	ID          int         `json:"id" example:"1" swaggertype:"integer" description:"任务ID"`                                   // 任务ID
	UserID      int         `json:"user_id" example:"1" swaggertype:"integer" description:"用户ID"`                              // 用户ID
	Title       string      `json:"title" example:"学习Go语言" swaggertype:"string" description:"任务标题"`                            // 任务标题
	Description string      `json:"description" example:"学习Go语言基础语法" swaggertype:"string" description:"任务描述"`                  // 任务描述
	Completed   bool        `json:"completed" example:"false" swaggertype:"boolean" description:"是否完成"`                        // 是否完成
	Priority    Priority    `json:"priority" example:"1" swaggertype:"integer" description:"优先级"`                              // 优先级
	DueDate     *time.Time  `json:"due_date,omitempty" example:"2023-12-31T23:59:59Z" swaggertype:"string" description:"截止日期"` // 截止日期
	Tags        StringSlice `json:"tags" example:"[\"工作\",\"重要\"]" swaggertype:"array,string" description:"标签"`                // 标签
	CategoryID  *int        `json:"category_id,omitempty" example:"1" swaggertype:"integer" description:"分类ID"`                // 分类ID
	Reminder    *time.Time  `json:"reminder,omitempty" example:"2023-12-30T09:00:00Z" swaggertype:"string" description:"提醒时间"` // 提醒时间
	CreatedAt   time.Time   `json:"created_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"创建时间"`         // 创建时间
	UpdatedAt   time.Time   `json:"updated_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"更新时间"`         // 更新时间
	IsDeleted   bool        `json:"is_deleted" example:"false" swaggertype:"boolean" description:"是否删除"`                       // 是否删除
	SyncVersion int64       `json:"sync_version" example:"1640995200000" swaggertype:"integer" description:"同步版本号"`            // 同步版本号
}
