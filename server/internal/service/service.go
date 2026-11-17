package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lanping/lp-server/internal/auth"
	"github.com/lanping/lp-server/internal/model"
	"github.com/lanping/lp-server/internal/repo"
	"github.com/lanping/lp-server/internal/sms"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *repo.Store
	JWT  *auth.Manager
	SMS  *sms.Client
}

func New(r *repo.Store, jwt *auth.Manager) *Service {
	return &Service{Repo: r, JWT: jwt}
}

func (s *Service) SetSMS(c *sms.Client) {
	s.SMS = c
}

func (s *Service) smsEnabled() bool {
	if s.SettingValue("sms_enabled", "true") == "false" {
		return false
	}
	return s.SMS != nil
}

// ReloadSMSFromSettings 从 lp_settings 热加载阿里云短信配置（管理后台保存后生效）。
// 库中为空的字段会保留当前客户端已有值（便于 yaml 启动引导）。
func (s *Service) ReloadSMSFromSettings() error {
	if s.SMS == nil {
		cli, err := sms.New(sms.Config{SignName: "兰瓶"})
		if err != nil {
			return err
		}
		s.SMS = cli
	}
	m, err := s.GetSettings()
	if err != nil {
		return err
	}
	cur := s.SMS.Config()
	pick := func(v, fallback string) string {
		v = strings.TrimSpace(v)
		if v == "" {
			return fallback
		}
		return v
	}
	enabled := cur.Enabled
	if v, ok := m["sms_enabled"]; ok {
		enabled = strings.ToLower(strings.TrimSpace(v)) != "false"
	}
	cfg := sms.Config{
		Enabled:         enabled,
		AccessKeyID:     pick(m["sms_access_key_id"], cur.AccessKeyID),
		AccessKeySecret: pick(m["sms_access_key_secret"], cur.AccessKeySecret),
		Endpoint:        pick(m["sms_endpoint"], cur.Endpoint),
		SignName:        pick(m["sms_sign"], cur.SignName),
		TplBirthday:     pick(m["sms_tpl_birthday"], cur.TplBirthday),
		TplConsume:      pick(m["sms_tpl_consume"], cur.TplConsume),
		TplAppoint:      pick(m["sms_tpl_appoint"], cur.TplAppoint),
	}
	if cfg.SignName == "" {
		cfg.SignName = "兰瓶"
	}
	return s.SMS.Reconfigure(cfg)
}

func (s *Service) logSMS(mobile, tpl, content, status string) {
	_, _ = s.Repo.DB.Exec(`INSERT INTO lp_sms_log(mobile,template_code,content,status) VALUES(?,?,?,?)`, mobile, tpl, content, status)
}

func (s *Service) notifyConsume(customerID int64, projectName string, amount float64) {
	if !s.smsEnabled() {
		return
	}
	var name, mobile string
	_ = s.Repo.DB.QueryRow(`SELECT name, mobile FROM lp_customer WHERE id=?`, customerID).Scan(&name, &mobile)
	if mobile == "" {
		return
	}
	amt := fmt.Sprintf("%.2f", amount)
	when := time.Now().Format("01-02 15:04")
	res, err := s.SMS.SendConsume(mobile, name, when, projectName, amt)
	st := "已发送"
	content := fmt.Sprintf("【%s】消费通知 %s %s元", s.SMS.SignName(), projectName, amt)
	if err != nil {
		st = "失败:" + err.Error()
	} else if res.Mock {
		st = "模拟未发"
		content = res.Message
	} else if res.BizID != "" {
		content += " bizId=" + res.BizID
	}
	s.logSMS(mobile, "SMS_CONSUME", content, st)
}

func (s *Service) notifyAppoint(customerID int64, date, timeStr, projectName, staffName string) {
	if !s.smsEnabled() {
		return
	}
	var name, mobile string
	_ = s.Repo.DB.QueryRow(`SELECT name, mobile FROM lp_customer WHERE id=?`, customerID).Scan(&name, &mobile)
	if mobile == "" {
		return
	}
	if staffName == "" {
		staffName = "门店美容师"
	}
	res, err := s.SMS.SendAppoint(mobile, name, date, timeStr, projectName, staffName)
	st := "已发送"
	content := fmt.Sprintf("【%s】预约 %s %s %s", s.SMS.SignName(), date, timeStr, projectName)
	if err != nil {
		st = "失败:" + err.Error()
	} else if res.Mock {
		st = "模拟未发"
		content = res.Message
	} else if res.BizID != "" {
		content += " bizId=" + res.BizID
	}
	s.logSMS(mobile, "SMS_APPOINT", content, st)
}

// NotifyBirthday 发送生日祝福；days 为「未来 N 天内生日」；customerID>0 则只发该客户。
func (s *Service) NotifyBirthday(days int, customerID int64) (sent int, failed int, err error) {
	if !s.smsEnabled() {
		return 0, 0, errors.New("短信未启用或未配置 AccessKey")
	}
	if days <= 0 {
		days = 7
	}
	q := `SELECT id,name,mobile,birthday FROM lp_customer WHERE deleted_at IS NULL AND birthday IS NOT NULL AND mobile<>''`
	args := []interface{}{}
	if customerID > 0 {
		q += " AND id=?"
		args = append(args, customerID)
	}
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()
	now := time.Now()
	for rows.Next() {
		var id int64
		var name, mobile string
		var birthday time.Time
		if err := rows.Scan(&id, &name, &mobile, &birthday); err != nil {
			continue
		}
		bThis := time.Date(now.Year(), birthday.Month(), birthday.Day(), 0, 0, 0, 0, now.Location())
		if bThis.Before(now.Truncate(24 * time.Hour)) {
			bThis = bThis.AddDate(1, 0, 0)
		}
		diff := int(bThis.Sub(now.Truncate(24*time.Hour)).Hours() / 24)
		if customerID == 0 && (diff < 0 || diff > days) {
			continue
		}
		res, e := s.SMS.SendBirthday(mobile, name)
		st := "已发送"
		content := fmt.Sprintf("【%s】生日祝福 %s", s.SMS.SignName(), name)
		if e != nil {
			failed++
			st = "失败:" + e.Error()
		} else {
			sent++
			if res.Mock {
				st = "模拟未发"
				content = res.Message
			} else if res.BizID != "" {
				content += " bizId=" + res.BizID
			}
		}
		s.logSMS(mobile, "SMS_BIRTHDAY", content, st)
	}
	return sent, failed, nil
}

func (s *Service) Login(phone, password string) (token string, user model.PublicUser, err error) {
	u, err := s.Repo.FindUserByPhone(phone)
	if err != nil {
		return "", user, err
	}
	if u == nil || u.Status != "active" {
		return "", user, errors.New("手机号或密码错误")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", user, errors.New("手机号或密码错误")
	}
	token, err = s.JWT.Issue(u.ID, u.Role, u.StoreIDValue())
	if err != nil {
		return "", user, err
	}
	return token, u.Public(), nil
}

func (s *Service) Register(name, phone, password, role string, storeID int64) error {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	if name == "" || phone == "" {
		return errors.New("请填写姓名和手机号")
	}
	if storeID <= 0 {
		return errors.New("请选择所在门店")
	}
	var storeCnt int
	_ = s.Repo.DB.Get(&storeCnt, `SELECT COUNT(*) FROM lp_store WHERE id=?`, storeID)
	if storeCnt == 0 {
		return errors.New("门店不存在")
	}
	exist, err := s.Repo.FindUserByPhone(phone)
	if err != nil {
		return err
	}
	if exist != nil {
		return errors.New("手机号已注册")
	}
	var pending int
	_ = s.Repo.DB.Get(&pending, `SELECT COUNT(*) FROM lp_audit_apply WHERE phone=? AND status IN ('pending','待审核')`, phone)
	if pending > 0 {
		return errors.New("该手机号已有待审核申请")
	}
	if role != "manager" && role != "staff" {
		role = "staff"
	}
	_, err = s.Repo.DB.Exec(`INSERT INTO lp_audit_apply(name,phone,store_id,role,status) VALUES(?,?,?,?, '待审核')`, name, phone, storeID, role)
	return err
}

func (s *Service) HashPassword(p string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(b), err
}

func pageParams(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 200 {
		size = 200
	}
	return page, size
}

func (s *Service) ScopeStore(claims *auth.Claims, queryStoreID int64) *int64 {
	if claims.Role == "admin" {
		if queryStoreID > 0 {
			return &queryStoreID
		}
		return nil
	}
	return claims.StoreID
}

// NormalizeVerifyMode 统一核销模式：staff=不发通知员工自助；notify=发通知客户亮码。
// 兼容旧值 mock→staff、real→notify。
func (s *Service) NormalizeVerifyMode(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "notify", "real":
		return "notify"
	default:
		return "staff"
	}
}

func (s *Service) VerifyMode() string {
	return s.NormalizeVerifyMode(s.SettingValue("verify_mode", "staff"))
}

func (s *Service) SettingValue(key, def string) string {
	m, err := s.GetSettings()
	if err != nil {
		return def
	}
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return def
}

func (s *Service) SettingBool(key string, def bool) bool {
	v := strings.ToLower(s.SettingValue(key, ""))
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes" || v == "on"
}

// ScopeCustomers 客户可见范围：开启 customer_share 后，门店端可见全部门店客户
func (s *Service) ScopeCustomers(claims *auth.Claims, queryStoreID int64) *int64 {
	if claims.Role == "admin" {
		return s.ScopeStore(claims, queryStoreID)
	}
	if s.SettingBool("customer_share", false) {
		return nil
	}
	return claims.StoreID
}

// ScopeBusinessData 经营数据范围与「门店客户共享」对齐：
// - 不共享：仅本店（开卡/核销/客户/员工等）
// - 共享：门店端看连锁汇总；管理端仍按 ScopeStore
func (s *Service) ScopeBusinessData(claims *auth.Claims, boundStoreID *int64) *int64 {
	if claims == nil {
		return boundStoreID
	}
	if claims.Role == "admin" {
		return s.ScopeStore(claims, 0)
	}
	if s.SettingBool("customer_share", false) {
		return nil
	}
	if boundStoreID != nil && *boundStoreID > 0 {
		return boundStoreID
	}
	return claims.StoreID
}

// Dashboard period: day | week | month（默认 month）
func (s *Service) Dashboard(storeID *int64, period string) (map[string]interface{}, error) {
	period = strings.TrimSpace(strings.ToLower(period))
	if period != "day" && period != "week" && period != "month" {
		period = "month"
	}
	now := time.Now()
	var rangeStart time.Time
	var periodLabel string
	switch period {
	case "day":
		rangeStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		periodLabel = "今日"
	case "week":
		// 周一为一周开始
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		rangeStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		periodLabel = "本周"
	default:
		rangeStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		periodLabel = "本月"
	}
	startStr := rangeStart.Format("2006-01-02 15:04:05")

	args := []interface{}{startStr}
	storeFilter := ""
	if storeID != nil {
		storeFilter = " AND store_id=?"
		args = append(args, *storeID)
	}
	var cardAmount, verifyAmount float64
	var verifyCount, customerTotal, staffTotal, pendingVerify, pendingAudit int64
	_ = s.Repo.DB.Get(&cardAmount, "SELECT COALESCE(SUM(total_amount),0) FROM lp_customer_card WHERE created_at >= ?"+storeFilter, args...)
	_ = s.Repo.DB.Get(&verifyAmount, "SELECT COALESCE(SUM(total_amount),0) FROM lp_verify_order WHERE status='已完成' AND created_at >= ?"+storeFilter, args...)
	_ = s.Repo.DB.Get(&verifyCount, "SELECT COUNT(*) FROM lp_verify_order WHERE status='已完成' AND created_at >= ?"+storeFilter, args...)
	custFilter := "deleted_at IS NULL"
	staffFilter := "status='active' AND role IN ('manager','staff')"
	custArgs := []interface{}{}
	if storeID != nil {
		custFilter += " AND store_id=?"
		staffFilter += " AND store_id=?"
		custArgs = append(custArgs, *storeID)
	}
	_ = s.Repo.DB.Get(&customerTotal, "SELECT COUNT(*) FROM lp_customer WHERE "+custFilter, custArgs...)
	_ = s.Repo.DB.Get(&staffTotal, "SELECT COUNT(*) FROM lp_user WHERE "+staffFilter, custArgs...)
	pendingArgs := []interface{}{}
	pendingStore := ""
	if storeID != nil {
		pendingStore = " AND store_id=?"
		pendingArgs = append(pendingArgs, *storeID)
	}
	_ = s.Repo.DB.Get(&pendingVerify, "SELECT COUNT(*) FROM lp_verify_order WHERE status='待确认'"+pendingStore, pendingArgs...)
	if storeID == nil {
		_ = s.Repo.DB.Get(&pendingAudit, "SELECT COUNT(*) FROM lp_audit_apply WHERE status IN ('pending','待审核')")
	} else {
		_ = s.Repo.DB.Get(&pendingAudit, "SELECT COUNT(*) FROM lp_audit_apply WHERE status IN ('pending','待审核') AND store_id=?", *storeID)
	}
	var newCustomers int64
	newCustQ := "SELECT COUNT(*) FROM lp_customer WHERE deleted_at IS NULL AND created_at >= ?"
	newCustArgs := []interface{}{startStr}
	if storeID != nil {
		newCustQ += " AND store_id=?"
		newCustArgs = append(newCustArgs, *storeID)
	}
	_ = s.Repo.DB.Get(&newCustomers, newCustQ, newCustArgs...)

	// 趋势：日=24小时聚合为当日单点也可展示近7日；周=近7日；月=近30日按日
	trendDays := 7
	trendLabel := "近 7 日核销趋势"
	if period == "month" {
		trendDays = 30
		trendLabel = "近 30 日核销趋势"
	} else if period == "day" {
		trendDays = 7
		trendLabel = "近 7 日核销趋势（含今日）"
	}
	trend := []map[string]interface{}{}
	for i := trendDays - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		var amt float64
		q := "SELECT COALESCE(SUM(total_amount),0) FROM lp_verify_order WHERE status='已完成' AND DATE(created_at)=?"
		a := []interface{}{day}
		if storeID != nil {
			q += " AND store_id=?"
			a = append(a, *storeID)
		}
		_ = s.Repo.DB.Get(&amt, q, a...)
		trend = append(trend, map[string]interface{}{"date": day, "amount": amt})
	}

	return map[string]interface{}{
		"period":         period,
		"period_label":   periodLabel,
		"range_start":    rangeStart.Format("2006-01-02"),
		"trend_label":    trendLabel,
		"revenue":        cardAmount + verifyAmount,
		"card_amount":    cardAmount,
		"verify_amount":  verifyAmount,
		"verify_count":   verifyCount,
		"new_customers":  newCustomers,
		"pending_verify": pendingVerify,
		"pending_audit":  pendingAudit,
		"customer_total": customerTotal,
		"staff_total":    staffTotal,
		"trend":          trend,
	}, nil
}

