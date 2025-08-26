// Package api TODO服务API接口
// @title TODO Service API
// @version 1.0
// @description TODO任务管理服务API文档
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host 127.0.0.1:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token with Bearer prefix
package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo-service/global"
	"todo-service/src/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "注册信息"
// @Success 200 {object} Response{data=map[string]string} "注册成功"
// @Failure 200 {object} Response "注册失败"
// @Router /api/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "密码加密失败"))
		return
	}

	// 插入用户
	_, err = global.Db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		req.Username, req.Email, string(hashedPassword))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusOK, ErrorResponse(CodeUserExists, "用户名或邮箱已存在"))
		} else {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "创建用户失败"))
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "用户创建成功"}))
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT token
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "登录凭据"
// @Success 200 {object} Response{data=map[string]interface{}} "登录成功，返回token和用户信息"
// @Failure 200 {object} Response "登录失败"
// @Router /api/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	defer func() {
		if r := recover(); r != nil {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "服务器错误"))
		}
	}()

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 查找用户
	var user repository.User
	var hashedPassword string
	err := global.Db.QueryRow("SELECT id, username, email, password FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.Email, &hashedPassword)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidCredentials, "账号密码错误"))
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidCredentials, "账号密码错误"))
		return
	}

	// 生成JWT token
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(global.JwtSecret)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "生成token失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{
		"token": tokenString,
		"user":  user,
	}))
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, ErrorResponse(CodeUnauthorized, "缺少Authorization头"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusOK, ErrorResponse(CodeUnauthorized, "需要Bearer token"))
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			return global.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusOK, ErrorResponse(CodeTokenError, "无效的token"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// LoggerMiddleware 统一日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = bodyBytes
				// 重新设置请求体，因为读取后会被消耗
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 记录请求信息
		var requestBodyStr string
		if len(requestBody) > 0 && json.Valid(requestBody) {
			// 如果是有效的JSON，格式化输出
			var jsonData any
			if err := json.Unmarshal(requestBody, &jsonData); err == nil {
				if formattedJSON, err := json.Marshal(jsonData); err == nil {
					requestBodyStr = string(formattedJSON)
				}
			}
		} else if len(requestBody) > 0 {
			requestBodyStr = string(requestBody)
		} else {
			requestBodyStr = "empty"
		}

		log.Printf("[REQUEST] %s %s | Body: %s | IP: %s | UserAgent: %s",
			c.Request.Method,
			c.Request.URL.Path,
			requestBodyStr,
			c.ClientIP(),
			c.Request.UserAgent(),
		)

		// 处理请求
		c.Next()

		// 记录响应信息
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// 获取用户信息（如果有的话）
		userID, userExists := c.Get("userID")
		username, usernameExists := c.Get("username")

		var userInfo string
		if userExists && usernameExists {
			userInfo = " | User: " + username.(string) + " (ID:" + strconv.Itoa(userID.(int)) + ")"
		} else {
			userInfo = " | User: anonymous"
		}

		log.Printf("[RESPONSE] %s %s | Status: %d | Duration: %v%s",
			c.Request.Method,
			c.Request.URL.Path,
			statusCode,
			duration,
			userInfo,
		)
	}
}

