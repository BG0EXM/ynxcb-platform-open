package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ynxcb-platform/internal/config"
	"ynxcb-platform/internal/database"
	"ynxcb-platform/internal/middleware"
	"ynxcb-platform/internal/models"
)

// UploadFile 文件上传
func UploadFile(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
		realName, _ := r.Context().Value(middleware.ContextRealName).(string)

		if err := r.ParseMultipartForm(cfg.Upload.MaxMB << 20); err != nil {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "文件过大，超过限制"})
			return
		}
		ownerType := r.FormValue("owner_type")
		ownerID := r.FormValue("owner_id")
		file, header, err := r.FormFile("file")
		if err != nil {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "未选择文件"})
			return
		}
		defer file.Close()

		// 安全校验：文件扩展名白名单
		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowedExt := map[string]bool{".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
			".pdf": true, ".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".txt": true, ".zip": true, ".rar": true, ".ppt": true, ".pptx": true}
		if !allowedExt[ext] {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "不支持的文件类型"})
			return
		}

		// 校验大小
		if header.Size > cfg.Upload.MaxMB<<20 {
			middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "文件超过大小限制"})
			return
		}

		// 保存文件
		dateDir := time.Now().Format("2006/01")
		saveDir := filepath.Join(cfg.Upload.Dir, dateDir)
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "创建目录失败"})
			return
		}

		fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
		savePath := filepath.Join(saveDir, fileName)
		dst, err := os.Create(savePath)
		if err != nil {
			middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "保存失败"})
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "写入失败"})
			return
		}

		// 记录到数据库
		oType := ownerType
		if oType == "" {
			oType = "document"
		}
		var oID int64
		if ownerID != "" {
			oID, _ = strconv.ParseInt(ownerID, 10, 64)
		}

		webPath := "/uploads/" + dateDir + "/" + fileName
		// 存储路径使用相对 webPath 便于访问
		dbPath := webPath
		absSavePath := savePath

		// 保存实际路径信息
		res, err := database.DB.Exec(
			"INSERT INTO attachments (owner_type, owner_id, file_name, file_path, file_size, uploader_id) VALUES (?, ?, ?, ?, ?, ?)",
			oType, oID, header.Filename, absSavePath, header.Size, userID)
		if err != nil {
			middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "记录失败"})
			return
		}
		attID, _ := res.LastInsertId()
		_ = dbPath

		middleware.JSON(w, http.StatusOK, map[string]interface{}{
			"message":   "上传成功",
			"id":        attID,
			"file_name": header.Filename,
			"file_path": "/api/uploads/" + strconv.FormatInt(attID, 10),
			"uploader":  realName,
		})
	}
}