func (s *Service) Workbench(claims *auth.Claims) (map[string]interface{}, error) {
	_, _ = s.ExpireStaleAppointments()
	u, err := s.Repo.FindUserByID(claims.UserID)
	if err != nil || u == nil {
		return nil, errors.New("用户不存在")
	}
	// 以库中绑定门店为准，再按客户共享开关决定经营数据范围
	bound := u.StoreIDValue()
	storeID := s.ScopeBusinessData(claims, bound)
	shareOn := s.SettingBool("customer_share", false)
	if !shareOn && claims.Role != "admin" && (storeID == nil || *storeID <= 0) {
		return nil, errors.New("账号未绑定门店，且未开启门店客户共享")
	}

	dash, err := s.Dashboard(storeID, "month")
	if err != nil {
		return nil, err
	}
	dash["user"] = u.Public()
	dash["month_performance"] = dash["verify_amount"]
	dash["customer_share"] = shareOn
	if storeID == nil {
		dash["data_scope"] = "all"
	} else {
		dash["data_scope"] = "store"
		dash["scope_store_id"] = *storeID
	}

	// 待办/今日预约始终按「本人绑定门店」，与经营汇总是否共享无关
	opsStore := bound
	if claims.Role == "admin" && storeID != nil {
		opsStore = storeID
	}

	today := time.Now().Format("2006-01-02")
	where := "a.appoint_date=? AND a.status IN ('待履约','已改约','待确认')"
	args := []interface{}{today}
	if opsStore != nil {
		where += " AND a.store_id=?"
		args = append(args, *opsStore)
	}
	rows, err := s.Repo.DB.Queryx(`SELECT a.id,a.appoint_time,a.status, COALESCE(c.name,''), COALESCE(p.name,''), COALESCE(u.name,'')
		FROM lp_appointment a
		LEFT JOIN lp_customer c ON c.id=a.customer_id
		LEFT JOIN lp_project p ON p.id=a.project_id
		LEFT JOIN lp_user u ON u.id=a.staff_id
		WHERE `+where+` ORDER BY a.appoint_time ASC LIMIT 8`, args...)
	todayAppts := []map[string]interface{}{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var tm, status, cname, pname, sname string
			if err := rows.Scan(&id, &tm, &status, &cname, &pname, &sname); err != nil {
				continue
			}
			todayAppts = append(todayAppts, map[string]interface{}{
				"id": id, "time": tm, "status": status,
				"customer_name": cname, "project_name": pname, "staff_name": sname,
			})
		}
	}
	dash["today_appointments"] = todayAppts

	var todayApptCount int64
	_ = s.Repo.DB.Get(&todayApptCount, `SELECT COUNT(*) FROM lp_appointment a WHERE `+where, args...)
	dash["today_appt_count"] = todayApptCount

	todos := []map[string]interface{}{}

	// 已履约（来店确认）且今日尚未完成核销 → 直达该客户核销页
	readyWhere := "a.appoint_date=? AND a.status='已履约'"
	readyArgs := []interface{}{today}
	if opsStore != nil {
		readyWhere += " AND a.store_id=?"
		readyArgs = append(readyArgs, *opsStore)
	}
	readyRows, rerr := s.Repo.DB.Queryx(`
		SELECT a.id, a.customer_id, a.project_id, COALESCE(c.name,''), COALESCE(p.name,'')
		FROM lp_appointment a
		LEFT JOIN lp_customer c ON c.id=a.customer_id
		LEFT JOIN lp_project p ON p.id=a.project_id
		WHERE `+readyWhere+`
		  AND NOT EXISTS (
		    SELECT 1 FROM lp_verify_order v
		    WHERE v.customer_id=a.customer_id AND v.project_id=a.project_id
		      AND v.status='已完成' AND DATE(v.created_at)=a.appoint_date
		  )
		ORDER BY a.appoint_time ASC LIMIT 5`, readyArgs...)
	if rerr == nil {
		defer readyRows.Close()
		for readyRows.Next() {
			var apptID, custID, projID int64
			var cname, pname string
			if err := readyRows.Scan(&apptID, &custID, &projID, &cname, &pname); err != nil {
				continue
			}
			if cname == "" {
				cname = "客户"
			}
			todos = append(todos, map[string]interface{}{
				"title": "待核销 · " + cname,
				"desc":  pname + " · 已来店确认",
				"path":  fmt.Sprintf("/verify?customer_id=%d&project_id=%d&appointment_id=%d", custID, projID, apptID),
				"tone":  "warn",
			})
		}
	}

	// 已有草稿待确认：直达确认页（按本店待办，不随共享汇总）
	var pendingOps int64
	{
		pq := "SELECT COUNT(*) FROM lp_verify_order WHERE status='待确认'"
		pa := []interface{}{}
		if opsStore != nil {
			pq += " AND store_id=?"
			pa = append(pa, *opsStore)
		}
		_ = s.Repo.DB.Get(&pendingOps, pq, pa...)
	}
	if pendingOps > 0 {
		var draftID, custID int64
		var custName string
		q := `SELECT v.id, v.customer_id, COALESCE(c.name,'') FROM lp_verify_order v
			LEFT JOIN lp_customer c ON c.id=v.customer_id
			WHERE v.status='待确认'`
		dArgs := []interface{}{}
		if opsStore != nil {
			q += ` AND v.store_id=?`
			dArgs = append(dArgs, *opsStore)
		}
		q += ` ORDER BY v.id DESC LIMIT 1`
		_ = s.Repo.DB.QueryRowx(q, dArgs...).Scan(&draftID, &custID, &custName)
		path := "/verify"
		desc := fmt.Sprintf("%d 笔待处理", pendingOps)
		if draftID > 0 {
			path = fmt.Sprintf("/verify?draft_id=%d", draftID)
			if custName != "" {
				desc = custName + " · " + desc
			}
		}
		todos = append(todos, map[string]interface{}{
			"title": "待确认核销", "desc": desc, "path": path, "tone": "warn",
		})
	}
	if todayApptCount > 0 {
		todos = append(todos, map[string]interface{}{
			"title": "今日预约", "desc": fmt.Sprintf("%d 个排期", todayApptCount), "path": "/appointments", "tone": "info",
		})
	}
	// 门店调度待收货：目标店店长确认入库（始终本店）
	if opsStore != nil && *opsStore > 0 && (claims.Role == "manager" || claims.Role == "admin") {
		pendingTransfers, _ := s.ListPendingStockTransfers(*opsStore)
		for _, t := range pendingTransfers {
			name, _ := t["item_name"].(string)
			from, _ := t["from_store_name"].(string)
			qty, _ := t["qty"].(int)
			unit, _ := t["unit"].(string)
			if name == "" {
				name = "物资"
			}
			todos = append(todos, map[string]interface{}{
				"title": "待确认入库 · " + name,
				"desc":  fmt.Sprintf("%s调入 %d%s · 去调度单查看", from, qty, unit),
				"path":  "/inventory?tab=orders",
				"tone":  "accent",
			})
		}
		dash["pending_transfers"] = pendingTransfers
		dash["pending_transfer_count"] = len(pendingTransfers)
	}
	// 门店调度待收货汇总（与上方按单待办去重：仅当未用 opsStore 列出时）
	if storeID != nil && *storeID > 0 && (claims.Role == "manager" || claims.Role == "admin") {
		if opsStore == nil || *opsStore <= 0 || (storeID != nil && opsStore != nil && *storeID != *opsStore) {
			pendingTransfers, _ := s.ListPendingStockTransfers(*storeID)
			if n := len(pendingTransfers); n > 0 {
				first := pendingTransfers[0]
				desc := fmt.Sprintf("%d 笔待收货", n)
				if name, _ := first["item_name"].(string); name != "" {
					desc = fmt.Sprintf("%s · %v%s · 来自%v", name, first["qty"], first["unit"], first["from_store_name"])
				}
				todos = append(todos, map[string]interface{}{
					"title": "待确认入库",
					"desc":  desc,
					"path":  "/inventory?tab=orders",
					"tone":  "accent",
				})
			}
		}
	}
	if claims.Role == "manager" {
		if pendingAudit, _ := dash["pending_audit"].(int64); pendingAudit > 0 {
			todos = append(todos, map[string]interface{}{
				"title": "待审核申请", "desc": fmt.Sprintf("%d 条待审", pendingAudit), "path": "/employees", "tone": "accent",
			})
		}
	}
	if len(todos) == 0 {
		todos = append(todos, map[string]interface{}{
			"title": "暂无待办", "desc": "今日事项已处理完毕", "path": "/appointments", "tone": "ok",
		})
	}
	dash["todos"] = todos
	return dash, nil
}

func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type DraftVerifyInput struct {
	EmployeeID      int64
	StoreID         int64
	CardID          int64
	ProjectID       int64   // 金额核销时可指定消费项目
	Times           int     // 次卡核销次数（金额模式忽略）
	Amount          float64 // 储值卡核销金额（消费金额）
	WelfareIssuedID int64   // 可选：福利抵扣
}

func (s *Service) expireWelfareIssued() {
	_, _ = s.Repo.DB.Exec(`UPDATE lp_welfare_issued SET status='已过期', remain_amount=0
		WHERE status='可用' AND expire_at IS NOT NULL AND expire_at < NOW()`)
}

func (s *Service) DraftVerify(in DraftVerifyInput) (*model.VerifyOrder, error) {
	card, err := s.Repo.GetCard(in.CardID)
	if err != nil || card == nil {
		return nil, errors.New("卡不存在")
	}
	if card.Status != "使用中" {
		return nil, errors.New("卡不可用")
	}
	kind := card.CardKind
	if kind == "" {
		kind = "times"
	}

	projectID := in.ProjectID
	if projectID == 0 {
		projectID = card.ProjectIDVal
		if projectID == 0 && card.ProjectID.Valid {
			projectID = card.ProjectID.Int64
		}
	}
	var proj *model.Project
	var projName string
	if projectID > 0 {
		proj, err = s.Repo.GetProject(projectID)
		if err != nil || proj == nil {
			return nil, errors.New("项目不存在")
		}
		projName = proj.Name
	} else {
		projName = "储值消费"
	}

	times := in.Times
	consumeAmount := 0.0
	if kind == "amount" {
		// 金额模式：按金额扣储值，可叠加福利；次卡路径暂不走福利
		if proj != nil {
			mode := NormalizeProjectCardMode(proj.CardMode)
			if mode == "times" {
				return nil, errors.New("该项目为次卡模式，请使用次卡核销")
			}
		}
		if in.Amount <= 0 {
			return nil, errors.New("请输入核销金额")
		}
		if card.TotalAmount+1e-9 < in.Amount {
			return nil, errors.New("储值余额不足")
		}
		consumeAmount = math.Round(in.Amount*100) / 100
		times = 0
	} else {
		return nil, errors.New("当前仅支持金额模式核销（储值卡），次卡暂不开放")
	}

	allowWelfare := true

	s.expireWelfareIssued()
	welfareOffset := 0.0
	var welfareID interface{}
	if in.WelfareIssuedID > 0 {
		if !allowWelfare {
			return nil, errors.New("当前卡/项目不支持福利抵扣（仅金额模式）")
		}
		var remain float64
		var wid, wStore int64
		var wCust sql.NullInt64
		var wStatus string
		var expireAt sql.NullTime
		err := s.Repo.DB.QueryRowx(`SELECT id,store_id,user_id,remain_amount,status,expire_at FROM lp_welfare_issued WHERE id=?`, in.WelfareIssuedID).
			Scan(&wid, &wStore, &wCust, &remain, &wStatus, &expireAt)
		if err != nil {
			return nil, errors.New("福利不存在")
		}
		if wStatus != "可用" || remain <= 0 {
			return nil, errors.New("福利不可用或已用完")
		}
		if expireAt.Valid && time.Now().After(expireAt.Time) {
			return nil, errors.New("福利已过期")
		}
		if wCust.Valid && wCust.Int64 > 0 {
			// 指定客户福利：仅校验归属客户，允许跨店核销
			if wCust.Int64 != card.CustomerID {
				return nil, errors.New("福利不属于该客户")
			}
		} else {
			// 门店通用福利：须匹配核销门店或卡所属门店
			if wStore != in.StoreID && wStore != card.StoreID {
				return nil, errors.New("福利不属于本门店")
			}
		}
		welfareOffset = remain
		if welfareOffset > consumeAmount {
			welfareOffset = consumeAmount
		}
		welfareOffset = math.Round(welfareOffset*100) / 100
		welfareID = wid
	} else {
		welfareID = nil
	}
	payAmount := math.Round((consumeAmount-welfareOffset)*100) / 100
	if payAmount < 0 {
		payAmount = 0
	}

	token := randomToken()
	expire := time.Now().Add(30 * time.Minute)
	verifyNo := fmt.Sprintf("HX%s%04d", time.Now().Format("20060102"), time.Now().Unix()%10000)

	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if kind == "amount" {
		// 草稿即预扣实付金额，确认时不再扣；取消则退回
		res, err := tx.Exec(`UPDATE lp_customer_card SET total_amount=total_amount-?, version=version+1,
			status=CASE WHEN total_amount-?<=0.00001 THEN '已完成' ELSE status END
			WHERE id=? AND total_amount>=? AND version=? AND status='使用中'`,
			payAmount, payAmount, card.ID, payAmount, card.Version)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, errors.New("卡状态已变更或余额不足，请重试")
		}
		_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark) VALUES(?,?,?,?, 'verify_hold',?,?)`,
			card.ID, card.CustomerID, 0, -payAmount, verifyNo,
			fmt.Sprintf("储值核销预扣 消费%.2f 福利抵%.2f", consumeAmount, welfareOffset))
	} else {
		res, err := tx.Exec(`UPDATE lp_customer_card SET frozen_times=frozen_times+?, version=version+1 WHERE id=? AND remain_times-frozen_times>=? AND version=?`,
			times, card.ID, times, card.Version)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, errors.New("卡状态已变更，请重试")
		}
	}

	if welfareOffset > 0 {
		res, err := tx.Exec(`UPDATE lp_welfare_issued SET remain_amount=remain_amount-?, used_amount=used_amount+?,
			status=CASE WHEN remain_amount-?<=0.00001 THEN '已用完' ELSE status END
			WHERE id=? AND status='可用' AND remain_amount>=?`,
			welfareOffset, welfareOffset, welfareOffset, in.WelfareIssuedID, welfareOffset)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil, errors.New("福利抵扣失败，请重试")
		}
	}

	r, err := tx.Exec(`INSERT INTO lp_verify_order(verify_no,customer_id,card_id,project_id,store_id,employee_id,total_amount,times,welfare_issued_id,welfare_offset,pay_amount,status,token,token_expire_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,'待确认',?,?)`,
		verifyNo, card.CustomerID, card.ID, projectID, in.StoreID, in.EmployeeID, consumeAmount, times, welfareID, welfareOffset, payAmount, token, expire)
	if err != nil {
		return nil, err
	}
	vid, _ := r.LastInsertId()
	_, _ = tx.Exec(`INSERT INTO lp_verify_item(verify_id,project_id,project_name,times,amount) VALUES(?,?,?,?,?)`,
		vid, projectID, projName, times, consumeAmount)
	mode := s.NormalizeVerifyMode(s.SettingValue("verify_mode", "staff"))
	var mobile string
	_ = tx.Get(&mobile, `SELECT mobile FROM lp_customer WHERE id=?`, card.CustomerID)
	if mode == "notify" {
		_, _ = tx.Exec(`INSERT INTO lp_sms_log(mobile,template_code,content,status) VALUES(?,?,?,?)`,
			mobile, "SMS_VERIFY", "核销确认链接 token="+token+"（待短信通道发送）", "待发送")
	} else {
		_, _ = tx.Exec(`INSERT INTO lp_sms_log(mobile,template_code,content,status) VALUES(?,?,?,?)`,
			mobile, "SMS_VERIFY", "员工自助核销，未通知客户", "无需发送")
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.Repo.GetVerify(vid)
}

func (s *Service) ConfirmVerify(id int64, token string) (*model.VerifyOrder, error) {
	v, err := s.Repo.GetVerify(id)
	if err != nil || v == nil {
		return nil, errors.New("核销单不存在")
	}
	if token != "" && v.Token != token {
		return nil, errors.New("核销令牌不匹配")
	}
	if v.Status != "待确认" {
		return nil, errors.New("核销单状态不可确认")
	}
	return s.finishVerify(v, true)
}

func (s *Service) ConfirmVerifyByToken(token string) (*model.VerifyOrder, error) {
	v, err := s.Repo.GetVerifyByToken(token)
	if err != nil || v == nil {
		return nil, errors.New("核销令牌无效")
	}
	if v.Status != "待确认" {
		return nil, errors.New("核销单状态不可确认")
	}
	if v.TokenExpireAt.Valid && time.Now().After(v.TokenExpireAt.Time) {
		return nil, errors.New("核销令牌已过期")
	}
	return s.finishVerify(v, true)
}

func (s *Service) CancelVerify(id int64) error {
	v, err := s.Repo.GetVerify(id)
	if err != nil || v == nil {
		return errors.New("核销单不存在")
	}
	if v.Status != "待确认" {
		return errors.New("仅待确认可取消")
	}
	_, err = s.finishVerify(v, false)
	return err
}

func (s *Service) finishVerify(v *model.VerifyOrder, confirm bool) (*model.VerifyOrder, error) {
	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var card model.Card
	if err := tx.Get(&card, `SELECT id,card_no,customer_id,project_id,store_id,employee_id,total_times,used_times,remain_times,frozen_times,total_amount,pay_method,COALESCE(card_kind,'times') AS card_kind,status,version,created_at FROM lp_customer_card WHERE id=? FOR UPDATE`, v.CardID); err != nil {
		return nil, err
	}
	kind := card.CardKind
	if kind == "" {
		kind = "times"
	}
	payAmount := v.PayAmount
	if payAmount <= 0 && v.WelfareOffset <= 0 {
		payAmount = v.TotalAmount
	}
	welfareID := int64(0)
	if v.WelfareIssuedID.Valid {
		welfareID = v.WelfareIssuedID.Int64
	} else if v.WelfareIssuedVal > 0 {
		welfareID = v.WelfareIssuedVal
	}

	if confirm {
		if kind == "amount" {
			// 余额已在草稿预扣
			_, _ = tx.Exec(`UPDATE lp_customer_card SET version=version+1 WHERE id=?`, card.ID)
			_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark) VALUES(?,?,?,?, 'verify',?,?)`,
				card.ID, card.CustomerID, 0, 0, v.VerifyNo,
				fmt.Sprintf("储值核销确认 消费%.2f 福利抵%.2f 实扣%.2f", v.TotalAmount, v.WelfareOffset, payAmount))
		} else {
			if card.FrozenTimes < v.Times || card.RemainTimes < v.Times {
				return nil, errors.New("卡次数不足")
			}
			res, err := tx.Exec(`UPDATE lp_customer_card SET frozen_times=frozen_times-?, remain_times=remain_times-?, used_times=used_times+?, version=version+1,
				status=CASE WHEN remain_times-?=0 THEN '已完成' ELSE status END WHERE id=? AND version=?`,
				v.Times, v.Times, v.Times, v.Times, card.ID, card.Version)
			if err != nil {
				return nil, err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return nil, errors.New("并发冲突，请重试")
			}
			_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark) VALUES(?,?,?,?, 'verify',?,?)`,
				card.ID, card.CustomerID, -v.Times, -payAmount, v.VerifyNo,
				fmt.Sprintf("次卡核销 福利抵%.2f", v.WelfareOffset))
		}
		_, _ = tx.Exec(`UPDATE lp_verify_order SET status='已完成', scanned_at=NOW() WHERE id=?`, v.ID)
		_, _ = tx.Exec(`UPDATE lp_customer SET last_visit_at=NOW() WHERE id=?`, v.CustomerID)
	} else {
		if kind == "amount" {
			// 退回预扣实付金额
			_, err := tx.Exec(`UPDATE lp_customer_card SET total_amount=total_amount+?, status='使用中', version=version+1 WHERE id=?`,
				payAmount, card.ID)
			if err != nil {
				return nil, err
			}
			_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark) VALUES(?,?,?,?, 'verify_cancel',?,?)`,
				card.ID, card.CustomerID, 0, payAmount, v.VerifyNo, "取消储值核销退回预扣")
		} else {
			_, err := tx.Exec(`UPDATE lp_customer_card SET frozen_times=GREATEST(frozen_times-?,0), version=version+1 WHERE id=?`, v.Times, card.ID)
			if err != nil {
				return nil, err
			}
		}
		if welfareID > 0 && v.WelfareOffset > 0 {
			_, _ = tx.Exec(`UPDATE lp_welfare_issued SET remain_amount=remain_amount+?, used_amount=GREATEST(used_amount-?,0),
				status=CASE WHEN expire_at IS NOT NULL AND expire_at<NOW() THEN '已过期' ELSE '可用' END
				WHERE id=?`, v.WelfareOffset, v.WelfareOffset, welfareID)
		}
		_, _ = tx.Exec(`UPDATE lp_verify_order SET status='已取消' WHERE id=?`, v.ID)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	out, err := s.Repo.GetVerify(v.ID)
	if err != nil {
		return nil, err
	}
	if confirm && out != nil {
		go s.notifyConsume(out.CustomerID, out.ProjectName, out.PayAmount)
	}
	return out, nil
}

