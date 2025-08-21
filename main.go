package main

import (
	"database/sql"
	"log"
	"todo-service/global"
	"todo-service/src/api"
	"todo-service/src/repository"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func initRouter(r *gin.Engine) {
	// CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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
	r.POST("/api/register", api.Register)
	r.POST("/api/login", api.Login)

	// 需要认证的路由
	auth := r.Group("/api")
	auth.Use(api.AuthMiddleware())
	{
		auth.GET("/todos", api.GetTodos)
		auth.POST("/todos", api.CreateTodo)
		auth.PUT("/todos/:id", api.UpdateTodo)
		auth.DELETE("/todos/:id", api.DeleteTodo)
		auth.GET("/profile", api.GetProfile)
	}
}

func main() {
	// 初始化数据库
	var err error
	global.Db, err = sql.Open("sqlite3", "db/todo.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer global.Db.Close()

	// 创建表
	repository.CreateTables()

	// 设置路由
	r := gin.Default()
	initRouter(r)

	r.Run(":8080")
}
