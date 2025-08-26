package main

import (
	"log"
	"net/http"
	"todo-service/global"
	"todo-service/src/api"
	"todo-service/src/repository"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func initRouter(r *gin.Engine) {
	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 添加日志中间件
	r.Use(api.LoggerMiddleware())

	// 公开路由
	r.POST("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.SuccessResponse(gin.H{"message": "Test POST endpoint"}))
	})
	r.POST("/api/register", api.Register)
	r.POST("/api/login", api.Login)

	// 需要认证的路由 - v1 API (保持向后兼容)
	auth := r.Group("/api")
	auth.Use(api.AuthMiddleware())
	{
		auth.POST("/todos/list", api.GetTodos)
		auth.POST("/todos/create", api.CreateTodo)
		auth.POST("/todos/update", api.UpdateTodo)
		auth.POST("/todos/delete", api.DeleteTodo)
		auth.POST("/profile", api.GetProfile)
	}

	// v2 API - 扩展功能
	v2 := r.Group("/api/v2")
	v2.Use(api.AuthMiddleware())
	{
		// 扩展TODO管理
		v2.POST("/todos/list", api.GetTodosExtended)
		v2.POST("/todos/create", api.CreateTodoExtended)
		v2.POST("/todos/update", api.UpdateTodoExtended)
		v2.POST("/todos/search", api.SearchTodos)

		// 分类管理
		v2.POST("/categories", api.GetCategories)
		v2.POST("/categories/create", api.CreateCategory)
		v2.POST("/categories/update", api.UpdateCategory)
		v2.POST("/categories/delete", api.DeleteCategory)

		// 用户设置
		v2.POST("/settings", api.GetUserSettings)
		v2.POST("/settings/update", api.UpdateUserSettings)

		// 数据同步
		v2.POST("/sync/version", api.GetSyncVersion)
		v2.POST("/sync/todos", api.IncrementalSync)
		v2.POST("/sync/batch", api.BatchSync)
	}
}

// @title TODO API
// @version 1.0
// @description TODO服务后端API接口文档

// @contact.name API Support
// @contact.url http://www.swagger.io/support

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:8080
// @BasePath
// @schemes http
func main() {
	// 初始化数据库
	config := repository.GetDatabaseConfig()
	var err error

	global.Db, err = repository.ConnectDatabase(config)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer global.Db.Close()

	// // 创建表
	// repository.CreateTables()

	// 设置路由
	r := gin.Default()
	initRouter(r)

	log.Printf("Server starting on port 8080 with PostgreSQL database")
	r.Run(":8080")
}