func (s *Service) RefundVerify(id int64) error {
	v, err := s.Repo.GetVerify(id)
	if err != nil || v == nil {
		return errors.New("核销单不存在")
	}
	if v.Status != "已完成" {
		return errors.New("仅已完成可退次")
	}
	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var version int
	if err := tx.Get(&version, `SELECT version FROM lp_customer_card WHERE id=?`, v.CardID); err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE lp_customer_card SET remain_times=remain_times+?, used_times=GREATEST(used_times-?,0), status='使用中', version=version+1 WHERE id=? AND version=?`,
		v.Times, v.Times, v.CardID, version)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("并发冲突，请重试")
	}
	_, _ = tx.Exec(`UPDATE lp_verify_order SET status='已退次' WHERE id=?`, v.ID)
	_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark) VALUES(?,?,?,?, 'refund',?,?)`,
		v.CardID, v.CustomerID, v.Times, -v.TotalAmount, v.VerifyNo, "退次")
	return tx.Commit()
}

type OpenCardInput struct {
	EmployeeID int64
	StoreID    int64
	CustomerID int64
	ProjectID  int64
	PayMethod  string
	CardKind   string  // times | amount
	Times      int     // 次卡次数，0 则取项目默认
	Amount     float64 // 储值金额
}

func NormalizeProjectCardMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case "times", "amount", "both":
		return mode
	default:
		return "amount"
	}
}

// UpdateProject 更新项目；若改为金额模式，将关联次卡剩余价值自动转入客户储值余额（反之不转换）。
func (s *Service) UpdateProject(id int64, p *model.Project) (map[string]interface{}, error) {
	old, err := s.Repo.GetProject(id)
	if err != nil || old == nil {
		return nil, errors.New("项目不存在")
	}
	oldMode := NormalizeProjectCardMode(old.CardMode)
	newMode := NormalizeProjectCardMode(p.CardMode)
	p.CardMode = newMode
	if err := s.Repo.UpdateProject(id, p); err != nil {
		return nil, err
	}
	converted := 0
	var convertedAmount float64
	if newMode == "amount" && oldMode != "amount" {
		converted, convertedAmount, err = s.convertProjectTimesCardsToBalance(id)
		if err != nil {
			return nil, fmt.Errorf("项目已更新，但次卡转余额失败: %w", err)
		}
	}
	return map[string]interface{}{
		"id":               id,
		"card_mode":        newMode,
		"converted_cards":  converted,
		"converted_amount": convertedAmount,
	}, nil
}

// convertProjectTimesCardsToBalance 将项目下使用中的次卡按剩余次数折算金额，并入客户储值卡。
func (s *Service) convertProjectTimesCardsToBalance(projectID int64) (int, float64, error) {
	type trow struct {
		ID          int64   `db:"id"`
		CustomerID  int64   `db:"customer_id"`
		StoreID     int64   `db:"store_id"`
		EmployeeID  sql.NullInt64 `db:"employee_id"`
		TotalTimes  int     `db:"total_times"`
		RemainTimes int     `db:"remain_times"`
		TotalAmount float64 `db:"total_amount"`
		Version     int     `db:"version"`
		CardNo      string  `db:"card_no"`
	}
	var rows []trow
	if err := s.Repo.DB.Select(&rows, `
		SELECT id,customer_id,store_id,employee_id,total_times,remain_times,total_amount,version,card_no
		FROM lp_customer_card
		WHERE project_id=? AND status='使用中' AND COALESCE(card_kind,'times')='times' AND remain_times>0
		ORDER BY id`, projectID); err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, nil
	}

	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return 0, 0, err
	}
	defer func() { _ = tx.Rollback() }()

	n := 0
	sum := 0.0
	for _, c := range rows {
		unit := 0.0
		if c.TotalTimes > 0 {
			unit = c.TotalAmount / float64(c.TotalTimes)
		}
		value := math.Round(unit*float64(c.RemainTimes)*100) / 100
		if value <= 0 {
			// 无剩余价值也关闭次卡，避免金额模式下残留
			_, _ = tx.Exec(`UPDATE lp_customer_card SET remain_times=0, frozen_times=0, used_times=total_times, status='已完成', version=version+1 WHERE id=?`, c.ID)
			continue
		}

		// 同客户同门店优先并入已有储值卡
		var amountCardID int64
		_ = tx.Get(&amountCardID, `
			SELECT id FROM lp_customer_card
			WHERE customer_id=? AND store_id=? AND status='使用中' AND card_kind='amount'
			ORDER BY id DESC LIMIT 1`, c.CustomerID, c.StoreID)

		if amountCardID > 0 {
			res, err := tx.Exec(`UPDATE lp_customer_card SET total_amount=total_amount+?, version=version+1 WHERE id=? AND status='使用中'`,
				value, amountCardID)
			if err != nil {
				return 0, 0, err
			}
			if aff, _ := res.RowsAffected(); aff == 0 {
				amountCardID = 0
			} else {
				_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark)
					VALUES(?,?,0,?,'mode_convert',?,?)`,
					amountCardID, c.CustomerID, value, c.CardNo,
					fmt.Sprintf("项目改金额模式：次卡%s 剩余%d次折算转入", c.CardNo, c.RemainTimes))
			}
		}
		if amountCardID == 0 {
			cardNo := fmt.Sprintf("CD%d", time.Now().UnixNano()%1e10)
			var emp interface{}
			if c.EmployeeID.Valid {
				emp = c.EmployeeID.Int64
			} else {
				emp = nil
			}
			res, err := tx.Exec(`INSERT INTO lp_customer_card(card_no,customer_id,project_id,store_id,employee_id,total_times,used_times,remain_times,frozen_times,total_amount,pay_method,card_kind,status,version,valid_from,valid_to)
				VALUES(?,?,NULL,?,?,0,0,0,0,?,'系统结转','amount','使用中',0,CURDATE(),DATE_ADD(CURDATE(), INTERVAL 365 DAY))`,
				cardNo, c.CustomerID, c.StoreID, emp, value)
			if err != nil {
				return 0, 0, err
			}
			nid, _ := res.LastInsertId()
			_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark)
				VALUES(?,?,0,?,'mode_convert',?,?)`,
				nid, c.CustomerID, value, c.CardNo,
				fmt.Sprintf("项目改金额模式：次卡%s 剩余%d次折算开储值", c.CardNo, c.RemainTimes))
		}

		res, err := tx.Exec(`UPDATE lp_customer_card SET remain_times=0, frozen_times=0, used_times=total_times, status='已完成', version=version+1
			WHERE id=? AND version=? AND status='使用中'`, c.ID, c.Version)
		if err != nil {
			return 0, 0, err
		}
		if aff, _ := res.RowsAffected(); aff == 0 {
			continue
		}
		_, _ = tx.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark)
			VALUES(?,?,?,?, 'mode_convert',?,?)`,
			c.ID, c.CustomerID, -c.RemainTimes, -value, c.CardNo,
			fmt.Sprintf("项目改金额模式：次卡结转余额 ¥%.2f", value))
		// 不累加客户 total_amount：开卡时已计入消费额，此处仅结转可用余额

		n++
		sum += value
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return n, math.Round(sum*100) / 100, nil
}

func projectAllowsCardKind(cardMode, kind string) bool {
	mode := NormalizeProjectCardMode(cardMode)
	switch mode {
	case "times":
		return kind == "times"
	case "amount":
		return kind == "amount"
	default: // both
		return kind == "times" || kind == "amount"
	}
}

func (s *Service) OpenCard(in OpenCardInput) (*model.Card, error) {
	if in.CustomerID == 0 || in.StoreID == 0 {
		return nil, errors.New("请选择客户与门店")
	}
	kind := in.CardKind
	if kind == "" {
		kind = "amount" // 系统默认金额模式
	}
	if kind != "times" && kind != "amount" {
		return nil, errors.New("开卡类型仅支持次卡或储值")
	}
	pay := in.PayMethod
	if pay == "" {
		pay = "微信"
	}
	c := &model.Card{
		CustomerID: in.CustomerID,
		StoreID:    in.StoreID,
		EmployeeID: sql.NullInt64{Int64: in.EmployeeID, Valid: in.EmployeeID > 0},
		PayMethod:  pay,
		CardKind:   kind,
	}
	if kind == "times" {
		if in.ProjectID == 0 {
			return nil, errors.New("请选择项目")
		}
		p, err := s.Repo.GetProject(in.ProjectID)
		if err != nil || p == nil {
			return nil, errors.New("项目不存在")
		}
		if !projectAllowsCardKind(p.CardMode, "times") {
			return nil, errors.New("该项目仅支持金额模式开卡，不能按次卡办理")
		}
		times := in.Times
		if times <= 0 {
			times = p.TotalTimes
		}
		if times <= 0 {
			times = 1
		}
		amount := in.Amount
		if amount <= 0 {
			amount = p.Price * float64(times)
		}
		c.ProjectID = sql.NullInt64{Int64: in.ProjectID, Valid: true}
		c.ProjectIDVal = in.ProjectID
		c.TotalTimes = times
		c.RemainTimes = times
		c.TotalAmount = amount
	} else {
		if in.Amount <= 0 {
			return nil, errors.New("请输入充值金额")
		}
		if in.ProjectID > 0 {
			p, err := s.Repo.GetProject(in.ProjectID)
			if err != nil || p == nil {
				return nil, errors.New("项目不存在")
			}
			if !projectAllowsCardKind(p.CardMode, "amount") {
				return nil, errors.New("该项目仅支持次卡模式，不能按金额充值")
			}
			c.ProjectID = sql.NullInt64{Int64: in.ProjectID, Valid: true}
			c.ProjectIDVal = in.ProjectID
		}
		c.TotalTimes = 0
		c.RemainTimes = 0
		c.TotalAmount = in.Amount
	}
	if err := s.Repo.CreateCard(c); err != nil {
		return nil, err
	}
	_, _ = s.Repo.DB.Exec(`UPDATE lp_customer SET total_amount=total_amount+? WHERE id=?`, c.TotalAmount, in.CustomerID)
	// 开卡计入员工业绩考核
	if in.EmployeeID > 0 {
		_ = s.SyncAssessmentProgress(time.Now().Format("2006-01"), &in.StoreID)
	}
	return s.Repo.GetCard(c.ID)
}

// ---- Phase 2/3 helpers ----

func (s *Service) ListAppointments(storeID *int64, page, size int) (model.Page[model.Appointment], error) {
	_, _ = s.ExpireStaleAppointments()
	page, size = pageParams(page, size)
	where := "1=1"
	args := []interface{}{}
	if storeID != nil {
		where += " AND a.store_id=?"
		args = append(args, *storeID)
	}
	var total int64
	_ = s.Repo.DB.Get(&total, "SELECT COUNT(*) FROM lp_appointment a WHERE "+where, args...)
	args2 := append(args, size, (page-1)*size)
	rows, err := s.Repo.DB.Queryx(`SELECT a.id,a.customer_id,a.project_id,a.store_id,a.staff_id,a.appoint_date,a.appoint_time,a.status,a.source,a.note,
		COALESCE(c.name,''), COALESCE(p.name,''), COALESCE(u.name,'')
		FROM lp_appointment a
		LEFT JOIN lp_customer c ON c.id=a.customer_id
		LEFT JOIN lp_project p ON p.id=a.project_id
		LEFT JOIN lp_user u ON u.id=a.staff_id
		WHERE `+where+` ORDER BY a.appoint_date DESC, a.appoint_time DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return model.Page[model.Appointment]{}, err
	}
	defer rows.Close()
	list := []model.Appointment{}
	for rows.Next() {
		var a model.Appointment
		if err := rows.Scan(&a.ID, &a.CustomerID, &a.ProjectID, &a.StoreID, &a.StaffID, &a.AppointDate, &a.AppointTime, &a.Status, &a.Source, &a.Note, &a.CustomerName, &a.ProjectName, &a.StaffName); err != nil {
			return model.Page[model.Appointment]{}, err
		}
		a.Date = a.AppointDate.Format("2006-01-02")
		if a.StaffID.Valid {
			a.StaffIDVal = a.StaffID.Int64
		}
		list = append(list, a)
	}
	return model.Page[model.Appointment]{List: list, Total: total, Page: page, PageSize: size}, nil
}

