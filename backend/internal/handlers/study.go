package handlers

import (
	"encoding/json"
	"net/http"

	"ynxcb-platform/internal/database"
	"ynxcb-platform/internal/middleware"
	"ynxcb-platform/internal/models"
)

// ListStudyMaterials 公共资料列表
func ListStudyMaterials(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	category := r.URL.Query().Get("category")

	query := `SELECT s.id, s.title, s.content, s.category, s.publisher_id, u.real_name, s.read_count, s.created_at, s.updated_at
		FROM study_materials s LEFT JOIN users u ON s.publisher_id = u.id WHERE 1=1`
	args := []interface{}{}
	if keyword != "" {
		query += ` AND s.title LIKE ?`
		args = append(args, "%"+keyword+"%")
	}
	if category != "" {
		query += ` AND s.category = ?`
		args = append(args, category)
	}
	query += ` ORDER BY s.id DESC`

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "查询失败"})
		return
	}
	defer rows.Close()

	materials := []models.StudyMaterial{}
	for rows.Next() {
		var s models.StudyMaterial
		rows.Scan(&s.ID, &s.Title, &s.Content, &s.Category, &s.PublisherID, &s.Publisher,
			&s.ReadCount, &s.CreatedAt, &s.UpdatedAt)
		materials = append(materials, s)
	}
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"list": materials})
}

// CreateStudyMaterial 发布公共资料
func CreateStudyMaterial(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.ContextUserID).(int64)
	var req models.StudyMaterial
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "请求格式错误"})
		return
	}
	if req.Title == "" {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "标题必填"})
		return
	}
	res, err := database.DB.Exec(
		`INSERT INTO study_materials (title, content, category, publisher_id) VALUES (?, ?, ?, ?)`,
		req.Title, req.Content, req.Category, userID)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "创建失败"})
		return
	}
	id, _ := res.LastInsertId()
	middleware.JSON(w, http.StatusOK, map[string]interface{}{"message": "发布成功", "id": id})
}

// GetStudyMaterial 公共资料详情
func GetStudyMaterial(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if id == 0 {
		middleware.JSON(w, http.StatusBadRequest, map[string]string{"error": "缺少资料ID"})
		return
	}
	var s models.StudyMaterial
	err := database.DB.QueryRow(
		`SELECT s.id, s.title, s.content, s.category, s.publisher_id, u.real_name, s.read_count, s.created_at
		 FROM study_materials s LEFT JOIN users u ON s.publisher_id = u.id WHERE s.id = ?`, id).
		Scan(&s.ID, &s.Title, &s.Content, &s.Category, &s.PublisherID, &s.Publisher,
			&s.ReadCount, &s.CreatedAt)
	if err != nil {
		middleware.JSON(w, http.StatusNotFound, map[string]string{"error": "资料不存在"})
		return
	}
	database.DB.Exec("UPDATE study_materials SET read_count = read_count + 1 WHERE id = ?", id)

	// 附件
	rows, err := database.DB.Query("SELECT id, file_name, file_path, file_size, created_at FROM attachments WHERE owner_type='study' AND owner_id=?", id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var a models.Attachment
			rows.Scan(&a.ID, &a.FileName, &a.FilePath, &a.FileSize, &a.CreatedAt)
			s.Attachments = append(s.Attachments, a)
		}
	}

	middleware.JSON(w, http.StatusOK, s)
}

// DeleteStudyMaterial 删除公共资料
func DeleteStudyMaterial(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	_, err := database.DB.Exec("DELETE FROM study_materials WHERE id=?", id)
	if err != nil {
		middleware.JSON(w, http.StatusInternalServerError, map[string]string{"error": "删除失败"})
		return
	}
	middleware.JSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
}