// DownloadAttachment 下载附件（通过数据库记录 ID）
func DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		http.Error(w, "缺少附件ID", http.StatusBadRequest)
		return
	}
	var a models.Attachment
	err := database.DB.QueryRow(
		"SELECT id, file_name, file_path, file_size FROM attachments WHERE id=?", id).
		Scan(&a.ID, &a.FileName, &a.FilePath, &a.FileSize)
	if err != nil {
		http.Error(w, "附件不存在", http.StatusNotFound)
		return
	}
	// 文件路径以 /uploads/ 开头则转实际路径
	fullPath := a.FilePath
	if strings.HasPrefix(fullPath, "/uploads/") {
		fullPath = "data" + fullPath
	}
	// 若数据库存的是绝对路径，直接使用
	if _, err := os.Stat(fullPath); err != nil {
		http.Error(w, "文件已被移除", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=%q; filename*=UTF-8''"+url.PathEscape(a.FileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, fullPath)
}

// LinkAttachment 将已上传附件关联到业务对象
func LinkAttachment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      int64  `json:"id"`
		OwnerID int64  `json:"owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.ID == 0 || req.OwnerID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "参数不完整"})
		return
	}
	_, err := database.DB.Exec("UPDATE attachments SET owner_id=? WHERE id=?", req.OwnerID, req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "关联失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "关联成功"})
}

// DashboardStats 首页统计
func DashboardStats(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)

	result := map[string]interface{}{}
	today := time.Now().Format("2006-01-02")
	month := time.Now().Format("2006-01")
	year := time.Now().Format("2006")

	// 今日值守（当天值守至21:00收文）
	var dutyName string
	var isDaWangYuan int
	err := database.DB.QueryRow(
		"SELECT u.real_name, s.is_dawangyuan FROM duty_schedules s JOIN users u ON s.user_id=u.id WHERE s.duty_date=?",
		today).Scan(&dutyName, &isDaWangYuan)
	if err == nil {
		result["today_duty"] = dutyName
		result["today_duty_dawangyuan"] = isDaWangYuan
	} else {
		result["today_duty"] = ""
		result["today_duty_dawangyuan"] = 0
	}

	// 今日考勤状态（当前用户）
	var todayStatus int
	var todayCheckin string
	err = database.DB.QueryRow(
		"SELECT status, IFNULL(remark,'') FROM attendances WHERE user_id=? AND attend_date=?", userID, today).Scan(&todayStatus, &todayCheckin)
	if err == nil {
		result["today_attendance"] = todayStatus
	} else {
		result["today_attendance"] = 0
	}

	// 本月考勤汇总（当前用户）
	var monthPresent, monthLeave int
	database.DB.QueryRow("SELECT COUNT(*) FROM attendances WHERE user_id=? AND attend_date LIKE ? AND status=1", userID, month+"%").Scan(&monthPresent)
	database.DB.QueryRow("SELECT COUNT(*) FROM attendances WHERE user_id=? AND attend_date LIKE ? AND status=2", userID, month+"%").Scan(&monthLeave)
	result["month_present"] = monthPresent
	result["month_leave"] = monthLeave

	// 本月请假天数
	var monthLeaveDays float64
	database.DB.QueryRow("SELECT COALESCE(SUM(days),0) FROM leave_records WHERE user_id=? AND strftime('%Y-%m', start_date)=?", userID, month).Scan(&monthLeaveDays)
	result["month_leave_days"] = monthLeaveDays

	// 今日用车报备
	var todayVehicle int
	database.DB.QueryRow("SELECT COUNT(*) FROM vehicle_applies WHERE use_date=?", today).Scan(&todayVehicle)
	result["today_vehicle"] = todayVehicle

	// 待处理收文（今日收到的未办结收文，管理员看全部）
	var pendingIncoming int
	if roleCode == "admin" {
		database.DB.QueryRow("SELECT COUNT(*) FROM incoming_docs WHERE status < 5").Scan(&pendingIncoming)
	} else {
		database.DB.QueryRow("SELECT COUNT(*) FROM incoming_docs WHERE status < 5 AND registrar_id=?", userID).Scan(&pendingIncoming)
	}
	result["pending_incoming"] = pendingIncoming

	// 最新收文列表（工作台展示）
	rows, err := database.DB.Query(
		`SELECT id, receive_no, received_date, from_unit, from_doc_no, title, status
		 FROM incoming_docs ORDER BY id DESC LIMIT 6`)
	if err == nil {
		type IncomingBrief struct {
			ID           int64  `json:"id"`
			ReceiveNo    string `json:"receive_no"`
			ReceivedDate string `json:"received_date"`
			FromUnit     string `json:"from_unit"`
			FromDocNo    string `json:"from_doc_no"`
			Title        string `json:"title"`
			Status       int    `json:"status"`
		}
		latest := []IncomingBrief{}
		for rows.Next() {
			var b IncomingBrief
			rows.Scan(&b.ID, &b.ReceiveNo, &b.ReceivedDate, &b.FromUnit, &b.FromDocNo, &b.Title, &b.Status)
			latest = append(latest, b)
		}
		rows.Close()
		result["latest_incoming"] = latest
	} else {
		result["latest_incoming"] = []interface{}{}
	}

	// 本周排班预览
	weekRows, err := database.DB.Query(
		`SELECT s.duty_date, u.real_name, s.is_dawangyuan FROM duty_schedules s
		 JOIN users u ON s.user_id=u.id
		 WHERE s.duty_date >= date('now','weekday 0','-6 days') AND s.duty_date <= date('now','weekday 0')
		 ORDER BY s.duty_date`)
	if err == nil {
		type DutyBrief struct {
			DutyDate    string `json:"duty_date"`
			UserName    string `json:"user_name"`
			IsDaWangYuan int   `json:"is_dawangyuan"`
		}
		weekDuty := []DutyBrief{}
		for weekRows.Next() {
			var d DutyBrief
			weekRows.Scan(&d.DutyDate, &d.UserName, &d.IsDaWangYuan)
			weekDuty = append(weekDuty, d)
		}
		weekRows.Close()
		result["week_duty"] = weekDuty
	} else {
		result["week_duty"] = []interface{}{}
	}

	// 报表待提交（本月）
	var reportDone int
	database.DB.QueryRow("SELECT COUNT(*) FROM reports WHERE submitter_id=? AND status=2", userID).Scan(&reportDone)
	result["report_submitted"] = reportDone

	// 全年请假汇总（管理员看全员）
	if roleCode == "admin" {
		var annualDays float64
		database.DB.QueryRow("SELECT COALESCE(SUM(days),0) FROM leave_records WHERE strftime('%Y', start_date)=?", year).Scan(&annualDays)
		result["year_leave_days"] = annualDays
	}

	middleware.JSON(w, http.StatusOK, result)
}

// Health 健康检查
func Health(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "time": time.Now().Format("2006-01-02 15:04:05")})
}

// pathID 提取路径中的ID
func pathID(r *http.Request) int64 {
	if v := r.PathValue("id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return id
		}
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return 0
	}
	id, _ := strconv.ParseInt(parts[len(parts)-1], 10, 64)
	return id
}