func parseHHMM(t string) (int, error) {
	t = strings.TrimSpace(t)
	if t == "" {
		return 0, errors.New("时间不能为空")
	}
	parts := strings.Split(t, ":")
	if len(parts) < 2 {
		return 0, errors.New("时间格式错误")
	}
	var h, m int
	if _, err := fmt.Sscanf(parts[0]+":"+parts[1], "%d:%d", &h, &m); err != nil {
		return 0, errors.New("时间格式错误")
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, errors.New("时间格式错误")
	}
	return h*60 + m, nil
}

func (s *Service) projectDurationMin(projectID int64) int {
	var dur int
	_ = s.Repo.DB.Get(&dur, `SELECT COALESCE(duration,0) FROM lp_project WHERE id=?`, projectID)
	if dur <= 0 {
		return 60
	}
	return dur
}

func (s *Service) IsStaffBusy(staffID int64, date, timeStr string, projectID, excludeID int64) (bool, error) {
	if staffID <= 0 || date == "" || timeStr == "" {
		return false, nil
	}
	start, err := parseHHMM(timeStr)
	if err != nil {
		return false, err
	}
	dur := s.projectDurationMin(projectID)
	end := start + dur
	rows, err := s.Repo.DB.Queryx(`SELECT a.id, a.appoint_time, COALESCE(p.duration,60)
		FROM lp_appointment a
		LEFT JOIN lp_project p ON p.id=a.project_id
		WHERE a.staff_id=? AND a.appoint_date=? AND a.status IN ('待履约','已改约','待确认') AND a.id<>?`, staffID, date, excludeID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var tm string
		var odur int
		if err := rows.Scan(&id, &tm, &odur); err != nil {
			return false, err
		}
		if odur <= 0 {
			odur = 60
		}
		os, err := parseHHMM(tm)
		if err != nil {
			continue
		}
		oe := os + odur
		if start < oe && os < end {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) ListStaffAvailability(storeID *int64, date, timeStr string, projectID, excludeID int64) ([]map[string]interface{}, error) {
	_, _ = s.ExpireStaleAppointments()
	q := `SELECT id,name,role,role_name,store_id,status FROM lp_user WHERE status='active' AND role IN ('manager','staff')`
	args := []interface{}{}
	if storeID != nil {
		q += " AND store_id=?"
		args = append(args, *storeID)
	}
	q += " ORDER BY role ASC, id ASC"
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	needDur := s.projectDurationMin(projectID)
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, sid int64
		var name, role, roleName, status string
		if err := rows.Scan(&id, &name, &role, &roleName, &sid, &status); err != nil {
			return nil, err
		}
		busy := false
		conflictHint := ""
		busySlots := []string{}
		if date != "" {
			slots, conflict, _ := s.staffDaySlots(id, date, timeStr, needDur, excludeID)
			busySlots = slots
			if timeStr != "" {
				busy = conflict != ""
				conflictHint = conflict
			}
		}
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "role": role, "role_name": roleName, "store_id": sid,
			"busy": busy, "available": !busy,
			"conflict_hint": conflictHint, "busy_slots": busySlots,
		})
	}
	return list, nil
}

// staffDaySlots returns occupied windows and a conflict hint if timeStr+needDur overlaps any.
func (s *Service) staffDaySlots(staffID int64, date, timeStr string, needDur int, excludeID int64) ([]string, string, error) {
	rows, err := s.Repo.DB.Queryx(`SELECT a.appoint_time, COALESCE(p.duration,60)
		FROM lp_appointment a
		LEFT JOIN lp_project p ON p.id=a.project_id
		WHERE a.staff_id=? AND a.appoint_date=? AND a.status IN ('待履约','已改约','待确认') AND a.id<>?
		ORDER BY a.appoint_time`, staffID, date, excludeID)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	slots := []string{}
	conflict := ""
	var needStart, needEnd int
	if timeStr != "" {
		if ns, err := parseHHMM(timeStr); err == nil {
			needStart = ns
			if needDur <= 0 {
				needDur = 60
			}
			needEnd = ns + needDur
		}
	}
	for rows.Next() {
		var tm string
		var odur int
		if err := rows.Scan(&tm, &odur); err != nil {
			return nil, "", err
		}
		if odur <= 0 {
			odur = 60
		}
		os, err := parseHHMM(tm)
		if err != nil {
			continue
		}
		oe := os + odur
		label := fmt.Sprintf("%s-%s", formatHHMM(os), formatHHMM(oe))
		slots = append(slots, label)
		if timeStr != "" && needEnd > 0 && needStart < oe && os < needEnd {
			conflict = "冲突 " + label
		}
	}
	return slots, conflict, nil
}

