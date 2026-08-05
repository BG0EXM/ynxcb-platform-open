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

// ListIncomingDocs 收文列表（分页）
func ListIncomingDocs(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	status := r.URL.Query().Get("status")
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	returned := r.URL.Query().Get("returned")
	needReturn := r.URL.Query().Get("need_return")

	// 构建 where 子句
	where := ` WHERE 1=1`
	args := []interface{}{}
	if keyword != "" {
		where += ` AND (d.title LIKE ? OR d.from_unit LIKE ? OR d.receive_no LIKE ? OR d.from_doc_no LIKE ? OR d.doc_no LIKE ?)`
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw, kw)
	}
	if status != "" && status != "0" {
		where += ` AND d.status = ?`
		args = append(args, status)
	}
	if returned != "" && returned != "0" {
		where += ` AND d.returned = ?`
		args = append(args, returned)
	}
	if needReturn != "" && needReturn != "0" {
		where += ` AND d.need_return = ?`
		args = append(args, needReturn)
	}
	if start != "" {
		where += ` AND d.received_date >= ?`
		args = append(args, start)
	}
	if end != "" {
		where += ` AND d.received_date <= ?`
		args = append(args, end)
	}

	// 分页
	p := parsePage(r)

	// 总数
	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM incoming_docs d"+where, args...).Scan(&total)

	// 数据
	query := `SELECT d.id, d.receive_no, d.received_date, d.from_unit, d.from_doc_no, d.doc_no, d.title,
		d.copies, d.secret_level, d.urgency, d.suggest, d.leader_comment, d.processing,
		d.return_date, d.returned, d.need_return,
		d.registrar_id, u.real_name, d.status, d.created_at, d.updated_at
		FROM incoming_docs d LEFT JOIN users u ON d.registrar_id = u.id` + where +
		` ORDER BY d.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	docs := []models.IncomingDoc{}
	for rows.Next() {
		var d models.IncomingDoc
		rows.Scan(&d.ID, &d.ReceiveNo, &d.ReceivedDate, &d.FromUnit, &d.FromDocNo, &d.DocNo, &d.Title,
			&d.Copies, &d.SecretLevel, &d.Urgency, &d.Suggest, &d.LeaderComment, &d.Processing,
			&d.ReturnDate, &d.Returned, &d.NeedReturn,
			&d.RegistrarID, &d.Registrar, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		docs = append(docs, d)
	}
	middleware.JSON(w, http.StatusOK, paginateResult(docs, total, p.Page, p.PageSize))
}

// GetIncomingDoc 收文详情（含传阅记录）
func GetIncomingDoc(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	var d models.IncomingDoc
	err := database.DB.QueryRow(
		`SELECT d.id, d.receive_no, d.received_date, d.from_unit, d.from_doc_no, d.doc_no, d.title,
		d.copies, d.secret_level, d.urgency, d.suggest, d.leader_comment, d.processing,
		d.return_date, d.returned, d.need_return,
		d.registrar_id, u.real_name, d.status, d.created_at, d.updated_at
		FROM incoming_docs d LEFT JOIN users u ON d.registrar_id = u.id WHERE d.id = ?`, id).
		Scan(&d.ID, &d.ReceiveNo, &d.ReceivedDate, &d.FromUnit, &d.FromDocNo, &d.DocNo, &d.Title,
			&d.Copies, &d.SecretLevel, &d.Urgency, &d.Suggest, &d.LeaderComment, &d.Processing,
			&d.ReturnDate, &d.Returned, &d.NeedReturn,
			&d.RegistrarID, &d.Registrar, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "文件不存在"})
		return
	}
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	d.CircList = getCirculations(d.ID)
	middleware.JSON(w, http.StatusOK, d)
}

// CreateIncomingDoc 登记收文（仅办公室部门用户可新增）
func CreateIncomingDoc(w http.ResponseWriter, r *http.Request) {
	if !isOfficeUser(r) {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅办公室用户可新增收文"})
		return
	}
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	var req models.IncomingDoc
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Title == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "文件标题必填"})
		return
	}
	if req.ReceivedDate == "" {
		req.ReceivedDate = time.Now().Format("2006-01-02")
	}
	if req.SecretLevel == "" {
		req.SecretLevel = "普通"
	}
	if req.Urgency == "" {
		req.Urgency = "一般"
	}
	if req.Status == 0 {
		req.Status = 1
	}
	res, err := database.DB.Exec(
		`INSERT INTO incoming_docs (receive_no, received_date, from_unit, from_doc_no, doc_no, title, copies,
			secret_level, urgency, suggest, leader_comment, processing, return_date, returned, need_return, registrar_id, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ReceiveNo, req.ReceivedDate, req.FromUnit, req.FromDocNo, req.DocNo, req.Title, req.Copies,
		req.SecretLevel, req.Urgency, req.Suggest, req.LeaderComment, req.Processing,
		req.ReturnDate, req.Returned, req.NeedReturn, userID, req.Status)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "登记失败"})
		return
	}
	id, _ := res.LastInsertId()
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"message": "登记成功", "id": id})
}

