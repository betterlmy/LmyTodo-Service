package global

import "database/sql"

var Db *sql.DB
var JwtSecret = []byte("your-secret-key-here") // 在生产环境中应该使用环境变量