func formatHHMM(min int) string {
	if min < 0 {
		min = 0
	}
	h := min / 60
	m := min % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func (s *Service) CreateAppointment(a *model.Appointment) error {
	if a.CustomerID == 0 || a.ProjectID == 0 || a.StoreID == 0 || a.Date == "" || a.AppointTime == "" {
		return errors.New("请完善预约客户、项目、日期和时间")
	}
	// 新建一律待履约；到店后再改为已履约
	a.Status = "待履约"
	staffID := a.StaffIDVal
	if staffID == 0 && a.StaffID.Valid {
		staffID = a.StaffID.Int64
	}
	if staffID > 0 {
		busy, err := s.IsStaffBusy(staffID, a.Date, a.AppointTime, a.ProjectID, 0)
		if err != nil {
			return err
		}
		if busy {
			return errors.New("该美容师此时段忙碌，请选择其他时间或美容师")
		}
	}
	var staffArg interface{}
	if staffID > 0 {
		staffArg = staffID
		a.StaffID = sql.NullInt64{Int64: staffID, Valid: true}
		a.StaffIDVal = staffID
	} else {
		staffArg = nil
	}
	res, err := s.Repo.DB.Exec(`INSERT INTO lp_appointment(customer_id,project_id,store_id,staff_id,appoint_date,appoint_time,status,source,note)
		VALUES(?,?,?,?,?,?,?,?,?)`, a.CustomerID, a.ProjectID, a.StoreID, staffArg, a.Date, a.AppointTime, a.Status, a.Source, a.Note)
	if err != nil {
		return err
	}
	a.ID, _ = res.LastInsertId()
	// 预约成功 → 预约通知短信
	var projectName, staffName string
	_ = s.Repo.DB.Get(&projectName, `SELECT name FROM lp_project WHERE id=?`, a.ProjectID)
	if staffID > 0 {
		_ = s.Repo.DB.Get(&staffName, `SELECT name FROM lp_user WHERE id=?`, staffID)
	}
	go s.notifyAppoint(a.CustomerID, a.Date, a.AppointTime, projectName, staffName)
	return nil
}

// 预约状态心智（员工/客户统一）：
// 1. 待履约 = 新建后默认（兼容旧名「待确认」），占用时段，等客户到店
// 2. 已履约 = 客户到店后，员工操作确认（终态）
// 3. 已改约 = 改过时间/人/项目，仍待到店（进行中）
// 4. 已取消 = 主动取消，释放时段（终态）
// 5. 已失效 = 超时未履约/取消/改约，系统自动置位（终态，≠取消）
var appointmentTransitions = map[string][]string{
	"待履约": {"已履约", "已取消", "已改约"},
	"待确认": {"已履约", "已取消", "已改约"}, // 兼容旧值
	"已改约": {"已履约", "已取消", "已改约"},
	"已履约": {},
	"已取消": {},
	"已失效": {},
}

func isAppointmentTerminal(status string) bool {
	return status == "已履约" || status == "已取消" || status == "已失效"
}

func normalizeAppointStatus(st string) string {
	switch strings.TrimSpace(st) {
	case "待确认", "已确认":
		return "待履约"
	case "失效":
		return "已失效"
	case "已完成":
		return "已履约"
	default:
		return st
	}
}

// AppointExpireMinutes 超时分钟数：超过「预约开始时间 + N 分钟」仍待履约/已改约 → 自动已失效。
func (s *Service) AppointExpireMinutes() int {
	n, err := strconv.Atoi(strings.TrimSpace(s.SettingValue("appoint_expire_minutes", "30")))
	if err != nil || n < 0 {
		return 30
	}
	if n > 24*60 {
		return 24 * 60
	}
	return n
}

// ExpireStaleAppointments 超时未到店履约/取消/改约 →「已失效」，释放时段。
func (s *Service) ExpireStaleAppointments() (int64, error) {
	mins := s.AppointExpireMinutes()
	res, err := s.Repo.DB.Exec(`
		UPDATE lp_appointment
		SET status='已失效'
		WHERE status IN ('待履约','已改约','待确认','已确认','失效')
		  AND TIMESTAMP(appoint_date, appoint_time) < DATE_SUB(NOW(), INTERVAL ? MINUTE)`, mins)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func canTransitAppointment(from, to string) bool {
	if from == to {
		return true
	}
	for _, s := range appointmentTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

type AppointmentUpdateInput struct {
	Status    string
	CustomerID int64
	ProjectID int64
	StoreID   int64
	StaffID   int64
	Date      string
	Time      string
	Note      *string
	Reschedule bool // 改约：改时间/人/项目
}

func (s *Service) GetAppointment(id int64) (*model.Appointment, error) {
	var a model.Appointment
	err := s.Repo.DB.QueryRowx(`SELECT a.id,a.customer_id,a.project_id,a.store_id,a.staff_id,a.appoint_date,a.appoint_time,a.status,a.source,a.note,
		COALESCE(c.name,''), COALESCE(p.name,''), COALESCE(u.name,'')
		FROM lp_appointment a
		LEFT JOIN lp_customer c ON c.id=a.customer_id
		LEFT JOIN lp_project p ON p.id=a.project_id
		LEFT JOIN lp_user u ON u.id=a.staff_id
		WHERE a.id=?`, id).Scan(
		&a.ID, &a.CustomerID, &a.ProjectID, &a.StoreID, &a.StaffID, &a.AppointDate, &a.AppointTime, &a.Status, &a.Source, &a.Note,
		&a.CustomerName, &a.ProjectName, &a.StaffName,
	)
	if err != nil {
		return nil, err
	}
	if a.StaffID.Valid {
		a.StaffIDVal = a.StaffID.Int64
	}
	a.Date = a.AppointDate.Format("2006-01-02")
	return &a, nil
}

func (s *Service) UpdateAppointment(id int64, in AppointmentUpdateInput) error {
	_, _ = s.ExpireStaleAppointments()
	cur, err := s.GetAppointment(id)
	if err != nil {
		return errors.New("预约不存在")
	}
	if isAppointmentTerminal(cur.Status) || cur.Status == "已完成" || cur.Status == "失效" {
		return errors.New("已结束的预约不能再改，请新建预约")
	}
	if st := strings.TrimSpace(in.Status); st == "已失效" || st == "失效" {
		return errors.New("「已失效」仅由系统超时自动置位，不可手动设置")
	}

	newStatus := strings.TrimSpace(in.Status)
	if newStatus == "" {
		newStatus = cur.Status
	}
	newStatus = normalizeAppointStatus(newStatus)
	// 「已履约」= 客户到店后员工确认（不是新建）
	if !canTransitAppointment(cur.Status, newStatus) && !canTransitAppointment(normalizeAppointStatus(cur.Status), newStatus) {
		return fmt.Errorf("不能从「%s」变更为「%s」", cur.Status, newStatus)
	}

	// 纯状态变更（到店履约 / 取消）
	if !in.Reschedule {
		_, err = s.Repo.DB.Exec(`UPDATE lp_appointment SET status=? WHERE id=?`, newStatus, id)
		return err
	}

	// 改约：改日期/时间/美容师/项目/备注 → 状态固定为「已改约」，继续占用新时段
	customerID := cur.CustomerID
	if in.CustomerID > 0 {
		customerID = in.CustomerID
	}
	projectID := cur.ProjectID
	if in.ProjectID > 0 {
		projectID = in.ProjectID
	}
	storeID := cur.StoreID
	if in.StoreID > 0 {
		storeID = in.StoreID
	}
	date := cur.Date
	if in.Date != "" {
		date = in.Date
	}
	timeStr := cur.AppointTime
	if in.Time != "" {
		timeStr = in.Time
	}
	staffID := cur.StaffIDVal
	if in.StaffID > 0 {
		staffID = in.StaffID
	}
	note := cur.Note
	if in.Note != nil {
		note = *in.Note
	}
	if customerID == 0 || projectID == 0 || storeID == 0 || date == "" || timeStr == "" {
		return errors.New("请完善预约客户、项目、日期和时间")
	}
	if staffID > 0 {
		busy, err := s.IsStaffBusy(staffID, date, timeStr, projectID, id)
		if err != nil {
			return err
		}
		if busy {
			return errors.New("该美容师此时段忙碌，请选择其他时间或美容师")
		}
	}
	newStatus = "已改约"
	var staffArg interface{}
	if staffID > 0 {
		staffArg = staffID
	} else {
		staffArg = nil
	}
	_, err = s.Repo.DB.Exec(`UPDATE lp_appointment SET customer_id=?, project_id=?, store_id=?, staff_id=?,
		appoint_date=?, appoint_time=?, status=?, note=? WHERE id=?`,
		customerID, projectID, storeID, staffArg, date, timeStr, newStatus, note, id)
	return err
}

func (s *Service) UpdateAppointmentStatus(id int64, status string) error {
	return s.UpdateAppointment(id, AppointmentUpdateInput{Status: status})
}

func (s *Service) DeleteAppointment(id int64) error {
	cur, err := s.GetAppointment(id)
	if err != nil {
		return errors.New("预约不存在")
	}
	if cur.Status == "已履约" || cur.Status == "已失效" || cur.Status == "已完成" {
		return errors.New("已结束的预约不可删除")
	}
	// 删除等同取消：优先软取消，避免误删历史
	if cur.Status != "已取消" {
		return s.UpdateAppointment(id, AppointmentUpdateInput{Status: "已取消"})
	}
	_, err = s.Repo.DB.Exec(`DELETE FROM lp_appointment WHERE id=?`, id)
	return err
}

// GeoSearch 地址搜索：优先高德 Web 服务（settings.amap_web_key），国内可用；无 Key 时提示配置。
func (s *Service) GeoSearch(q string) (map[string]interface{}, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, errors.New("请输入地址关键词")
	}
	key := strings.TrimSpace(s.SettingValue("amap_web_key", ""))
	if key == "" {
		return map[string]interface{}{
			"list":    []interface{}{},
			"message": "未配置高德 Web Key",
		}, nil
	}
	u := "https://restapi.amap.com/v3/place/text?key=" + url.QueryEscape(key) +
		"&keywords=" + url.QueryEscape(q) + "&offset=8&page=1&extensions=base"
	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.New("高德搜索请求失败，请检查网络")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw struct {
		Status   string `json:"status"`
		Info     string `json:"info"`
		Pois     []struct {
			Name     string `json:"name"`
			Address  string `json:"address"`
			Location string `json:"location"`
			Pname    string `json:"pname"`
			Cityname string `json:"cityname"`
			Adname   string `json:"adname"`
		} `json:"pois"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, errors.New("高德返回解析失败")
	}
	if raw.Status != "1" {
		msg := raw.Info
		if msg == "" {
			msg = "高德搜索失败"
		}
		return nil, errors.New(msg)
	}
	list := []map[string]interface{}{}
	for _, p := range raw.Pois {
		parts := strings.Split(p.Location, ",")
		if len(parts) != 2 {
			continue
		}
		lng, _ := strconv.ParseFloat(parts[0], 64)
		lat, _ := strconv.ParseFloat(parts[1], 64)
		// 高德返回 GCJ-02，转 WGS84 后与打卡/落库一致
		wlat, wlng := gcj02ToWgs84(lat, lng)
		name := p.Name
		addr := strings.TrimSpace(p.Pname + p.Cityname + p.Adname + p.Address)
		if addr != "" {
			name = name + " · " + addr
		}
		list = append(list, map[string]interface{}{
			"name": name, "address": addr, "lat": wlat, "lng": wlng,
		})
	}
	return map[string]interface{}{"list": list, "total": len(list)}, nil
}

// gcj02ToWgs84 近似逆转换，供高德 POI 落库为手机 GPS 同系坐标。
func gcj02ToWgs84(lat, lng float64) (float64, float64) {
	if outOfChina(lat, lng) {
		return lat, lng
	}
	dLat, dLng := gcjDelta(lat, lng)
	return lat - dLat, lng - dLng
}

func outOfChina(lat, lng float64) bool {
	return lng < 72.004 || lng > 137.8347 || lat < 0.8293 || lat > 55.8271
}

func gcjDelta(lat, lng float64) (dLat, dLng float64) {
	const a = 6378245.0
	const ee = 0.00669342162296594323
	dLat = transformLat(lng-105.0, lat-35.0)
	dLng = transformLng(lng-105.0, lat-35.0)
	radLat := lat / 180.0 * math.Pi
	magic := math.Sin(radLat)
	magic = 1 - ee*magic*magic
	sqrtMagic := math.Sqrt(magic)
	dLat = (dLat * 180.0) / (((a * (1 - ee)) / (magic * sqrtMagic)) * math.Pi)
	dLng = (dLng * 180.0) / ((a / sqrtMagic) * math.Cos(radLat) * math.Pi)
	return
}

func transformLat(x, y float64) float64 {
	r := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	r += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	r += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	r += (160.0*math.Sin(y/12.0*math.Pi) + 320*math.Sin(y*math.Pi/30.0)) * 2.0 / 3.0
	return r
}

func transformLng(x, y float64) float64 {
	r := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	r += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	r += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
	r += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
	return r
}

func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earth = 6371000.0
	toRad := math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earth * math.Asin(math.Min(1, math.Sqrt(a)))
}

// CheckIn 门店端打卡：默认强制定位围栏（对照门店 lat/lng + 设置 clockin_range）。
func (s *Service) CheckIn(userID, storeID int64, method string, lat, lng *float64) (map[string]interface{}, error) {
	if storeID <= 0 {
		u, _ := s.Repo.FindUserByID(userID)
		if u != nil && u.StoreID.Valid {
			storeID = u.StoreID.Int64
		}
	}
	if storeID <= 0 {
		return nil, errors.New("账号未绑定门店，无法打卡")
	}

	method = strings.TrimSpace(method)
	if method == "" || method == "人脸" {
		method = "定位"
	}
	// H5 无法可靠读取 Wi-Fi SSID，统一走定位围栏
	if method == "Wi-Fi" {
		method = "定位"
	}

	var storeLat, storeLng float64
	var storeName string
	err := s.Repo.DB.QueryRow(`SELECT COALESCE(lat,0), COALESCE(lng,0), name FROM lp_store WHERE id=?`, storeID).
		Scan(&storeLat, &storeLng, &storeName)
	if err != nil {
		return nil, errors.New("门店不存在")
	}
	if storeLat == 0 && storeLng == 0 {
		return nil, errors.New("门店未配置打卡坐标，请管理员在门店管理中填写经纬度")
	}
	if lat == nil || lng == nil {
		return nil, errors.New("请开启定位后再打卡")
	}
	if *lat < -90 || *lat > 90 || *lng < -180 || *lng > 180 {
		return nil, errors.New("定位坐标无效")
	}

	rangeM := 200.0
	if v := strings.TrimSpace(s.SettingValue("clockin_range", "200")); v != "" {
		if n, e := strconv.ParseFloat(v, 64); e == nil && n > 0 {
			rangeM = n
		}
	}
	if rangeM < 50 {
		rangeM = 50
	}
	if rangeM > 2000 {
		rangeM = 2000
	}
	dist := haversineMeters(*lat, *lng, storeLat, storeLng)
	if dist > rangeM {
		return nil, fmt.Errorf("不在打卡范围内（距「%s」约 %.0f 米，允许 %.0f 米）", storeName, dist, rangeM)
	}
	distInt := int(math.Round(dist))

	today := time.Now().Format("2006-01-02")
	var id int64
	var checkOut sql.NullTime
	err = s.Repo.DB.QueryRow(`SELECT id, check_out FROM lp_attendance WHERE user_id=? AND work_date=?`, userID, today).
		Scan(&id, &checkOut)
	if err == sql.ErrNoRows {
		_, err = s.Repo.DB.Exec(`INSERT INTO lp_attendance(user_id,store_id,work_date,check_in,method,status,lat,lng,distance_m)
			VALUES(?,?,?,NOW(),?,'正常',?,?,?)`, userID, storeID, today, method, *lat, *lng, distInt)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"ok": true, "action": "check_in", "distance_m": distInt, "store_name": storeName,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if checkOut.Valid {
		return nil, errors.New("今日已完成上下班打卡")
	}
	_, err = s.Repo.DB.Exec(`UPDATE lp_attendance SET check_out=NOW(), method=?, lat=?, lng=?, distance_m=? WHERE id=?`,
		method, *lat, *lng, distInt, id)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"ok": true, "action": "check_out", "distance_m": distInt, "store_name": storeName,
	}, nil
}

func (s *Service) AttendanceToday(userID int64) (map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	u, _ := s.Repo.FindUserByID(userID)
	storeID := int64(0)
	storeName := ""
	storeLat, storeLng := 0.0, 0.0
	if u != nil && u.StoreID.Valid {
		storeID = u.StoreID.Int64
		_ = s.Repo.DB.QueryRow(`SELECT name, COALESCE(lat,0), COALESCE(lng,0) FROM lp_store WHERE id=?`, storeID).
			Scan(&storeName, &storeLat, &storeLng)
	}
	rangeM := 200
	if v := strings.TrimSpace(s.SettingValue("clockin_range", "200")); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			rangeM = n
		}
	}

	var checkIn, checkOut sql.NullTime
	var status string
	err := s.Repo.DB.QueryRow(`SELECT check_in,check_out,status FROM lp_attendance WHERE user_id=? AND work_date=?`, userID, today).Scan(&checkIn, &checkOut, &status)
	out := map[string]interface{}{
		"date": today, "checked_in": false, "checked_out": false,
		"next_action": "check_in", "method": "定位",
		"store_id": storeID, "store_name": storeName,
		"store_lat": storeLat, "store_lng": storeLng,
		"range_m": rangeM, "geo_ready": storeLat != 0 || storeLng != 0,
	}
	if err == sql.ErrNoRows {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	out["status"] = status
	if checkIn.Valid {
		out["checked_in"] = true
		out["check_in"] = checkIn.Time.Format("15:04:05")
		out["next_action"] = "check_out"
	}
	if checkOut.Valid {
		out["checked_out"] = true
		out["check_out"] = checkOut.Time.Format("15:04:05")
		out["next_action"] = "done"
	}
	return out, nil
}

// ListAttendanceMonth 按月汇总在职店长/员工考勤，供管理端员工维/日历维查看。
func (s *Service) ListAttendanceMonth(month string, storeID *int64) (map[string]interface{}, error) {
	month = strings.TrimSpace(month)
	if len(month) < 7 {
		month = time.Now().Format("2006-01")
	}
	month = month[:7]
	start, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return nil, errors.New("月份格式应为 YYYY-MM")
	}
	end := start.AddDate(0, 1, 0)
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	uq := `SELECT u.id,u.name,u.role,u.role_name,COALESCE(u.store_id,0),COALESCE(st.name,'')
		FROM lp_user u LEFT JOIN lp_store st ON st.id=u.store_id
		WHERE u.status='active' AND u.role IN ('manager','staff')`
	uArgs := []interface{}{}
	if storeID != nil && *storeID > 0 {
		uq += ` AND u.store_id=?`
		uArgs = append(uArgs, *storeID)
	}
	uq += ` ORDER BY u.store_id, u.role, u.id`
	uRows, err := s.Repo.DB.Queryx(uq, uArgs...)
	if err != nil {
		return nil, err
	}
	defer uRows.Close()

	type emp struct {
		ID        int64
		Name      string
		Role      string
		RoleName  string
		StoreID   int64
		StoreName string
		Days      map[string]map[string]interface{}
	}
	emps := []*emp{}
	byID := map[int64]*emp{}
	for uRows.Next() {
		var e emp
		if err := uRows.Scan(&e.ID, &e.Name, &e.Role, &e.RoleName, &e.StoreID, &e.StoreName); err != nil {
			return nil, err
		}
		e.Days = map[string]map[string]interface{}{}
		emps = append(emps, &e)
		byID[e.ID] = &e
	}

	aq := `SELECT a.user_id,a.store_id,a.work_date,a.check_in,a.check_out,a.method,a.status
		FROM lp_attendance a
		JOIN lp_user u ON u.id=a.user_id
		WHERE a.work_date>=? AND a.work_date<? AND u.role IN ('manager','staff') AND u.status='active'`
	aArgs := []interface{}{startStr, endStr}
	if storeID != nil && *storeID > 0 {
		aq += ` AND a.store_id=?`
		aArgs = append(aArgs, *storeID)
	}
	aRows, err := s.Repo.DB.Queryx(aq, aArgs...)
	if err != nil {
		return nil, err
	}
	defer aRows.Close()
	for aRows.Next() {
		var uid, sid int64
		var workDate time.Time
		var checkIn, checkOut sql.NullTime
		var method, status string
		if err := aRows.Scan(&uid, &sid, &workDate, &checkIn, &checkOut, &method, &status); err != nil {
			return nil, err
		}
		e := byID[uid]
		if e == nil {
			continue
		}
		day := map[string]interface{}{
			"date": workDate.Format("2006-01-02"), "method": method, "status": status, "store_id": sid,
		}
		if checkIn.Valid {
			day["check_in"] = checkIn.Time.Format("15:04")
		}
		if checkOut.Valid {
			day["check_out"] = checkOut.Time.Format("15:04")
		}
		e.Days[workDate.Format("2006-01-02")] = day
	}

	list := make([]map[string]interface{}, 0, len(emps))
	for _, e := range emps {
		days := make([]map[string]interface{}, 0, len(e.Days))
		present := 0
		for _, d := range e.Days {
			days = append(days, d)
			if d["check_in"] != nil {
				present++
			}
		}
		// 按日期排序
		for i := 0; i < len(days); i++ {
			for j := i + 1; j < len(days); j++ {
				if days[i]["date"].(string) > days[j]["date"].(string) {
					days[i], days[j] = days[j], days[i]
				}
			}
		}
		list = append(list, map[string]interface{}{
			"user_id": e.ID, "user_name": e.Name, "role": e.Role, "role_name": e.RoleName,
			"store_id": e.StoreID, "store_name": e.StoreName,
			"present_days": present, "days": days,
		})
	}

	daysInMonth := end.AddDate(0, 0, -1).Day()
	return map[string]interface{}{
		"month": month, "days_in_month": daysInMonth,
		"list": list, "total": len(list),
	}, nil
}

func (s *Service) ListInventory(storeID *int64) ([]map[string]interface{}, error) {
	q := `SELECT s.id, s.store_id, s.item_id, s.qty, i.name, i.sku, i.unit, st.name AS store_name
		FROM lp_stock s JOIN lp_stock_item i ON i.id=s.item_id JOIN lp_store st ON st.id=s.store_id`
	args := []interface{}{}
	if storeID != nil {
		q += " WHERE s.store_id=?"
		args = append(args, *storeID)
	}
	q += " ORDER BY s.store_id ASC, i.id ASC"
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, storeIDV, itemID int64
		var qty int
		var name, sku, unit, storeName string
		if err := rows.Scan(&id, &storeIDV, &itemID, &qty, &name, &sku, &unit, &storeName); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"id": id, "store_id": storeIDV, "item_id": itemID, "qty": qty, "name": name, "sku": sku, "unit": unit, "store_name": storeName,
		})
	}
	return list, nil
}

func (s *Service) ListStockItems() ([]map[string]interface{}, error) {
	rows, err := s.Repo.DB.Queryx(`SELECT i.id, i.name, i.sku, i.unit, i.created_at,
		COALESCE((SELECT SUM(qty) FROM lp_stock WHERE item_id=i.id),0) AS total_qty,
		COALESCE((SELECT COUNT(*) FROM lp_stock WHERE item_id=i.id),0) AS store_count
		FROM lp_stock_item i ORDER BY i.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var name, sku, unit string
		var createdAt time.Time
		var totalQty, storeCount int64
		if err := rows.Scan(&id, &name, &sku, &unit, &createdAt, &totalQty, &storeCount); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "sku": sku, "unit": unit, "created_at": createdAt,
			"total_qty": totalQty, "store_count": storeCount,
		})
	}
	return list, nil
}

type StockItemInput struct {
	Name    string
	SKU     string
	Unit    string
	StoreID int64
	Qty     int
}

func (s *Service) CreateStockItem(in StockItemInput) (map[string]interface{}, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errors.New("请填写产品名称")
	}
	unit := strings.TrimSpace(in.Unit)
	if unit == "" {
		unit = "件"
	}
	sku := strings.TrimSpace(in.SKU)
	if sku == "" {
		sku = fmt.Sprintf("P%d", time.Now().UnixNano()%1e10)
	}
	res, err := s.Repo.DB.Exec(`INSERT INTO lp_stock_item(name,sku,unit) VALUES(?,?,?)`, name, sku, unit)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, errors.New("SKU 已存在")
		}
		return nil, err
	}
	id, _ := res.LastInsertId()
	if in.StoreID > 0 {
		qty := in.Qty
		if qty < 0 {
			qty = 0
		}
		_, _ = s.Repo.DB.Exec(`INSERT INTO lp_stock(store_id,item_id,qty) VALUES(?,?,?)
			ON DUPLICATE KEY UPDATE qty=VALUES(qty)`, in.StoreID, id, qty)
		if qty > 0 {
			_ = s.recordInitStock(in.StoreID, id, qty, 0)
		}
	}
	return map[string]interface{}{"id": id, "name": name, "sku": sku, "unit": unit, "store_id": in.StoreID, "qty": in.Qty}, nil
}

// recordInitStock 产品初始库存写入出入库单与流水，渠道标记为「系统初始化」。
func (s *Service) recordInitStock(storeID, itemID int64, qty int, operatorID int64) error {
	if storeID <= 0 || itemID <= 0 || qty <= 0 {
		return nil
	}
	orderNo := fmt.Sprintf("INI%d", time.Now().UnixNano()%1e12)
	_, err := s.Repo.DB.Exec(`INSERT INTO lp_stock_order(order_no,store_id,item_id,qty,direction,channel,status,operator_id)
		VALUES(?,?,?,?, '入库','系统初始化','已完成',?)`, orderNo, storeID, itemID, qty, operatorID)
	if err != nil {
		return err
	}
	_, _ = s.Repo.DB.Exec(`INSERT INTO lp_stock_flow(store_id,item_id,flow_type,change_qty,before_qty,after_qty,biz_no,operator_id)
		VALUES(?,?, '入库', ?, 0, ?, ?, ?)`, storeID, itemID, qty, qty, orderNo, operatorID)
	return nil
}

