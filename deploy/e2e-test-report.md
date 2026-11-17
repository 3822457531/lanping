# 兰瓶闭环测试报告

- 时间：2026-11-14 14:10:45
- 环境：API `http://127.0.0.1:8080`；管理端 `5173`；门店端 `5174`
- 结果：**50/50 通过**，失败 **0**
- 耗时：平均 **116.2ms**，最大 **554ms**

## 汇总

| 模块 | 通过 | 失败 |
|------|------|------|
| env | 1 | 0 |
| auth | 5 | 0 |
| customer | 3 | 0 |
| project | 1 | 0 |
| store | 1 | 0 |
| appt | 7 | 0 |
| card | 3 | 0 |
| settings | 1 | 0 |
| verify | 6 | 0 |
| workbench | 2 | 0 |
| inventory | 5 | 0 |
| register | 4 | 0 |
| salary | 4 | 0 |
| attendance | 2 | 0 |
| misc | 3 | 0 |
| perm | 2 | 0 |

## 明细

| 结果 | 模块 | 用例 | 耗时ms | 说明 |
|------|------|------|--------|------|
| PASS | env | health | 64 | status=200 |
| PASS | auth | login-admin | 247 | role=admin store=1 |
| PASS | auth | login-manager | 297 | role=manager store=2 |
| PASS | auth | login-staff | 150 | role=staff store=1 |
| PASS | auth | bad-password | 82 | status=401 |
| PASS | auth | me | 3 | name=张美丽 |
| PASS | customer | create | 363 | id=17 phone=13994505563 |
| PASS | customer | list | 175 | total=9 |
| PASS | customer | detail | 2 | name=E2E客户 |
| PASS | project | list | 2 | count=6 pid=6 |
| PASS | store | list | 1 | count=4 sid=2 |
| PASS | appt | availability | 236 | n=3 free=2 |
| PASS | appt | create | 80 | id=6 2026-11-14 16:10 |
| PASS | appt | conflict-reject | 4 | status=400 err=该美容师此时段忙碌，请选择其他时间或美容师 |
| PASS | appt | same-staff-other-slot | 103 | id=7 time=19:10 |
| PASS | appt | confirm | 316 | id=6 |
| PASS | appt | cancel | 84 | id=6 |
| PASS | appt | list | 7 | total=3 |
| PASS | card | open-times | 390 | id=12 remain=3 kind=times |
| PASS | card | open-amount | 323 | id=13 amount=500 |
| PASS | card | list | 3 | total=11 |
| PASS | settings | verify-mock | 175 | verify_mode=mock |
| PASS | verify | draft-mock-sms | 340 | id=3 no=HX202611146643 |
| PASS | verify | sms-log-exists | 2 | last_status=模拟未发 |
| PASS | verify | confirm | 147 | id=3 |
| PASS | verify | remain-decrement | 3 | remain=2 expect=2 |
| PASS | verify | amount-card-reject | 2 | err=储值卡请走金额核销，暂不支持次数核销 |
| PASS | verify | over-times-reject | 62 | err=剩余次数不足 |
| PASS | workbench | store | 39 | appt=1 todos=2 |
| PASS | workbench | admin-dashboard | 8 | customers=9 |
| PASS | inventory | items | 53 | count=6 |
| PASS | inventory | create-item | 218 | id=7 sku=E2E9988 |
| PASS | inventory | stock-out | 306 | qty=2 |
| PASS | inventory | insufficient-reject | 64 | err=库存不足 |
| PASS | inventory | qty-after-out | 3 | qty=8 |
| PASS | register | apply | 91 | phone=13763036395 |
| PASS | register | audit-visible | 54 | auditId=5 |
| PASS | register | approve | 231 | id=5 |
| PASS | register | login-after-approve | 137 | role=staff |
| PASS | salary | rules | 54 | has=True |
| PASS | salary | compute | 554 | month=2026-11 |
| PASS | salary | records | 53 | count=8 |
| PASS | salary | assessments | 29 | total=2 |
| PASS | attendance | checkin | 109 | method=定位 |
| PASS | attendance | today | 2 | in=14:10:45 |
| PASS | misc | welfare | 96 | total=5 |
| PASS | misc | messages | 39 | total=1 |
| PASS | misc | operate-logs | 5 | total=64 |
| PASS | perm | staff-create-item-denied | 1 | status=403 |
| PASS | perm | no-token-denied | 1 | status=401 |

## 失败项

无

## 业务结论

- 已覆盖：三角色登录、错误密码、客户CRUD、预约创建/冲突拒绝/同时段可约/确认取消、次卡与储值开卡、模拟核销与短信日志、余次扣减、储值卡次数核销拒绝、库存出入库与不足拒绝、注册审核登录、薪资计算、工作台、日志与权限。
- 短信为模拟写库（按约定忽略真发）。
- 打卡仅记流水（按约定忽略 Wi-Fi/定位校验）。

## 上线建议

- 主业务闭环可用，可作为门店试用上线。
- 上线前仍建议：改 JWT/库密码、关演示账号弱口令、HTTPS、API 进程守护。
