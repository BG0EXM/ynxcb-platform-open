package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"ynxcb-platform/internal/database"
	"ynxcb-platform/internal/middleware"
	"ynxcb-platform/internal/models"
)

// MarkAttendance 管理员晨会点到（单人或批量）
// body: {"attend_date":"2026-08-05","records":[{"user_id":1,"status":1,"remark":""},...]}
// 或单条: {"attend_date":"2026-08-05","user_id":1,"status":1,"remark":""}
func MarkAttendance(w http.ResponseWriter, r *http.Request) {
	operatorID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)
	if roleCode != "admin" {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅管理员可点到"})
		return
	}

	var req struct {
		AttendDate string `json:"attend_date"`
		UserID     int64  `json:"user_id"`
		Status     int    `json:"status"`
		LeaveType  string `json:"leave_type"`
		Remark     string `json:"remark"`
		Records    []struct {
			UserID    int64  `json:"user_id"`
			Status    int    `json:"status"`
			LeaveType string `json:"leave_type"`
			Remark    string `json:"remark"`
		} `json:"records"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.AttendDate == "" {
		req.AttendDate = time.Now().Format("2006-01-02")
	}

	// 批量
	if len(req.Records) > 0 {
		for _, rec := range req.Records {
			status := rec.Status
			if status == 0 {
				status = 1
			}
			_, err := database.DB.Exec(
				`INSERT INTO attendances (user_id, attend_date, status, leave_type, remark, marked_by) VALUES (?, ?, ?, ?, ?, ?)
				 ON CONFLICT(user_id, attend_date) DO UPDATE SET status=?, leave_type=?, remark=?, marked_by=?, updated_at=CURRENT_TIMESTAMP`,
				rec.UserID, req.AttendDate, status, rec.LeaveType, rec.Remark, operatorID,
				status, rec.LeaveType, rec.Remark, operatorID)
			if err != nil {
				middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "点到失败"})
				return
			}
		}
		middleware.JSON(w, http.StatusOK, map[string]string{"message": "点到成功"})
		return
	}

	// 单条
	if req.UserID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少人员"})
		return
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	_, err := database.DB.Exec(
		`INSERT INTO attendances (user_id, attend_date, status, leave_type, remark, marked_by) VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, attend_date) DO UPDATE SET status=?, leave_type=?, remark=?, marked_by=?, updated_at=CURRENT_TIMESTAMP`,
		req.UserID, req.AttendDate, status, req.LeaveType, req.Remark, operatorID,
		status, req.LeaveType, req.Remark, operatorID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "点到失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "点到成功"})
}

// ListAttendances 考勤记录（管理员看全部，普通用户看自己，分页）
func ListAttendances(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)

	date := r.URL.Query().Get("date")
	month := r.URL.Query().Get("month")
	userIDFilter := r.URL.Query().Get("user_id")

	where := ` WHERE 1=1`
	args := []interface{}{}
	if date != "" {
		where += ` AND a.attend_date = ?`
		args = append(args, date)
	}
	if month != "" {
		where += ` AND a.attend_date LIKE ?`
		args = append(args, month+"%")
	}
	if roleCode != "admin" {
		where += ` AND a.user_id = ?`
		args = append(args, userID)
	}
	if userIDFilter != "" && roleCode == "admin" {
		where += ` AND a.user_id = ?`
		args = append(args, userIDFilter)
	}

	p := parsePage(r)

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM attendances a"+where, args...).Scan(&total)

	query := `SELECT a.id, a.user_id, u.real_name, a.attend_date, a.status, a.leave_type, a.remark, a.created_at, a.updated_at
		FROM attendances a LEFT JOIN users u ON a.user_id = u.id` + where +
		` ORDER BY a.attend_date DESC, a.user_id LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	list := []models.Attendance{}
	for rows.Next() {
		var a models.Attendance
		rows.Scan(&a.ID, &a.UserID, &a.UserName, &a.AttendDate, &a.Status, &a.LeaveType, &a.Remark, &a.CreatedAt, &a.UpdatedAt)
		list = append(list, a)
	}
	middleware.JSON(w, http.StatusOK, paginateResult(list, total, p.Page, p.PageSize))
}

