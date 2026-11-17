package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
)

// 销售提成 = 开卡提成；消耗提成 = 核销提成
// 规则以 DB lp_salary_rule.config_json 为准，管理端可改。

type SalesTier struct {
	Min  float64 `json:"min"`            // 含下限（元）
	Max  float64 `json:"max"`            // 不含上限；0=无上限
	Rate float64 `json:"rate"`           // 提成率 0~1
	Label string `json:"label,omitempty"`
}

type CapByHeadcount struct {
	MaxPeople *int    `json:"max_people,omitempty"` // ≤N
	People    *int    `json:"people,omitempty"`     // =N
	MinPeople *int    `json:"min_people,omitempty"` // ≥N
	CapRatio  float64 `json:"cap_ratio"`
}

type RoleFixedItems struct {
	BaseSalary   float64 `json:"base_salary"`   // 基本工资
	Overtime     float64 `json:"overtime"`      // 固定加班费
	NonCompete   float64 `json:"non_compete"`   // 禁业补贴（店长）
	SocialSubsidy float64 `json:"social_subsidy"` // 社保补贴
}

func (f RoleFixedItems) Total(role string) float64 {
	sum := f.BaseSalary + f.Overtime + f.SocialSubsidy
	if role == "manager" {
		sum += f.NonCompete
	}
	return round2(sum)
}

type RoleSalaryRule struct {
	FixedItems  RoleFixedItems `json:"fixed_items"`
	SalesTiers  []SalesTier    `json:"sales_tiers"`
	VerifyRate  float64        `json:"verify_rate"`
}

type ManagerExtraRule struct {
	MinStoreShare       float64          `json:"min_store_share"`
	MinShortfallPenalty float64          `json:"min_shortfall_penalty"`
	CapByHeadcount      []CapByHeadcount `json:"cap_by_headcount"`
	AttendDaysGt        int              `json:"attend_days_gt"`
}

type SalaryConfig struct {
	Name    string           `json:"name"`
	Note    string           `json:"note"`
	Manager RoleSalaryRule   `json:"manager"`
	Staff   RoleSalaryRule   `json:"staff"`
	Extra   ManagerExtraRule `json:"manager_extra"`
}

func DefaultSalaryConfig() SalaryConfig {
	tiers := []SalesTier{
		{Min: 0, Max: 10000, Rate: 0.02, Label: "1万以下"},
		{Min: 10000, Max: 15000, Rate: 0.05, Label: "1万含-1.5万"},
		{Min: 15000, Max: 20000, Rate: 0.08, Label: "1.5万含-2万"},
		{Min: 20000, Max: 25000, Rate: 0.11, Label: "2万含-2.5万"},
		{Min: 25000, Max: 30000, Rate: 0.12, Label: "2.5万含-3万"},
		{Min: 30000, Max: 0, Rate: 0.13, Label: "3万及以上"},
	}
	two, three, four := 2, 3, 4
	return SalaryConfig{
		Name: "兰瓶美业固定薪酬+开卡/核销提成",
		Note: "销售提成=开卡提成；消耗提成=核销提成",
		Manager: RoleSalaryRule{
			FixedItems: RoleFixedItems{BaseSalary: 1600, Overtime: 700, NonCompete: 100, SocialSubsidy: 1100},
			SalesTiers: append([]SalesTier{}, tiers...),
			VerifyRate: 0.05,
		},
		Staff: RoleSalaryRule{
			FixedItems: RoleFixedItems{BaseSalary: 1200, Overtime: 700, SocialSubsidy: 1100},
			SalesTiers: append([]SalesTier{}, tiers...),
			VerifyRate: 0.05,
		},
		Extra: ManagerExtraRule{
			MinStoreShare:       0.10,
			MinShortfallPenalty: 0.08,
			AttendDaysGt:        20,
			CapByHeadcount: []CapByHeadcount{
				{MaxPeople: &two, CapRatio: 0.60},
				{People: &three, CapRatio: 0.40},
				{MinPeople: &four, CapRatio: 0.30},
			},
		},
	}
}

func (c SalaryConfig) RoleRule(role string) RoleSalaryRule {
	if role == "manager" {
		return c.Manager
	}
	return c.Staff
}

func (c SalaryConfig) FixedTotal(role string) float64 {
	return c.RoleRule(role).FixedItems.Total(role)
}

