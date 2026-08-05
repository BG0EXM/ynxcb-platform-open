package models

import "time"

// User 系统用户
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	RealName     string    `json:"real_name"`
	Phone        string    `json:"phone"`
	DepartmentID int64     `json:"department_id"`
	Department   string    `json:"department_name,omitempty"`
	RoleID       int64     `json:"role_id"`
	RoleName     string    `json:"role_name,omitempty"`
	RoleCode     string    `json:"role_code,omitempty"`
	Status       int       `json:"status"` // 1启用 0禁用
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Role 角色
type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"` // admin / staff / reporter
	Description string `json:"description"`
}

// Department 部门/科室
type Department struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID int64  `json:"parent_id"`
	Sort     int    `json:"sort"`
}

// StudyMaterial 学习资料
type StudyMaterial struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Content     string       `json:"content"`
	Category    string       `json:"category"`
	PublisherID int64        `json:"publisher_id"`
	Publisher   string       `json:"publisher_name,omitempty"`
	ReadCount   int          `json:"read_count"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Contact 通讯录
type Contact struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	DepartmentID int64  `json:"department_id"`
	Department   string `json:"department_name,omitempty"`
	Position     string `json:"position"`
	IsPublic     int    `json:"is_public"`
	Sort         int    `json:"sort"`
}

// DutySchedule 值守排班（当天值守，至晚上9点收文；可标记县委大院附加排班）
type DutySchedule struct {
	ID           int64     `json:"id"`
	DutyDate     string    `json:"duty_date"` // YYYY-MM-DD
	UserID       int64     `json:"user_id"`
	UserName     string    `json:"user_name,omitempty"`
	IsDaWangYuan int       `json:"is_dawangyuan"` // 1=当天同时有县委大院排班
	Note         string    `json:"note"`
	Status       int       `json:"status"` // 1正常 2已完成
	CreatedAt    time.Time `json:"created_at"`
}

// Report 周报/月报/年报
type Report struct {
	ID          int64     `json:"id"`
	ReportType  string    `json:"report_type"` // weekly / monthly / yearly
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Period      string    `json:"period"` // 周期标识,如 2026-W32
	SubmitterID int64     `json:"submitter_id"`
	Submitter   string    `json:"submitter_name,omitempty"`
	Status      int       `json:"status"` // 1草稿 2已提交 3已审阅
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment 文件附件
type Attachment struct {
	ID           int64     `json:"id"`
	OwnerType    string    `json:"owner_type"` // document / study / report
	OwnerID      int64     `json:"owner_id"`
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	UploaderID   int64     `json:"uploader_id"`
	UploaderName string    `json:"uploader_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token      string `json:"token"`
	User       User   `json:"user"`
	MustChange bool   `json:"must_change"` // 是否需强制修改默认密码
}

// IncomingDoc 收文登记（上级来文）
type IncomingDoc struct {
	ID           int64     `json:"id"`
	ReceiveNo    string    `json:"receive_no"`     // 收文编号
	ReceivedDate string    `json:"received_date"`  // 收文日期 YYYY-MM-DD
	FromUnit     string    `json:"from_unit"`      // 来文单位
	FromDocNo    string    `json:"from_doc_no"`    // 来文字号
	DocNo        string    `json:"doc_no"`         // 文件编号
	Title        string    `json:"title"`          // 文件标题
	Copies       int       `json:"copies"`         // 份数
	SecretLevel  string    `json:"secret_level"`   // 密级
	Urgency      string    `json:"urgency"`        // 紧急程度
	Suggest      string    `json:"suggest"`        // 拟办意见
	LeaderComment string   `json:"leader_comment"` // 领导批示
	Processing   string    `json:"processing"`     // 办理情况
	ReturnDate   string    `json:"return_date"`    // 退回日期
	Returned     int       `json:"returned"`       // 是否已退 1已退 0未退
	NeedReturn   int       `json:"need_return"`    // 是否需要退回 1需要 0不需要
	RegistrarID  int64     `json:"registrar_id"`
	Registrar    string    `json:"registrar_name,omitempty"`
	Status       int       `json:"status"` // 1待登记 2拟办中 3待批示 4办理中 5已办结
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CircList     []CirculationRecord `json:"circulations,omitempty"`
}

// CirculationRecord 传阅记录
type CirculationRecord struct {
	ID        int64  `json:"id"`
	DocID     int64  `json:"doc_id"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name,omitempty"`
	OrderNo   int    `json:"order_no"`
	ReadDate  string `json:"read_date"`
	Signature string `json:"signature"`
}

// Attendance 考勤记录（管理员晨会点到）
type Attendance struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name,omitempty"`
	AttendDate string    `json:"attend_date"` // YYYY-MM-DD
	Status     int       `json:"status"` // 1出勤 2请假 3出差 4未到
	LeaveType  string    `json:"leave_type"` // 请假类型（年假/事假/病假/婚假/产假/丧假/其他）
	Remark     string    `json:"remark"`
	MarkedBy   int64     `json:"marked_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// LeaveRecord 请假记录（年假、特殊假期等）
type LeaveRecord struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	LeaveType string    `json:"leave_type"` // annual年假 / sick病假 / personal事假 / marriage婚假 / maternity产假 / bereavement丧假 / other其他
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Days      float64   `json:"days"`
	Reason    string    `json:"reason"`
	Status    int       `json:"status"` // 1已登记
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Vehicle 公车信息
type Vehicle struct {
	ID            int64  `json:"id"`
	PlateNo       string `json:"plate_no"`       // 车牌号
	Brand         string `json:"brand"`          // 品牌型号
	Seats         int    `json:"seats"`          // 座位数
	Driver        string `json:"driver"`         // 司机
	Status        int    `json:"status"`         // 1可用 2使用中 3维修中 4已报废
	Vin           string `json:"vin"`            // 车架号(VIN)
	EngineNo      string `json:"engine_no"`      // 发动机号
	InsuranceDate string `json:"insurance_date"` // 保险到期日期
	InspectDate   string `json:"inspect_date"`   // 年检到期日期
	RegisterDate  string `json:"register_date"`  // 登记日期
	PurchaseAt    string `json:"purchase_at"`    // 购置日期
	Note          string `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

// VehicleApply 用车报备（无需审批）
type VehicleApply struct {
	ID          int64     `json:"id"`
	VehicleID   int64     `json:"vehicle_id"`
	VehicleNo   string    `json:"vehicle_no,omitempty"`
	VehicleBrand string   `json:"vehicle_brand,omitempty"`
	VehicleDriver string  `json:"vehicle_driver,omitempty"`
	ReporterID  int64     `json:"reporter_id"`
	Reporter    string    `json:"reporter_name,omitempty"`
	UserName    string    `json:"user_name"`    // 用车人
	Purpose     string    `json:"purpose"`      // 用车事由
	Destination string    `json:"destination"`  // 目的地
	UseDate     string    `json:"use_date"`     // 用车日期
	UseTime     string    `json:"use_time"`     // 用车时间
	Passengers  int       `json:"passengers"`
	CreatedAt   time.Time `json:"created_at"`
}
