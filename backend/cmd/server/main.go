package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"ynxcb-platform/internal/auth"
	"ynxcb-platform/internal/config"
	"ynxcb-platform/internal/database"
	"ynxcb-platform/internal/router"
)

func main() {
	configPath := flag.String("config", "config.json", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 若配置文件不存在则创建默认配置
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		if err := cfg.Save(*configPath); err != nil {
			log.Fatalf("创建默认配置失败: %v", err)
		}
		log.Printf("已生成默认配置文件 %s，请修改 JWT 密钥和管理员密码", *configPath)
	}

	// 初始化 JWT
	auth.Init(cfg.JWT.Secret)

	// 初始化数据库
	if err := database.Init(cfg.Database.Path); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 应用配置中的管理员账号覆盖默认管理员
	applyAdmin(cfg)

	// 确保上传目录存在
	os.MkdirAll(cfg.Upload.Dir, 0755)

	// 构建路由
	r := router.NewRouter(cfg)
	handler := r

	addr := ":" + cfg.Server.Port
	log.Printf("伊宁县委宣传部部务工作平台已启动，监听 %s", addr)
	log.Printf("数据库: %s", cfg.Database.Path)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// applyAdmin 用配置的管理员信息创建/更新管理员账号
// 密码非空即生效，用户名缺省时使用默认 admin，方便通过 config.json 修改管理员密码
func applyAdmin(cfg *config.Config) {
	if cfg.Admin.Password == "" {
		return
	}
	username := cfg.Admin.Username
	if username == "" {
		username = "admin"
	}
	hash, err := database.HashPassword(cfg.Admin.Password)
	if err != nil {
		log.Printf("加密管理员密码失败: %v", err)
		return
	}
	_, err = database.DB.Exec(
		`INSERT INTO users (username, password_hash, real_name, phone, department_id, role_id, status)
		 VALUES (?, ?, '系统管理员', '', 1, 1, 1)
		 ON CONFLICT(username) DO UPDATE SET password_hash = ?`,
		username, hash, hash)
	if err != nil {
		log.Printf("配置管理员账号失败: %v", err)
	}
}
