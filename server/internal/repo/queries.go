package repo

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lanping/lp-server/internal/model"
)

type Store struct{ DB *sqlx.DB }

func New(db *sqlx.DB) *Store { return &Store{DB: db} }

func (s *Store) FindUserByPhone(phone string) (*model.User, error) {
	var u model.User
	err := s.DB.Get(&u, `SELECT u.id,u.name,u.phone,u.password_hash,u.role,u.role_name,u.store_id,u.position,u.hire_date,u.status,u.level_code,u.work_start,u.work_end,u.base_salary,u.avatar,u.created_at,
		COALESCE(st.name,'') AS store_name
		FROM lp_user u LEFT JOIN lp_store st ON st.id = u.store_id WHERE u.phone=? LIMIT 1`, phone)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) FindUserByID(id int64) (*model.User, error) {
	var u model.User
	err := s.DB.Get(&u, `SELECT u.id,u.name,u.phone,u.password_hash,u.role,u.role_name,u.store_id,u.position,u.hire_date,u.status,u.level_code,u.work_start,u.work_end,u.base_salary,u.avatar,u.created_at,
		COALESCE(st.name,'') AS store_name
		FROM lp_user u LEFT JOIN lp_store st ON st.id = u.store_id WHERE u.id=?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (s *Store) ListStores() ([]model.Store, error) {
	var list []model.Store
	err := s.DB.Select(&list, `SELECT s.id,s.name,s.address,COALESCE(s.lat,0) AS lat,COALESCE(s.lng,0) AS lng,s.phone,s.manager_name,s.status,s.created_at,
		(SELECT u.id FROM lp_user u WHERE u.store_id=s.id AND u.role='manager' AND u.status='active' ORDER BY u.id LIMIT 1) AS manager_id
		FROM lp_store s ORDER BY s.id`)
	return list, err
}

func (s *Store) CreateStore(st *model.Store) error {
	if st.Status == "" {
		st.Status = "营业中"
	}
	res, err := s.DB.Exec(`INSERT INTO lp_store(name,address,lat,lng,phone,manager_name,status) VALUES(?,?,?,?,?,?,?)`,
		st.Name, st.Address, st.Lat, st.Lng, st.Phone, st.ManagerName, st.Status)
	if err != nil {
		return err
	}
	st.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) UpdateStore(id int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	allowed := map[string]bool{"name": true, "address": true, "lat": true, "lng": true, "phone": true, "manager_name": true, "status": true}
	sets := []string{}
	args := []interface{}{}
	for k, v := range fields {
		if !allowed[k] {
			continue
		}
		sets = append(sets, k+"=?")
		args = append(args, v)
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	_, err := s.DB.Exec(`UPDATE lp_store SET `+strings.Join(sets, ",")+` WHERE id=?`, args...)
	return err
}

func (s *Store) ListUsers(storeID *int64, q string, page, size int) ([]model.User, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if storeID != nil {
		where = append(where, "u.store_id=?")
		args = append(args, *storeID)
	}
	if q != "" {
		where = append(where, "(u.name LIKE ? OR u.phone LIKE ?)")
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	w := strings.Join(where, " AND ")
	var total int64
	if err := s.DB.Get(&total, "SELECT COUNT(*) FROM lp_user u WHERE "+w, args...); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	args2 := append(args, size, offset)
	var list []model.User
	err := s.DB.Select(&list, `SELECT u.id,u.name,u.phone,u.password_hash,u.role,u.role_name,u.store_id,u.position,u.hire_date,u.status,u.level_code,u.work_start,u.work_end,u.base_salary,u.avatar,u.created_at,
		COALESCE(st.name,'') AS store_name FROM lp_user u LEFT JOIN lp_store st ON st.id=u.store_id WHERE `+w+` ORDER BY u.id DESC LIMIT ? OFFSET ?`, args2...)
	return list, total, err
}

func (s *Store) CreateUser(u *model.User) error {
	res, err := s.DB.Exec(`INSERT INTO lp_user(name,phone,password_hash,role,role_name,store_id,position,status,level_code,base_salary)
		VALUES(?,?,?,?,?,?,?,?,?,?)`, u.Name, u.Phone, u.PasswordHash, u.Role, u.RoleName, u.StoreID, u.Position, u.Status, u.LevelCode, u.BaseSalary)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = id
	return nil
}

func (s *Store) UpdateUser(id int64, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}
	sets := []string{}
	args := []interface{}{}
	for k, v := range fields {
		sets = append(sets, k+"=?")
		args = append(args, v)
	}
	args = append(args, id)
	_, err := s.DB.Exec("UPDATE lp_user SET "+strings.Join(sets, ",")+" WHERE id=?", args...)
	return err
}

func (s *Store) ListProjects(q, status string, page, size int) ([]model.Project, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if q != "" {
		where = append(where, "(name LIKE ? OR category LIKE ?)")
		args = append(args, "%"+q+"%", "%"+q+"%")
	}
	if status != "" {
		where = append(where, "status=?")
		args = append(args, status)
	}
	w := strings.Join(where, " AND ")
	var total int64
	_ = s.DB.Get(&total, "SELECT COUNT(*) FROM lp_project WHERE "+w, args...)
	args2 := append(args, size, (page-1)*size)
	var list []model.Project
	err := s.DB.Select(&list, `SELECT id,name,category,price,member_price,total_times,COALESCE(NULLIF(card_mode,''),'amount') AS card_mode,valid_days,commission_rate,duration,status FROM lp_project WHERE `+w+` ORDER BY id DESC LIMIT ? OFFSET ?`, args2...)
	return list, total, err
}

func (s *Store) GetProject(id int64) (*model.Project, error) {
	var p model.Project
	err := s.DB.Get(&p, `SELECT id,name,category,price,member_price,total_times,COALESCE(NULLIF(card_mode,''),'amount') AS card_mode,valid_days,commission_rate,duration,status FROM lp_project WHERE id=?`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func normalizeProjectCardMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case "times", "amount", "both":
		return mode
	default:
		return "amount"
	}
}

func (s *Store) CreateProject(p *model.Project) error {
	p.CardMode = normalizeProjectCardMode(p.CardMode)
	res, err := s.DB.Exec(`INSERT INTO lp_project(name,category,price,member_price,total_times,card_mode,valid_days,commission_rate,duration,status)
		VALUES(?,?,?,?,?,?,?,?,?,?)`, p.Name, p.Category, p.Price, p.MemberPrice, p.TotalTimes, p.CardMode, p.ValidDays, p.CommissionRate, p.Duration, p.Status)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) UpdateProject(id int64, p *model.Project) error {
	p.CardMode = normalizeProjectCardMode(p.CardMode)
	_, err := s.DB.Exec(`UPDATE lp_project SET name=?,category=?,price=?,member_price=?,total_times=?,card_mode=?,valid_days=?,commission_rate=?,duration=?,status=? WHERE id=?`,
		p.Name, p.Category, p.Price, p.MemberPrice, p.TotalTimes, p.CardMode, p.ValidDays, p.CommissionRate, p.Duration, p.Status, id)
	return err
}

func (s *Store) DeleteProject(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM lp_project WHERE id=?`, id)
	return err
}