func (s *Service) UpdateStockItem(id int64, in StockItemInput) error {
	if id <= 0 {
		return errors.New("产品不存在")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return errors.New("请填写产品名称")
	}
	unit := strings.TrimSpace(in.Unit)
	if unit == "" {
		unit = "件"
	}
	sku := strings.TrimSpace(in.SKU)
	if sku == "" {
		// 编辑时 SKU 可空：保留原 SKU（与新增「可空自动生成」一致）
		var existing string
		if err := s.Repo.DB.Get(&existing, `SELECT sku FROM lp_stock_item WHERE id=?`, id); err != nil {
			return errors.New("产品不存在")
		}
		sku = strings.TrimSpace(existing)
		if sku == "" {
			sku = fmt.Sprintf("P%d", time.Now().UnixNano()%1e10)
		}
	}
	_, err := s.Repo.DB.Exec(`UPDATE lp_stock_item SET name=?, sku=?, unit=? WHERE id=?`, name, sku, unit, id)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return errors.New("SKU 已存在")
		}
		return err
	}
	var exists int
	if err := s.Repo.DB.Get(&exists, `SELECT 1 FROM lp_stock_item WHERE id=?`, id); err != nil || exists != 1 {
		return errors.New("产品不存在")
	}
	if in.StoreID > 0 {
		qty := in.Qty
		if qty < 0 {
			qty = 0
		}
		var before int
		errBefore := s.Repo.DB.Get(&before, `SELECT qty FROM lp_stock WHERE store_id=? AND item_id=?`, in.StoreID, id)
		if errBefore == sql.ErrNoRows {
			before = 0
		} else if errBefore != nil {
			return errBefore
		}
		_, err = s.Repo.DB.Exec(`INSERT INTO lp_stock(store_id,item_id,qty) VALUES(?,?,?)
			ON DUPLICATE KEY UPDATE qty=VALUES(qty)`, in.StoreID, id, qty)
		if err != nil {
			return err
		}
		delta := qty - before
		if delta != 0 {
			orderNo := fmt.Sprintf("ADJ%d", time.Now().UnixNano()%1e12)
			dir := "入库"
			change := delta
			if delta < 0 {
				dir = "出库"
				change = -delta
			}
			_, _ = s.Repo.DB.Exec(`INSERT INTO lp_stock_order(order_no,store_id,item_id,qty,direction,channel,status,operator_id)
				VALUES(?,?,?,?,?,'系统初始化','已完成',0)`, orderNo, in.StoreID, id, change, dir)
			_, _ = s.Repo.DB.Exec(`INSERT INTO lp_stock_flow(store_id,item_id,flow_type,change_qty,before_qty,after_qty,biz_no,operator_id)
				VALUES(?,?,?,?,?,?,?,0)`, in.StoreID, id, dir, delta, before, qty, orderNo)
		}
	}
	return nil
}

func (s *Service) DeleteStockItem(id int64) error {
	var used int64
	_ = s.Repo.DB.Get(&used, `SELECT COUNT(*) FROM lp_stock WHERE item_id=? AND qty<>0`, id)
	if used > 0 {
		return errors.New("该产品仍有门店库存，无法删除")
	}
	_, _ = s.Repo.DB.Exec(`DELETE FROM lp_stock WHERE item_id=?`, id)
	_, err := s.Repo.DB.Exec(`DELETE FROM lp_stock_item WHERE id=?`, id)
	return err
}

type StockOrderInput struct {
	FromStoreID int64
	ToStoreID   int64
	ItemID      int64
	OperatorID  int64
	Qty         int
	Direction   string
	Channel     string
}

// CreateStockOrder 门店调度仅做出库：本店扣库存，目标店待确认入库。
func (s *Service) CreateStockOrder(in StockOrderInput) error {
	if in.ItemID <= 0 || in.Qty <= 0 {
		return errors.New("请选择产品并填写数量")
	}
	if in.FromStoreID <= 0 {
		return errors.New("请选择门店")
	}
	channel := strings.TrimSpace(in.Channel)
	if channel == "" {
		channel = "门店调度"
	}
	direction := strings.TrimSpace(in.Direction)
	if direction == "" {
		direction = "出库"
	}
	// 门店端闭环：调度场景禁止主动入库，只能出库后由对方确认
	if channel == "门店调度" {
		if direction != "出库" {
			return errors.New("门店调度仅支持出库，入库需对方店长确认")
		}
		if in.ToStoreID <= 0 {
			return errors.New("请选择调入门店")
		}
		if in.ToStoreID == in.FromStoreID {
			return errors.New("调入门店不能与本店相同")
		}
		var n int
		_ = s.Repo.DB.Get(&n, `SELECT COUNT(*) FROM lp_store WHERE id=?`, in.ToStoreID)
		if n == 0 {
			return errors.New("调入门店不存在")
		}
	}

	orderNo := fmt.Sprintf("IO%d", time.Now().UnixNano()%1e12)
	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	status := "已完成"
	var toStore interface{}
	if channel == "门店调度" && direction == "出库" && in.ToStoreID > 0 {
		status = "待收货"
		toStore = in.ToStoreID
	} else {
		toStore = nil
	}

	_, err = tx.Exec(`INSERT INTO lp_stock_order(order_no,store_id,to_store_id,item_id,qty,direction,channel,status,operator_id)
		VALUES(?,?,?,?,?,?,?,?,?)`,
		orderNo, in.FromStoreID, toStore, in.ItemID, in.Qty, direction, channel, status, in.OperatorID)
	if err != nil {
		return err
	}

	var before int
	err = tx.Get(&before, `SELECT qty FROM lp_stock WHERE store_id=? AND item_id=?`, in.FromStoreID, in.ItemID)
	if err == sql.ErrNoRows {
		before = 0
		_, _ = tx.Exec(`INSERT INTO lp_stock(store_id,item_id,qty) VALUES(?,?,0)`, in.FromStoreID, in.ItemID)
	} else if err != nil {
		return err
	}
	change := in.Qty
	if direction == "出库" {
		change = -in.Qty
	}
	after := before + change
	if after < 0 {
		return errors.New("库存不足")
	}
	_, err = tx.Exec(`UPDATE lp_stock SET qty=? WHERE store_id=? AND item_id=?`, after, in.FromStoreID, in.ItemID)
	if err != nil {
		return err
	}
	_, _ = tx.Exec(`INSERT INTO lp_stock_flow(store_id,item_id,flow_type,change_qty,before_qty,after_qty,biz_no,operator_id) VALUES(?,?,?,?,?,?,?,?)`,
		in.FromStoreID, in.ItemID, direction, change, before, after, orderNo, in.OperatorID)
	return tx.Commit()
}

// ConfirmStockTransfer 目标门店确认收货入库，完成调度闭环。
func (s *Service) ConfirmStockTransfer(orderID, operatorID int64, storeID *int64) error {
	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var id, fromStore, itemID int64
	var toStore sql.NullInt64
	var qty int
	var status, channel, orderNo string
	err = tx.QueryRowx(`SELECT id,store_id,to_store_id,item_id,qty,status,channel,order_no FROM lp_stock_order WHERE id=? FOR UPDATE`, orderID).
		Scan(&id, &fromStore, &toStore, &itemID, &qty, &status, &channel, &orderNo)
	if err != nil {
		return errors.New("调度单不存在")
	}
	if status != "待收货" {
		return errors.New("该单已处理")
	}
	if !toStore.Valid || toStore.Int64 <= 0 {
		return errors.New("缺少调入门店")
	}
	if storeID != nil && *storeID > 0 && toStore.Int64 != *storeID {
		return errors.New("仅目标门店可确认收货")
	}

	var before int
	err = tx.Get(&before, `SELECT qty FROM lp_stock WHERE store_id=? AND item_id=?`, toStore.Int64, itemID)
	if err == sql.ErrNoRows {
		before = 0
		_, _ = tx.Exec(`INSERT INTO lp_stock(store_id,item_id,qty) VALUES(?,?,0)`, toStore.Int64, itemID)
	} else if err != nil {
		return err
	}
	after := before + qty
	_, err = tx.Exec(`UPDATE lp_stock SET qty=? WHERE store_id=? AND item_id=?`, after, toStore.Int64, itemID)
	if err != nil {
		return err
	}
	_, _ = tx.Exec(`INSERT INTO lp_stock_flow(store_id,item_id,flow_type,change_qty,before_qty,after_qty,biz_no,operator_id) VALUES(?,?,?,?,?,?,?,?)`,
		toStore.Int64, itemID, "入库", qty, before, after, orderNo, operatorID)
	_, err = tx.Exec(`UPDATE lp_stock_order SET status='已完成', confirm_by=?, confirm_at=NOW() WHERE id=? AND status='待收货'`,
		operatorID, orderID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) ListStockOrders(storeID *int64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `SELECT o.id,o.order_no,o.store_id,o.to_store_id,o.item_id,o.qty,o.direction,o.channel,o.status,o.created_at,
		COALESCE(i.name,''), COALESCE(i.unit,''),
		COALESCE(sf.name,''), COALESCE(st.name,'')
		FROM lp_stock_order o
		LEFT JOIN lp_stock_item i ON i.id=o.item_id
		LEFT JOIN lp_store sf ON sf.id=o.store_id
		LEFT JOIN lp_store st ON st.id=o.to_store_id
		WHERE 1=1`
	args := []interface{}{}
	if storeID != nil && *storeID > 0 {
		q += ` AND (o.store_id=? OR o.to_store_id=?)`
		args = append(args, *storeID, *storeID)
	}
	q += ` ORDER BY o.id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, fromID, itemID int64
		var toID sql.NullInt64
		var qty int
		var orderNo, dir, channel, status, name, unit, fromName, toName string
		var createdAt time.Time
		if err := rows.Scan(&id, &orderNo, &fromID, &toID, &itemID, &qty, &dir, &channel, &status, &createdAt, &name, &unit, &fromName, &toName); err != nil {
			return nil, err
		}
		item := map[string]interface{}{
			"id": id, "order_no": orderNo, "store_id": fromID, "item_id": itemID, "qty": qty,
			"direction": dir, "channel": channel, "status": status,
			"item_name": name, "unit": unit, "from_store_name": fromName, "to_store_name": toName,
			"created_at": createdAt.Format("2006-01-02 15:04:05"),
		}
		if toID.Valid {
			item["to_store_id"] = toID.Int64
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *Service) ListPendingStockTransfers(storeID int64) ([]map[string]interface{}, error) {
	if storeID <= 0 {
		return []map[string]interface{}{}, nil
	}
	rows, err := s.Repo.DB.Queryx(`
		SELECT o.id,o.order_no,o.store_id,o.to_store_id,o.item_id,o.qty,o.created_at,
		       COALESCE(i.name,''), COALESCE(i.unit,''), COALESCE(sf.name,'')
		FROM lp_stock_order o
		LEFT JOIN lp_stock_item i ON i.id=o.item_id
		LEFT JOIN lp_store sf ON sf.id=o.store_id
		WHERE o.status='待收货' AND o.to_store_id=?
		ORDER BY o.id ASC`, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, fromID, toID, itemID int64
		var qty int
		var orderNo, name, unit, fromName string
		var createdAt time.Time
		if err := rows.Scan(&id, &orderNo, &fromID, &toID, &itemID, &qty, &createdAt, &name, &unit, &fromName); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"id": id, "order_no": orderNo, "store_id": fromID, "to_store_id": toID, "item_id": itemID,
			"qty": qty, "item_name": name, "unit": unit, "from_store_name": fromName,
			"created_at": createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	return list, nil
}

func (s *Service) ListAudits(storeID *int64) ([]map[string]interface{}, error) {
	q := `SELECT a.id,a.name,a.phone,a.store_id,a.role,a.status,a.apply_at, COALESCE(st.name,'') AS store_name
		FROM lp_audit_apply a LEFT JOIN lp_store st ON st.id=a.store_id WHERE 1=1`
	args := []interface{}{}
	if storeID != nil {
		q += " AND a.store_id=?"
		args = append(args, *storeID)
	}
	q += " ORDER BY a.id DESC"
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, storeIDV int64
		var name, phone, role, status, storeName string
		var applyAt time.Time
		if err := rows.Scan(&id, &name, &phone, &storeIDV, &role, &status, &applyAt, &storeName); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "phone": phone, "store_id": storeIDV, "role": role, "status": status,
			"apply_at": applyAt.Format("2006-01-02 15:04"), "store_name": storeName,
		})
	}
	return list, nil
}

func (s *Service) ReviewAudit(id, reviewerID int64, approve bool) error {
	var name, phone, role string
	var storeID int64
	var status string
	err := s.Repo.DB.QueryRow(`SELECT name,phone,store_id,role,status FROM lp_audit_apply WHERE id=?`, id).Scan(&name, &phone, &storeID, &role, &status)
	if err != nil {
		return errors.New("申请不存在")
	}
	if status != "pending" && status != "待审核" {
		return errors.New("申请已处理")
	}
	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	newStatus := "已拒绝"
	var newUID int64
	if approve {
		newStatus = "已通过"
		hash, _ := s.HashPassword("123456")
		roleName := "美业员工"
		if role == "manager" {
			roleName = "店长"
		}
		res, err2 := tx.Exec(`INSERT INTO lp_user(name,phone,password_hash,role,role_name,store_id,position,status,base_salary) VALUES(?,?,?,?,?,?,?, 'active', ?)`,
			name, phone, hash, role, roleName, storeID, roleName, s.FixedForRole(role))
		if err2 != nil {
			return err2
		}
		newUID, _ = res.LastInsertId()
	}
	_, err = tx.Exec(`UPDATE lp_audit_apply SET status=?, reviewed_at=NOW(), reviewer_id=? WHERE id=?`, newStatus, reviewerID, id)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if approve && role == "manager" && newUID > 0 && storeID > 0 {
		return s.AssignStoreManager(storeID, newUID)
	}
	return nil
}

func (s *Service) GetSettings() (map[string]string, error) {
	rows, err := s.Repo.DB.Queryx(`SELECT config_key, config_value FROM lp_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	if v, ok := m["verify_mode"]; ok {
		m["verify_mode"] = s.NormalizeVerifyMode(v)
	} else {
		m["verify_mode"] = "staff"
	}
	return m, nil
}

func (s *Service) SaveSettings(data map[string]string) error {
	if v, ok := data["verify_mode"]; ok {
		data["verify_mode"] = s.NormalizeVerifyMode(v)
	}
	if v, ok := data["appoint_expire_minutes"]; ok {
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil || n < 0 {
			n = 30
		}
		if n > 24*60 {
			n = 24 * 60
		}
		data["appoint_expire_minutes"] = strconv.Itoa(n)
	}
	for k, v := range data {
		_, err := s.Repo.DB.Exec(`INSERT INTO lp_settings(config_key,config_value) VALUES(?,?) ON DUPLICATE KEY UPDATE config_value=VALUES(config_value)`, k, v)
		if err != nil {
			return err
		}
	}
	if err := s.ReloadSMSFromSettings(); err != nil {
		return fmt.Errorf("设置已保存，但短信客户端刷新失败: %w", err)
	}
	return nil
}

func (s *Service) ListAssessments(storeID *int64, month string) ([]map[string]interface{}, error) {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	// 列表前按业务数据刷新可自动统计的完成值（业绩/新客/考勤/核销）
	_ = s.SyncAssessmentProgress(month, storeID)

	q := `SELECT a.id,a.user_id,a.store_id,a.month_key,a.name,a.weight,a.target_value,a.done_value,a.passed, u.name AS user_name
		FROM lp_assessment a JOIN lp_user u ON u.id=a.user_id WHERE a.month_key=?`
	args := []interface{}{month}
	if storeID != nil {
		q += " AND a.store_id=?"
		args = append(args, *storeID)
	}
	q += " ORDER BY a.store_id, a.user_id, a.id"
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, uid, sid int64
		var monthKey, name, userName string
		var weight, target, done float64
		var passed int
		if err := rows.Scan(&id, &uid, &sid, &monthKey, &name, &weight, &target, &done, &passed, &userName); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"id": id, "user_id": uid, "store_id": sid, "month": monthKey, "name": name, "weight": weight,
			"target": target, "done": done, "pass": passed == 1, "user_name": userName,
			"auto": assessmentMetricAuto(name),
		})
	}
	return list, nil
}

