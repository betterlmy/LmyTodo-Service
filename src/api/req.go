package api

import "github.com/golang-jwt/jwt/v5"

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

// Claims JWT Claims
type Claims struct {
	UserID   int    `json:"user_id"`  // 用户ID
	Username string `json:"username"` // 用户名
	jwt.RegisteredClaims
}