func (s *Store) ListCustomers(storeID *int64, q string, page, size int) ([]model.Customer, int64, error) {
	where := []string{"c.deleted_at IS NULL"}
	args := []interface{}{}
	if storeID != nil {
		where = append(where, "c.store_id=?")
		args = append(args, *storeID)
	}
	if q != "" {
		where = append(where, "(c.name LIKE ? OR c.mobile LIKE ? OR c.customer_no LIKE ?)")
		args = append(args, "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}
	w := strings.Join(where, " AND ")
	var total int64
	_ = s.DB.Get(&total, "SELECT COUNT(*) FROM lp_customer c WHERE "+w, args...)
	args2 := append(args, size, (page-1)*size)
	rows, err := s.DB.Queryx(`SELECT c.id,c.customer_no,c.name,c.mobile,c.gender,c.birthday,c.level_name,c.source,c.store_id,c.employee_id,c.total_amount,c.last_visit_at,c.status,c.remark,
		COALESCE(st.name,'') AS store_name,
		COALESCE((SELECT SUM(remain_times) FROM lp_customer_card WHERE customer_id=c.id AND status='使用中' AND COALESCE(card_kind,'times')='times'),0) AS remain,
		COALESCE((SELECT SUM(total_amount) FROM lp_customer_card WHERE customer_id=c.id AND status='使用中' AND card_kind='amount'),0) AS balance,
		COALESCE((SELECT COUNT(*) FROM lp_customer_card WHERE customer_id=c.id AND status='使用中'),0) AS cards_count
		FROM lp_customer c LEFT JOIN lp_store st ON st.id=c.store_id WHERE `+w+` ORDER BY c.id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Customer{}
	for rows.Next() {
		var c model.Customer
		var remain, cardsCount int
		var balance float64
		var storeName string
		if err := rows.Scan(&c.ID, &c.CustomerNo, &c.Name, &c.Mobile, &c.Gender, &c.Birthday, &c.LevelName, &c.Source, &c.StoreID, &c.EmployeeID, &c.TotalAmount, &c.LastVisitAt, &c.Status, &c.Remark, &storeName, &remain, &balance, &cardsCount); err != nil {
			return nil, 0, err
		}
		c.StoreName = storeName
		c.Remain = remain
		c.Balance = balance
		c.CardsCount = cardsCount
		if c.Birthday.Valid {
			c.BirthdayStr = c.Birthday.Time.Format("01-02")
		}
		if c.LastVisitAt.Valid {
			c.LastVisit = c.LastVisitAt.Time.Format("2006-01-02")
		}
		list = append(list, c)
	}
	// 列表摘要：从开卡数据拼项目次数 / 余额
	if len(list) > 0 {
		ids := make([]interface{}, 0, len(list))
		idx := map[int64]int{}
		for i := range list {
			ids = append(ids, list[i].ID)
			idx[list[i].ID] = i
		}
		ph := strings.Repeat("?,", len(ids))
		ph = ph[:len(ph)-1]
		crows, err := s.DB.Queryx(`SELECT customer_id, COALESCE(card_kind,'times'), COALESCE(p.name,''), remain_times, total_times, total_amount
			FROM lp_customer_card cd
			LEFT JOIN lp_project p ON p.id=cd.project_id
			WHERE cd.status='使用中' AND cd.customer_id IN (`+ph+`)
			ORDER BY cd.id DESC`, ids...)
		if err == nil {
			defer crows.Close()
			briefs := map[int64][]string{}
			for crows.Next() {
				var cid int64
				var kind, pname string
				var remainT, totalT int
				var amt float64
				if err := crows.Scan(&cid, &kind, &pname, &remainT, &totalT, &amt); err != nil {
					continue
				}
				var bit string
				if kind == "amount" {
					bit = fmt.Sprintf("储值¥%.0f", amt)
				} else {
					if pname == "" {
						pname = "次卡"
					}
					bit = fmt.Sprintf("%s %d/%d次", pname, remainT, totalT)
				}
				briefs[cid] = append(briefs[cid], bit)
			}
			for cid, parts := range briefs {
				i, ok := idx[cid]
				if !ok {
					continue
				}
				if len(parts) > 3 {
					list[i].CardsBrief = strings.Join(parts[:3], " · ") + fmt.Sprintf(" 等%d项", len(parts))
				} else {
					list[i].CardsBrief = strings.Join(parts, " · ")
				}
			}
		}
	}
	return list, total, nil
}

func (s *Store) GetCustomer(id int64) (*model.Customer, error) {
	list, _, err := s.ListCustomers(nil, "", 1, 100000)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			cards, _, _ := s.ListCards(nil, id, 1, 200)
			list[i].Cards = cards
			return &list[i], nil
		}
	}
	return nil, nil
}

func (s *Store) CreateCustomer(c *model.Customer) error {
	if c.CustomerNo == "" {
		c.CustomerNo = fmt.Sprintf("LP%d", time.Now().UnixNano()%1e12)
	}
	res, err := s.DB.Exec(`INSERT INTO lp_customer(customer_no,name,mobile,gender,birthday,level_name,source,store_id,employee_id,total_amount,status,remark)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, c.CustomerNo, c.Name, c.Mobile, c.Gender, nullDate(c.BirthdayStr), c.LevelName, c.Source, c.StoreID, c.EmployeeID, c.TotalAmount, c.Status, c.Remark)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func nullDate(s string) interface{} {
	if s == "" {
		return nil
	}
	if len(s) == 5 {
		s = fmt.Sprintf("%d-%s", time.Now().Year(), s)
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return t
}

func (s *Store) UpdateCustomer(id int64, c *model.Customer) error {
	_, err := s.DB.Exec(`UPDATE lp_customer SET name=?,mobile=?,gender=?,level_name=?,source=?,store_id=?,remark=?,status=? WHERE id=? AND deleted_at IS NULL`,
		c.Name, c.Mobile, c.Gender, c.LevelName, c.Source, c.StoreID, c.Remark, c.Status, id)
	return err
}

func (s *Store) SoftDeleteCustomer(id int64) error {
	_, err := s.DB.Exec(`UPDATE lp_customer SET deleted_at=NOW() WHERE id=?`, id)
	return err
}

func (s *Store) ListCards(storeID *int64, customerID int64, page, size int) ([]model.Card, int64, error) {
	return s.ListCardsQuery(storeID, customerID, "", page, size)
}

func (s *Store) ListCardsQuery(storeID *int64, customerID int64, q string, page, size int) ([]model.Card, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if storeID != nil {
		where = append(where, "cd.store_id=?")
		args = append(args, *storeID)
	}
	if customerID > 0 {
		where = append(where, "cd.customer_id=?")
		args = append(args, customerID)
	}
	q = strings.TrimSpace(q)
	if q != "" {
		like := "%" + q + "%"
		where = append(where, "(cu.name LIKE ? OR cu.mobile LIKE ? OR cd.card_no LIKE ? OR COALESCE(p.name,'') LIKE ?)")
		args = append(args, like, like, like, like)
	}
	w := strings.Join(where, " AND ")
	join := `FROM lp_customer_card cd
		LEFT JOIN lp_customer cu ON cu.id=cd.customer_id
		LEFT JOIN lp_project p ON p.id=cd.project_id
		LEFT JOIN lp_store st ON st.id=cd.store_id`
	var total int64
	_ = s.DB.Get(&total, "SELECT COUNT(*) "+join+" WHERE "+w, args...)
	args2 := append(args, size, (page-1)*size)
	rows, err := s.DB.Queryx(`SELECT cd.id,cd.card_no,cd.customer_id,cd.project_id,cd.store_id,cd.employee_id,cd.total_times,cd.used_times,cd.remain_times,cd.frozen_times,cd.total_amount,cd.pay_method,COALESCE(cd.card_kind,'times'),cd.status,cd.version,cd.created_at,
		COALESCE(cu.name,''), COALESCE(p.name,''), COALESCE(st.name,'')
		`+join+` WHERE `+w+` ORDER BY cd.id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Card{}
	for rows.Next() {
		var c model.Card
		if err := rows.Scan(&c.ID, &c.CardNo, &c.CustomerID, &c.ProjectID, &c.StoreID, &c.EmployeeID, &c.TotalTimes, &c.UsedTimes, &c.RemainTimes, &c.FrozenTimes, &c.TotalAmount, &c.PayMethod, &c.CardKind, &c.Status, &c.Version, &c.CreatedAt, &c.CustomerName, &c.ProjectName, &c.StoreName); err != nil {
			return nil, 0, err
		}
		if c.ProjectID.Valid {
			c.ProjectIDVal = c.ProjectID.Int64
		}
		if c.CardKind == "" {
			c.CardKind = "times"
		}
		list = append(list, c)
	}
	return list, total, nil
}

func (s *Store) GetCard(id int64) (*model.Card, error) {
	list, _, err := s.ListCards(nil, 0, 1, 100000)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			return &list[i], nil
		}
	}
	return nil, nil
}

func (s *Store) CreateCard(c *model.Card) error {
	if c.CardNo == "" {
		c.CardNo = fmt.Sprintf("CD%d", time.Now().UnixNano()%1e10)
	}
	if c.CardKind == "" {
		c.CardKind = "amount"
	}
	var projectArg interface{}
	if c.ProjectID.Valid {
		projectArg = c.ProjectID.Int64
	} else if c.ProjectIDVal > 0 {
		projectArg = c.ProjectIDVal
		c.ProjectID = sql.NullInt64{Int64: c.ProjectIDVal, Valid: true}
	} else {
		projectArg = nil
	}
	res, err := s.DB.Exec(`INSERT INTO lp_customer_card(card_no,customer_id,project_id,store_id,employee_id,total_times,used_times,remain_times,frozen_times,total_amount,pay_method,card_kind,status,version,valid_from,valid_to)
		VALUES(?,?,?,?,?,?,0,?,0,?,?,?, '使用中',0,CURDATE(),DATE_ADD(CURDATE(), INTERVAL 365 DAY))`,
		c.CardNo, c.CustomerID, projectArg, c.StoreID, c.EmployeeID, c.TotalTimes, c.RemainTimes, c.TotalAmount, c.PayMethod, c.CardKind)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	remark := "开卡"
	if c.CardKind == "amount" {
		remark = "充值开卡"
	}
	_, _ = s.DB.Exec(`INSERT INTO lp_card_flow(card_id,customer_id,change_times,change_amount,biz_type,biz_no,remark,operator_id)
		VALUES(?,?,?,?, 'open',?,?,?)`, c.ID, c.CustomerID, c.TotalTimes, c.TotalAmount, c.CardNo, remark, c.EmployeeID)
	return nil
}

func (s *Store) ListVerifies(storeID *int64, page, size int) ([]model.VerifyOrder, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if storeID != nil {
		where = append(where, "v.store_id=?")
		args = append(args, *storeID)
	}
	w := strings.Join(where, " AND ")
	var total int64
	_ = s.DB.Get(&total, "SELECT COUNT(*) FROM lp_verify_order v WHERE "+w, args...)
	args2 := append(args, size, (page-1)*size)
	rows, err := s.DB.Queryx(`SELECT v.id,v.verify_no,v.customer_id,v.card_id,v.project_id,v.store_id,v.employee_id,v.total_amount,v.times,
		v.welfare_issued_id,COALESCE(v.welfare_offset,0),COALESCE(v.pay_amount,v.total_amount),
		v.status,v.token,v.token_expire_at,v.scanned_at,v.remark,v.created_at,
		COALESCE(cu.name,''), COALESCE(p.name,''), COALESCE(u.name,''), COALESCE(st.name,''), COALESCE(cd.card_kind,'times')
		FROM lp_verify_order v
		LEFT JOIN lp_customer cu ON cu.id=v.customer_id
		LEFT JOIN lp_project p ON p.id=v.project_id
		LEFT JOIN lp_user u ON u.id=v.employee_id
		LEFT JOIN lp_store st ON st.id=v.store_id
		LEFT JOIN lp_customer_card cd ON cd.id=v.card_id
		WHERE `+w+` ORDER BY v.id DESC LIMIT ? OFFSET ?`, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.VerifyOrder{}
	for rows.Next() {
		var v model.VerifyOrder
		if err := rows.Scan(&v.ID, &v.VerifyNo, &v.CustomerID, &v.CardID, &v.ProjectID, &v.StoreID, &v.EmployeeID, &v.TotalAmount, &v.Times,
			&v.WelfareIssuedID, &v.WelfareOffset, &v.PayAmount,
			&v.Status, &v.Token, &v.TokenExpireAt, &v.ScannedAt, &v.Remark, &v.CreatedAt,
			&v.CustomerName, &v.ProjectName, &v.StaffName, &v.StoreName, &v.CardKind); err != nil {
			return nil, 0, err
		}
		if v.WelfareIssuedID.Valid {
			v.WelfareIssuedVal = v.WelfareIssuedID.Int64
		}
		if v.PayAmount <= 0 && v.TotalAmount > 0 {
			v.PayAmount = v.TotalAmount
		}
		list = append(list, v)
	}
	return list, total, nil
}

func (s *Store) GetVerifyByToken(token string) (*model.VerifyOrder, error) {
	var v model.VerifyOrder
	err := s.DB.Get(&v, `SELECT id,verify_no,customer_id,card_id,project_id,store_id,employee_id,total_amount,times,
		welfare_issued_id,COALESCE(welfare_offset,0) AS welfare_offset,COALESCE(pay_amount,total_amount) AS pay_amount,
		status,token,token_expire_at,scanned_at,remark,created_at
		FROM lp_verify_order WHERE token=? LIMIT 1`, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if v.WelfareIssuedID.Valid {
		v.WelfareIssuedVal = v.WelfareIssuedID.Int64
	}
	return &v, nil
}

func (s *Store) GetVerify(id int64) (*model.VerifyOrder, error) {
	list, _, err := s.ListVerifies(nil, 1, 100000)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == id {
			return &list[i], nil
		}
	}
	return nil, nil
}

func (s *Store) LogOperate(userID *int64, module, action, detail, ip string) {
	_, _ = s.DB.Exec(`INSERT INTO lp_operate_log(user_id,module,action,detail,ip) VALUES(?,?,?,?,?)`, userID, module, action, detail, ip)
}