// assessmentMetricAuto 是否由业务数据自动汇总（不可手工覆盖，列表时刷新）。
func assessmentMetricAuto(name string) bool {
	n := strings.TrimSpace(name)
	switch {
	case strings.Contains(n, "业绩"):
		return true
	case strings.Contains(n, "新增客户"), strings.Contains(n, "新客"):
		return true
	case strings.Contains(n, "考勤"):
		return true
	case strings.Contains(n, "核销"):
		return true
	default:
		return false
	}
}

func (s *Service) assessmentAutoDone(userID int64, month, name string) (float64, bool) {
	n := strings.TrimSpace(name)
	switch {
	case strings.Contains(n, "业绩"):
		var v float64
		_ = s.Repo.DB.Get(&v, `
			SELECT COALESCE(SUM(total_amount),0) FROM lp_customer_card
			WHERE employee_id=? AND DATE_FORMAT(created_at,'%Y-%m')=?`, userID, month)
		return v, true
	case strings.Contains(n, "新增客户"), strings.Contains(n, "新客"):
		var v float64
		_ = s.Repo.DB.Get(&v, `
			SELECT COUNT(*) FROM lp_customer
			WHERE employee_id=? AND deleted_at IS NULL AND DATE_FORMAT(created_at,'%Y-%m')=?`, userID, month)
		return v, true
	case strings.Contains(n, "考勤"):
		var v float64
		_ = s.Repo.DB.Get(&v, `
			SELECT COUNT(DISTINCT work_date) FROM lp_attendance
			WHERE user_id=? AND DATE_FORMAT(work_date,'%Y-%m')=?`, userID, month)
		return v, true
	case strings.Contains(n, "核销"):
		var v float64
		_ = s.Repo.DB.Get(&v, `
			SELECT COALESCE(SUM(times),0) FROM lp_verify_order
			WHERE employee_id=? AND status='已完成' AND DATE_FORMAT(created_at,'%Y-%m')=?`, userID, month)
		return v, true
	default:
		return 0, false
	}
}

// SyncAssessmentProgress 按开卡/新客/考勤/核销实绩回写考核完成值。
func (s *Service) SyncAssessmentProgress(month string, storeID *int64) error {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	// 补齐本月新建客户归属：取同客户本月最早一笔开卡的员工
	_, _ = s.Repo.DB.Exec(`
		UPDATE lp_customer c
		INNER JOIN (
			SELECT cc.customer_id, cc.employee_id
			FROM lp_customer_card cc
			INNER JOIN (
				SELECT customer_id, MIN(id) AS mid
				FROM lp_customer_card
				WHERE employee_id IS NOT NULL AND DATE_FORMAT(created_at,'%Y-%m')=?
				GROUP BY customer_id
			) x ON x.mid = cc.id
		) t ON t.customer_id = c.id
		SET c.employee_id = t.employee_id
		WHERE c.employee_id IS NULL
		  AND c.deleted_at IS NULL
		  AND DATE_FORMAT(c.created_at,'%Y-%m')=?`, month, month)

	q := `SELECT id, user_id, name, target_value, done_value FROM lp_assessment WHERE month_key=?`
	args := []interface{}{month}
	if storeID != nil {
		q += " AND store_id=?"
		args = append(args, *storeID)
	}
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, uid int64
		var name string
		var target, done float64
		if err := rows.Scan(&id, &uid, &name, &target, &done); err != nil {
			continue
		}
		autoDone, ok := s.assessmentAutoDone(uid, month, name)
		if !ok {
			continue
		}
		passed := assessmentPassed(autoDone, target)
		if autoDone == done {
			_, _ = s.Repo.DB.Exec(`UPDATE lp_assessment SET passed=? WHERE id=?`, passed, id)
			continue
		}
		_, _ = s.Repo.DB.Exec(`UPDATE lp_assessment SET done_value=?, passed=? WHERE id=?`, autoDone, passed, id)
	}
	return nil
}

type AssessmentInput struct {
	UserID   int64
	StoreID  int64
	Month    string
	Name     string
	Weight   float64
	Target   float64
	Done     float64
	Unified  bool // 统一指标：自动套到全体店长/员工
}

func assessmentPassed(done, target float64) int {
	if target > 0 && done >= target {
		return 1
	}
	if target <= 0 && done > 0 {
		return 1
	}
	return 0
}

func (s *Service) UpsertAssessment(id int64, in AssessmentInput) (int64, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return 0, errors.New("请填写考核指标名称")
	}
	month := strings.TrimSpace(in.Month)
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	weight := in.Weight
	if weight <= 0 {
		weight = 1
	}

	// 统一新增：给全体（或指定门店）店长/员工各建一条相同指标
	if id == 0 && (in.Unified || in.UserID <= 0) {
		return s.createUnifiedAssessment(in.StoreID, month, name, weight, in.Target)
	}

	if in.UserID <= 0 {
		return 0, errors.New("请指定员工")
	}
	storeID := in.StoreID
	if storeID <= 0 {
		_ = s.Repo.DB.Get(&storeID, `SELECT COALESCE(store_id,0) FROM lp_user WHERE id=?`, in.UserID)
	}
	if storeID <= 0 {
		return 0, errors.New("员工未绑定门店")
	}
	passed := assessmentPassed(in.Done, in.Target)
	if id > 0 {
		_, err := s.Repo.DB.Exec(`UPDATE lp_assessment SET user_id=?, store_id=?, month_key=?, name=?, weight=?, target_value=?, done_value=?, passed=? WHERE id=?`,
			in.UserID, storeID, month, name, weight, in.Target, in.Done, passed, id)
		return id, err
	}
	res, err := s.Repo.DB.Exec(`INSERT INTO lp_assessment(user_id,store_id,month_key,name,weight,target_value,done_value,passed) VALUES(?,?,?,?,?,?,?,?)`,
		in.UserID, storeID, month, name, weight, in.Target, in.Done, passed)
	if err != nil {
		return 0, err
	}
	nid, _ := res.LastInsertId()
	return nid, nil
}