func (c SalaryConfig) SalesCommissionRate(role string, salesAmount float64) float64 {
	if salesAmount < 0 {
		salesAmount = 0
	}
	tiers := c.RoleRule(role).SalesTiers
	if len(tiers) == 0 {
		tiers = DefaultSalaryConfig().RoleRule(role).SalesTiers
	}
	for _, t := range tiers {
		if t.Max == 0 {
			if salesAmount >= t.Min {
				return t.Rate
			}
			continue
		}
		if salesAmount >= t.Min && salesAmount < t.Max {
			return t.Rate
		}
	}
	return tiers[len(tiers)-1].Rate
}

func (c SalaryConfig) ManagerSalesCapRatio(workingHeadcount int) float64 {
	caps := c.Extra.CapByHeadcount
	if len(caps) == 0 {
		caps = DefaultSalaryConfig().Extra.CapByHeadcount
	}
	for _, item := range caps {
		if item.People != nil && workingHeadcount == *item.People {
			return item.CapRatio
		}
	}
	for _, item := range caps {
		if item.MaxPeople != nil && workingHeadcount <= *item.MaxPeople {
			return item.CapRatio
		}
	}
	// 找最大的 min_people 规则中匹配的
	var best *CapByHeadcount
	for i := range caps {
		item := &caps[i]
		if item.MinPeople != nil && workingHeadcount >= *item.MinPeople {
			if best == nil || *item.MinPeople > *best.MinPeople {
				best = item
			}
		}
	}
	if best != nil {
		return best.CapRatio
	}
	return 0.30
}

func looksCorruptedText(s string) bool {
	if s == "" {
		return true
	}
	q, han := 0, 0
	for _, r := range s {
		if r == '?' || r == '？' {
			q++
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			han++
		}
	}
	// 中文业务文案被写成 ???? 时：问号多且几乎无汉字
	return q >= 2 && han == 0
}

func (c SalaryConfig) Normalize() SalaryConfig {
	def := DefaultSalaryConfig()
	if c.Name == "" || looksCorruptedText(c.Name) {
		c.Name = def.Name
	}
	if c.Note == "" || looksCorruptedText(c.Note) {
		c.Note = def.Note
	}
	if c.Manager.VerifyRate <= 0 {
		c.Manager.VerifyRate = 0.05
	}
	if c.Staff.VerifyRate <= 0 {
		c.Staff.VerifyRate = 0.05
	}
	if len(c.Manager.SalesTiers) == 0 {
		c.Manager.SalesTiers = def.Manager.SalesTiers
	} else {
		for i := range c.Manager.SalesTiers {
			if looksCorruptedText(c.Manager.SalesTiers[i].Label) && i < len(def.Manager.SalesTiers) {
				c.Manager.SalesTiers[i].Label = def.Manager.SalesTiers[i].Label
			}
		}
	}
	if len(c.Staff.SalesTiers) == 0 {
		c.Staff.SalesTiers = def.Staff.SalesTiers
	} else {
		for i := range c.Staff.SalesTiers {
			if looksCorruptedText(c.Staff.SalesTiers[i].Label) && i < len(def.Staff.SalesTiers) {
				c.Staff.SalesTiers[i].Label = def.Staff.SalesTiers[i].Label
			}
		}
	}
	if c.Extra.AttendDaysGt <= 0 {
		c.Extra.AttendDaysGt = 20
	}
	if c.Extra.MinStoreShare <= 0 {
		c.Extra.MinStoreShare = 0.10
	}
	if c.Extra.MinShortfallPenalty <= 0 {
		c.Extra.MinShortfallPenalty = 0.08
	}
	if len(c.Extra.CapByHeadcount) == 0 {
		c.Extra.CapByHeadcount = DefaultSalaryConfig().Extra.CapByHeadcount
	}
	sort.Slice(c.Manager.SalesTiers, func(i, j int) bool { return c.Manager.SalesTiers[i].Min < c.Manager.SalesTiers[j].Min })
	sort.Slice(c.Staff.SalesTiers, func(i, j int) bool { return c.Staff.SalesTiers[i].Min < c.Staff.SalesTiers[j].Min })
	return c
}

