package service

import "testing"

func TestSalesCommissionRate(t *testing.T) {
	cfg := DefaultSalaryConfig()
	cases := []struct {
		amt  float64
		rate float64
	}{
		{0, 0.02},
		{9999.99, 0.02},
		{10000, 0.05},
		{14999, 0.05},
		{15000, 0.08},
		{20000, 0.11},
		{25000, 0.12},
		{30000, 0.13},
		{88000, 0.13},
	}
	for _, c := range cases {
		got := cfg.SalesCommissionRate("staff", c.amt)
		if got != c.rate {
			t.Fatalf("sales=%.2f want rate=%.2f got=%.2f", c.amt, c.rate, got)
		}
	}
}

func TestStaffSalary(t *testing.T) {
	cfg := DefaultSalaryConfig()
	r := ComputeRoleSalary(SalaryCalcInput{
		Role: "staff", SalesAmount: 12000, VerifyAmount: 8000, Config: &cfg,
	})
	if r.FixedSalary != 3000 {
		t.Fatalf("fixed=%v", r.FixedSalary)
	}
	if r.SalesCommission != 600 {
		t.Fatalf("salesComm=%v", r.SalesCommission)
	}
	if r.VerifyCommission != 400 {
		t.Fatalf("verifyComm=%v", r.VerifyCommission)
	}
	if r.TotalPayable != 4000 {
		t.Fatalf("total=%v", r.TotalPayable)
	}
}

func TestManagerCapAndPenalty(t *testing.T) {
	cfg := DefaultSalaryConfig()
	store := 100000.0
	r := ComputeRoleSalary(SalaryCalcInput{
		Role: "manager", SalesAmount: 5000, VerifyAmount: 2000,
		StoreSalesAmount: store, WorkingHeadcount: 4, Config: &cfg,
	})
	if r.SalesCapRatio != 0.30 {
		t.Fatalf("cap=%v", r.SalesCapRatio)
	}
	if r.ManagerPenalty != 400 {
		t.Fatalf("penalty=%v", r.ManagerPenalty)
	}
	if r.SalesCommission != 100 {
		t.Fatalf("salesComm=%v", r.SalesCommission)
	}
	if r.VerifyCommission != 100 {
		t.Fatalf("verifyComm=%v", r.VerifyCommission)
	}
	if r.TotalPayable != 3300 {
		t.Fatalf("total=%v", r.TotalPayable)
	}

	r2 := ComputeRoleSalary(SalaryCalcInput{
		Role: "manager", SalesAmount: 90000, VerifyAmount: 0,
		StoreSalesAmount: store, WorkingHeadcount: 2, Config: &cfg,
	})
	if r2.SalesAmountUsed != 60000 {
		t.Fatalf("used=%v", r2.SalesAmountUsed)
	}
	if r2.SalesRate != 0.13 {
		t.Fatalf("rate=%v", r2.SalesRate)
	}
}

func TestConfigurableRates(t *testing.T) {
	cfg := DefaultSalaryConfig()
	cfg.Staff.VerifyRate = 0.08
	cfg.Staff.FixedItems.BaseSalary = 1500
	r := ComputeRoleSalary(SalaryCalcInput{
		Role: "staff", SalesAmount: 5000, VerifyAmount: 1000, Config: &cfg,
	})
	if r.FixedSalary != 3300 { // 1500+700+1100
		t.Fatalf("fixed=%v", r.FixedSalary)
	}
	if r.VerifyCommission != 80 {
		t.Fatalf("verifyComm=%v", r.VerifyCommission)
	}
}
