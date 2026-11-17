-- +goose Up
UPDATE lp_user SET base_salary=3500 WHERE role='manager';
UPDATE lp_user SET base_salary=3000 WHERE role='staff';

UPDATE lp_salary_rule SET
  name='兰瓶美业-开卡/核销提成',
  base_ratio=1,
  commission_rate=0.05,
  pass_bonus=0,
  absent_deduct=0,
  full_attendance_bonus=0,
  late_deduct=0
WHERE id=1;

-- +goose Down
UPDATE lp_user SET base_salary=4600 WHERE role='manager';
UPDATE lp_user SET base_salary=3400 WHERE role='staff';
