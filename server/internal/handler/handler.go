package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lanping/lp-server/internal/middleware"
	"github.com/lanping/lp-server/internal/model"
	"github.com/lanping/lp-server/internal/response"
	"github.com/lanping/lp-server/internal/service"
)

type API struct {
	Svc *service.Service
}

func (a *API) Health(c *gin.Context) {
	response.OK(c, gin.H{"status": "up"})
}

func (a *API) Login(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Phone == "" {
		response.Fail(c, http.StatusBadRequest, "请输入手机号和密码")
		return
	}
	token, user, err := a.Svc.Login(req.Phone, req.Password)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	uid := user.ID
	a.Svc.Repo.LogOperate(&uid, "auth", "login", user.Name, c.ClientIP())
	response.OK(c, gin.H{"token": token, "user": user})
}

func (a *API) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Role     string `json:"role"`
		StoreID  int64  `json:"store_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "请完善注册信息")
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Phone) == "" || req.StoreID <= 0 {
		response.Fail(c, http.StatusBadRequest, "请填写姓名、手机号并选择门店")
		return
	}
	if err := a.Svc.Register(req.Name, req.Phone, req.Password, req.Role, req.StoreID); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, gin.H{"message": "已提交审核，请等待管理员通过"})
}

func (a *API) Me(c *gin.Context) {
	claims := middleware.GetClaims(c)
	u, err := a.Svc.Repo.FindUserByID(claims.UserID)
	if err != nil || u == nil {
		response.Fail(c, http.StatusUnauthorized, "用户不存在")
		return
	}
	response.OK(c, u.Public())
}

func (a *API) CustomerLogin(c *gin.Context) {
	var req struct {
		Phone string `json:"phone"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Phone == "" {
		response.Fail(c, http.StatusBadRequest, "请输入手机号")
		return
	}
	token, user, err := a.Svc.CustomerLogin(req.Phone)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}

func (a *API) CustomerMe(c *gin.Context) {
	claims := middleware.GetClaims(c)
	cu, err := a.Svc.Repo.GetCustomer(claims.UserID)
	if err != nil || cu == nil {
		response.Fail(c, http.StatusUnauthorized, "会员不存在")
		return
	}
	response.OK(c, cu)
}

func queryPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", c.DefaultQuery("page_size", "20")))
	return page, size
}

