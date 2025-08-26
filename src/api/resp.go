package api

import (
	"todo-service/src/repository"
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code" example:"0" swaggertype:"integer" description:"响应码，0表示成功"`   // 响应码，0表示成功
	Message string `json:"message" example:"成功" swaggertype:"string" description:"响应消息"`     // 响应消息
	Data    any    `json:"data,omitempty" swaggertype:"object" description:"响应数据，成功时包含具体数据"` // 响应数据，成功时包含具体数据
}

// 错误码定义
const (
	CodeSuccess            = 0     // 成功
	CodeInvalidParams      = 10001 // 参数错误
	CodeUserExists         = 10002 // 用户已存在
	CodeInvalidCredentials = 10003 // 凭据无效
	CodeTokenError         = 10004 // Token错误
	CodeNotFound           = 10005 // 资源不存在
	CodeInternalError      = 10006 // 内部错误
	CodeUnauthorized       = 10007 // 未授权
)

// SuccessResponse 成功响应
func SuccessResponse(data any) Response {
	return Response{
		Code:    CodeSuccess,
		Message: "成功",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// ===== 数据同步相关响应 =====

// SyncResponse 同步响应
type SyncResponse struct {
	Todos         []repository.TodoSyncItem        `json:"todos" description:"TODO同步数据"`
	Categories    []repository.CategorySyncItem    `json:"categories" description:"分类同步数据"`
	Settings      *repository.UserSettingsSyncItem `json:"settings,omitempty" description:"用户设置同步数据"`
	ServerVersion int64                            `json:"server_version" example:"1640995200000" swaggertype:"integer" description:"服务器当前版本号"`
}

// BatchSyncResponse 批量同步响应
type BatchSyncResponse struct {
	Success   []repository.SyncResult `json:"success" description:"成功同步的项目"`
	Conflicts []repository.SyncResult `json:"conflicts" description:"存在冲突的项目"`
	Errors    []repository.SyncResult `json:"errors" description:"同步失败的项目"`
}

// ConflictResolution 冲突解决策略
type ConflictResolution struct {
	Strategy string `json:"strategy" example:"server_wins" swaggertype:"string" description:"解决策略（server_wins/client_wins/merge）"`
	TodoID   int    `json:"todo_id,omitempty" example:"1" swaggertype:"integer" description:"TODO ID"`
}

// SyncVersionResponse 同步版本响应
type SyncVersionResponse struct {
	Version int64 `json:"version" example:"1640995200000" swaggertype:"integer" description:"当前服务器版本号"`
}