func (c SalaryConfig) Validate() error {
	c = c.Normalize()
	for _, role := range []string{"manager", "staff"} {
		rr := c.RoleRule(role)
		if rr.FixedItems.Total(role) <= 0 {
			return fmt.Errorf("%s 固定薪酬合计必须大于0", role)
		}
		if rr.VerifyRate < 0 || rr.VerifyRate > 1 {
			return errors.New("核销提成率须在 0~1 之间")
		}
		for _, t := range rr.SalesTiers {
			if t.Rate < 0 || t.Rate > 1 {
				return errors.New("开卡提成率须在 0~1 之间")
			}
			if t.Max > 0 && t.Max <= t.Min {
				return errors.New("开卡业绩区间上限须大于下限")
			}
		}
	}
	for _, cap := range c.Extra.CapByHeadcount {
		if cap.CapRatio < 0 || cap.CapRatio > 1 {
			return errors.New("店长业绩上限占比须在 0~1 之间")
		}
	}
	return nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// 兼容旧常量名：FixedPackage 使用默认配置（无 DB 时 / 单测）
func FixedPackage(role string) float64 {
	return DefaultSalaryConfig().FixedTotal(role)
}

var (
	ManagerFixedTotal = DefaultSalaryConfig().FixedTotal("manager")
	StaffFixedTotal   = DefaultSalaryConfig().FixedTotal("staff")
)

type SalaryCalcInput struct {
	Role             string
	SalesAmount      float64
	VerifyAmount     float64
	StoreSalesAmount float64
	WorkingHeadcount int
	Config           *SalaryConfig
}

type SalaryCalcResult struct {
	FixedSalary      float64                `json:"fixed_salary"`
	SalesAmount      float64                `json:"sales_amount"`
	SalesAmountUsed  float64                `json:"sales_amount_used"`
	VerifyAmount     float64                `json:"verify_amount"`
	SalesRate        float64                `json:"sales_rate"`
	VerifyRate       float64                `json:"verify_rate"`
	SalesCommission  float64                `json:"sales_commission"`
	VerifyCommission float64                `json:"verify_commission"`
	Commission       float64                `json:"commission"`
	ManagerPenalty   float64                `json:"manager_penalty"`
	WorkingHeadcount int                    `json:"working_headcount"`
	SalesCapRatio    float64                `json:"sales_cap_ratio,omitempty"`
	StoreSalesAmount float64                `json:"store_sales_amount,omitempty"`
	TotalPayable     float64                `json:"total_payable"`
	RuleNote         string                 `json:"rule_note"`
	Breakdown        map[string]interface{} `json:"breakdown"`
}

func ComputeRoleSalary(in SalaryCalcInput) SalaryCalcResult {
	cfg := DefaultSalaryConfig()
	if in.Config != nil {
		cfg = in.Config.Normalize()
	}
	rr := cfg.RoleRule(in.Role)
	fixed := rr.FixedItems.Total(in.Role)
	salesUsed := in.SalesAmount
	capRatio := 0.0
	penalty := 0.0
	note := cfg.Note

	if in.Role == "manager" {
		capRatio = cfg.ManagerSalesCapRatio(in.WorkingHeadcount)
		if in.StoreSalesAmount > 0 {
			cap := in.StoreSalesAmount * capRatio
			if salesUsed > cap {
				salesUsed = cap
			}
			minNeed := in.StoreSalesAmount * cfg.Extra.MinStoreShare
			if in.SalesAmount < minNeed {
				penalty = round2((minNeed - in.SalesAmount) * cfg.Extra.MinShortfallPenalty)
			}
		}
		note = fmt.Sprintf("店长：固定%.0f + 开卡阶梯提成 + 核销%.0f%%；含业绩上下限", fixed, rr.VerifyRate*100)
	} else {
		note = fmt.Sprintf("店员：固定%.0f + 开卡阶梯提成 + 核销%.0f%%", fixed, rr.VerifyRate*100)
	}

	rate := cfg.SalesCommissionRate(in.Role, salesUsed)
	salesComm := round2(salesUsed * rate)
	verifyComm := round2(in.VerifyAmount * rr.VerifyRate)
	commission := round2(salesComm + verifyComm)
	total := round2(fixed + commission - penalty)

	res := SalaryCalcResult{
		FixedSalary:      round2(fixed),
		SalesAmount:      round2(in.SalesAmount),
		SalesAmountUsed:  round2(salesUsed),
		VerifyAmount:     round2(in.VerifyAmount),
		SalesRate:        rate,
		VerifyRate:       rr.VerifyRate,
		SalesCommission:  salesComm,
		VerifyCommission: verifyComm,
		Commission:       commission,
		ManagerPenalty:   penalty,
		WorkingHeadcount: in.WorkingHeadcount,
		SalesCapRatio:    capRatio,
		StoreSalesAmount: round2(in.StoreSalesAmount),
		TotalPayable:     total,
		RuleNote:         note,
	}
	res.Breakdown = map[string]interface{}{
		"fixed_salary":      res.FixedSalary,
		"sales_amount":      res.SalesAmount,
		"sales_amount_used": res.SalesAmountUsed,
		"verify_amount":     res.VerifyAmount,
		"sales_rate":        res.SalesRate,
		"verify_rate":       res.VerifyRate,
		"sales_commission":  res.SalesCommission,
		"verify_commission": res.VerifyCommission,
		"manager_penalty":   res.ManagerPenalty,
		"working_headcount": res.WorkingHeadcount,
		"sales_cap_ratio":   res.SalesCapRatio,
		"store_sales":       res.StoreSalesAmount,
		"total_payable":     res.TotalPayable,
		"note":              res.RuleNote,
	}
	return res
}

func (c SalaryConfig) ToAPIMap() map[string]interface{} {
	c = c.Normalize()
	b, _ := json.Marshal(c)
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m == nil {
		m = map[string]interface{}{}
	}
	if mg, ok := m["manager"].(map[string]interface{}); ok {
		mg["fixed_total"] = c.FixedTotal("manager")
	}
	if st, ok := m["staff"].(map[string]interface{}); ok {
		st["fixed_total"] = c.FixedTotal("staff")
	}
	return m
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (s *Service) LoadSalaryConfig() (SalaryConfig, error) {
	def := DefaultSalaryConfig()
	var raw sql.NullString
	err := s.Repo.DB.Get(&raw, `SELECT config_json FROM lp_salary_rule ORDER BY id LIMIT 1`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return def, nil
		}
		// 列可能尚未迁移：回退默认
		return def, nil
	}
	if !raw.Valid || raw.String == "" || raw.String == "{}" {
		return def, nil
	}
	var cfg SalaryConfig
	if err := json.Unmarshal([]byte(raw.String), &cfg); err != nil {
		return def, nil
	}
	before := cfg
	cfg = cfg.Normalize()
	// 文案曾被错误编码写成 ???，自动回写正确中文，避免管理端一直显示问号
	if looksCorruptedText(before.Name) || looksCorruptedText(before.Note) {
		_ = s.SaveSalaryConfig(cfg)
	}
	return cfg, nil
}

func (s *Service) SaveSalaryConfig(cfg SalaryConfig) error {
	cfg = cfg.Normalize()
	if err := cfg.Validate(); err != nil {
		return err
	}
	raw := mustJSON(cfg)
	var id int64
	err := s.Repo.DB.Get(&id, `SELECT id FROM lp_salary_rule ORDER BY id LIMIT 1`)
	if err != nil {
		_, err = s.Repo.DB.Exec(`INSERT INTO lp_salary_rule(name, base_ratio, commission_rate, pass_bonus, absent_deduct, full_attendance_bonus, late_deduct, config_json)
			VALUES(?,1,?,0,0,0,0,?)`, cfg.Name, cfg.Staff.VerifyRate, raw)
		if err != nil {
			return err
		}
	} else {
		_, err = s.Repo.DB.Exec(`UPDATE lp_salary_rule SET name=?, config_json=?, commission_rate=? WHERE id=?`,
			cfg.Name, raw, cfg.Staff.VerifyRate, id)
		if err != nil {
			return err
		}
	}
	_, _ = s.Repo.DB.Exec(`UPDATE lp_user SET base_salary=? WHERE role='manager' AND status='active'`, cfg.FixedTotal("manager"))
	_, _ = s.Repo.DB.Exec(`UPDATE lp_user SET base_salary=? WHERE role='staff' AND status='active'`, cfg.FixedTotal("staff"))
	return nil
}

func (s *Service) SalaryRules() map[string]interface{} {
	cfg, _ := s.LoadSalaryConfig()
	return cfg.ToAPIMap()
}

func (s *Service) FixedForRole(role string) float64 {
	cfg, _ := s.LoadSalaryConfig()
	return cfg.FixedTotal(role)
}