func (a *API) ListStores(c *gin.Context) {
	list, err := a.Svc.Repo.ListStores()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) CreateStore(c *gin.Context) {
	var req struct {
		Name      string  `json:"name"`
		Address   string  `json:"address"`
		Phone     string  `json:"phone"`
		Status    string  `json:"status"`
		Lat       float64 `json:"lat"`
		Lng       float64 `json:"lng"`
		ManagerID *int64  `json:"manager_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		response.Fail(c, 400, "请填写门店名称")
		return
	}
	st := &model.Store{Name: req.Name, Address: req.Address, Phone: req.Phone, Status: req.Status, Lat: req.Lat, Lng: req.Lng}
	if err := a.Svc.Repo.CreateStore(st); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if req.ManagerID != nil {
		if err := a.Svc.AssignStoreManager(st.ID, *req.ManagerID); err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
	}
	list, _ := a.Svc.Repo.ListStores()
	for _, s := range list {
		if s.ID == st.ID {
			response.OK(c, s)
			return
		}
	}
	response.OK(c, st)
}

func (a *API) UpdateStore(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	fields := map[string]interface{}{}
	for _, k := range []string{"name", "address", "phone", "status", "lat", "lng"} {
		if v, ok := req[k]; ok {
			fields[k] = v
		}
	}
	if err := a.Svc.Repo.UpdateStore(id, fields); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if _, hasManager := req["manager_id"]; hasManager {
		var managerID int64
		switch v := req["manager_id"].(type) {
		case float64:
			managerID = int64(v)
		case int64:
			managerID = v
		case int:
			managerID = int64(v)
		case string:
			managerID, _ = strconv.ParseInt(v, 10, 64)
		case nil:
			managerID = 0
		}
		if err := a.Svc.AssignStoreManager(id, managerID); err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
	}
	list, _ := a.Svc.Repo.ListStores()
	for _, s := range list {
		if s.ID == id {
			response.OK(c, s)
			return
		}
	}
	response.OK(c, gin.H{"id": id})
}

func (a *API) Dashboard(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	period := c.DefaultQuery("period", "month")
	storeID := a.Svc.ScopeStore(claims, sid)
	data, err := a.Svc.Dashboard(storeID, period)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) Workbench(c *gin.Context) {
	data, err := a.Svc.Workbench(middleware.GetClaims(c))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) ListCustomers(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	page, size := queryPage(c)
	list, total, err := a.Svc.Repo.ListCustomers(a.Svc.ScopeCustomers(claims, sid), c.Query("q"), page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, service.PageOf(list, total, page, size))
}

func (a *API) GetCustomer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	cu, err := a.Svc.Repo.GetCustomer(id)
	if err != nil || cu == nil {
		response.Fail(c, 404, "客户不存在")
		return
	}
	response.OK(c, cu)
}

func (a *API) CreateCustomer(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	name, _ := raw["name"].(string)
	phone, _ := raw["phone"].(string)
	if phone == "" {
		phone, _ = raw["mobile"].(string)
	}
	if name == "" || phone == "" {
		response.Fail(c, 400, "请填写姓名和手机号")
		return
	}
	req := model.Customer{Name: name, Mobile: phone, Status: 1, LevelName: "银卡会员", Source: "到店"}
	if v, ok := raw["level_name"].(string); ok && v != "" {
		req.LevelName = v
	} else if v, ok := raw["level"].(string); ok && v != "" {
		req.LevelName = v
	}
	if v, ok := raw["source"].(string); ok && v != "" {
		req.Source = v
	}
	if v, ok := raw["remark"].(string); ok {
		req.Remark = v
	}
	if v, ok := raw["store_id"].(float64); ok && v > 0 {
		req.StoreID = int64(v)
	}
	if req.StoreID == 0 && claims.StoreID != nil {
		req.StoreID = *claims.StoreID
	}
	if req.StoreID == 0 {
		response.Fail(c, 400, "请选择所属门店")
		return
	}
	// 归属到建档员工，计入「新增客户」考核
	if claims.UserID > 0 && (claims.Role == "staff" || claims.Role == "manager" || claims.Role == "admin") {
		req.EmployeeID = sql.NullInt64{Int64: claims.UserID, Valid: true}
	}
	if err := a.Svc.Repo.CreateCustomer(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "customer", "create", req.Name, c.ClientIP())
	if req.StoreID > 0 {
		sid := req.StoreID
		_ = a.Svc.SyncAssessmentProgress(time.Now().Format("2006-01"), &sid)
	}
	response.OK(c, req)
}

func (a *API) UpdateCustomer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.Customer
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := a.Svc.Repo.UpdateCustomer(id, &req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	cu, _ := a.Svc.Repo.GetCustomer(id)
	response.OK(c, cu)
}

func (a *API) DeleteCustomer(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.Repo.SoftDeleteCustomer(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListUsers(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	page, size := queryPage(c)
	list, total, err := a.Svc.Repo.ListUsers(a.Svc.ScopeStore(claims, sid), c.Query("q"), page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	out := make([]model.PublicUser, 0, len(list))
	for _, u := range list {
		out = append(out, u.Public())
	}
	response.OK(c, service.PageOf(out, total, page, size))
}

func (a *API) CreateUser(c *gin.Context) {
	var req struct {
		Name       string  `json:"name"`
		Phone      string  `json:"phone"`
		Password   string  `json:"password"`
		Role       string  `json:"role"`
		StoreID    *int64  `json:"store_id"`
		Position   string  `json:"position"`
		BaseSalary float64 `json:"base_salary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Phone) == "" || strings.TrimSpace(req.Name) == "" {
		response.Fail(c, 400, "请填写姓名和手机号")
		return
	}
	if req.Role != "manager" && req.Role != "staff" {
		response.Fail(c, 400, "请选择角色：店长或员工")
		return
	}
	if req.StoreID == nil || *req.StoreID <= 0 {
		response.Fail(c, 400, "请选择所属门店")
		return
	}
	if req.Password == "" {
		req.Password = "123456"
	}
	if req.Position == "" {
		if req.Role == "manager" {
			req.Position = "店长"
		} else {
			req.Position = "美容师"
		}
	}
	hash, _ := a.Svc.HashPassword(req.Password)
	roleName := map[string]string{"manager": "店长", "staff": "美业员工"}[req.Role]
	if req.BaseSalary <= 0 {
		req.BaseSalary = a.Svc.FixedForRole(req.Role)
	}
	u := &model.User{Name: req.Name, Phone: req.Phone, PasswordHash: hash, Role: req.Role, RoleName: roleName, Position: req.Position, Status: "active", BaseSalary: req.BaseSalary}
	u.StoreID = sql.NullInt64{Int64: *req.StoreID, Valid: true}
	if err := a.Svc.Repo.CreateUser(u); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	if req.Role == "manager" {
		if err := a.Svc.AssignStoreManager(*req.StoreID, u.ID); err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
	}
	uid := middleware.GetClaims(c).UserID
	a.Svc.Repo.LogOperate(&uid, "user", "create", req.Name+"/"+req.Role, c.ClientIP())
	full, _ := a.Svc.Repo.FindUserByID(u.ID)
	if full != nil {
		response.OK(c, full.Public())
		return
	}
	response.OK(c, u.Public())
}

