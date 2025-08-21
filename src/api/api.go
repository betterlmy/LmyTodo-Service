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

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 插入用户
	_, err = global.Db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		req.Username, req.Email, string(hashedPassword))
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	time.Sleep(time.Second)

	// 查找用户
	var user repository.User
	var hashedPassword string
	err := global.Db.QueryRow("SELECT id, username, email, password FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.Email, &hashedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return global.JwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
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
			var jsonData interface{}
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

func GetTodos(c *gin.Context) {
	userID := c.GetInt("userID")

	rows, err := global.Db.Query(`
		SELECT id, title, description, completed, created_at, updated_at 
		FROM todos WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch todos"})
		return
	}
	defer rows.Close()

	var todos []repository.Todo
	for rows.Next() {
		var todo repository.Todo
		todo.UserID = userID
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan todo"})
			return
		}
		todos = append(todos, todo)
	}

	c.JSON(http.StatusOK, todos)
}

func CreateTodo(c *gin.Context) {
	userID := c.GetInt("userID")
	var req TodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := global.Db.Exec(`
		INSERT INTO todos (user_id, title, description) 
		VALUES (?, ?, ?)`, userID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create todo"})
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

	c.JSON(http.StatusCreated, todo)
}

func UpdateTodo(c *gin.Context) {
	userID := c.GetInt("userID")
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Completed   *bool   `json:"completed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建动态更新查询
	updates := []string{}
	args := []interface{}{}

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, userID, todoID)

	query := "UPDATE todos SET " + strings.Join(updates, ", ") + " WHERE user_id = ? AND id = ?"
	result, err := global.Db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update todo"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo updated successfully"})
}

func DeleteTodo(c *gin.Context) {
	userID := c.GetInt("userID")
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid todo ID"})
		return
	}

	result, err := global.Db.Exec("DELETE FROM todos WHERE user_id = ? AND id = ?", userID, todoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete todo"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

func GetProfile(c *gin.Context) {
	userID := c.GetInt("userID")

	var user repository.User
	err := global.Db.QueryRow("SELECT id, username, email FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