// UpdateIncomingDoc 更新收文（仅办公室部门用户可编辑）
func UpdateIncomingDoc(w http.ResponseWriter, r *http.Request) {
	if !isOfficeUser(r) {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅办公室用户可编辑收文"})
		return
	}
	var req models.IncomingDoc
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.ID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec(
		`UPDATE incoming_docs SET receive_no=?, received_date=?, from_unit=?, from_doc_no=?, doc_no=?, title=?, copies=?,
			secret_level=?, urgency=?, suggest=?, leader_comment=?, processing=?, return_date=?, returned=?, need_return=?, status=?, updated_at=? WHERE id=?`,
		req.ReceiveNo, req.ReceivedDate, req.FromUnit, req.FromDocNo, req.DocNo, req.Title, req.Copies,
		req.SecretLevel, req.Urgency, req.Suggest, req.LeaderComment, req.Processing,
		req.ReturnDate, req.Returned, req.NeedReturn, req.Status,
		time.Now(), req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

// DeleteIncomingDoc 删除收文（仅办公室部门用户可删除）
func DeleteIncomingDoc(w http.ResponseWriter, r *http.Request) {
	if !isOfficeUser(r) {
		middleware.JSON(w, http.StatusForbidden, map[string]string{"error": "仅办公室用户可删除收文"})
		return
	}
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	database.DB.Exec("DELETE FROM circulation_records WHERE doc_id=?", id)
	_, err := database.DB.Exec("DELETE FROM incoming_docs WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// 传阅记录管理

func getCirculations(docID int64) []models.CirculationRecord {
	rows, err := database.DB.Query(
		`SELECT c.id, c.doc_id, c.user_id, u.real_name, c.order_no, c.read_date, c.signature
		 FROM circulation_records c LEFT JOIN users u ON c.user_id = u.id
		 WHERE c.doc_id = ? ORDER BY c.order_no, c.id`, docID)
	if err != nil {
		return []models.CirculationRecord{}
	}
	defer rows.Close()
	list := []models.CirculationRecord{}
	for rows.Next() {
		var c models.CirculationRecord
		rows.Scan(&c.ID, &c.DocID, &c.UserID, &c.UserName, &c.OrderNo, &c.ReadDate, &c.Signature)
		list = append(list, c)
	}
	return list
}

// AddCirculation 添加传阅人
func AddCirculation(w http.ResponseWriter, r *http.Request) {
	var req models.CirculationRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.DocID == 0 || req.UserID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "参数不完整"})
		return
	}
	var maxOrder int
	database.DB.QueryRow("SELECT COALESCE(MAX(order_no),0) FROM circulation_records WHERE doc_id=?", req.DocID).Scan(&maxOrder)

	var exists int
	database.DB.QueryRow("SELECT COUNT(*) FROM circulation_records WHERE doc_id=? AND user_id=?", req.DocID, req.UserID).Scan(&exists)
	if exists > 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "该人员已在传阅名单中"})
		return
	}

	_, err := database.DB.Exec(
		"INSERT INTO circulation_records (doc_id, user_id, order_no) VALUES (?, ?, ?)",
		req.DocID, req.UserID, maxOrder+1)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "添加失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "已添加"})
}

// UpdateCirculation 更新传阅记录（传阅日期/签名）
func UpdateCirculation(w http.ResponseWriter, r *http.Request) {
	var req models.CirculationRecord
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.ID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec(
		"UPDATE circulation_records SET read_date=?, signature=? WHERE id=?",
		req.ReadDate, req.Signature, req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

// DeleteCirculation 删除传阅记录
func DeleteCirculation(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec("DELETE FROM circulation_records WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// IncomingDocStats 收文统计
func IncomingDocStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{"total": 0, "draft": 0, "pending": 0, "approving": 0, "processing": 0, "done": 0}
	rows, err := database.DB.Query("SELECT status, COUNT(*) FROM incoming_docs GROUP BY status")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var status, count int
		rows.Scan(&status, &count)
		switch status {
		case 1:
			stats["draft"] = count
		case 2:
			stats["pending"] = count
		case 3:
			stats["approving"] = count
		case 4:
			stats["processing"] = count
		case 5:
			stats["done"] = count
		}
		stats["total"] += count
	}
	middleware.JSON(w, http.StatusOK, stats)
}

// isOfficeUser 判断当前用户是否属于"办公室"部门
func isOfficeUser(r *http.Request) bool {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	if userID == 0 {
		return false
	}
	var deptName string
	err := database.DB.QueryRow(
		"SELECT d.name FROM users u LEFT JOIN departments d ON u.department_id = d.id WHERE u.id = ?", userID).Scan(&deptName)
	if err != nil {
		return false
	}
	return deptName == "办公室"
}
