package model

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int64          `db:"id" json:"id"`
	Name         string         `db:"name" json:"name"`
	Phone        string         `db:"phone" json:"phone"`
	PasswordHash string         `db:"password_hash" json:"-"`
	Role         string         `db:"role" json:"role"`
	RoleName     string         `db:"role_name" json:"role_name"`
	StoreID      sql.NullInt64  `db:"store_id" json:"-"`
	StoreName    string         `db:"store_name" json:"store_name,omitempty"`
	Position     string         `db:"position" json:"position"`
	HireDate     sql.NullTime   `db:"hire_date" json:"hire_date,omitempty"`
	Status       string         `db:"status" json:"status"`
	LevelCode    string         `db:"level_code" json:"level"`
	WorkStart    string         `db:"work_start" json:"work_start"`
	WorkEnd      string         `db:"work_end" json:"work_end"`
	BaseSalary   float64        `db:"base_salary" json:"base_salary"`
	Avatar       string         `db:"avatar" json:"avatar"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

func (u User) StoreIDValue() *int64 {
	if u.StoreID.Valid {
		v := u.StoreID.Int64
		return &v
	}
	return nil
}

type PublicUser struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	Role      string  `json:"role"`
	RoleName  string  `json:"role_name"`
	StoreID   *int64  `json:"store_id,omitempty"`
	StoreName string  `json:"store_name,omitempty"`
	Position  string  `json:"position"`
	Status    string  `json:"status"`
	Level     string  `json:"level"`
	BaseSalary float64 `json:"base_salary"`
}

func (u User) Public() PublicUser {
	return PublicUser{
		ID: u.ID, Name: u.Name, Phone: u.Phone, Role: u.Role, RoleName: u.RoleName,
		StoreID: u.StoreIDValue(), StoreName: u.StoreName, Position: u.Position,
		Status: u.Status, Level: u.LevelCode, BaseSalary: u.BaseSalary,
	}
}

type Store struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Address     string    `db:"address" json:"address"`
	Lat         float64   `db:"lat" json:"lat"`
	Lng         float64   `db:"lng" json:"lng"`
	Phone       string    `db:"phone" json:"phone"`
	ManagerName string    `db:"manager_name" json:"manager_name"`
	ManagerID   *int64    `db:"manager_id" json:"manager_id,omitempty"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Project struct {
	ID              int64   `db:"id" json:"id"`
	Name            string  `db:"name" json:"name"`
	Category        string  `db:"category" json:"category"`
	Price           float64 `db:"price" json:"price"`
	MemberPrice     float64 `db:"member_price" json:"member_price"`
	TotalTimes      int     `db:"total_times" json:"total_times"`
	CardMode        string  `db:"card_mode" json:"card_mode"` // times | amount | both
	ValidDays       int     `db:"valid_days" json:"valid_days"`
	CommissionRate  float64 `db:"commission_rate" json:"commission_rate"`
	Duration        int     `db:"duration" json:"duration"`
	Status          string  `db:"status" json:"status"`
}

type Customer struct {
	ID          int64          `db:"id" json:"id"`
	CustomerNo  string         `db:"customer_no" json:"code"`
	Name        string         `db:"name" json:"name"`
	Mobile      string         `db:"mobile" json:"phone"`
	Gender      int            `db:"gender" json:"gender"`
	Birthday    sql.NullTime   `db:"birthday" json:"-"`
	BirthdayStr string         `json:"birthday,omitempty"`
	LevelName   string         `db:"level_name" json:"level"`
	Source      string         `db:"source" json:"source"`
	StoreID     int64          `db:"store_id" json:"store_id"`
	StoreName   string         `json:"store_name,omitempty"`
	EmployeeID  sql.NullInt64  `db:"employee_id" json:"employee_id,omitempty"`
	TotalAmount float64        `db:"total_amount" json:"spent"`
	LastVisitAt sql.NullTime   `db:"last_visit_at" json:"-"`
	LastVisit   string         `json:"last_visit,omitempty"`
	Status      int            `db:"status" json:"status"`
	Remark      string         `db:"remark" json:"remark"`
	Remain      int            `json:"remain"`             // 次卡剩余次数合计（来自开卡）
	Balance     float64        `json:"balance"`            // 储值卡余额合计（来自开卡）
	CardsBrief  string         `json:"cards_brief"`        // 列表摘要：项目次数/余额
	CardsCount  int            `json:"cards_count"`        // 使用中卡数
	Cards       []Card         `json:"cards,omitempty"`    // 详情卡项
}