func (a *API) UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	old, _ := a.Svc.Repo.FindUserByID(id)
	var req map[string]interface{}
	_ = c.ShouldBindJSON(&req)
	fields := map[string]interface{}{}
	for _, k := range []string{"name", "role", "position", "status", "base_salary", "store_id"} {
		if v, ok := req[k]; ok {
			fields[k] = v
		}
	}
	if role, ok := fields["role"].(string); ok {
		fields["role_name"] = map[string]string{"admin": "超级管理员", "manager": "店长", "staff": "美业员工"}[role]
	}
	if err := a.Svc.Repo.UpdateUser(id, fields); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	u, _ := a.Svc.Repo.FindUserByID(id)
	if u != nil {
		_, roleTouched := fields["role"]
		_, storeTouched := fields["store_id"]
		if roleTouched || storeTouched {
			role := u.Role
			storeID := u.StoreIDValue()
			if err := a.Svc.SyncStoreManagerFromUser(id, role, storeID, old); err != nil {
				response.Fail(c, 400, err.Error())
				return
			}
			u, _ = a.Svc.Repo.FindUserByID(id)
		} else if name, ok := fields["name"].(string); ok && u.Role == "manager" && u.StoreID.Valid {
			_, _ = a.Svc.Repo.DB.Exec(`UPDATE lp_store SET manager_name=? WHERE id=?`, name, u.StoreID.Int64)
		}
	}
	if u != nil {
		response.OK(c, u.Public())
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (a *API) ListProjects(c *gin.Context) {
	page, size := queryPage(c)
	list, total, err := a.Svc.Repo.ListProjects(c.Query("q"), c.Query("status"), page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, service.PageOf(list, total, page, size))
}

func (a *API) CreateProject(c *gin.Context) {
	var p model.Project
	if err := c.ShouldBindJSON(&p); err != nil || p.Name == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	if p.Status == "" {
		p.Status = "上架"
	}
	if p.TotalTimes == 0 {
		p.TotalTimes = 1
	}
	p.CardMode = service.NormalizeProjectCardMode(p.CardMode)
	if err := a.Svc.Repo.CreateProject(&p); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, p)
}

func (a *API) UpdateProject(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var p model.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	p.CardMode = service.NormalizeProjectCardMode(p.CardMode)
	res, err := a.Svc.UpdateProject(id, &p)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	p.ID = id
	response.OK(c, gin.H{"project": p, "convert": res})
}

