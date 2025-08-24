package api

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
