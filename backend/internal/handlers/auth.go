package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"ynxcb-platform/internal/auth"
	"ynxcb-platform/internal/config"
	"ynxcb-platform/internal/database"
	"ynxcb-platform/internal/middleware"
	"ynxcb-platform/internal/models"
)

// Login 用户登录
func Login(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
			return
		}
		if req.Username == "" || req.Password == "" {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "用户名和密码不能为空"})
			return
		}

		var user models.User
		err := database.DB.QueryRow(
			`SELECT u.id, u.username, u.password_hash, u.real_name, u.phone, u.department_id, d.name, u.role_id, r.name, r.code, u.status
			 FROM users u
			 LEFT JOIN departments d ON u.department_id = d.id
			 LEFT JOIN roles r ON u.role_id = r.id
			 WHERE u.username = ?`, req.Username).
			Scan(&user.ID, &user.Username, &user.PasswordHash, &user.RealName, &user.Phone,
				&user.DepartmentID, &user.Department, &user.RoleID, &user.RoleName, &user.RoleCode,
				&user.Status)

		if err == sql.ErrNoRows {
			middleware.JSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
			return
		}
		if err != nil {
			middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "系统错误"})
			return
		}
		if user.Status != 1 {
			middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "账号已被禁用"})
			return
		}
		if !database.CheckPassword(user.PasswordHash, req.Password) {
			middleware.JSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
			return
		}

		// 特殊：管理员首次登录使用配置的默认密码
		if user.Username == cfg.Admin.Username && req.Password == cfg.Admin.Password {
			// 使用配置中的管理员密码登录
			hash, _ := database.HashPassword(cfg.Admin.Password)
			database.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", hash, user.ID)
		}

		token, err := auth.GenerateToken(user.ID, user.Username, user.RealName, user.RoleCode)
		if err != nil {
			middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "生成令牌失败"})
			return
		}

		resp := models.LoginResponse{
			Token: token,
			User:  user,
		}
		resp.User.PasswordHash = ""
		middleware.JSON(w, http.StatusOK, resp)
	}
}

// GetProfile 获取当前用户信息
func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	var user models.User
	err := database.DB.QueryRow(
		`SELECT u.id, u.username, u.real_name, u.phone, u.department_id, d.name, u.role_id, r.name, r.code
		 FROM users u
		 LEFT JOIN departments d ON u.department_id = d.id
		 LEFT JOIN roles r ON u.role_id = r.id
		 WHERE u.id = ?`, userID).
		Scan(&user.ID, &user.Username, &user.RealName, &user.Phone,
			&user.DepartmentID, &user.Department, &user.RoleID, &user.RoleName, &user.RoleCode)
	if err != nil {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	middleware.JSON(w, http.StatusOK, user)
}

// ChangePassword 修改密码
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if len(req.NewPassword) < 6 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "新密码至少6位"})
		return
	}
	var hash string
	err := database.DB.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&hash)
	if err != nil {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	if !database.CheckPassword(hash, req.OldPassword) {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "原密码错误"})
		return
	}
	newHash, _ := database.HashPassword(req.NewPassword)
	database.DB.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", newHash, time.Now(), userID)
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "密码修改成功"})
}

// ListUsers 用户列表（管理员）
func ListUsers(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	query := `SELECT u.id, u.username, u.real_name, u.phone, u.department_id, d.name, u.role_id, r.name, u.status, u.created_at
			  FROM users u
			  LEFT JOIN departments d ON u.department_id = d.id
			  LEFT JOIN roles r ON u.role_id = r.id`
	var rows *sql.Rows
	var err error
	if keyword != "" {
		query += ` WHERE u.real_name LIKE ? OR u.username LIKE ? ORDER BY u.id DESC`
		kw := "%" + keyword + "%"
		rows, err = database.DB.Query(query, kw, kw)
	} else {
		query += ` ORDER BY u.id DESC`
		rows, err = database.DB.Query(query)
	}
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.RealName, &u.Phone, &u.DepartmentID, &u.Department,
			&u.RoleID, &u.RoleName, &u.Status, &u.CreatedAt)
		users = append(users, u)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": users})
}