func (a *API) DeleteProject(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.Repo.DeleteProject(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListCards(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	cid, _ := strconv.ParseInt(c.Query("customer_id"), 10, 64)
	q := strings.TrimSpace(c.Query("q"))
	page, size := queryPage(c)
	scope := a.Svc.ScopeCustomers(claims, sid)
	if claims.Role == "customer" {
		cid = claims.UserID
		scope = nil
	}
	list, total, err := a.Svc.Repo.ListCardsQuery(scope, cid, q, page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, service.PageOf(list, total, page, size))
}

func (a *API) CreateCard(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req struct {
		CustomerID int64   `json:"customer_id"`
		ProjectID  int64   `json:"project_id"`
		StoreID    int64   `json:"store_id"`
		PayMethod  string  `json:"pay_method"`
		CardKind   string  `json:"card_kind"`
		Times      int     `json:"times"`
		Amount     float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CustomerID == 0 {
		response.Fail(c, 400, "请选择客户")
		return
	}
	storeID := req.StoreID
	if storeID == 0 && claims.StoreID != nil {
		storeID = *claims.StoreID
	}
	card, err := a.Svc.OpenCard(service.OpenCardInput{
		EmployeeID: claims.UserID,
		StoreID:    storeID,
		CustomerID: req.CustomerID,
		ProjectID:  req.ProjectID,
		PayMethod:  req.PayMethod,
		CardKind:   req.CardKind,
		Times:      req.Times,
		Amount:     req.Amount,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "card", "open", card.CardNo, c.ClientIP())
	response.OK(c, card)
}

func (a *API) ListVerifies(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	page, size := queryPage(c)
	list, total, err := a.Svc.Repo.ListVerifies(a.Svc.ScopeStore(claims, sid), page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, service.PageOf(list, total, page, size))
}

func (a *API) DraftVerify(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req struct {
		CardID          int64   `json:"card_id"`
		ProjectID       int64   `json:"project_id"`
		Times           int     `json:"times"`
		Amount          float64 `json:"amount"`
		WelfareIssuedID int64   `json:"welfare_issued_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CardID == 0 {
		response.Fail(c, 400, "请选择卡项")
		return
	}
	storeID := int64(0)
	if claims.StoreID != nil {
		storeID = *claims.StoreID
	}
	v, err := a.Svc.DraftVerify(service.DraftVerifyInput{
		EmployeeID:      claims.UserID,
		StoreID:         storeID,
		CardID:          req.CardID,
		ProjectID:       req.ProjectID,
		Times:           req.Times,
		Amount:          req.Amount,
		WelfareIssuedID: req.WelfareIssuedID,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, v)
}

func (a *API) GetVerify(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	v, err := a.Svc.Repo.GetVerify(id)
	if err != nil || v == nil {
		response.Fail(c, 404, "核销单不存在")
		return
	}
	response.OK(c, v)
}

func (a *API) ConfirmVerify(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Token string `json:"token"`
	}
	_ = c.ShouldBindJSON(&req)
	v, err := a.Svc.ConfirmVerify(id, req.Token)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, v)
}

func (a *API) CancelVerify(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.CancelVerify(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) RefundVerify(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.RefundVerify(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) PublicVerification(c *gin.Context) {
	data, err := a.Svc.PublicVerification(c.Param("token"))
	if err != nil {
		response.Fail(c, 404, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) ConfirmPublicVerification(c *gin.Context) {
	v, err := a.Svc.ConfirmVerifyByToken(c.Param("token"))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, v)
}

func (a *API) ListAppointments(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	page, size := queryPage(c)
	data, err := a.Svc.ListAppointments(a.Svc.ScopeStore(claims, sid), page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) CreateAppointment(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var raw struct {
		CustomerID int64  `json:"customer_id"`
		ProjectID  int64  `json:"project_id"`
		StoreID    int64  `json:"store_id"`
		StaffID    int64  `json:"staff_id"`
		Date       string `json:"date"`
		Time       string `json:"time"`
		Note       string `json:"note"`
		Status     string `json:"status"`
		Source     string `json:"source"`
	}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if raw.CustomerID == 0 || raw.ProjectID == 0 || raw.Date == "" || raw.Time == "" {
		response.Fail(c, 400, "请选择客户、项目、日期和时间")
		return
	}
	req := model.Appointment{
		CustomerID:  raw.CustomerID,
		ProjectID:   raw.ProjectID,
		StoreID:     raw.StoreID,
		StaffIDVal:  raw.StaffID,
		Date:        raw.Date,
		AppointTime: raw.Time,
		Note:        raw.Note,
		Status:      raw.Status,
		Source:      raw.Source,
	}
	if req.StoreID == 0 && claims.StoreID != nil {
		req.StoreID = *claims.StoreID
	}
	if req.StaffIDVal == 0 {
		req.StaffIDVal = claims.UserID
	}
	if req.Source == "" {
		if claims.Role == "customer" {
			req.Source = "顾客端"
		} else {
			req.Source = "门店"
		}
	}
	// 心智：新建一律「待履约」；客户到店后员工再点「确认履约」→ 已履约
	if req.Status == "" || req.Status == "已确认" || req.Status == "待确认" || req.Status == "已履约" {
		req.Status = "待履约"
	}
	if err := a.Svc.CreateAppointment(&req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, req)
}

func (a *API) ListStaffAvailability(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	pid, _ := strconv.ParseInt(c.Query("project_id"), 10, 64)
	excludeID, _ := strconv.ParseInt(c.Query("exclude_id"), 10, 64)
	list, err := a.Svc.ListStaffAvailability(a.Svc.ScopeStore(claims, sid), c.Query("date"), c.Query("time"), pid, excludeID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) UpdateAppointment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status     string  `json:"status"`
		CustomerID int64   `json:"customer_id"`
		ProjectID  int64   `json:"project_id"`
		StoreID    int64   `json:"store_id"`
		StaffID    int64   `json:"staff_id"`
		Date       string  `json:"date"`
		Time       string  `json:"time"`
		Note       *string `json:"note"`
		Reschedule bool    `json:"reschedule"`
	}
	_ = c.ShouldBindJSON(&req)
	// 带了时间/人/项目任一字段时视为改约
	if !req.Reschedule && (req.Date != "" || req.Time != "" || req.StaffID > 0 || req.ProjectID > 0 || req.CustomerID > 0) {
		req.Reschedule = true
	}
	if err := a.Svc.UpdateAppointment(id, service.AppointmentUpdateInput{
		Status:     req.Status,
		CustomerID: req.CustomerID,
		ProjectID:  req.ProjectID,
		StoreID:    req.StoreID,
		StaffID:    req.StaffID,
		Date:       req.Date,
		Time:       req.Time,
		Note:       req.Note,
		Reschedule: req.Reschedule,
	}); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) DeleteAppointment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.DeleteAppointment(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListInventory(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	list, err := a.Svc.ListInventory(a.Svc.ScopeStore(claims, sid))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ListStockItems(c *gin.Context) {
	list, err := a.Svc.ListStockItems()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) CreateStockItem(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req struct {
		Name    string `json:"name"`
		SKU     string `json:"sku"`
		Unit    string `json:"unit"`
		StoreID int64  `json:"store_id"`
		Qty     int    `json:"qty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.StoreID == 0 && claims.StoreID != nil {
		req.StoreID = *claims.StoreID
	}
	item, err := a.Svc.CreateStockItem(service.StockItemInput{
		Name: req.Name, SKU: req.SKU, Unit: req.Unit, StoreID: req.StoreID, Qty: req.Qty,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "inventory", "create_item", fmt.Sprintf("%v", item["name"]), c.ClientIP())
	response.OK(c, item)
}

func (a *API) UpdateStockItem(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name    string `json:"name"`
		SKU     string `json:"sku"`
		Unit    string `json:"unit"`
		StoreID int64  `json:"store_id"`
		Qty     int    `json:"qty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || id == 0 {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.StoreID == 0 && claims.StoreID != nil {
		req.StoreID = *claims.StoreID
	}
	if err := a.Svc.UpdateStockItem(id, service.StockItemInput{
		Name: req.Name, SKU: req.SKU, Unit: req.Unit, StoreID: req.StoreID, Qty: req.Qty,
	}); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "inventory", "update_item", req.Name, c.ClientIP())
	response.OK(c, gin.H{"ok": true})
}

func (a *API) DeleteStockItem(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := a.Svc.DeleteStockItem(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "inventory", "delete_item", fmt.Sprintf("%d", id), c.ClientIP())
	response.OK(c, gin.H{"ok": true})
}

func (a *API) CreateInventoryOrder(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req struct {
		StoreID   int64  `json:"store_id"`
		ToStoreID int64  `json:"to_store_id"`
		ItemID    int64  `json:"item_id"`
		Qty       int    `json:"qty"`
		Direction string `json:"direction"`
		Channel   string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ItemID == 0 || req.Qty <= 0 {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.StoreID == 0 && claims.StoreID != nil {
		req.StoreID = *claims.StoreID
	}
	if req.StoreID == 0 {
		response.Fail(c, 400, "请选择门店")
		return
	}
	if req.Direction == "" {
		req.Direction = "出库"
	}
	if req.Channel == "" {
		req.Channel = "门店调度"
	}
	if err := a.Svc.CreateStockOrder(service.StockOrderInput{
		FromStoreID: req.StoreID,
		ToStoreID:   req.ToStoreID,
		ItemID:      req.ItemID,
		OperatorID:  claims.UserID,
		Qty:         req.Qty,
		Direction:   req.Direction,
		Channel:     req.Channel,
	}); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "inventory", "order", req.Direction, c.ClientIP())
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListInventoryOrders(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	limit := 200
	if claims.Role != "admin" {
		limit = 80
	}
	list, err := a.Svc.ListStockOrders(a.Svc.ScopeStore(claims, sid), limit)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ListPendingInventoryTransfers(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid := int64(0)
	if claims.StoreID != nil {
		sid = *claims.StoreID
	}
	qSid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	if qSid > 0 && (claims.Role == "admin" || (claims.StoreID != nil && *claims.StoreID == qSid)) {
		sid = qSid
	}
	list, err := a.Svc.ListPendingStockTransfers(sid)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ConfirmInventoryTransfer(c *gin.Context) {
	claims := middleware.GetClaims(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := a.Svc.ConfirmStockTransfer(id, claims.UserID, claims.StoreID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	uid := claims.UserID
	a.Svc.Repo.LogOperate(&uid, "inventory", "confirm_transfer", fmt.Sprintf("%d", id), c.ClientIP())
	response.OK(c, gin.H{"ok": true})
}

func (a *API) CheckIn(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req struct {
		Method string   `json:"method"`
		Lat    *float64 `json:"lat"`
		Lng    *float64 `json:"lng"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Method == "" {
		req.Method = "定位"
	}
	storeID := int64(0)
	if claims.StoreID != nil {
		storeID = *claims.StoreID
	}
	data, err := a.Svc.CheckIn(claims.UserID, storeID, req.Method, req.Lat, req.Lng)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) GeoSearch(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	data, err := a.Svc.GeoSearch(q)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) AttendanceToday(c *gin.Context) {
	claims := middleware.GetClaims(c)
	data, err := a.Svc.AttendanceToday(claims.UserID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) ListAttendance(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	data, err := a.Svc.ListAttendanceMonth(month, a.Svc.ScopeStore(claims, sid))
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) ListAudits(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	list, err := a.Svc.ListAudits(a.Svc.ScopeStore(claims, sid))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ApproveAudit(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.ReviewAudit(id, middleware.GetClaims(c).UserID, true); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) RejectAudit(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.ReviewAudit(id, middleware.GetClaims(c).UserID, false); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) GetSettings(c *gin.Context) {
	claims := middleware.GetClaims(c)
	data, err := a.Svc.GetSettings()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	// 非管理员不返回 AccessKey 明文
	if claims == nil || claims.Role != "admin" {
		delete(data, "sms_access_key_id")
		delete(data, "sms_access_key_secret")
		delete(data, "amap_web_key")
	}
	response.OK(c, data)
}

func (a *API) SaveSettings(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 核销模式、客户共享、阿里云密钥仅管理员可改
	if claims.Role != "admin" {
		delete(req, "verify_mode")
		delete(req, "customer_share")
		delete(req, "sms_access_key_id")
		delete(req, "sms_access_key_secret")
		delete(req, "sms_tpl_birthday")
		delete(req, "sms_tpl_consume")
		delete(req, "sms_tpl_appoint")
		delete(req, "sms_endpoint")
		delete(req, "amap_web_key")
	}
	if err := a.Svc.SaveSettings(req); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, req)
}

func (a *API) CreateAssessment(c *gin.Context) {
	var req struct {
		UserID  int64   `json:"user_id"`
		StoreID int64   `json:"store_id"`
		Month   string  `json:"month"`
		Name    string  `json:"name"`
		Weight  float64 `json:"weight"`
		Target  float64 `json:"target"`
		Done    float64 `json:"done"`
		Unified bool    `json:"unified"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 默认统一下发，不选人
	if !req.Unified && req.UserID <= 0 {
		req.Unified = true
	}
	id, err := a.Svc.UpsertAssessment(0, service.AssessmentInput{
		UserID: req.UserID, StoreID: req.StoreID, Month: req.Month, Name: req.Name,
		Weight: req.Weight, Target: req.Target, Done: req.Done, Unified: req.Unified,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id, "unified": req.Unified})
}

func (a *API) UpdateAssessment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		UserID  int64   `json:"user_id"`
		StoreID int64   `json:"store_id"`
		Month   string  `json:"month"`
		Name    string  `json:"name"`
		Weight  float64 `json:"weight"`
		Target  float64 `json:"target"`
		Done    float64 `json:"done"`
		Unified bool    `json:"unified"` // true：按指标名批量改目标/权重
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Unified {
		n, err := a.Svc.UpdateUnifiedAssessment(req.Month, req.Name, req.Weight, req.Target, req.StoreID)
		if err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
		response.OK(c, gin.H{"updated": n})
		return
	}
	_, err := a.Svc.UpsertAssessment(id, service.AssessmentInput{
		UserID: req.UserID, StoreID: req.StoreID, Month: req.Month, Name: req.Name,
		Weight: req.Weight, Target: req.Target, Done: req.Done,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (a *API) DeleteAssessment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	// ?unified=1&month=&name= 删除整项统一指标
	if c.Query("unified") == "1" || c.Query("unified") == "true" {
		n, err := a.Svc.DeleteUnifiedAssessment(c.Query("month"), c.Query("name"), 0)
		if err != nil {
			response.Fail(c, 400, err.Error())
			return
		}
		response.OK(c, gin.H{"deleted": n})
		return
	}
	if err := a.Svc.DeleteAssessment(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListAssessments(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	list, err := a.Svc.ListAssessments(a.Svc.ScopeStore(claims, sid), c.Query("month"))
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ComputeSalary(c *gin.Context) {
	claims := middleware.GetClaims(c)
	var req struct {
		Month   string `json:"month"`
		StoreID int64  `json:"store_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := a.Svc.ComputeSalary(req.Month, a.Svc.ScopeStore(claims, req.StoreID)); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListSalary(c *gin.Context) {
	claims := middleware.GetClaims(c)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	page, size := queryPage(c)
	data, err := a.Svc.ListSalary(a.Svc.ScopeStore(claims, sid), c.Query("month"), page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) SalaryStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status string `json:"status"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := a.Svc.SetSalaryStatus(id, req.Status); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) GetSalaryRules(c *gin.Context) {
	response.OK(c, a.Svc.SalaryRules())
}

func (a *API) SaveSalaryRules(c *gin.Context) {
	var cfg service.SalaryConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := a.Svc.SaveSalaryConfig(cfg); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, a.Svc.SalaryRules())
}

func (a *API) ListWelfare(c *gin.Context) {
	list, err := a.Svc.ListWelfare()
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) CreateWelfare(c *gin.Context) {
	var req struct {
		Name        string  `json:"name"`
		Amount      float64 `json:"amount"`
		ValidDays   int     `json:"valid_days"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	id, err := a.Svc.CreateWelfare(req.Name, req.Amount, req.ValidDays, req.Description, req.Status)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (a *API) UpdateWelfare(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name        string  `json:"name"`
		Amount      float64 `json:"amount"`
		ValidDays   int     `json:"valid_days"`
		Description string  `json:"description"`
		Status      string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := a.Svc.UpdateWelfare(id, req.Name, req.Amount, req.ValidDays, req.Description, req.Status); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) ListAvailableWelfare(c *gin.Context) {
	claims := middleware.GetClaims(c)
	cid, _ := strconv.ParseInt(c.Query("customer_id"), 10, 64)
	sid, _ := strconv.ParseInt(c.Query("store_id"), 10, 64)
	if sid == 0 && claims.StoreID != nil {
		sid = *claims.StoreID
	}
	list, err := a.Svc.ListAvailableWelfare(cid, sid)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) DeleteWelfare(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := a.Svc.DeleteWelfare(id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (a *API) IssueWelfare(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		UserID     int64 `json:"user_id"`      // 兼容旧字段：客户 ID
		CustomerID int64 `json:"customer_id"`  // 可选客户
		StoreID    int64 `json:"store_id"`     // 单独门店
		AllStores  bool  `json:"all_stores"`   // 统一全部门店
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	custID := req.CustomerID
	if custID == 0 {
		custID = req.UserID
	}
	n, err := a.Svc.IssueWelfare(service.IssueWelfareInput{
		WelfareID:  id,
		CustomerID: custID,
		StoreID:    req.StoreID,
		AllStores:  req.AllStores,
	})
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true, "issued_stores": n})
}

func (a *API) ListWelfareIssued(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("pageSize"))
	list, err := a.Svc.ListWelfareIssued(limit)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ListMessages(c *gin.Context) {
	list, err := a.Svc.ListMessages(middleware.GetClaims(c).UserID)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

func (a *API) ListSMS(c *gin.Context) {
	page, size := queryPage(c)
	var total int64
	_ = a.Svc.Repo.DB.Get(&total, `SELECT COUNT(*) FROM lp_sms_log`)
	rows, err := a.Svc.Repo.DB.Queryx(`SELECT id,mobile,template_code,content,status,created_at FROM lp_sms_log ORDER BY id DESC LIMIT ? OFFSET ?`, size, (page-1)*size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var mobile, tpl, content, status string
		var at sql.NullTime
		_ = rows.Scan(&id, &mobile, &tpl, &content, &status, &at)
		item := map[string]interface{}{"id": id, "mobile": mobile, "template_code": tpl, "content": content, "status": status}
		if at.Valid {
			item["created_at"] = at.Time.Format("2006-01-02 15:04:05")
		}
		list = append(list, item)
	}
	response.OK(c, service.PageOf(list, total, page, size))
}

func (a *API) ListOperateLogs(c *gin.Context) {
	page, size := queryPage(c)
	data, err := a.Svc.ListOperateLogs(page, size)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

func (a *API) BirthdayUpcoming(c *gin.Context) {
	claims := middleware.GetClaims(c)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	q := `SELECT id,name,mobile,birthday,store_id FROM lp_customer WHERE deleted_at IS NULL AND birthday IS NOT NULL`
	args := []interface{}{}
	if claims.Role != "admin" && claims.StoreID != nil {
		q += " AND store_id=?"
		args = append(args, *claims.StoreID)
	}
	rows, err := a.Svc.Repo.DB.Queryx(q, args...)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	now := time.Now()
	for rows.Next() {
		var id, storeID int64
		var name, mobile string
		var birthday sql.NullTime
		if err := rows.Scan(&id, &name, &mobile, &birthday, &storeID); err != nil {
			continue
		}
		if !birthday.Valid {
			continue
		}
		b := birthday.Time
		bThis := time.Date(now.Year(), b.Month(), b.Day(), 0, 0, 0, 0, now.Location())
		if bThis.Before(now.Truncate(24 * time.Hour)) {
			bThis = bThis.AddDate(1, 0, 0)
		}
		daysLeft := int(bThis.Sub(now.Truncate(24*time.Hour)).Hours() / 24)
		if daysLeft < 0 || daysLeft > days {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "phone": mobile, "date": b.Format("01-02"), "store_id": storeID, "days_left": daysLeft,
		})
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

// BirthdayNotify 发送生日短信：body { days?:7, customer_id?:0 }
func (a *API) BirthdayNotify(c *gin.Context) {
	var req struct {
		Days       int   `json:"days"`
		CustomerID int64 `json:"customer_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Days <= 0 {
		req.Days = 7
		if v := a.Svc.SettingValue("birthday_days", "7"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				req.Days = n
			}
		}
	}
	if a.Svc.SettingValue("birthday_remind", "true") == "false" && req.CustomerID == 0 {
		response.Fail(c, 400, "生日提醒已关闭")
		return
	}
	sent, failed, err := a.Svc.NotifyBirthday(req.Days, req.CustomerID)
	if err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.OK(c, gin.H{"sent": sent, "failed": failed})
}