// createUnifiedAssessment 统一考核指标：写入所有在职店长/员工（可按门店过滤）。
func (s *Service) createUnifiedAssessment(storeID int64, month, name string, weight, target float64) (int64, error) {
	q := `SELECT id, COALESCE(store_id,0) FROM lp_user WHERE status='active' AND role IN ('manager','staff')`
	args := []interface{}{}
	if storeID > 0 {
		q += " AND store_id=?"
		args = append(args, storeID)
	}
	rows, err := s.Repo.DB.Queryx(q, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var firstID int64
	n := 0
	for rows.Next() {
		var uid, sid int64
		if err := rows.Scan(&uid, &sid); err != nil {
			continue
		}
		if sid <= 0 {
			continue
		}
		// 同月同人同指标已存在则更新目标/权重，保留完成值
		var existID int64
		_ = s.Repo.DB.Get(&existID, `SELECT id FROM lp_assessment WHERE user_id=? AND month_key=? AND name=? LIMIT 1`, uid, month, name)
		if existID > 0 {
			_, _ = s.Repo.DB.Exec(`UPDATE lp_assessment SET store_id=?, weight=?, target_value=?, passed=CASE WHEN done_value>=? AND ?>0 THEN 1 ELSE passed END WHERE id=?`,
				sid, weight, target, target, target, existID)
			if firstID == 0 {
				firstID = existID
			}
		} else {
			res, err := s.Repo.DB.Exec(`INSERT INTO lp_assessment(user_id,store_id,month_key,name,weight,target_value,done_value,passed) VALUES(?,?,?,?,?,?,0,0)`,
				uid, sid, month, name, weight, target)
			if err != nil {
				return 0, err
			}
			nid, _ := res.LastInsertId()
			if firstID == 0 {
				firstID = nid
			}
		}
		n++
	}
	if n == 0 {
		return 0, errors.New("没有可下发的在职店长/员工")
	}
	return firstID, nil
}

// UpdateUnifiedAssessment 按月份+指标名批量更新目标/权重（统一口径）。
func (s *Service) UpdateUnifiedAssessment(month, name string, weight, target float64, storeID int64) (int64, error) {
	name = strings.TrimSpace(name)
	month = strings.TrimSpace(month)
	if name == "" || month == "" {
		return 0, errors.New("缺少月份或指标名")
	}
	if weight <= 0 {
		weight = 1
	}
	q := `UPDATE lp_assessment SET weight=?, target_value=?, passed=CASE WHEN done_value>=? AND ?>0 THEN 1 ELSE 0 END WHERE month_key=? AND name=?`
	args := []interface{}{weight, target, target, target, month, name}
	if storeID > 0 {
		q += " AND store_id=?"
		args = append(args, storeID)
	}
	res, err := s.Repo.DB.Exec(q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (s *Service) DeleteAssessment(id int64) error {
	_, err := s.Repo.DB.Exec(`DELETE FROM lp_assessment WHERE id=?`, id)
	return err
}

// DeleteUnifiedAssessment 删除某月某统一指标下的全部人员记录。
func (s *Service) DeleteUnifiedAssessment(month, name string, storeID int64) (int64, error) {
	name = strings.TrimSpace(name)
	month = strings.TrimSpace(month)
	if name == "" || month == "" {
		return 0, errors.New("缺少月份或指标名")
	}
	q := `DELETE FROM lp_assessment WHERE month_key=? AND name=?`
	args := []interface{}{month, name}
	if storeID > 0 {
		q += " AND store_id=?"
		args = append(args, storeID)
	}
	res, err := s.Repo.DB.Exec(q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (s *Service) ComputeSalary(month string, storeID *int64) error {
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	cfg, err := s.LoadSalaryConfig()
	if err != nil {
		return err
	}
	attendGt := cfg.Extra.AttendDaysGt
	if attendGt <= 0 {
		attendGt = 20
	}

	q := `SELECT id,store_id,role,base_salary FROM lp_user WHERE status='active' AND role IN ('manager','staff')`
	args := []interface{}{}
	if storeID != nil {
		q += " AND store_id=?"
		args = append(args, *storeID)
	}
	type urow struct {
		ID         int64   `db:"id"`
		StoreID    int64   `db:"store_id"`
		Role       string  `db:"role"`
		BaseSalary float64 `db:"base_salary"`
	}
	var users []urow
	if err := s.Repo.DB.Select(&users, q, args...); err != nil {
		return err
	}

	storeSalesCache := map[int64]float64{}
	storeHeadCache := map[int64]int{}

	for _, u := range users {
		if _, ok := storeSalesCache[u.StoreID]; !ok {
			var storeSales float64
			_ = s.Repo.DB.Get(&storeSales, `
				SELECT COALESCE(SUM(total_amount),0) FROM lp_customer_card
				WHERE store_id=? AND DATE_FORMAT(created_at,'%Y-%m')=?`, u.StoreID, month)
			storeSalesCache[u.StoreID] = storeSales

			var head int
			_ = s.Repo.DB.Get(&head, `
				SELECT COUNT(*) FROM (
					SELECT a.user_id FROM lp_attendance a
					JOIN lp_user u2 ON u2.id=a.user_id
					WHERE a.store_id=? AND DATE_FORMAT(a.work_date,'%Y-%m')=?
					  AND u2.role IN ('manager','staff') AND u2.status='active'
					GROUP BY a.user_id
					HAVING COUNT(*) > ?
				) t`, u.StoreID, month, attendGt)
			storeHeadCache[u.StoreID] = head
		}

		var salesAmt, verifyAmt float64
		_ = s.Repo.DB.Get(&salesAmt, `
			SELECT COALESCE(SUM(total_amount),0) FROM lp_customer_card
			WHERE employee_id=? AND DATE_FORMAT(created_at,'%Y-%m')=?`, u.ID, month)
		_ = s.Repo.DB.Get(&verifyAmt, `
			SELECT COALESCE(SUM(total_amount),0) FROM lp_verify_order
			WHERE employee_id=? AND status='已完成' AND DATE_FORMAT(created_at,'%Y-%m')=?`, u.ID, month)

		calc := ComputeRoleSalary(SalaryCalcInput{
			Role:             u.Role,
			SalesAmount:      salesAmt,
			VerifyAmount:     verifyAmt,
			StoreSalesAmount: storeSalesCache[u.StoreID],
			WorkingHeadcount: storeHeadCache[u.StoreID],
			Config:           &cfg,
		})
		snap := mustJSON(calc.Breakdown)

		_, err := s.Repo.DB.Exec(`INSERT INTO lp_salary_detail(
			month_key,user_id,store_id,base_salary,commission,assess_bonus,attendance_deduct,total_payable,status,calc_snapshot)
			VALUES(?,?,?,?,?,0,?,?,'已计算',?)
			ON DUPLICATE KEY UPDATE
			  base_salary=VALUES(base_salary),
			  commission=VALUES(commission),
			  assess_bonus=0,
			  attendance_deduct=VALUES(attendance_deduct),
			  total_payable=VALUES(total_payable),
			  status='已计算',
			  calc_snapshot=VALUES(calc_snapshot)`,
			month, u.ID, u.StoreID, calc.FixedSalary, calc.Commission, calc.ManagerPenalty, calc.TotalPayable, snap)
		if err != nil {
			return err
		}
		_, _ = s.Repo.DB.Exec(`UPDATE lp_user SET base_salary=? WHERE id=?`, cfg.FixedTotal(u.Role), u.ID)
	}
	return nil
}

func (s *Service) ListSalary(storeID *int64, month string, page, size int) (model.Page[map[string]interface{}], error) {
	page, size = pageParams(page, size)
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	where := "d.month_key=?"
	args := []interface{}{month}
	if storeID != nil {
		where += " AND d.store_id=?"
		args = append(args, *storeID)
	}
	var total int64
	_ = s.Repo.DB.Get(&total, "SELECT COUNT(*) FROM lp_salary_detail d WHERE "+where, args...)
	args2 := append(args, size, (page-1)*size)
	rows, err := s.Repo.DB.Queryx(`SELECT d.id,d.month_key,d.user_id,d.store_id,d.base_salary,d.commission,d.assess_bonus,d.attendance_deduct,d.total_payable,d.status,d.calc_snapshot,u.name,u.role,u.role_name
		FROM lp_salary_detail d JOIN lp_user u ON u.id=d.user_id WHERE `+where+` ORDER BY d.id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return model.Page[map[string]interface{}]{}, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, uid, sid int64
		var monthKey, status, name, role, roleName string
		var base, commission, bonus, deduct, totalPay float64
		var snap sql.NullString
		if err := rows.Scan(&id, &monthKey, &uid, &sid, &base, &commission, &bonus, &deduct, &totalPay, &status, &snap, &name, &role, &roleName); err != nil {
			return model.Page[map[string]interface{}]{}, err
		}
		item := map[string]interface{}{
			"id": id, "month": monthKey, "user_id": uid, "store_id": sid, "user_name": name,
			"role": role, "role_name": roleName,
			"base_salary": base, "fixed_salary": base,
			"commission": commission, "assess_bonus": bonus, "attendance_deduct": deduct,
			"manager_penalty": deduct, "total_payable": totalPay, "status": status,
			"sales_commission": 0.0, "verify_commission": 0.0,
			"sales_amount": 0.0, "verify_amount": 0.0,
		}
		if snap.Valid && snap.String != "" {
			var br map[string]interface{}
			if json.Unmarshal([]byte(snap.String), &br) == nil {
				item["snapshot"] = br
				for _, k := range []string{"sales_amount", "verify_amount", "sales_commission", "verify_commission", "sales_rate", "verify_rate", "sales_amount_used", "working_headcount", "sales_cap_ratio", "store_sales", "note"} {
					if v, ok := br[k]; ok {
						item[k] = v
					}
				}
				if v, ok := br["manager_penalty"]; ok {
					item["manager_penalty"] = v
				}
			}
		}
		list = append(list, item)
	}
	return model.Page[map[string]interface{}]{List: list, Total: total, Page: page, PageSize: size}, nil
}

func (s *Service) SetSalaryStatus(id int64, status string) error {
	_, err := s.Repo.DB.Exec(`UPDATE lp_salary_detail SET status=? WHERE id=?`, status, id)
	return err
}

func (s *Service) ListWelfare() ([]map[string]interface{}, error) {
	rows, err := s.Repo.DB.Queryx(`SELECT id,name,amount,COALESCE(valid_days,30),description,status FROM lp_welfare ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var name, desc, status string
		var amount float64
		var validDays int
		if err := rows.Scan(&id, &name, &amount, &validDays, &desc, &status); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "amount": amount, "valid_days": validDays,
			"description": desc, "status": status,
		})
	}
	return list, nil
}

func (s *Service) CreateWelfare(name string, amount float64, validDays int, description, status string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("请填写福利名称")
	}
	if status == "" {
		status = "启用"
	}
	if validDays <= 0 {
		validDays = 30
	}
	res, err := s.Repo.DB.Exec(`INSERT INTO lp_welfare(name,amount,valid_days,description,status) VALUES(?,?,?,?,?)`,
		name, amount, validDays, description, status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Service) UpdateWelfare(id int64, name string, amount float64, validDays int, description, status string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("请填写福利名称")
	}
	if status == "" {
		status = "启用"
	}
	if validDays <= 0 {
		validDays = 30
	}
	_, err := s.Repo.DB.Exec(`UPDATE lp_welfare SET name=?, amount=?, valid_days=?, description=?, status=? WHERE id=?`,
		name, amount, validDays, description, status, id)
	return err
}

func (s *Service) DeleteWelfare(id int64) error {
	_, err := s.Repo.DB.Exec(`DELETE FROM lp_welfare WHERE id=?`, id)
	return err
}

type IssueWelfareInput struct {
	WelfareID  int64
	CustomerID int64 // 可选：指定客户；0 表示不绑客户
	StoreID    int64 // 单独门店时必填
	AllStores  bool  // true=统一发放到全部门店
}

func (s *Service) IssueWelfare(in IssueWelfareInput) (int, error) {
	var amount float64
	var status string
	var validDays int
	if err := s.Repo.DB.QueryRowx(`SELECT amount,status,COALESCE(valid_days,30) FROM lp_welfare WHERE id=?`, in.WelfareID).
		Scan(&amount, &status, &validDays); err != nil {
		return 0, errors.New("福利不存在")
	}
	if status == "停用" {
		return 0, errors.New("该福利已停用，无法发放")
	}
	if validDays <= 0 {
		validDays = 30
	}
	expireAt := time.Now().AddDate(0, 0, validDays)

	var storeIDs []int64
	if in.AllStores {
		if err := s.Repo.DB.Select(&storeIDs, `SELECT id FROM lp_store ORDER BY id`); err != nil {
			return 0, err
		}
		if len(storeIDs) == 0 {
			return 0, errors.New("暂无门店可发放")
		}
	} else {
		if in.StoreID <= 0 {
			return 0, errors.New("请选择发放门店，或勾选全部门店")
		}
		var n int
		_ = s.Repo.DB.Get(&n, `SELECT COUNT(*) FROM lp_store WHERE id=?`, in.StoreID)
		if n == 0 {
			return 0, errors.New("门店不存在")
		}
		storeIDs = []int64{in.StoreID}
	}

	if in.CustomerID > 0 {
		var n int
		_ = s.Repo.DB.Get(&n, `SELECT COUNT(*) FROM lp_customer WHERE id=? AND deleted_at IS NULL`, in.CustomerID)
		if n == 0 {
			return 0, errors.New("客户不存在")
		}
	}

	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	for _, sid := range storeIDs {
		var cust interface{}
		if in.CustomerID > 0 {
			cust = in.CustomerID
		} else {
			cust = nil
		}
		if _, err := tx.Exec(`INSERT INTO lp_welfare_issued(welfare_id,user_id,store_id,amount,remain_amount,used_amount,expire_at,status)
			VALUES(?,?,?,?,?,0,?,'可用')`,
			in.WelfareID, cust, sid, amount, amount, expireAt); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(storeIDs), nil
}

// ListAvailableWelfare 客户可用福利：
// 1) 指定发给该客户的福利（跨店可用）；
// 2) 核销门店的通用福利（未指定客户）。
func (s *Service) ListAvailableWelfare(customerID, storeID int64) ([]map[string]interface{}, error) {
	s.expireWelfareIssued()
	if customerID <= 0 {
		return []map[string]interface{}{}, nil
	}
	rows, err := s.Repo.DB.Queryx(`
		SELECT i.id, i.welfare_id, i.amount, i.remain_amount, i.expire_at, i.store_id,
		       COALESCE(w.name,'') AS welfare_name
		FROM lp_welfare_issued i
		LEFT JOIN lp_welfare w ON w.id=i.welfare_id
		WHERE i.status='可用' AND i.remain_amount>0
		  AND (i.expire_at IS NULL OR i.expire_at>=NOW())
		  AND (
		    (i.user_id=? AND i.user_id>0)
		    OR ((i.user_id IS NULL OR i.user_id=0) AND i.store_id=? AND ?>0)
		  )
		ORDER BY i.remain_amount DESC, i.expire_at ASC, i.id ASC`, customerID, storeID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, wid, sid int64
		var amount, remain float64
		var expireAt sql.NullTime
		var name string
		if err := rows.Scan(&id, &wid, &amount, &remain, &expireAt, &sid, &name); err != nil {
			return nil, err
		}
		item := map[string]interface{}{
			"id": id, "welfare_id": wid, "amount": amount, "remain_amount": remain,
			"store_id": sid, "name": name,
		}
		if expireAt.Valid {
			item["expire_at"] = expireAt.Time.Format("2006-01-02 15:04:05")
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *Service) ListWelfareIssued(limit int) ([]map[string]interface{}, error) {
	s.expireWelfareIssued()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.Repo.DB.Queryx(`
		SELECT i.id, i.welfare_id, i.user_id, i.store_id, i.amount,
		       COALESCE(i.remain_amount,i.amount), COALESCE(i.used_amount,0), i.expire_at, COALESCE(i.status,'可用'),
		       i.issued_at,
		       COALESCE(w.name,'') AS welfare_name,
		       COALESCE(s.name,'') AS store_name,
		       COALESCE(c.name,'') AS customer_name
		FROM lp_welfare_issued i
		LEFT JOIN lp_welfare w ON w.id=i.welfare_id
		LEFT JOIN lp_store s ON s.id=i.store_id
		LEFT JOIN lp_customer c ON c.id=i.user_id
		ORDER BY i.id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, wid, sid int64
		var uid sql.NullInt64
		var amount, remain, used float64
		var expireAt sql.NullTime
		var status string
		var issuedAt time.Time
		var wName, sName, cName string
		if err := rows.Scan(&id, &wid, &uid, &sid, &amount, &remain, &used, &expireAt, &status, &issuedAt, &wName, &sName, &cName); err != nil {
			return nil, err
		}
		item := map[string]interface{}{
			"id": id, "welfare_id": wid, "store_id": sid, "amount": amount,
			"remain_amount": remain, "used_amount": used, "status": status,
			"issued_at": issuedAt.Format("2006-01-02 15:04:05"),
			"welfare_name": wName, "store_name": sName,
			"customer_name": cName, "customer_id": nil,
		}
		if expireAt.Valid {
			item["expire_at"] = expireAt.Time.Format("2006-01-02 15:04:05")
		}
		if uid.Valid && uid.Int64 > 0 {
			item["customer_id"] = uid.Int64
			item["user_id"] = uid.Int64
		}
		list = append(list, item)
	}
	return list, nil
}

func (s *Service) ListMessages(userID int64) ([]map[string]interface{}, error) {
	rows, err := s.Repo.DB.Queryx(`SELECT id,title,content,is_read,created_at FROM lp_message WHERE user_id=? ORDER BY id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var title, content string
		var isRead int
		var at time.Time
		if err := rows.Scan(&id, &title, &content, &isRead, &at); err != nil {
			return nil, err
		}
		list = append(list, map[string]interface{}{"id": id, "title": title, "content": content, "is_read": isRead == 1, "created_at": at.Format("2006-01-02 15:04")})
	}
	return list, nil
}

func operateModuleZH(m string) string {
	return map[string]string{
		"auth": "登录认证", "customer": "客户", "card": "开卡", "project": "项目",
		"verify": "核销", "appointment": "预约", "inventory": "库存", "user": "员工",
		"salary": "薪资", "settings": "设置", "audit": "审核",
	}[m]
}

func operateActionZH(a string) string {
	return map[string]string{
		"login": "登录", "create": "新建", "update": "更新", "delete": "删除",
		"open": "开卡", "confirm": "确认", "cancel": "取消", "refund": "退次",
		"approve": "通过", "reject": "拒绝", "compute": "计算", "save": "保存",
	}[a]
}

func (s *Service) ListOperateLogs(page, size int) (model.Page[map[string]interface{}], error) {
	page, size = pageParams(page, size)
	var total int64
	_ = s.Repo.DB.Get(&total, `SELECT COUNT(*) FROM lp_operate_log`)
	rows, err := s.Repo.DB.Queryx(`SELECT l.id,l.user_id,l.module,l.action,l.detail,l.ip,l.created_at, COALESCE(u.name,'系统')
		FROM lp_operate_log l LEFT JOIN lp_user u ON u.id=l.user_id
		ORDER BY l.id DESC LIMIT ? OFFSET ?`, size, (page-1)*size)
	if err != nil {
		return model.Page[map[string]interface{}]{}, err
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var uid sql.NullInt64
		var module, action, detail, ip, userName string
		var at time.Time
		if err := rows.Scan(&id, &uid, &module, &action, &detail, &ip, &at, &userName); err != nil {
			return model.Page[map[string]interface{}]{}, err
		}
		modZH := operateModuleZH(module)
		if modZH == "" {
			modZH = module
		}
		actZH := operateActionZH(action)
		if actZH == "" {
			actZH = action
		}
		item := map[string]interface{}{
			"id": id, "module": modZH, "module_code": module,
			"action": actZH, "action_code": action,
			"detail": detail, "ip": ip,
			"user_name": userName, "created_at": at.Format("2006-01-02 15:04:05"),
		}
		if uid.Valid {
			item["user_id"] = uid.Int64
		}
		list = append(list, item)
	}
	return model.Page[map[string]interface{}]{List: list, Total: total, Page: page, PageSize: size}, nil
}

func (s *Service) CustomerLogin(mobile string) (token string, data map[string]interface{}, err error) {
	var id int64
	var name, level string
	err = s.Repo.DB.QueryRow(`SELECT id,name,level_name FROM lp_customer WHERE mobile=? AND deleted_at IS NULL`, mobile).Scan(&id, &name, &level)
	if err == sql.ErrNoRows {
		return "", nil, errors.New("会员不存在")
	}
	if err != nil {
		return "", nil, err
	}
	role := "customer"
	token, err = s.JWT.Issue(id, role, nil)
	if err != nil {
		return "", nil, err
	}
	var remain int
	_ = s.Repo.DB.Get(&remain, `SELECT COALESCE(SUM(remain_times),0) FROM lp_customer_card WHERE customer_id=? AND status='使用中'`, id)
	data = map[string]interface{}{"id": id, "name": name, "phone": mobile, "level": level, "role": "customer", "remain": remain}
	return token, data, nil
}

func (s *Service) PublicVerification(token string) (map[string]interface{}, error) {
	v, err := s.Repo.GetVerifyByToken(token)
	if err != nil || v == nil {
		return nil, errors.New("核销码无效")
	}
	full, _ := s.Repo.GetVerify(v.ID)
	expire := ""
	if v.TokenExpireAt.Valid {
		expire = v.TokenExpireAt.Time.Format(time.RFC3339)
	}
	return map[string]interface{}{
		"id": v.ID, "verify_no": v.VerifyNo, "status": v.Status, "token": v.Token,
		"expire_at": expire, "amount": v.TotalAmount, "times": v.Times,
		"customer_name": full.CustomerName, "project_name": full.ProjectName, "store_name": full.StoreName,
	}, nil
}

// AssignStoreManager 指定门店负责人：同步 lp_store.manager_name，并将员工升为店长、归属该门店；
// 同店原店长降为员工。managerID<=0 时清空负责人并降级本店店长。
func (s *Service) AssignStoreManager(storeID, managerID int64) error {
	if storeID <= 0 {
		return errors.New("门店无效")
	}
	tx, err := s.Repo.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var storeCnt int
	if err := tx.Get(&storeCnt, `SELECT COUNT(*) FROM lp_store WHERE id=?`, storeID); err != nil {
		return err
	}
	if storeCnt == 0 {
		return errors.New("门店不存在")
	}

	demoteSQL := `UPDATE lp_user SET role='staff', role_name='美业员工',
		position=CASE WHEN position='店长' OR position='' THEN '美容师' ELSE position END
		WHERE store_id=? AND role='manager' AND status='active'`

	if managerID <= 0 {
		if _, err := tx.Exec(demoteSQL, storeID); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE lp_store SET manager_name='' WHERE id=?`, storeID); err != nil {
			return err
		}
		return tx.Commit()
	}

	var u struct {
		ID     int64         `db:"id"`
		Name   string        `db:"name"`
		Role   string        `db:"role"`
		Status string        `db:"status"`
		Store  sql.NullInt64 `db:"store_id"`
	}
	if err := tx.Get(&u, `SELECT id,name,role,status,store_id FROM lp_user WHERE id=?`, managerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("员工不存在")
		}
		return err
	}
	if u.Role == "admin" {
		return errors.New("不能指定管理员为负责人")
	}
	if u.Status != "active" {
		return errors.New("该员工已停用")
	}
	if u.Role == "manager" && u.Store.Valid && u.Store.Int64 > 0 && u.Store.Int64 != storeID {
		return errors.New("该员工已是其他门店店长")
	}

	if _, err := tx.Exec(demoteSQL+` AND id<>?`, storeID, managerID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE lp_user SET role='manager', role_name='店长', store_id=?, position='店长' WHERE id=?`,
		storeID, managerID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE lp_store SET manager_name=? WHERE id=?`, u.Name, storeID); err != nil {
		return err
	}
	return tx.Commit()
}

// SyncStoreManagerFromUser 员工侧设为店长时，同步门店负责人；从店长改回员工时清空对应门店负责人名。
func (s *Service) SyncStoreManagerFromUser(userID int64, role string, storeID *int64, old *model.User) error {
	if role == "manager" {
		if storeID == nil || *storeID <= 0 {
			return errors.New("店长必须指定所属门店")
		}
		return s.AssignStoreManager(*storeID, userID)
	}
	if old != nil && old.Role == "manager" && old.StoreID.Valid {
		_, err := s.Repo.DB.Exec(`UPDATE lp_store SET manager_name='' WHERE id=? AND manager_name=?`,
			old.StoreID.Int64, old.Name)
		return err
	}
	return nil
}

func PageOf[T any](list []T, total int64, page, size int) model.Page[T] {
	page, size = pageParams(page, size)
	return model.Page[T]{List: list, Total: total, Page: page, PageSize: size}
}