// GetTodos 获取TODO列表
// @Summary 获取用户的TODO列表
// @Description 获取当前用户的所有TODO任务，按创建时间倒序排列
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=[]repository.Todo} "获取成功"
// @Failure 200 {object} Response "获取失败"
// @Router /api/todos/list [post]
func GetTodos(c *gin.Context) {
	userID := c.GetInt("userID")

	rows, err := global.Db.Query(`
		SELECT id, title, description, completed, created_at, updated_at 
		FROM todos WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取TODO列表失败"))
		return
	}
	defer rows.Close()

	var todos []repository.Todo
	for rows.Next() {
		var todo repository.Todo
		todo.UserID = userID
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "解析TODO数据失败"))
			return
		}
		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, SuccessResponse(todos))
}

// CreateTodo 创建TODO
// @Summary 创建新的TODO任务
// @Description 为当前用户创建一个新的TODO任务
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body TodoRequest true "TODO信息"
// @Success 200 {object} Response{data=repository.Todo} "创建成功"
// @Failure 200 {object} Response "创建失败"
// @Router /api/todos/create [post]
func CreateTodo(c *gin.Context) {
	userID := c.GetInt("userID")
	var req TodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	result, err := global.Db.Exec(`
		INSERT INTO todos (user_id, title, description) 
		VALUES (?, ?, ?)`, userID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "创建TODO失败"))
		return
	}

	id, _ := result.LastInsertId()
	todo := repository.Todo{
		ID:          int(id),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	c.JSON(http.StatusOK, SuccessResponse(todo))
}

// UpdateTodo 更新TODO
// @Summary 更新TODO任务
// @Description 更新指定的TODO任务信息，支持部分字段更新
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body UpdateTodoRequest true "更新信息"
// @Success 200 {object} Response{data=map[string]string} "更新成功"
// @Failure 200 {object} Response "更新失败"
// @Router /api/todos/update [post]
func UpdateTodo(c *gin.Context) {
	userID := c.GetInt("userID")
	var req UpdateTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 构建动态更新查询
	updates := []string{}
	args := []any{}

	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Completed != nil {
		updates = append(updates, "completed = ?")
		args = append(args, *req.Completed)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "没有要更新的字段"))
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID, req.ID)

	query := "UPDATE todos SET " + strings.Join(updates, ", ") + " WHERE user_id = ? AND id = ?"
	result, err := global.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "更新TODO失败"))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusOK, ErrorResponse(CodeNotFound, "TODO不存在"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "TODO更新成功"}))
}

// DeleteTodo 删除TODO
// @Summary 删除TODO任务
// @Description 删除指定的TODO任务
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body DeleteTodoRequest true "删除信息"
// @Success 200 {object} Response{data=map[string]string} "删除成功"
// @Failure 200 {object} Response "删除失败"
// @Router /api/todos/delete [post]
func DeleteTodo(c *gin.Context) {
	userID := c.GetInt("userID")
	var req DeleteTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	result, err := global.Db.Exec("DELETE FROM todos WHERE user_id = ? AND id = ?", userID, req.ID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "删除TODO失败"))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusOK, ErrorResponse(CodeNotFound, "TODO不存在"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "TODO删除成功"}))
}

// GetProfile 获取用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的基本信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=repository.User} "获取成功"
// @Failure 200 {object} Response "获取失败"
// @Router /api/profile [post]
func GetProfile(c *gin.Context) {
	userID := c.GetInt("userID")

	var user repository.User
	err := global.Db.QueryRow("SELECT id, username, email FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeNotFound, "用户不存在"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(user))
}

// ===== 扩展API接口 =====

// GetTodosExtended 获取扩展TODO列表
// @Summary 获取用户的扩展TODO列表
// @Description 获取当前用户的所有TODO任务，支持分页和扩展字段
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GetTodosRequest true "分页参数"
// @Success 200 {object} Response{data=[]repository.Todo} "获取成功"
// @Failure 200 {object} Response "获取失败"
// @Router /api/v2/todos/list [post]
func GetTodosExtended(c *gin.Context) {
	userID := c.GetInt("userID")
	var req GetTodosRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	repo := repository.NewExtendedTodoRepository()
	todos, err := repo.GetTodosByUserIDExtended(userID, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取TODO列表失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(todos))
}

// CreateTodoExtended 创建扩展TODO
// @Summary 创建新的扩展TODO任务
// @Description 为当前用户创建一个新的TODO任务，支持优先级、标签等扩展字段
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body ExtendedTodoRequest true "TODO信息"
// @Success 200 {object} Response{data=repository.Todo} "创建成功"
// @Failure 200 {object} Response "创建失败"
// @Router /api/v2/todos/create [post]
func CreateTodoExtended(c *gin.Context) {
	userID := c.GetInt("userID")
	var req ExtendedTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 验证优先级范围
	if req.Priority < 0 || req.Priority > 3 {
		req.Priority = 0 // 默认低优先级
	}

	todo := &repository.Todo{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    repository.Priority(req.Priority),
		Tags:        repository.StringSlice(req.Tags),
		CategoryID:  req.CategoryID,
		Completed:   false,
		IsDeleted:   false,
	}

	// 解析时间字段
	if req.DueDate != nil {
		if dueDate, err := time.Parse(time.RFC3339, *req.DueDate); err == nil {
			todo.DueDate = &dueDate
		}
	}
	if req.Reminder != nil {
		if reminder, err := time.Parse(time.RFC3339, *req.Reminder); err == nil {
			todo.Reminder = &reminder
		}
	}

	repo := repository.NewExtendedTodoRepository()
	if err := repo.CreateTodoExtended(todo); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "创建TODO失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(todo))
}

// UpdateTodoExtended 更新扩展TODO
// @Summary 更新扩展TODO任务
// @Description 更新指定的TODO任务信息，支持扩展字段的部分更新
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param todo body UpdateExtendedTodoRequest true "更新信息"
// @Success 200 {object} Response{data=map[string]string} "更新成功"
// @Failure 200 {object} Response "更新失败"
// @Router /api/v2/todos/update [post]
func UpdateTodoExtended(c *gin.Context) {
	userID := c.GetInt("userID")
	var req UpdateExtendedTodoRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 首先获取现有的TODO
	repo := repository.NewExtendedTodoRepository()
	todo, err := repo.GetTodoByID(req.ID, userID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeNotFound, "TODO不存在"))
		return
	}

	// 更新字段
	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}
	if req.Priority != nil {
		if *req.Priority >= 0 && *req.Priority <= 3 {
			todo.Priority = repository.Priority(*req.Priority)
		}
	}
	if req.Tags != nil {
		todo.Tags = repository.StringSlice(req.Tags)
	}
	if req.CategoryID != nil {
		todo.CategoryID = req.CategoryID
	}

	// 解析时间字段
	if req.DueDate != nil {
		if dueDate, err := time.Parse(time.RFC3339, *req.DueDate); err == nil {
			todo.DueDate = &dueDate
		} else {
			todo.DueDate = nil
		}
	}
	if req.Reminder != nil {
		if reminder, err := time.Parse(time.RFC3339, *req.Reminder); err == nil {
			todo.Reminder = &reminder
		} else {
			todo.Reminder = nil
		}
	}

	if err := repo.UpdateTodoExtended(todo); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "更新TODO失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "TODO更新成功"}))
}

// SearchTodos 搜索TODO
// @Summary 搜索TODO任务
// @Description 根据关键词搜索用户的TODO任务，支持标题和描述搜索
// @Tags TODO管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param search body SearchTodosRequest true "搜索参数"
// @Success 200 {object} Response{data=[]repository.Todo} "搜索成功"
// @Failure 200 {object} Response "搜索失败"
// @Router /api/v2/todos/search [post]
func SearchTodos(c *gin.Context) {
	userID := c.GetInt("userID")
	var req SearchTodosRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	repo := repository.NewExtendedTodoRepository()
	todos, err := repo.SearchTodos(userID, req.Keyword, req.Limit, req.Offset)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "搜索TODO失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(todos))
}

// ===== 分类管理API =====

// GetCategories 获取分类列表
// @Summary 获取用户的分类列表
// @Description 获取当前用户的所有分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=[]repository.Category} "获取成功"
// @Failure 200 {object} Response "获取失败"
// @Router /api/v2/categories [post]
func GetCategories(c *gin.Context) {
	userID := c.GetInt("userID")

	repo := repository.NewCategoryRepository()
	categories, err := repo.GetCategoriesByUserID(userID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取分类列表失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(categories))
}

// CreateCategory 创建分类
// @Summary 创建新分类
// @Description 为当前用户创建一个新的分类
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body CategoryRequest true "分类信息"
// @Success 200 {object} Response{data=repository.Category} "创建成功"
// @Failure 200 {object} Response "创建失败"
// @Router /api/v2/categories/create [post]
func CreateCategory(c *gin.Context) {
	userID := c.GetInt("userID")
	var req CategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 设置默认值
	if req.Color == "" {
		req.Color = "#2196F3"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	category := &repository.Category{
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
		Icon:   req.Icon,
	}

	repo := repository.NewCategoryRepository()
	if err := repo.CreateCategory(category); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "分类名称已存在"))
		} else {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "创建分类失败"))
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(category))
}

// UpdateCategory 更新分类
// @Summary 更新分类
// @Description 更新指定的分类信息
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body UpdateCategoryRequest true "更新信息"
// @Success 200 {object} Response{data=map[string]string} "更新成功"
// @Failure 200 {object} Response "更新失败"
// @Router /api/v2/categories/update [post]
func UpdateCategory(c *gin.Context) {
	userID := c.GetInt("userID")
	var req UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	category := &repository.Category{
		ID:     req.ID,
		UserID: userID,
		Name:   req.Name,
		Color:  req.Color,
		Icon:   req.Icon,
	}

	repo := repository.NewCategoryRepository()
	if err := repo.UpdateCategory(category); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") || strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "分类名称已存在"))
		} else {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "更新分类失败"))
		}
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "分类更新成功"}))
}

// DeleteCategory 删除分类
// @Summary 删除分类
// @Description 删除指定的分类（软删除）
// @Tags 分类管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body DeleteCategoryRequest true "删除信息"
// @Success 200 {object} Response{data=map[string]string} "删除成功"
// @Failure 200 {object} Response "删除失败"
// @Router /api/v2/categories/delete [post]
func DeleteCategory(c *gin.Context) {
	userID := c.GetInt("userID")
	var req DeleteCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	repo := repository.NewCategoryRepository()
	if err := repo.DeleteCategory(req.ID, userID); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "删除分类失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "分类删除成功"}))
}

// ===== 用户设置API =====

// GetUserSettings 获取用户设置
// @Summary 获取用户设置
// @Description 获取当前用户的个性化设置
// @Tags 用户设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=repository.UserSettings} "获取成功"
// @Failure 200 {object} Response "获取失败"
// @Router /api/v2/settings [post]
func GetUserSettings(c *gin.Context) {
	userID := c.GetInt("userID")

	repo := repository.NewUserSettingsRepository()
	settings, err := repo.GetUserSettings(userID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取用户设置失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(settings))
}

// UpdateUserSettings 更新用户设置
// @Summary 更新用户设置
// @Description 更新当前用户的个性化设置
// @Tags 用户设置
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body UserSettingsRequest true "设置信息"
// @Success 200 {object} Response{data=map[string]string} "更新成功"
// @Failure 200 {object} Response "更新失败"
// @Router /api/v2/settings/update [post]
func UpdateUserSettings(c *gin.Context) {
	userID := c.GetInt("userID")
	var req UserSettingsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 验证主题设置
	if req.Theme != "" && req.Theme != "light" && req.Theme != "dark" && req.Theme != "auto" {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "无效的主题设置"))
		return
	}

	settings := &repository.UserSettings{
		UserID:           userID,
		Theme:            req.Theme,
		NotificationTime: req.NotificationTime,
		Language:         req.Language,
		TimeZone:         req.TimeZone,
	}

	repo := repository.NewUserSettingsRepository()
	if err := repo.UpdateUserSettings(settings); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "更新用户设置失败"))
		return
	}

	c.JSON(http.StatusOK, SuccessResponse(gin.H{"message": "用户设置更新成功"}))
}

// ===== 数据同步API =====

// GetSyncVersion 获取同步版本
// @Summary 获取当前同步版本号
// @Description 获取服务器当前的同步版本号，用于增量同步
// @Tags 数据同步
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=SyncVersionResponse} "获取成功"
// @Failure 200 {object} Response "获取失败"
// @Router /api/v2/sync/version [post]
func GetSyncVersion(c *gin.Context) {
	userID := c.GetInt("userID")

	version, err := repository.GetCurrentSyncVersion(global.Db, userID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取同步版本失败"))
		return
	}

	response := SyncVersionResponse{
		Version: version,
	}

	c.JSON(http.StatusOK, SuccessResponse(response))
}

// IncrementalSync 增量同步
// @Summary 增量同步数据
// @Description 基于时间戳获取增量更新的数据
// @Tags 数据同步
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body IncrementalSyncRequest true "同步参数"
// @Success 200 {object} Response{data=SyncResponse} "同步成功"
// @Failure 200 {object} Response "同步失败"
// @Router /api/v2/sync/todos [post]
func IncrementalSync(c *gin.Context) {
	userID := c.GetInt("userID")
	var req IncrementalSyncRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	// 获取增量TODO数据
	todoRepo := repository.NewExtendedTodoRepository()
	todos, err := todoRepo.GetTodosSince(userID, req.Since)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取TODO增量数据失败"))
		return
	}

	// 获取增量分类数据
	categoryRepo := repository.NewCategoryRepository()
	categories, err := categoryRepo.GetCategoriesSince(userID, req.Since)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取分类增量数据失败"))
		return
	}

	// 获取增量用户设置数据
	settingsRepo := repository.NewUserSettingsRepository()
	settings, err := settingsRepo.GetUserSettingsSince(userID, req.Since)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取用户设置增量数据失败"))
		return
	}

	// 获取当前服务器版本
	serverVersion, err := repository.GetCurrentSyncVersion(global.Db, userID)
	if err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "获取服务器版本失败"))
		return
	}

	// 转换为同步格式
	var todoSyncItems []repository.TodoSyncItem
	for _, todo := range todos {
		item := repository.TodoSyncItem{
			ID:          todo.ID,
			Title:       todo.Title,
			Description: todo.Description,
			Completed:   todo.Completed,
			Priority:    int(todo.Priority),
			Tags:        []string(todo.Tags),
			CategoryID:  todo.CategoryID,
			IsDeleted:   todo.IsDeleted,
			SyncVersion: todo.SyncVersion,
			UpdatedAt:   todo.UpdatedAt.Format(time.RFC3339),
		}
		if todo.DueDate != nil {
			dueDateStr := todo.DueDate.Format(time.RFC3339)
			item.DueDate = &dueDateStr
		}
		if todo.Reminder != nil {
			reminderStr := todo.Reminder.Format(time.RFC3339)
			item.Reminder = &reminderStr
		}
		todoSyncItems = append(todoSyncItems, item)
	}

	var categorySyncItems []repository.CategorySyncItem
	for _, category := range categories {
		item := repository.CategorySyncItem{
			ID:          category.ID,
			Name:        category.Name,
			Color:       category.Color,
			Icon:        category.Icon,
			IsDeleted:   category.IsDeleted,
			SyncVersion: category.SyncVersion,
			UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
		}
		categorySyncItems = append(categorySyncItems, item)
	}

	var settingsSyncItem *repository.UserSettingsSyncItem
	if settings != nil {
		settingsSyncItem = &repository.UserSettingsSyncItem{
			Theme:            settings.Theme,
			NotificationTime: settings.NotificationTime,
			Language:         settings.Language,
			TimeZone:         settings.TimeZone,
			SyncVersion:      settings.SyncVersion,
			UpdatedAt:        settings.UpdatedAt.Format(time.RFC3339),
		}
	}

	response := SyncResponse{
		Todos:         todoSyncItems,
		Categories:    categorySyncItems,
		Settings:      settingsSyncItem,
		ServerVersion: serverVersion,
	}

	c.JSON(http.StatusOK, SuccessResponse(response))
}

// BatchSync 批量同步
// @Summary 批量同步数据
// @Description 批量上传客户端数据并处理冲突
// @Tags 数据同步
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchSyncRequest true "批量同步数据"
// @Success 200 {object} Response{data=BatchSyncResponse} "同步成功"
// @Failure 200 {object} Response "同步失败"
// @Router /api/v2/sync/batch [post]
func BatchSync(c *gin.Context) {
	userID := c.GetInt("userID")
	var req BatchSyncRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ErrorResponse(CodeInvalidParams, "参数错误: "+err.Error()))
		return
	}

	var allResults []repository.SyncResult
	var successResults []repository.SyncResult
	var conflictResults []repository.SyncResult
	var errorResults []repository.SyncResult

	// 处理TODO同步
	if len(req.Todos) > 0 {
		todoRepo := repository.NewExtendedTodoRepository()
		todoResults, err := todoRepo.BatchCreateOrUpdateTodos(userID, req.Todos)
		if err != nil {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "批量同步TODO失败"))
			return
		}
		allResults = append(allResults, todoResults...)
	}

	// 处理分类同步
	if len(req.Categories) > 0 {
		categoryRepo := repository.NewCategoryRepository()
		categoryResults, err := categoryRepo.BatchCreateOrUpdateCategories(userID, req.Categories)
		if err != nil {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "批量同步分类失败"))
			return
		}
		allResults = append(allResults, categoryResults...)
	}

	// 处理用户设置同步
	if req.Settings != nil {
		settingsRepo := repository.NewUserSettingsRepository()
		settingsResult, err := settingsRepo.BatchUpdateUserSettings(userID, req.Settings)
		if err != nil {
			c.JSON(http.StatusOK, ErrorResponse(CodeInternalError, "批量同步用户设置失败"))
			return
		}
		allResults = append(allResults, *settingsResult)
	}

	// 分类结果
	for _, result := range allResults {
		switch result.Action {
		case "created", "updated", "deleted":
			successResults = append(successResults, result)
		case "conflict":
			conflictResults = append(conflictResults, result)
		case "error":
			errorResults = append(errorResults, result)
		}
	}

	response := BatchSyncResponse{
		Success:   successResults,
		Conflicts: conflictResults,
		Errors:    errorResults,
	}

	c.JSON(http.StatusOK, SuccessResponse(response))
}