// CreateUser 创建用户
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Username == "" || req.RealName == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "用户名和姓名必填"})
		return
	}
	password := "123456" // 默认初始密码
	if req.Username != "" {
		// 允许自定义初始密码
		// 这里简单起见使用默认 123456
	}
	hash, _ := database.HashPassword(password)
	_, err := database.DB.Exec(
		"INSERT INTO users (username, password_hash, real_name, phone, department_id, role_id, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		req.Username, hash, req.RealName, req.Phone, req.DepartmentID, req.RoleID, 1)
	if err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "创建失败，用户名可能已存在"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "创建成功，初始密码 123456"})
}

// UpdateUser 更新用户
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.ID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少用户ID"})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE users SET real_name = ?, phone = ?, department_id = ?, role_id = ?, status = ?, updated_at = ? WHERE id = ?",
		req.RealName, req.Phone, req.DepartmentID, req.RoleID, req.Status, time.Now(), req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

// ResetPassword 重置密码
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	hash, _ := database.HashPassword("123456")
	_, err := database.DB.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", hash, time.Now(), req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "重置失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "密码已重置为 123456"})
}

// ListRoles 角色列表
func ListRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, name, code, description FROM roles ORDER BY id")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	roles := []models.Role{}
	for rows.Next() {
		var role models.Role
		rows.Scan(&role.ID, &role.Name, &role.Code, &role.Description)
		roles = append(roles, role)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": roles})
}

// ListDepartments 部门列表
func ListDepartments(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, name, parent_id, sort FROM departments ORDER BY sort")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	depts := []models.Department{}
	for rows.Next() {
		var d models.Department
		rows.Scan(&d.ID, &d.Name, &d.ParentID, &d.Sort)
		depts = append(depts, d)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": depts})
}

// DeleteUser 真删除用户（管理员）
// 从数据库物理删除，释放 ID，同时清理该用户在其他表中的关联数据
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	operatorID, _ := r.Context().Value(middleware.ContextUserID).(int64)

	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少用户ID"})
		return
	}
	// 不能删除自己
	if id == operatorID {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "不能删除当前登录账号"})
		return
	}
	// 检查用户是否存在
	var exists int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE id=?", id).Scan(&exists)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	if exists == 0 {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "用户不存在"})
		return
	}
	// 防止删除最后一个管理员
	var roleCode string
	var adminCount int
	database.DB.QueryRow("SELECT code FROM roles WHERE id=(SELECT role_id FROM users WHERE id=?)", id).Scan(&roleCode)
	if roleCode == "admin" {
		database.DB.QueryRow("SELECT COUNT(*) FROM users u JOIN roles r ON u.role_id=r.id WHERE r.code='admin'").Scan(&adminCount)
		if adminCount <= 1 {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "不能删除最后一个管理员"})
			return
		}
	}

	// 清理该用户的关联数据（物理删除，避免残留）
	tables := []string{
		"attendances", "leave_records", "circulation_records",
		"duty_schedules", "vehicle_applies",
	}
	for _, t := range tables {
		database.DB.Exec("DELETE FROM "+t+" WHERE user_id=?", id)
	}
	database.DB.Exec("DELETE FROM vehicle_applies WHERE reporter_id=?", id)
	database.DB.Exec("DELETE FROM incoming_docs WHERE registrar_id=?", id)

	// 物理删除用户
	_, err = database.DB.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}

	// 重置自增序列，使被删除的最大 ID 可被复用（真删除 + 释放 ID）
	resetAutoIncrement()

	middleware.JSON(w, http.StatusOK, map[string]string{"message": "用户已删除"})
}

// resetAutoIncrement 将自增序列重置为各表当前最大 ID，实现删除后 ID 复用
func resetAutoIncrement() {
	tables := []string{"users", "attendances", "leave_records", "vehicles", "vehicle_applies",
		"contacts", "duty_schedules", "reports", "incoming_docs", "circulation_records",
		"study_materials", "attachments"}
	for _, t := range tables {
		// 将序列设为当前最大 ID（若表为空则为 0）
		database.DB.Exec("DELETE FROM sqlite_sequence WHERE name=?", t)
		database.DB.Exec("INSERT OR REPLACE INTO sqlite_sequence (name, seq) VALUES (?, (SELECT COALESCE(MAX(id),0) FROM "+t+"))", t)
	}
}
