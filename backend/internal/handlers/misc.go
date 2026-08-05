package handlers

import (
	"encoding/json"
	"net/http"

	"ynxcb-platform/internal/database"
	"ynxcb-platform/internal/middleware"
	"ynxcb-platform/internal/models"
)

// ListContacts 通讯录（分页）
func ListContacts(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	departmentID := r.URL.Query().Get("department_id")

	where := ` WHERE 1=1`
	args := []interface{}{}
	if keyword != "" {
		where += ` AND (c.name LIKE ? OR c.position LIKE ? OR c.phone LIKE ?)`
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if departmentID != "" {
		where += ` AND c.department_id = ?`
		args = append(args, departmentID)
	}

	p := parsePage(r)

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM contacts c"+where, args...).Scan(&total)

	query := `SELECT c.id, c.name, c.phone, c.department_id, d.name, c.position, c.is_public, c.sort
		FROM contacts c LEFT JOIN departments d ON c.department_id = d.id` + where +
		` ORDER BY c.sort, c.id LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	contacts := []models.Contact{}
	for rows.Next() {
		var c models.Contact
		rows.Scan(&c.ID, &c.Name, &c.Phone, &c.DepartmentID, &c.Department, &c.Position, &c.IsPublic, &c.Sort)
		contacts = append(contacts, c)
	}
	middleware.JSON(w, http.StatusOK, paginateResult(contacts, total, p.Page, p.PageSize))
}

// CreateContact 添加通讯录
func CreateContact(w http.ResponseWriter, r *http.Request) {
	var req models.Contact
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Name == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "姓名必填"})
		return
	}
	_, err := database.DB.Exec(
		"INSERT INTO contacts (name, phone, department_id, position, is_public, sort) VALUES (?, ?, ?, ?, ?, ?)",
		req.Name, req.Phone, req.DepartmentID, req.Position, req.IsPublic, req.Sort)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "添加成功"})
}

// UpdateContact 更新通讯录
func UpdateContact(w http.ResponseWriter, r *http.Request) {
	var req models.Contact
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.ID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE contacts SET name=?, phone=?, department_id=?, position=?, is_public=?, sort=? WHERE id=?",
		req.Name, req.Phone, req.DepartmentID, req.Position, req.IsPublic, req.Sort, req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

// DeleteContact 删除通讯录
func DeleteContact(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec("DELETE FROM contacts WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// ListDutySchedules 排班列表
func ListDutySchedules(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month") // YYYY-MM
	userID := r.URL.Query().Get("user_id")

	query := `SELECT s.id, s.duty_date, s.user_id, u.real_name, s.is_dawangyuan, s.note, s.status
		FROM duty_schedules s LEFT JOIN users u ON s.user_id = u.id WHERE 1=1`
	args := []interface{}{}
	if month != "" {
		query += ` AND s.duty_date LIKE ?`
		args = append(args, month+"%")
	}
	if userID != "" {
		query += ` AND s.user_id = ?`
		args = append(args, userID)
	}
	query += ` ORDER BY s.duty_date`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	schedules := []models.DutySchedule{}
	for rows.Next() {
		var s models.DutySchedule
		rows.Scan(&s.ID, &s.DutyDate, &s.UserID, &s.UserName, &s.IsDaWangYuan, &s.Note, &s.Status)
		schedules = append(schedules, s)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": schedules})
}

// SaveDutySchedule 保存排班（按日期+人员 upsert，支持一天多人）
func SaveDutySchedule(w http.ResponseWriter, r *http.Request) {
	var req models.DutySchedule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.DutyDate == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "日期必填"})
		return
	}
	if req.UserID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请选择值守人员"})
		return
	}
	_, err := database.DB.Exec(
		`INSERT INTO duty_schedules (duty_date, user_id, is_dawangyuan, note, status) VALUES (?, ?, ?, ?, 1)
		 ON CONFLICT(duty_date, user_id) DO UPDATE SET is_dawangyuan=?, note=?`,
		req.DutyDate, req.UserID, req.IsDaWangYuan, req.Note,
		req.IsDaWangYuan, req.Note)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "保存成功"})
}

// DeleteDutySchedule 删除排班
func DeleteDutySchedule(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec("DELETE FROM duty_schedules WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// ListReports 报表列表（分页）
func ListReports(w http.ResponseWriter, r *http.Request) {
	reportType := r.URL.Query().Get("report_type")
	status := r.URL.Query().Get("status")
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)

	where := ` WHERE 1=1`
	args := []interface{}{}
	if reportType != "" {
		where += ` AND r.report_type = ?`
		args = append(args, reportType)
	}
	if status != "" {
		where += ` AND r.status = ?`
		args = append(args, status)
	}
	// 非管理员只看到自己提交的
	if roleCode != "admin" {
		where += ` AND r.submitter_id = ?`
		args = append(args, userID)
	}

	p := parsePage(r)

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM reports r"+where, args...).Scan(&total)

	query := `SELECT r.id, r.report_type, r.title, r.period, r.submitter_id, u.real_name, r.status, r.created_at, r.updated_at
		FROM reports r LEFT JOIN users u ON r.submitter_id = u.id` + where +
		` ORDER BY r.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	reports := []models.Report{}
	for rows.Next() {
		var rep models.Report
		rows.Scan(&rep.ID, &rep.ReportType, &rep.Title, &rep.Period, &rep.SubmitterID,
			&rep.Submitter, &rep.Status, &rep.CreatedAt, &rep.UpdatedAt)
		reports = append(reports, rep)
	}
	middleware.JSON(w, http.StatusOK, paginateResult(reports, total, p.Page, p.PageSize))
}

// CreateReport 提交报表
func CreateReport(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	var req models.Report
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Title == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "标题必填"})
		return
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	res, err := database.DB.Exec(
		`INSERT INTO reports (report_type, title, content, period, submitter_id, status) VALUES (?, ?, ?, ?, ?, ?)`,
		req.ReportType, req.Title, req.Content, req.Period, userID, status)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "提交失败"})
		return
	}
	id, _ := res.LastInsertId()
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"message": "提交成功", "id": id})
}

// GetReport 报表详情
func GetReport(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	var rep models.Report
	err := database.DB.QueryRow(
		`SELECT r.id, r.report_type, r.title, r.content, r.period, r.submitter_id, u.real_name, r.status, r.created_at, r.updated_at
		 FROM reports r LEFT JOIN users u ON r.submitter_id = u.id WHERE r.id = ?`, id).
		Scan(&rep.ID, &rep.ReportType, &rep.Title, &rep.Content, &rep.Period, &rep.SubmitterID,
			&rep.Submitter, &rep.Status, &rep.CreatedAt, &rep.UpdatedAt)
	if err != nil {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "报表不存在"})
		return
	}
	middleware.JSON(w, http.StatusOK, rep)
}

// UpdateReportStatus 审阅报表
func UpdateReportStatus(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req struct {
		Status int `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	_, err := database.DB.Exec("UPDATE reports SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.Status, id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

// ReportStats 报表统计（按类型和状态）
func ReportStats(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT report_type, status, COUNT(*) FROM reports GROUP BY report_type, status")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	result := map[string]map[string]int{}
	for rows.Next() {
		var t, s string
		var count int
		rows.Scan(&t, &s, &count)
		if _, ok := result[t]; !ok {
			result[t] = map[string]int{}
		}
		result[t][s] = count
	}
	middleware.JSON(w, http.StatusOK, result)
}
