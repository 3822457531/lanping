-- +goose Up
INSERT INTO lp_store (id, name, address, phone, manager_name, status) VALUES
(1, '兰瓶总店', '怀化市鹤城区人民路 88 号', '0745-2288001', '张美丽', '营业中'),
(2, '万象城店', '怀化市鹤城区万象城 3F', '0745-2288002', '李静', '营业中'),
(3, '天河店', '怀化市鹤城区天河北路 12 号', '0745-2288003', '王芳', '营业中');

-- password for all demo accounts: 123456 (bcrypt)
INSERT INTO lp_user (id, name, phone, password_hash, role, role_name, store_id, position, hire_date, status, level_code, base_salary) VALUES
(1, '张美丽', '13800000001', '$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.', 'admin', '超级管理员', 1, '总经理', '2021-03-01', 'active', 'P7', 5200),
(2, '李静', '13800000002', '$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.', 'manager', '店长', 2, '店长', '2022-05-12', 'active', 'P5', 3500),
(3, '王芳', '13800000003', '$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.', 'manager', '店长', 3, '店长', '2022-08-20', 'active', 'P5', 3500),
(4, '陈晓', '13800000004', '$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.', 'staff', '美业员工', 1, '高级美容师', '2023-02-06', 'active', 'P3', 3000),
(5, '刘婷', '13800000005', '$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.', 'staff', '美业员工', 1, '美容师', '2023-06-15', 'active', 'P2', 3000),
(6, '赵敏', '13800000006', '$2a$10$juEexCLUMeMRwMcx97U33uQ7Ic.Mszpnkm4t.KcoiB2xJmNUU1jL.', 'staff', '美业员工', 2, '美容师', '2024-01-08', 'active', 'P2', 3000);

INSERT INTO lp_project (id, name, category, price, member_price, total_times, valid_days, commission_rate, duration, status) VALUES
(1, '小气泡深层清洁', '面部清洁', 398, 358, 10, 365, 0.02, 60, '上架'),
(2, '水光补水护理', '面部护理', 698, 628, 10, 365, 0.02, 75, '上架'),
(3, '光子嫩肤', '光电项目', 1980, 1780, 5, 365, 0.03, 45, '上架'),
(4, '肩颈舒缓', '身体护理', 298, 268, 10, 365, 0.02, 45, '上架'),
(5, '背部精油SPA', '身体护理', 458, 418, 8, 365, 0.02, 60, '上架'),
(6, '热玛吉紧致', '光电项目', 6800, 6200, 3, 540, 0.04, 90, '上架');

INSERT INTO lp_customer (id, customer_no, name, mobile, gender, birthday, level_name, source, store_id, employee_id, total_amount, last_visit_at, status) VALUES
(1, 'LP2025000001', '王小美', '13900000001', 2, '1995-04-28', '钻石会员', '到店', 1, 4, 12680, NOW() - INTERVAL 1 DAY, 1),
(2, 'LP2025000002', '李婷婷', '13900000002', 2, '1993-04-29', '金卡会员', '转介绍', 1, 5, 6420, NOW() - INTERVAL 6 DAY, 1),
(3, 'LP2025000003', '陈静', '13900000003', 2, '1998-05-01', '银卡会员', '线上', 2, 6, 2180, NOW() - INTERVAL 39 DAY, 1),
(4, 'LP2025000004', '周雪', '13900000004', 2, '1992-07-12', '钻石会员', '到店', 2, 6, 18960, NOW() - INTERVAL 2 DAY, 1),
(5, 'LP2025000005', '赵梦琪', '13900000005', 2, '1996-03-15', '金卡会员', '活动', 3, 3, 8600, NOW() - INTERVAL 12 DAY, 1),
(6, 'LP2025000006', '孙悦', '13900000006', 2, '1994-09-08', '银卡会员', '到店', 1, 4, 3200, NOW() - INTERVAL 5 DAY, 1);

