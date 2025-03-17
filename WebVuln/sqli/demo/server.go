package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

/*
# 测试SQLi:
$ curl "http://localhost:8080/vuln?id=5"
> No user found

$ curl "http://localhost:8080/vuln?id=5%20OR%201=1"
> User: admin, Password: admin123

# 测试使用参数占位符的函数
curl "http://localhost:8080/safe?id=5%20OR%201=1"
> No user found
*/

// 初始化数据库并创建测试表
func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}

	// 创建 users 表
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            password TEXT NOT NULL
        );
        INSERT OR IGNORE INTO users (id, username, password) VALUES
            (1, 'admin', 'admin123'),
            (2, 'user', 'user456');
    `)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func vulnerableHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		query := fmt.Sprintf("SELECT username, password FROM users WHERE id = %s", id)
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var username, password string
		if rows.Next() {
			err = rows.Scan(&username, &password)
			if err != nil {
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "User: %s, Password: %s\n", username, password)
		} else {
			fmt.Fprintf(w, "No user found\n")
		}
	}
}

// 安全的处理器（使用参数化查询）
func safeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		// 安全：使用参数化查询
		query := "SELECT username, password FROM users WHERE id = ?"
		rows, err := db.Query(query, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var username, password string
		if rows.Next() {
			err = rows.Scan(&username, &password)
			if err != nil {
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "User: %s, Password: %s\n", username, password)
		} else {
			fmt.Fprintf(w, "No user found\n")
		}
	}
}

func main() {
	// 初始化数据库
	db := initDB()
	defer db.Close()

	// 设置路由
	http.HandleFunc("/vuln", vulnerableHandler(db))
	http.HandleFunc("/safe", safeHandler(db))

	// 启动服务器
	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
