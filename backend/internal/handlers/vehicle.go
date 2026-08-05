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

func timeNow() time.Time { return time.Now() }

// ListVehicles 公车列表
func ListVehicles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")

	query := `SELECT id, plate_no, brand, seats, driver, status, vin, engine_no, insurance_date, inspect_date, register_date, purchase_at, note, created_at
		FROM vehicles WHERE 1=1`
	args := []interface{}{}
	if status != "" && status != "0" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	if keyword != "" {
		query += ` AND (plate_no LIKE ? OR brand LIKE ? OR driver LIKE ? OR vin LIKE ?)`
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
	query += ` ORDER BY id DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	list := []models.Vehicle{}
	for rows.Next() {
		var v models.Vehicle
		rows.Scan(&v.ID, &v.PlateNo, &v.Brand, &v.Seats, &v.Driver, &v.Status, &v.Vin, &v.EngineNo,
			&v.InsuranceDate, &v.InspectDate, &v.RegisterDate, &v.PurchaseAt, &v.Note, &v.CreatedAt)
		list = append(list, v)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": list})
}

// CreateVehicle 新增公车
func CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req models.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.PlateNo == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "车牌号必填"})
		return
	}
	if req.Seats == 0 {
		req.Seats = 5
	}
	if req.Status == 0 {
		req.Status = 1
	}
	_, err := database.DB.Exec(
		`INSERT INTO vehicles (plate_no, brand, seats, driver, status, vin, engine_no, insurance_date, inspect_date, register_date, purchase_at, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.PlateNo, req.Brand, req.Seats, req.Driver, req.Status, req.Vin, req.EngineNo,
		req.InsuranceDate, req.InspectDate, req.RegisterDate, req.PurchaseAt, req.Note)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "添加失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "添加成功"})
}

// UpdateVehicle 更新公车
func UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	var req models.Vehicle
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.ID == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec(
		`UPDATE vehicles SET plate_no=?, brand=?, seats=?, driver=?, status=?, vin=?, engine_no=?,
			insurance_date=?, inspect_date=?, register_date=?, purchase_at=?, note=? WHERE id=?`,
		req.PlateNo, req.Brand, req.Seats, req.Driver, req.Status, req.Vin, req.EngineNo,
		req.InsuranceDate, req.InspectDate, req.RegisterDate, req.PurchaseAt, req.Note, req.ID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "更新失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "更新成功"})
}