INSERT INTO lp_customer_card (id, card_no, customer_id, project_id, store_id, employee_id, total_times, used_times, remain_times, frozen_times, total_amount, pay_method, status, version, valid_from, valid_to) VALUES
(1, 'CD20250001', 1, 1, 1, 4, 10, 2, 8, 0, 3980, '微信', '使用中', 0, CURDATE() - INTERVAL 30 DAY, CURDATE() + INTERVAL 335 DAY),
(2, 'CD20250002', 2, 2, 1, 5, 10, 3, 7, 0, 6980, '刷卡', '使用中', 0, CURDATE() - INTERVAL 40 DAY, CURDATE() + INTERVAL 325 DAY),
(3, 'CD20250003', 3, 3, 2, 6, 5, 1, 4, 0, 9900, '微信', '使用中', 0, CURDATE() - INTERVAL 20 DAY, CURDATE() + INTERVAL 345 DAY),
(4, 'CD20250004', 4, 2, 2, 6, 10, 5, 5, 0, 6980, '微信', '使用中', 0, CURDATE() - INTERVAL 60 DAY, CURDATE() + INTERVAL 305 DAY),
(5, 'CD20250005', 5, 4, 3, 3, 10, 2, 8, 0, 2980, '刷卡', '使用中', 0, CURDATE() - INTERVAL 15 DAY, CURDATE() + INTERVAL 350 DAY),
(6, 'CD20250006', 6, 1, 1, 4, 10, 0, 10, 0, 3980, '微信', '使用中', 0, CURDATE(), CURDATE() + INTERVAL 365 DAY);

INSERT INTO lp_salary_rule (id, name, base_ratio, commission_rate, pass_bonus, absent_deduct, full_attendance_bonus, late_deduct) VALUES
(1, '兰瓶美业-开卡/核销提成', 1.0000, 0.0500, 0, 0, 0, 0);

INSERT INTO lp_welfare (id, name, amount, description, status) VALUES
(1, '生日礼金', 200, '员工生日当月发放', '启用'),
(2, '全勤奖', 200, '当月无迟到早退缺勤', '启用'),
(3, '节日福利', 500, '春节/中秋统一发放', '启用');

INSERT INTO lp_settings (config_key, config_value) VALUES
('brand_name', '兰瓶美业'),
('birthday_remind', 'true'),
('birthday_days', '7'),
('sms_enabled', 'true'),
('sms_sign', '兰瓶美业'),
('clockin_wifi', 'LanPing-Office'),
('clockin_range', '200'),
('clockin_face', 'true'),
('verify_mode', 'mock'),
('customer_share', 'false');

INSERT INTO lp_stock_item (id, name, sku, unit) VALUES
(1, '玻尿酸补水面膜', 'M-001', '片'),
(2, '水杨酸精华', 'S-014', '瓶'),
(3, '舒缓修护乳', 'C-007', '瓶');

INSERT INTO lp_stock (store_id, item_id, qty) VALUES
(1, 1, 120), (1, 2, 40), (2, 1, 80), (2, 2, 60), (3, 3, 40);

INSERT INTO lp_assessment (user_id, store_id, month_key, name, weight, target_value, done_value, passed) VALUES
(4, 1, DATE_FORMAT(CURDATE(), '%Y-%m'), '月度业绩（元）', 1, 8000, 9200, 1),
(5, 1, DATE_FORMAT(CURDATE(), '%Y-%m'), '月度业绩（元）', 1, 8000, 6500, 0),
(6, 2, DATE_FORMAT(CURDATE(), '%Y-%m'), '月度业绩（元）', 1, 8000, 8800, 1),
(2, 2, DATE_FORMAT(CURDATE(), '%Y-%m'), '月度业绩（元）', 1, 20000, 21000, 1);

INSERT INTO lp_message (user_id, title, content, is_read) VALUES
(4, '欢迎使用兰瓶门店端', '今日待办与核销入口已就绪', 0),
(2, '库存提醒', '万象城店水杨酸精华库存偏低', 0);

INSERT INTO lp_audit_apply (name, phone, store_id, role, status) VALUES
('何佳', '13800009901', 2, 'staff', '待审核'),
('林晓', '13800009902', 1, 'staff', '待审核');

-- +goose Down
DELETE FROM lp_audit_apply;
DELETE FROM lp_message;
DELETE FROM lp_assessment;
DELETE FROM lp_stock;
DELETE FROM lp_stock_item;
DELETE FROM lp_settings;
DELETE FROM lp_welfare;
DELETE FROM lp_salary_rule;
DELETE FROM lp_customer_card;
DELETE FROM lp_customer;
DELETE FROM lp_project;
DELETE FROM lp_user;
DELETE FROM lp_store;