// AttendanceStats 考勤统计（指定日期，管理员用）
func AttendanceStats(w http.ResponseWriter, r *http.Request) {
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)
	if roleCode != "admin" {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅管理员可查看统计"})
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	stats := map[string]int{"total": 0, "present": 0, "leave": 0, "trip": 0, "absent": 0}
	rows, err := database.DB.Query("SELECT status, COUNT(*) FROM attendances WHERE attend_date=? GROUP BY status", date)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	total := 0
	for rows.Next() {
		var status, count int
		rows.Scan(&status, &count)
		total += count
		switch status {
		case 1:
			stats["present"] = count
		case 2:
			stats["leave"] = count
		case 3:
			stats["trip"] = count
		case 4:
			stats["absent"] = count
		}
	}
	stats["total"] = total
	middleware.JSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats, "date": date,
	})
}

// AttendanceDates 已点到日期列表（管理员选择日期用）
func AttendanceDates(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT DISTINCT attend_date FROM attendances ORDER BY attend_date DESC LIMIT 60")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	dates := []string{}
	for rows.Next() {
		var d string
		rows.Scan(&d)
		dates = append(dates, d)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"dates": dates})
}

// MarkUsers 管理员点到人员列表（全部启用人员 + 当日状态）
func MarkUsers(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	query := `SELECT u.id, u.real_name, d.name,
			COALESCE(a.status, 0) as status, COALESCE(a.leave_type, '') as leave_type, COALESCE(a.remark, '') as remark
		FROM users u
		LEFT JOIN departments d ON u.department_id = d.id
		LEFT JOIN attendances a ON a.user_id = u.id AND a.attend_date = ?
		WHERE u.status = 1
		ORDER BY u.id`
	rows, err := database.DB.Query(query, date)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	type UserItem struct {
		ID         int64  `json:"id"`
		RealName   string `json:"real_name"`
		Department string `json:"department"`
		Status     int    `json:"status"`
		LeaveType  string `json:"leave_type"`
		Remark     string `json:"remark"`
	}
	list := []UserItem{}
	for rows.Next() {
		var u UserItem
		var dept sql.NullString
		rows.Scan(&u.ID, &u.RealName, &dept, &u.Status, &u.LeaveType, &u.Remark)
		if dept.Valid {
			u.Department = dept.String
		}
		list = append(list, u)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": list})
}

// 请假类型常量
var LeaveTypes = []string{"annual", "sick", "personal", "marriage", "maternity", "bereavement", "other"}

// CreateLeaveRecord 登记请假（管理员）
func CreateLeaveRecord(w http.ResponseWriter, r *http.Request) {
	operatorID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)
	if roleCode != "admin" {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅管理员可登记请假"})
		return
	}
	var req models.LeaveRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.UserID == 0 || req.LeaveType == "" || req.StartDate == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请填写完整信息"})
		return
	}
	if req.EndDate == "" {
		req.EndDate = req.StartDate
	}
	if req.Days <= 0 {
		req.Days = 1
	}
	_, err := database.DB.Exec(
		`INSERT INTO leave_records (user_id, leave_type, start_date, end_date, days, reason, status) VALUES (?, ?, ?, ?, ?, ?, 1)`,
		req.UserID, req.LeaveType, req.StartDate, req.EndDate, req.Days, req.Reason)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "登记失败"})
		return
	}
	_ = operatorID
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "请假登记成功"})
}