type Card struct {
	ID           int64         `db:"id" json:"id"`
	CardNo       string        `db:"card_no" json:"card_no"`
	CustomerID   int64         `db:"customer_id" json:"customer_id"`
	ProjectID    sql.NullInt64 `db:"project_id" json:"-"`
	ProjectIDVal int64         `json:"project_id"`
	StoreID      int64         `db:"store_id" json:"store_id"`
	EmployeeID   sql.NullInt64 `db:"employee_id" json:"employee_id,omitempty"`
	TotalTimes   int           `db:"total_times" json:"total"`
	UsedTimes    int           `db:"used_times" json:"used"`
	RemainTimes  int           `db:"remain_times" json:"remain"`
	FrozenTimes  int           `db:"frozen_times" json:"frozen"`
	TotalAmount  float64       `db:"total_amount" json:"amount"`
	PayMethod    string        `db:"pay_method" json:"pay_method"`
	CardKind     string        `db:"card_kind" json:"card_kind"`
	Status       string        `db:"status" json:"status"`
	Version      int           `db:"version" json:"version"`
	CustomerName string        `json:"customer_name,omitempty"`
	ProjectName  string        `json:"project_name,omitempty"`
	StoreName    string        `json:"store_name,omitempty"`
	CreatedAt    time.Time     `db:"created_at" json:"created_at"`
}

type VerifyOrder struct {
	ID               int64          `db:"id" json:"id"`
	VerifyNo         string         `db:"verify_no" json:"verify_no"`
	CustomerID       int64          `db:"customer_id" json:"customer_id"`
	CardID           int64          `db:"card_id" json:"card_id"`
	ProjectID        int64          `db:"project_id" json:"project_id"`
	StoreID          int64          `db:"store_id" json:"store_id"`
	EmployeeID       int64          `db:"employee_id" json:"employee_id"`
	TotalAmount      float64        `db:"total_amount" json:"amount"`
	Times            int            `db:"times" json:"times"`
	WelfareIssuedID  sql.NullInt64  `db:"welfare_issued_id" json:"-"`
	WelfareIssuedVal int64          `json:"welfare_issued_id,omitempty"`
	WelfareOffset    float64        `db:"welfare_offset" json:"welfare_offset"`
	PayAmount        float64        `db:"pay_amount" json:"pay_amount"`
	Status           string         `db:"status" json:"status"`
	Token            string         `db:"token" json:"token,omitempty"`
	TokenExpireAt    sql.NullTime   `db:"token_expire_at" json:"-"`
	ScannedAt        sql.NullTime   `db:"scanned_at" json:"-"`
	Remark           string         `db:"remark" json:"remark"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	CustomerName     string         `json:"customer_name,omitempty"`
	ProjectName      string         `json:"project_name,omitempty"`
	StaffName        string         `json:"staff_name,omitempty"`
	StoreName        string         `json:"store_name,omitempty"`
	CardKind         string         `json:"card_kind,omitempty"`
}

type Appointment struct {
	ID           int64         `db:"id" json:"id"`
	CustomerID   int64         `db:"customer_id" json:"customer_id"`
	ProjectID    int64         `db:"project_id" json:"project_id"`
	StoreID      int64         `db:"store_id" json:"store_id"`
	StaffID      sql.NullInt64 `db:"staff_id" json:"-"`
	StaffIDVal   int64         `json:"staff_id"`
	AppointDate  time.Time     `db:"appoint_date" json:"-"`
	Date         string        `json:"date"`
	AppointTime  string        `db:"appoint_time" json:"time"`
	Status       string        `db:"status" json:"status"`
	Source       string        `db:"source" json:"source"`
	Note         string        `db:"note" json:"note"`
	CustomerName string        `json:"customer_name,omitempty"`
	ProjectName  string        `json:"project_name,omitempty"`
	StaffName    string        `json:"staff_name,omitempty"`
}

type Page[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}