// DeleteVehicle 删除公车
func DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	database.DB.Exec("UPDATE vehicle_applies SET vehicle_id=0 WHERE vehicle_id=?", id)
	_, err := database.DB.Exec("DELETE FROM vehicles WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// CreateVehicleApply 用车报备（无需审批）
func CreateVehicleApply(w http.ResponseWriter, r *http.Request) {
	reporterID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	var req models.VehicleApply
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.VehicleID == 0 || req.Purpose == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请选择车辆并填写事由"})
		return
	}
	if req.UserName == "" {
		req.UserName, _ = r.Context().Value(middleware.ContextRealName).(string)
	}
	if req.Passengers == 0 {
		req.Passengers = 1
	}
	res, err := database.DB.Exec(
		`INSERT INTO vehicle_applies (vehicle_id, reporter_id, user_name, purpose, destination, use_date, use_time, passengers)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.VehicleID, reporterID, req.UserName, req.Purpose, req.Destination, req.UseDate, req.UseTime, req.Passengers)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "报备失败"})
		return
	}
	id, _ := res.LastInsertId()
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"message": "报备成功", "id": id})
}

// ListVehicleApplies 用车报备列表（分页）
func ListVehicleApplies(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	roleCode, _ := r.Context().Value(middleware.ContextRoleCode).(string)

	date := r.URL.Query().Get("date")
	mine := r.URL.Query().Get("mine")

	where := ` WHERE 1=1`
	args := []interface{}{}
	if date != "" {
		where += ` AND a.use_date = ?`
		args = append(args, date)
	}
	if mine == "1" {
		where += ` AND a.reporter_id = ?`
		args = append(args, userID)
	} else if roleCode != "admin" {
		where += ` AND a.reporter_id = ?`
		args = append(args, userID)
	}

	p := parsePage(r)

	var total int
	database.DB.QueryRow("SELECT COUNT(*) FROM vehicle_applies a"+where, args...).Scan(&total)

	query := `SELECT a.id, a.vehicle_id, v.plate_no, v.brand, v.driver, a.reporter_id, u.real_name,
		a.user_name, a.purpose, a.destination, a.use_date, a.use_time, a.passengers, a.created_at
		FROM vehicle_applies a
		LEFT JOIN vehicles v ON a.vehicle_id = v.id
		LEFT JOIN users u ON a.reporter_id = u.id` + where +
		` ORDER BY a.id DESC LIMIT ? OFFSET ?`
	args = append(args, p.PageSize, (p.Page-1)*p.PageSize)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	list := []models.VehicleApply{}
	for rows.Next() {
		var a models.VehicleApply
		var plateNo, brand, driver, reporterName sql.NullString
		rows.Scan(&a.ID, &a.VehicleID, &plateNo, &brand, &driver, &a.ReporterID, &reporterName,
			&a.UserName, &a.Purpose, &a.Destination, &a.UseDate, &a.UseTime, &a.Passengers, &a.CreatedAt)
		if plateNo.Valid {
			a.VehicleNo = plateNo.String
		}
		if brand.Valid {
			a.VehicleBrand = brand.String
		}
		if driver.Valid {
			a.VehicleDriver = driver.String
		}
		if reporterName.Valid {
			a.Reporter = reporterName.String
		}
		list = append(list, a)
	}
	middleware.JSON(w, http.StatusOK, paginateResult(list, total, p.Page, p.PageSize))
}

// GetVehicleApply 报备详情（派车单用）
func GetVehicleApply(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	var a models.VehicleApply
	var plateNo, brand, driver, reporterName sql.NullString
	err := database.DB.QueryRow(
		`SELECT a.id, a.vehicle_id, v.plate_no, v.brand, v.driver, a.reporter_id, u.real_name,
			a.user_name, a.purpose, a.destination, a.use_date, a.use_time, a.passengers, a.created_at
		 FROM vehicle_applies a
		 LEFT JOIN vehicles v ON a.vehicle_id = v.id
		 LEFT JOIN users u ON a.reporter_id = u.id
		 WHERE a.id = ?`, id).
		Scan(&a.ID, &a.VehicleID, &plateNo, &brand, &driver, &a.ReporterID, &reporterName,
			&a.UserName, &a.Purpose, &a.Destination, &a.UseDate, &a.UseTime, &a.Passengers, &a.CreatedAt)
	if err != nil {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "报备不存在"})
		return
	}
	if plateNo.Valid {
		a.VehicleNo = plateNo.String
	}
	if brand.Valid {
		a.VehicleBrand = brand.String
	}
	if driver.Valid {
		a.VehicleDriver = driver.String
	}
	if reporterName.Valid {
		a.Reporter = reporterName.String
	}
	middleware.JSON(w, http.StatusOK, a)
}

// DeleteVehicleApply 删除报备
func DeleteVehicleApply(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少ID"})
		return
	}
	_, err := database.DB.Exec("DELETE FROM vehicle_applies WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}

// VehicleStats 车辆统计
func VehicleStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{"total": 0, "available": 0, "in_use": 0, "repair": 0, "today_use": 0}
	rows, err := database.DB.Query("SELECT status, COUNT(*) FROM vehicles GROUP BY status")
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var status, count int
		rows.Scan(&status, &count)
		stats["total"] += count
		switch status {
		case 1:
			stats["available"] = count
		case 2:
			stats["in_use"] = count
		case 3:
			stats["repair"] = count
		}
	}
	// 今日报备数
	today := timeNow().Format("2006-01-02")
	var todayUse int
	database.DB.QueryRow("SELECT COUNT(*) FROM vehicle_applies WHERE use_date=?", today).Scan(&todayUse)
	stats["today_use"] = todayUse
	middleware.JSON(w, http.StatusOK, stats)
}