// ListLeaveRecords 请假记录列表（分页）
func ListLeaveRecords(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)

	leaveType := r.URL.Query().Get("leave_type")
	userIDFilter := r.URL.Query().Get("user_id")

	where := ` WHERE 1=1`
	args := []interface{}{}
	if leaveType != "" {
		where += ` AND l.leave_type = ?`
		args = append(args, leaveType)
	}
	if roleCode != "admin" {
		where += ` AND l.user_id = ?`
		args = append(args, userID)
	}
	if userIDFilter != "" && roleCode == "admin" {
		where += ` AND l.user_id = ?`
		args = append(args, userIDFilter)
	}

	p := parsePage(r)

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM leave_records l"+where, args...).Scan(&total)

	query := `SELECT l.id, l.user_id, u.real_name, l.leave_type, l.start_date, l.end_date, l.days, l.reason, l.status, l.created_at, l.updated_at
		FROM leave_records l LEFT JOIN users u ON l.user_id = u.id` + where +
		` ORDER BY l.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	list := []models.LeaveRecord{}
	for rows.Next() {
		var l models.LeaveRecord
		rows.Scan(&l.ID, &l.UserID, &l.UserName, &l.LeaveType, &l.StartDate, &l.EndDate,
			&l.Days, &l.Reason, &l.Status, &l.CreatedAt, &l.UpdatedAt)
		list = append(list, l)
	}
	middleware.JSON(w, http.StatusOK, paginateResult(list, total, p.Page, p.PageSize))
}

// DeleteLeaveRecord 删除请假记录
func DeleteLeaveRecord(w http.ResponseWriter, r *http.Request) {
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)
	if roleCode != "admin" {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅管理员可删除"})
		return
	}
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec("DELETE FROM leave_records WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// LeaveStats 请假统计（各类请假 + 总计）
func LeaveStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)

	year := r.URL.Query().Get("year")
	if year == "" {
		year = time.Now().Format("2006")
	}

	// 各类型天数统计
	query := `SELECT leave_type, COUNT(*) as cnt, SUM(days) as total_days FROM leave_records
		WHERE strftime('%Y', start_date) = ?`
	args := []interface{}{year}
	if roleCode != "admin" {
		query += ` AND user_id = ?`
		args = append(args, userID)
	}
	query += ` GROUP BY leave_type`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	result := map[string]interface{}{"year": year}
	totalDays := 0.0
	totalCount := 0
	for _, lt := range LeaveTypes {
		result[lt+"_count"] = 0
		result[lt+"_days"] = 0
	}
	for rows.Next() {
		var lt string
		var cnt int
		var days sql.NullFloat64
		rows.Scan(&lt, &cnt, &days)
		d := 0.0
		if days.Valid {
			d = days.Float64
		}
		result[lt+"_count"] = cnt
		result[lt+"_days"] = d
		totalDays += d
		totalCount += cnt
	}
	result["total_count"] = totalCount
	result["total_days"] = totalDays
	middleware.JSON(w, http.StatusOK, result)
}

// AttendanceMonthly 月度考勤统计（按人员）
// 返回每个人员的出勤/请假/出差/未到天数 + 各类请假天数
func AttendanceMonthly(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month") // YYYY-MM
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	// 先查询出勤汇总
	query := `SELECT a.user_id, u.real_name, d.name,
			SUM(CASE WHEN a.status=1 THEN 1 ELSE 0 END) as present,
			SUM(CASE WHEN a.status=2 THEN 1 ELSE 0 END) as leave,
			SUM(CASE WHEN a.status=3 THEN 1 ELSE 0 END) as trip,
			SUM(CASE WHEN a.status=4 THEN 1 ELSE 0 END) as absent
		FROM attendances a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN departments d ON u.department_id = d.id
		WHERE a.attend_date LIKE ?
		GROUP BY a.user_id ORDER BY a.user_id`
	rows, err := database.DB.Query(query, month+"%")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}

	type Row struct {
		UserID     int64   `json:"user_id"`
		UserName   string  `json:"user_name"`
		Department string  `json:"department"`
		Present    int     `json:"present"`
		Leave      int     `json:"leave"`
		Trip       int     `json:"trip"`
		Absent     int     `json:"absent"`
		AnnualDays float64 `json:"annual_days"`
		SickDays   float64 `json:"sick_days"`
		PersonalDays float64 `json:"personal_days"`
		OtherDays  float64 `json:"other_days"`
	}

	// 收集 user_id 列表，先关闭外层 rows 再查请假明细（避免 MaxOpenConns=1 死锁）
	userIDs := []int64{}
	base := map[int64]*Row{}
	for rows.Next() {
		var rw Row
		var dept sql.NullString
		rows.Scan(&rw.UserID, &rw.UserName, &dept, &rw.Present, &rw.Leave, &rw.Trip, &rw.Absent)
		if dept.Valid {
			rw.Department = dept.String
		}
		base[rw.UserID] = &rw
		userIDs = append(userIDs, rw.UserID)
	}
	rows.Close()

	// 单独查询请假明细（一次性）
	leaveQuery := `SELECT user_id,
			COALESCE(SUM(CASE WHEN leave_type='annual' THEN days ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN leave_type='sick' THEN days ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN leave_type='personal' THEN days ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN leave_type NOT IN ('annual','sick','personal') THEN days ELSE 0 END),0)
		FROM leave_records WHERE strftime('%Y-%m', start_date)=? GROUP BY user_id`
	lrows, err := database.DB.Query(leaveQuery, month)
	if err == nil {
		for lrows.Next() {
			var uid int64
			var annual, sick, personal, other float64
			if lrows.Scan(&uid, &annual, &sick, &personal, &other) == nil {
				if rw, ok := base[uid]; ok {
					rw.AnnualDays = annual
					rw.SickDays = sick
					rw.PersonalDays = personal
					rw.OtherDays = other
				}
			}
		}
		lrows.Close()
	}

	list := []Row{}
	var totalPresent, totalLeave, totalTrip, totalAbsent int
	for _, uid := range userIDs {
		if rw, ok := base[uid]; ok {
			list = append(list, *rw)
			totalPresent += rw.Present
			totalLeave += rw.Leave
			totalTrip += rw.Trip
			totalAbsent += rw.Absent
		}
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{
		"month": month, "list": list,
		"total": map[string]int{
			"present": totalPresent, "leave": totalLeave, "trip": totalTrip, "absent": totalAbsent,
		},
	})
}

// AttendanceYearly 年度考勤统计（按月汇总）
// 返回全年各月份的出勤/请假/出差/未到人数汇总
func AttendanceYearly(w http.ResponseWriter, r *http.Request) {
	year := r.URL.Query().Get("year") // YYYY
	if year == "" {
		year = time.Now().Format("2006")
	}

	query := `SELECT substr(a.attend_date, 1, 7) as ym,
			SUM(CASE WHEN a.status=1 THEN 1 ELSE 0 END) as present,
			SUM(CASE WHEN a.status=2 THEN 1 ELSE 0 END) as leave,
			SUM(CASE WHEN a.status=3 THEN 1 ELSE 0 END) as trip,
			SUM(CASE WHEN a.status=4 THEN 1 ELSE 0 END) as absent
		FROM attendances a
		WHERE a.attend_date LIKE ?
		GROUP BY ym ORDER BY ym`
	rows, err := database.DB.Query(query, year+"%")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	type Row struct {
		Month   string `json:"month"`
		Present int    `json:"present"`
		Leave   int    `json:"leave"`
		Trip    int    `json:"trip"`
		Absent  int    `json:"absent"`
	}
	monthly := []Row{}
	var totalPresent, totalLeave, totalTrip, totalAbsent int
	for rows.Next() {
		var rw Row
		rows.Scan(&rw.Month, &rw.Present, &rw.Leave, &rw.Trip, &rw.Absent)
		monthly = append(monthly, rw)
		totalPresent += rw.Present
		totalLeave += rw.Leave
		totalTrip += rw.Trip
		totalAbsent += rw.Absent
	}

	// 全年请假类型汇总
	var annualDays, sickDays, personalDays, otherDays, totalDays, totalCount float64
	var cnt int
	database.DB.QueryRow(
		`SELECT
			COALESCE(SUM(CASE WHEN leave_type='annual' THEN days ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN leave_type='sick' THEN days ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN leave_type='personal' THEN days ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN leave_type NOT IN ('annual','sick','personal') THEN days ELSE 0 END),0),
			COALESCE(SUM(days),0),
			COUNT(*)
		 FROM leave_records WHERE strftime('%Y', start_date)=?`,
		year).Scan(&annualDays, &sickDays, &personalDays, &otherDays, &totalDays, &cnt)
	totalCount = float64(cnt)

	middleware.JSON(w, http.StatusOK, map[string]interface{}{
		"year": year, "monthly": monthly,
		"total": map[string]int{
			"present": totalPresent, "leave": totalLeave, "trip": totalTrip, "absent": totalAbsent,
		},
		"leave": map[string]float64{
			"annual": annualDays, "sick": sickDays, "personal": personalDays, "other": otherDays,
			"total_days": totalDays, "total_count": totalCount,
		},
	})
}

// GetAssignees 可分配/可选人员列表（启用用户）
func GetAssignees(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(
		`SELECT u.id, u.real_name, d.name FROM users u LEFT JOIN departments d ON u.department_id = d.id WHERE u.status = 1 ORDER BY u.id`)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	type Assignee struct {
		ID         int64  `json:"id"`
		RealName   string `json:"real_name"`
		Department string `json:"department"`
	}
	list := []Assignee{}
	for rows.Next() {
		var a Assignee
		var dept sql.NullString
		rows.Scan(&a.ID, &a.RealName, &dept)
		if dept.Valid {
			a.Department = dept.String
		}
		list = append(list, a)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": list})
}
