package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"ynxcb-platform/internal/config"
	"ynxcb-platform/internal/handlers"
	"ynxcb-platform/internal/middleware"
)

func NewRouter(cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	// ---- 公开路由 ----
	mux.HandleFunc("GET /api/health", handlers.Health)
	mux.HandleFunc("POST /api/auth/login", handlers.Login(cfg))

	// ---- 认证与用户 ----
	mux.Handle("GET /api/auth/profile", middleware.Auth(http.HandlerFunc(handlers.GetProfile)))
	mux.Handle("POST /api/auth/change-password", middleware.Auth(http.HandlerFunc(handlers.ChangePassword)))

	// ---- 用户管理（管理员）----
	mux.Handle("GET /api/users", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.ListUsers))))
	mux.Handle("POST /api/users", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateUser))))
	mux.Handle("PUT /api/users", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.UpdateUser))))
	mux.Handle("DELETE /api/users/{id}", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteUser))))
	mux.Handle("POST /api/users/reset-password", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.ResetPassword))))
	mux.Handle("GET /api/roles", middleware.Auth(http.HandlerFunc(handlers.ListRoles)))
	mux.Handle("GET /api/departments", middleware.Auth(http.HandlerFunc(handlers.ListDepartments)))
	mux.Handle("GET /api/assignees", middleware.Auth(http.HandlerFunc(handlers.GetAssignees)))

	// ---- 学习资料 ----
	mux.Handle("GET /api/study-materials", middleware.Auth(http.HandlerFunc(handlers.ListStudyMaterials)))
	mux.Handle("POST /api/study-materials", middleware.Auth(http.HandlerFunc(handlers.CreateStudyMaterial)))
	mux.Handle("GET /api/study-materials/{id}", middleware.Auth(http.HandlerFunc(handlers.GetStudyMaterial)))
	mux.Handle("DELETE /api/study-materials/{id}", middleware.Auth(http.HandlerFunc(handlers.DeleteStudyMaterial)))

	// ---- 通讯录 ----
	mux.Handle("GET /api/contacts", middleware.Auth(http.HandlerFunc(handlers.ListContacts)))
	mux.Handle("POST /api/contacts", middleware.Auth(http.HandlerFunc(handlers.CreateContact)))
	mux.Handle("PUT /api/contacts", middleware.Auth(http.HandlerFunc(handlers.UpdateContact)))
	mux.Handle("DELETE /api/contacts/{id}", middleware.Auth(http.HandlerFunc(handlers.DeleteContact)))

	// ---- 排班 ----
	mux.Handle("GET /api/duty-schedules", middleware.Auth(http.HandlerFunc(handlers.ListDutySchedules)))
	mux.Handle("POST /api/duty-schedules", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.SaveDutySchedule))))
	mux.Handle("DELETE /api/duty-schedules/{id}", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteDutySchedule))))

	// ---- 报表 ----
	mux.Handle("GET /api/reports", middleware.Auth(http.HandlerFunc(handlers.ListReports)))
	mux.Handle("POST /api/reports", middleware.Auth(http.HandlerFunc(handlers.CreateReport)))
	mux.Handle("GET /api/reports/{id}", middleware.Auth(http.HandlerFunc(handlers.GetReport)))
	mux.Handle("POST /api/reports/{id}/status", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.UpdateReportStatus))))
	mux.Handle("GET /api/report-stats", middleware.Auth(http.HandlerFunc(handlers.ReportStats)))

	// ---- 考勤（管理员晨会点到）----
	mux.Handle("POST /api/attendance/mark", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.MarkAttendance))))
	mux.Handle("GET /api/attendance/list", middleware.Auth(http.HandlerFunc(handlers.ListAttendances)))
	mux.Handle("GET /api/attendance/stats", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.AttendanceStats))))
	mux.Handle("GET /api/attendance/dates", middleware.Auth(http.HandlerFunc(handlers.AttendanceDates)))
	mux.Handle("GET /api/attendance/mark-users", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.MarkUsers))))
	mux.Handle("GET /api/attendance/monthly", middleware.Auth(http.HandlerFunc(handlers.AttendanceMonthly)))
	mux.Handle("GET /api/attendance/yearly", middleware.Auth(http.HandlerFunc(handlers.AttendanceYearly)))

	// ---- 请假管理 ----
	mux.Handle("POST /api/leave-records", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateLeaveRecord))))
	mux.Handle("GET /api/leave-records", middleware.Auth(http.HandlerFunc(handlers.ListLeaveRecords)))
	mux.Handle("DELETE /api/leave-records/{id}", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteLeaveRecord))))
	mux.Handle("GET /api/leave-stats", middleware.Auth(http.HandlerFunc(handlers.LeaveStats)))

	// ---- 公车管理 ----
	mux.Handle("GET /api/vehicles", middleware.Auth(http.HandlerFunc(handlers.ListVehicles)))
	mux.Handle("POST /api/vehicles", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.CreateVehicle))))
	mux.Handle("PUT /api/vehicles", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.UpdateVehicle))))
	mux.Handle("DELETE /api/vehicles/{id}", middleware.Auth(middleware.RequireRole("admin")(http.HandlerFunc(handlers.DeleteVehicle))))
	mux.Handle("POST /api/vehicle-applies", middleware.Auth(http.HandlerFunc(handlers.CreateVehicleApply)))
	mux.Handle("GET /api/vehicle-applies", middleware.Auth(http.HandlerFunc(handlers.ListVehicleApplies)))
	mux.Handle("GET /api/vehicle-applies/{id}", middleware.Auth(http.HandlerFunc(handlers.GetVehicleApply)))
	mux.Handle("DELETE /api/vehicle-applies/{id}", middleware.Auth(http.HandlerFunc(handlers.DeleteVehicleApply)))
	mux.Handle("GET /api/vehicle-stats", middleware.Auth(http.HandlerFunc(handlers.VehicleStats)))

	// ---- 文件上传 ----
	mux.Handle("POST /api/uploads", middleware.Auth(handlers.UploadFile(cfg)))
	mux.Handle("GET /api/uploads/{id}", middleware.Auth(http.HandlerFunc(handlers.DownloadAttachment)))
	mux.Handle("PUT /api/uploads/link", middleware.Auth(http.HandlerFunc(handlers.LinkAttachment)))

	// ---- 收文登记 ----
	mux.Handle("GET /api/incoming-docs", middleware.Auth(http.HandlerFunc(handlers.ListIncomingDocs)))
	mux.Handle("POST /api/incoming-docs", middleware.Auth(http.HandlerFunc(handlers.CreateIncomingDoc)))
	mux.Handle("GET /api/incoming-docs/{id}", middleware.Auth(http.HandlerFunc(handlers.GetIncomingDoc)))
	mux.Handle("PUT /api/incoming-docs", middleware.Auth(http.HandlerFunc(handlers.UpdateIncomingDoc)))
	mux.Handle("DELETE /api/incoming-docs/{id}", middleware.Auth(http.HandlerFunc(handlers.DeleteIncomingDoc)))
	mux.Handle("GET /api/incoming-doc-stats", middleware.Auth(http.HandlerFunc(handlers.IncomingDocStats)))
	// 传阅记录
	mux.Handle("POST /api/circulations", middleware.Auth(http.HandlerFunc(handlers.AddCirculation)))
	mux.Handle("PUT /api/circulations", middleware.Auth(http.HandlerFunc(handlers.UpdateCirculation)))
	mux.Handle("DELETE /api/circulations/{id}", middleware.Auth(http.HandlerFunc(handlers.DeleteCirculation)))

	// ---- 首页统计 ----
	mux.Handle("GET /api/dashboard-stats", middleware.Auth(http.HandlerFunc(handlers.DashboardStats)))

	// ---- 数据导出（Excel）----
	mux.Handle("GET /api/export/vehicle-applies", middleware.Auth(http.HandlerFunc(handlers.ExportVehicleApplies)))
	mux.Handle("GET /api/export/leave-records", middleware.Auth(http.HandlerFunc(handlers.ExportLeaveRecords)))
	mux.Handle("GET /api/export/attendances", middleware.Auth(http.HandlerFunc(handlers.ExportAttendances)))
	mux.Handle("GET /api/export/duty-schedules", middleware.Auth(http.HandlerFunc(handlers.ExportDutySchedules)))
	mux.Handle("GET /api/export/incoming-docs", middleware.Auth(http.HandlerFunc(handlers.ExportIncomingDocs)))

	// 静态文件服务（前端构建产物）+ SPA 回退
	staticDir := "static"
	mux.Handle("GET /", spaHandler{staticDir: staticDir})

	return mux
}

// spaHandler 支持 history 路由的静态文件服务
type spaHandler struct {
	staticDir string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 路径穿越防护
	path := filepath.Clean(r.URL.Path)
	rel := strings.TrimPrefix(path, string(os.PathSeparator))
	fullPath := filepath.Join(h.staticDir, rel)
	absStatic, _ := filepath.Abs(h.staticDir)
	absFull, _ := filepath.Abs(fullPath)
	if absFull != absStatic && !strings.HasPrefix(absFull, absStatic+string(os.PathSeparator)) {
		http.NotFound(w, r)
		return
	}
	// 如果请求的文件存在则直接提供，否则回退到 index.html
	if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, fullPath)
		return
	}
	http.ServeFile(w, r, filepath.Join(h.staticDir, "index.html"))
}
