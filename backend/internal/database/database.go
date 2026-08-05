package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Init 初始化数据库连接和表结构
func Init(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("创建数据库目录失败: %v", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	DB.SetMaxOpenConns(1) // SQLite 单写者，避免锁竞争
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(time.Hour)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 开启 WAL 模式提升并发读性能
	DB.Exec("PRAGMA journal_mode=WAL")
	DB.Exec("PRAGMA busy_timeout=5000")

	if err := createTables(); err != nil {
		return err
	}

	if err := migrate(); err != nil {
		return err
	}

	if err := seed(); err != nil {
		return err
	}

	return nil
}

// migrate 数据库结构迁移
func migrate() error {
	// 迁移条件1：旧结构（含 shift 列）或含 is_xianweiban 列的表，重建为最新结构
	// 迁移条件2：duty_schedules 仅有单列 UNIQUE(duty_date)，需升级为 UNIQUE(duty_date, user_id) 支持一天两人
	if hasColumn("duty_schedules", "shift") || hasColumn("duty_schedules", "is_xianweiban") || hasSingleDateUnique("duty_schedules") {
		DB.Exec(`CREATE TABLE duty_schedules_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			duty_date TEXT,
			user_id INTEGER,
			is_dawangyuan INTEGER DEFAULT 0,
			note TEXT,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(duty_date, user_id)
		)`)
		DB.Exec(`INSERT OR IGNORE INTO duty_schedules_new (id, duty_date, user_id, is_dawangyuan, note, status, created_at)
			SELECT id, duty_date, user_id, is_dawangyuan, note, status, created_at FROM duty_schedules`)
		DB.Exec("DROP TABLE duty_schedules")
		DB.Exec("ALTER TABLE duty_schedules_new RENAME TO duty_schedules")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_schedule_date ON duty_schedules(duty_date)")
	}

	// 考勤表迁移：旧结构含 check_in_time 打卡字段 → 重建为点到结构
	if hasColumn("attendances", "check_in_time") {
		DB.Exec(`CREATE TABLE attendances_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			attend_date TEXT,
			status INTEGER DEFAULT 1,
			leave_type TEXT,
			remark TEXT,
			marked_by INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, attend_date)
		)`)
		DB.Exec(`INSERT OR IGNORE INTO attendances_new (id, user_id, attend_date, status, remark, created_at, updated_at)
			SELECT id, user_id, attend_date, status, remark, created_at, updated_at FROM attendances`)
		DB.Exec("DROP TABLE attendances")
		DB.Exec("ALTER TABLE attendances_new RENAME TO attendances")
	}

	// 考勤表补充列：已有表但缺 leave_type 时补充
	if !hasColumn("attendances", "leave_type") {
		DB.Exec("ALTER TABLE attendances ADD COLUMN leave_type TEXT")
	}

	// 用车表迁移：旧结构含审批字段 → 重建为报备结构
	if hasColumn("vehicle_applies", "approver_id") {
		DB.Exec(`CREATE TABLE vehicle_applies_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vehicle_id INTEGER,
			reporter_id INTEGER,
			user_name TEXT,
			purpose TEXT,
			destination TEXT,
			use_date TEXT,
			use_time TEXT,
			passengers INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`)
		DB.Exec(`INSERT OR IGNORE INTO vehicle_applies_new (id, vehicle_id, reporter_id, purpose, destination, created_at)
			SELECT id, vehicle_id, applicant_id, purpose, destination, created_at FROM vehicle_applies`)
		DB.Exec("DROP TABLE vehicle_applies")
		DB.Exec("ALTER TABLE vehicle_applies_new RENAME TO vehicle_applies")
	}

	// 确保新结构索引存在（放在迁移之后，避免旧表结构下建索引报错）
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_att_user ON attendances(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_att_date ON attendances(attend_date)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_apply_vehicle ON vehicle_applies(vehicle_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_apply_reporter ON vehicle_applies(reporter_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_apply_date ON vehicle_applies(use_date)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_leave_user ON leave_records(user_id)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_leave_type ON leave_records(leave_type)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_leave_date ON leave_records(start_date)")

	// 车辆表补充列：已有表缺车架号等字段时补充
	for _, col := range []string{"vin", "engine_no", "insurance_date", "inspect_date", "register_date"} {
		if !hasColumn("vehicles", col) {
			DB.Exec("ALTER TABLE vehicles ADD COLUMN " + col + " TEXT")
		}
	}

	// 收文表补充列：文件编号/退回相关字段
	if !hasColumn("incoming_docs", "doc_no") {
		DB.Exec("ALTER TABLE incoming_docs ADD COLUMN doc_no TEXT")
	}
	if !hasColumn("incoming_docs", "return_date") {
		DB.Exec("ALTER TABLE incoming_docs ADD COLUMN return_date TEXT")
	}
	if !hasColumn("incoming_docs", "returned") {
		DB.Exec("ALTER TABLE incoming_docs ADD COLUMN returned INTEGER DEFAULT 0")
	}
	if !hasColumn("incoming_docs", "need_return") {
		DB.Exec("ALTER TABLE incoming_docs ADD COLUMN need_return INTEGER DEFAULT 0")
	}

	return nil
}

// hasSingleDateUnique 检测 duty_schedules 是否仅有单列 UNIQUE(duty_date) 约束
// 通过 sqlite_master 中的建表语句判断，避免 PRAGMA 嵌套查询死锁
func hasSingleDateUnique(table string) bool {
	var sqlText string
	err := DB.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&sqlText)
	if err != nil {
		return false
	}
	// 单列约束形式：UNIQUE(duty_date)（后面紧跟 ) 或 ,）区别于 UNIQUE(duty_date, user_id)
	if strings.Contains(sqlText, "UNIQUE(duty_date)") {
		return true
	}
	return false
}

// hasColumn 检查表是否存在某列
// 使用 pragma_table_info 表值函数，单条查询避免 PRAGMA 嵌套死锁
// 表名和列名均为代码内常量，无注入风险
func hasColumn(table, col string) bool {
	var cnt int
	err := DB.QueryRow("SELECT COUNT(*) FROM pragma_table_info('" + table + "') WHERE name='" + col + "'").Scan(&cnt)
	if err != nil {
		return false
	}
	return cnt > 0
}

func createTables() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS departments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			parent_id INTEGER DEFAULT 0,
			sort INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			code TEXT NOT NULL UNIQUE,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			real_name TEXT NOT NULL,
			phone TEXT,
			department_id INTEGER DEFAULT 0,
			role_id INTEGER DEFAULT 3,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS study_materials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT,
			category TEXT,
			publisher_id INTEGER,
			read_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			phone TEXT,
			department_id INTEGER DEFAULT 0,
			position TEXT,
			is_public INTEGER DEFAULT 1,
			sort INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS duty_schedules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			duty_date TEXT,
			user_id INTEGER,
			is_dawangyuan INTEGER DEFAULT 0,
			note TEXT,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(duty_date, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			report_type TEXT,
			title TEXT NOT NULL,
			content TEXT,
			period TEXT,
			submitter_id INTEGER,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS leave_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			leave_type TEXT,
			start_date TEXT,
			end_date TEXT,
			days REAL DEFAULT 0,
			reason TEXT,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS attendances (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			attend_date TEXT,
			status INTEGER DEFAULT 1,
			leave_type TEXT,
			remark TEXT,
			marked_by INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, attend_date)
		)`,
		`CREATE TABLE IF NOT EXISTS vehicles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			plate_no TEXT NOT NULL,
			brand TEXT,
			seats INTEGER DEFAULT 5,
			driver TEXT,
			status INTEGER DEFAULT 1,
			vin TEXT,
			engine_no TEXT,
			insurance_date TEXT,
			inspect_date TEXT,
			register_date TEXT,
			purchase_at TEXT,
			note TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS vehicle_applies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vehicle_id INTEGER,
			reporter_id INTEGER,
			user_name TEXT,
			purpose TEXT,
			destination TEXT,
			use_date TEXT,
			use_time TEXT,
			passengers INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS incoming_docs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			receive_no TEXT,
			received_date TEXT,
			from_unit TEXT,
			from_doc_no TEXT,
			doc_no TEXT,
			title TEXT NOT NULL,
			copies INTEGER DEFAULT 1,
			secret_level TEXT DEFAULT '普通',
			urgency TEXT DEFAULT '一般',
			suggest TEXT,
			leader_comment TEXT,
			processing TEXT,
			return_date TEXT,
			returned INTEGER DEFAULT 0,
			need_return INTEGER DEFAULT 0,
			registrar_id INTEGER,
			status INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS circulation_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			doc_id INTEGER,
			user_id INTEGER,
			order_no INTEGER DEFAULT 0,
			read_date TEXT,
			signature TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS attachments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner_type TEXT,
			owner_id INTEGER,
			file_name TEXT,
			file_path TEXT,
			file_size INTEGER DEFAULT 0,
			uploader_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_submitter ON reports(submitter_id)`,
		`CREATE INDEX IF NOT EXISTS idx_schedule_date ON duty_schedules(duty_date)`,
		`CREATE INDEX IF NOT EXISTS idx_incoming_status ON incoming_docs(status)`,
		`CREATE INDEX IF NOT EXISTS idx_incoming_date ON incoming_docs(received_date)`,
		`CREATE INDEX IF NOT EXISTS idx_circ_doc ON circulation_records(doc_id)`,
	}

	for _, s := range stmts {
		if _, err := DB.Exec(s); err != nil {
			return fmt.Errorf("建表失败: %v\nSQL: %s", err, s)
		}
	}
	return nil
}

// seed 初始化基础数据
func seed() error {
	// 角色
	roles := []struct {
		name, code, desc string
	}{
		{"系统管理员", "admin", "系统管理，用户管理，全部权限"},
		{"科室工作人员", "staff", "业务办理，日常操作"},
		{"乡镇/通讯员", "reporter", "投稿、接收通知、上报材料"},
	}

	// 检查 roles 是否为空
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM roles").Scan(&count)
	if count == 0 {
		for _, r := range roles {
			_, err := DB.Exec("INSERT INTO roles (name, code, description) VALUES (?, ?, ?)", r.name, r.code, r.desc)
			if err != nil {
				return fmt.Errorf("初始化角色失败: %v", err)
			}
		}
	}

	// 默认部门
	var deptCount int
	DB.QueryRow("SELECT COUNT(*) FROM departments").Scan(&deptCount)
	if deptCount == 0 {
		defaultDepts := []string{"办公室", "新闻宣传科", "理论教育科", "文明创建科", "网络舆情科", "文化文艺科", "乡镇通讯员"}
		for i, name := range defaultDepts {
			_, err := DB.Exec("INSERT INTO departments (name, parent_id, sort) VALUES (?, 0, ?)", name, i+1)
			if err != nil {
				return fmt.Errorf("初始化部门失败: %v", err)
			}
		}
	}

	// 默认管理员 admin / admin123
	var userCount int
	DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if userCount == 0 {
		hash, err := HashPassword("admin123")
		if err != nil {
			return fmt.Errorf("生成默认密码失败: %v", err)
		}
		var roleID int64
		DB.QueryRow("SELECT id FROM roles WHERE code='admin'").Scan(&roleID)
		var deptID int64
		DB.QueryRow("SELECT id FROM departments WHERE name='办公室'").Scan(&deptID)
		_, err = DB.Exec("INSERT INTO users (username, password_hash, real_name, phone, department_id, role_id, status) VALUES (?, ?, ?, ?, ?, ?, 1)",
			"admin", hash, "系统管理员", "13800000000", deptID, roleID)
		if err != nil {
			return fmt.Errorf("初始化管理员失败: %v", err)
		}
	}

	return nil
}
