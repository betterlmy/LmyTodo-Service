package repository

import "time"

// User 用户模型
type User struct {
	ID       int    `json:"id" example:"1" swaggertype:"integer" description:"用户ID"`                 // 用户ID
	Username string `json:"username" example:"admin" swaggertype:"string" description:"用户名"`         // 用户名
	Email    string `json:"email" example:"admin@example.com" swaggertype:"string" description:"邮箱"` // 邮箱
	Password string `json:"-" swaggerignore:"true"`                                                  // 密码（不返回给客户端）
}

// Todo TODO任务模型
type Todo struct {
	ID          int       `json:"id" example:"1" swaggertype:"integer" description:"任务ID"`                           // 任务ID
	UserID      int       `json:"user_id" example:"1" swaggertype:"integer" description:"用户ID"`                      // 用户ID
	Title       string    `json:"title" example:"学习Go语言" swaggertype:"string" description:"任务标题"`                    // 任务标题
	Description string    `json:"description" example:"学习Go语言基础语法" swaggertype:"string" description:"任务描述"`          // 任务描述
	Completed   bool      `json:"completed" example:"false" swaggertype:"boolean" description:"是否完成"`                // 是否完成
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"创建时间"` // 创建时间
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z" swaggertype:"string" description:"更新时间"` // 更新时间
	IsDeleted   bool      `json:"is_deleted" example:"false" swaggertype:"boolean" description:"是否删除"`               // 是否删除
}
