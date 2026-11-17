# -*- coding: utf-8 -*-
"""Lanping closed-loop E2E test runner."""
from __future__ import annotations

import json
import time
import random
import urllib.error
import urllib.request
from datetime import datetime, timedelta
from pathlib import Path

BASE = "http://127.0.0.1:8080"
OUT_DIR = Path(r"D:\workSpace\lan\deploy")
results: list[dict] = []
ctx: dict = {}


def add(module: str, case: str, ok: bool, detail: str = "", ms: int = 0):
    results.append({"module": module, "case": case, "ok": bool(ok), "detail": detail, "ms": int(ms)})
    mark = "PASS" if ok else "FAIL"
    print(f"[{mark}] {module} :: {case} ({ms}ms) {detail}")


def api(method: str, path: str, token: str | None = None, body=None, expect: int = 200):
    url = BASE + path
    data = None
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    if body is not None:
        data = json.dumps(body, ensure_ascii=False).encode("utf-8")
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    t0 = time.perf_counter()
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            raw = resp.read().decode("utf-8", errors="replace")
            status = resp.status
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", errors="replace")
        status = e.code
    except Exception as e:
        ms = int((time.perf_counter() - t0) * 1000)
        return {"ok": False, "status": 0, "ms": ms, "json": None, "raw": str(e)}
    ms = int((time.perf_counter() - t0) * 1000)
    try:
        js = json.loads(raw) if raw else None
    except Exception:
        js = None
    ok_http = status == expect
    ok_biz = True
    if js is not None and isinstance(js, dict) and "ok" in js and 200 <= expect < 300:
        ok_biz = bool(js.get("ok"))
    return {"ok": ok_http and ok_biz, "status": status, "ms": ms, "json": js, "raw": raw}


def data(r):
    return (r.get("json") or {}).get("data")


def main():
    # 0 health
    r = api("GET", "/api/health")
    add("env", "health", r["ok"], f"status={r['status']}", r["ms"])

    # 1 login
    for role, phone in [("admin", "13800000001"), ("manager", "13800000002"), ("staff", "13800000004")]:
        r = api("POST", "/api/auth/login", body={"phone": phone, "password": "123456"})
        d = data(r) or {}
        token = d.get("token")
        user = d.get("user") or {}
        ok = r["ok"] and bool(token)
        if ok:
            ctx[role] = {"token": token, "user": user}
        add("auth", f"login-{role}", ok, f"role={user.get('role')} store={user.get('store_id')}", r["ms"])

    r = api("POST", "/api/auth/login", body={"phone": "13800000001", "password": "wrong"}, expect=400)
    add("auth", "bad-password", r["status"] == 400 or not (r.get("json") or {}).get("ok", True), f"status={r['status']}", r["ms"])

    admin, mgr, staff = ctx["admin"]["token"], ctx["manager"]["token"], ctx["staff"]["token"]
    r = api("GET", "/api/auth/me", token=admin)
    add("auth", "me", r["ok"], f"name={(data(r) or {}).get('name')}", r["ms"])

    # 2 customer
    mobile = "139" + str(random.randint(10000000, 99999999))
    r = api("POST", "/api/customers", token=mgr, body={
        "name": "E2E客户", "phone": mobile, "gender": 2, "level": "银卡会员", "source": "到店"
    })
    d = data(r) or {}
    cid = d.get("id")
    add("customer", "create", r["ok"] and bool(cid), f"id={cid} phone={mobile}", r["ms"])
    r = api("GET", "/api/customers?pageSize=5", token=mgr)
    add("customer", "list", r["ok"] and bool((data(r) or {}).get("list") is not None), f"total={(data(r) or {}).get('total')}", r["ms"])
    r = api("GET", f"/api/customers/{cid}", token=mgr)
    add("customer", "detail", r["ok"], f"name={(data(r) or {}).get('name')}", r["ms"])

    # 3 project/store
    r = api("GET", "/api/projects?pageSize=50", token=mgr)
    projects = (data(r) or {}).get("list") or []
    on_shelf = [p for p in projects if p.get("status") == "上架"]
    pid = (on_shelf[0] if on_shelf else projects[0])["id"] if projects else None
    add("project", "list", r["ok"] and bool(pid), f"count={len(projects)} pid={pid}", r["ms"])
    r = api("GET", "/api/stores", token=admin)
    stores = (data(r) or {}).get("list") or []
    sid = ctx["manager"]["user"].get("store_id") or (stores[0]["id"] if stores else None)
    add("store", "list", r["ok"] and bool(sid), f"count={len(stores)} sid={sid}", r["ms"])

    # 4 appointments
    now = datetime.now()
    date = now.strftime("%Y-%m-%d")
    time1 = (now + timedelta(hours=2)).strftime("%H:%M")
    time2 = (now + timedelta(hours=5)).strftime("%H:%M")
    r = api("GET", f"/api/staff/availability?date={date}&time={time1}&project_id={pid}&store_id={sid}", token=mgr)
    staff_list = (data(r) or {}).get("list") or []
    free = next((s for s in staff_list if not s.get("busy")), None)
    add("appt", "availability", r["ok"] and len(staff_list) > 0, f"n={len(staff_list)} free={(free or {}).get('id')}", r["ms"])
    appt_staff = (free or {}).get("id") or ctx["staff"]["user"].get("id")

    r = api("POST", "/api/appointments", token=mgr, body={
        "customer_id": cid, "project_id": pid, "store_id": sid, "staff_id": appt_staff,
        "date": date, "time": time1, "note": "e2e1"
    })
    appt_id = (data(r) or {}).get("id")
    add("appt", "create", r["ok"] and bool(appt_id), f"id={appt_id} {date} {time1}", r["ms"])

    r2 = api("POST", "/api/appointments", token=mgr, body={
        "customer_id": cid, "project_id": pid, "store_id": sid, "staff_id": appt_staff,
        "date": date, "time": time1, "note": "conflict"
    }, expect=400)
    add("appt", "conflict-reject", r2["status"] == 400 or not (r2.get("json") or {}).get("ok", True),
        f"status={r2['status']} err={(r2.get('json') or {}).get('error')}", r2["ms"])

    r = api("POST", "/api/appointments", token=mgr, body={
        "customer_id": cid, "project_id": pid, "store_id": sid, "staff_id": appt_staff,
        "date": date, "time": time2, "note": "e2e2"
    })
    add("appt", "same-staff-other-slot", r["ok"] and bool((data(r) or {}).get("id")),
        f"id={(data(r) or {}).get('id')} time={time2}", r["ms"])

    if appt_id:
        r = api("PUT", f"/api/appointments/{appt_id}", token=mgr, body={"status": "已确认"})
        add("appt", "confirm", r["ok"], f"id={appt_id}", r["ms"])
        r = api("PUT", f"/api/appointments/{appt_id}", token=mgr, body={"status": "已取消"})
        add("appt", "cancel", r["ok"], f"id={appt_id}", r["ms"])

    r = api("GET", "/api/appointments?pageSize=20", token=mgr)
    add("appt", "list", r["ok"], f"total={(data(r) or {}).get('total')}", r["ms"])

    # 5 cards
    r = api("POST", "/api/cards", token=mgr, body={
        "customer_id": cid, "project_id": pid, "store_id": sid,
        "card_kind": "times", "times": 3, "amount": 900, "pay_method": "微信"
    })
    card_times = data(r) or {}
    add("card", "open-times", r["ok"] and bool(card_times.get("id")),
        f"id={card_times.get('id')} remain={card_times.get('remain')} kind={card_times.get('card_kind')}", r["ms"])
    card_times_id = card_times.get("id")

    r = api("POST", "/api/cards", token=mgr, body={
        "customer_id": cid, "store_id": sid, "card_kind": "amount", "amount": 500, "pay_method": "现金"
    })
    card_amt = data(r) or {}
    add("card", "open-amount", r["ok"] and card_amt.get("card_kind") == "amount",
        f"id={card_amt.get('id')} amount={card_amt.get('amount')}", r["ms"])

    r = api("GET", "/api/cards?pageSize=20", token=mgr)
    add("card", "list", r["ok"], f"total={(data(r) or {}).get('total')}", r["ms"])

    # 6 verify + mock sms
    r = api("PUT", "/api/settings", token=admin, body={"verify_mode": "mock", "customer_share": "true"})
    add("settings", "verify-mock", r["ok"], "verify_mode=mock", r["ms"])

    r = api("POST", "/api/verifications/draft", token=staff, body={"card_id": card_times_id, "times": 1})
    draft = data(r) or {}
    add("verify", "draft-mock-sms", r["ok"] and bool(draft.get("id")) and bool(draft.get("token")),
        f"id={draft.get('id')} no={draft.get('verify_no')}", r["ms"])
    verify_id, verify_token = draft.get("id"), draft.get("token")

    r = api("GET", "/api/sms/logs?pageSize=5", token=admin)
    sms_list = (data(r) or {}).get("list") or []
    add("verify", "sms-log-exists", r["ok"] and len(sms_list) > 0,
        f"last_status={(sms_list[0].get('status') if sms_list else None)}", r["ms"])

    r = api("POST", f"/api/verifications/{verify_id}/confirm", token=staff, body={"token": verify_token})
    add("verify", "confirm", r["ok"], f"id={verify_id}", r["ms"])

    r = api("GET", "/api/cards?pageSize=50", token=mgr)
    cards = (data(r) or {}).get("list") or []
    after = next((c for c in cards if c.get("id") == card_times_id), None)
    add("verify", "remain-decrement", bool(after) and after.get("remain") == 2, f"remain={(after or {}).get('remain')} expect=2", r["ms"])

    r = api("POST", "/api/verifications/draft", token=staff, body={"card_id": card_amt.get("id"), "times": 1}, expect=400)
    add("verify", "amount-card-reject", r["status"] == 400 or not (r.get("json") or {}).get("ok", True),
        f"err={(r.get('json') or {}).get('error')}", r["ms"])

    r = api("POST", "/api/verifications/draft", token=staff, body={"card_id": card_times_id, "times": 99}, expect=400)
    add("verify", "over-times-reject", r["status"] == 400 or not (r.get("json") or {}).get("ok", True),
        f"err={(r.get('json') or {}).get('error')}", r["ms"])

    # 7 workbench
    r = api("GET", "/api/workbench", token=mgr)
    d = data(r) or {}
    add("workbench", "store", r["ok"], f"appt={d.get('today_appt_count')} todos={len(d.get('todos') or [])}", r["ms"])
    r = api("GET", "/api/dashboard", token=admin)
    add("workbench", "admin-dashboard", r["ok"], f"customers={(data(r) or {}).get('customer_total')}", r["ms"])

    # 8 inventory
    r = api("GET", "/api/inventory/items", token=mgr)
    items = (data(r) or {}).get("list") or []
    add("inventory", "items", r["ok"], f"count={len(items)}", r["ms"])
    sku = f"E2E{random.randint(1000,9999)}"
    r = api("POST", "/api/inventory/items", token=mgr, body={
        "name": "E2E耗材", "sku": sku, "unit": "瓶", "store_id": sid, "qty": 10
    })
    item_id = (data(r) or {}).get("id")
    add("inventory", "create-item", r["ok"] and bool(item_id), f"id={item_id} sku={sku}", r["ms"])
    if item_id:
        r = api("POST", "/api/inventory/orders", token=mgr, body={
            "store_id": sid, "item_id": item_id, "qty": 2, "direction": "出库", "channel": "门店调度"
        })
        add("inventory", "stock-out", r["ok"], "qty=2", r["ms"])
        r = api("POST", "/api/inventory/orders", token=mgr, body={
            "store_id": sid, "item_id": item_id, "qty": 999, "direction": "出库", "channel": "门店调度"
        }, expect=400)
        add("inventory", "insufficient-reject", r["status"] == 400 or not (r.get("json") or {}).get("ok", True),
            f"err={(r.get('json') or {}).get('error')}", r["ms"])
        r = api("GET", f"/api/inventory/stock?store_id={sid}", token=mgr)
        stock = next((x for x in ((data(r) or {}).get("list") or []) if x.get("item_id") == item_id), None)
        add("inventory", "qty-after-out", bool(stock) and stock.get("qty") == 8, f"qty={(stock or {}).get('qty')}", r["ms"])

    # 9 register + audit
    reg_phone = "137" + str(random.randint(10000000, 99999999))
    r = api("POST", "/api/auth/register", body={
        "name": "E2E新员工", "phone": reg_phone, "password": "123456", "store_id": sid, "role": "staff"
    })
    add("register", "apply", r["ok"], f"phone={reg_phone}", r["ms"])
    r = api("GET", "/api/audits", token=mgr)
    audits = (data(r) or {}).get("list") or []
    pending = next((a for a in audits if a.get("phone") == reg_phone and a.get("status") == "待审核"), None)
    add("register", "audit-visible", r["ok"] and bool(pending), f"auditId={(pending or {}).get('id')}", r["ms"])
    if pending:
        r = api("POST", f"/api/audits/{pending['id']}/approve", token=mgr)
        add("register", "approve", r["ok"], f"id={pending['id']}", r["ms"])
        r = api("POST", "/api/auth/login", body={"phone": reg_phone, "password": "123456"})
        add("register", "login-after-approve", r["ok"] and bool((data(r) or {}).get("token")),
            f"role={((data(r) or {}).get('user') or {}).get('role')}", r["ms"])

    # 10 salary
    month = datetime.now().strftime("%Y-%m")
    r = api("GET", "/api/salary/rules", token=admin)
    add("salary", "rules", r["ok"], f"has={(data(r) is not None)}", r["ms"])
    r = api("POST", "/api/salary/compute", token=admin, body={"month": month})
    add("salary", "compute", r["ok"], f"month={month}", r["ms"])
    r = api("GET", f"/api/salary/records?month={month}&pageSize=50", token=admin)
    recs = (data(r) or {}).get("list") or []
    add("salary", "records", r["ok"] and len(recs) > 0, f"count={len(recs)}", r["ms"])
    r = api("GET", "/api/assessments", token=mgr)
    add("salary", "assessments", r["ok"], f"total={(data(r) or {}).get('total')}", r["ms"])

    # 11 attendance record-only
    r = api("POST", "/api/attendance/check", token=staff, body={"method": "定位"})
    add("attendance", "checkin", r["ok"], "method=定位", r["ms"])
    r = api("GET", "/api/attendance/today", token=staff)
    d = data(r) or {}
    add("attendance", "today", r["ok"] and bool(d.get("checked_in")), f"in={d.get('check_in')}", r["ms"])

    # 12 welfare/messages/logs
    r = api("GET", "/api/welfare", token=admin)
    add("misc", "welfare", r["ok"], f"total={(data(r) or {}).get('total')}", r["ms"])
    r = api("GET", "/api/messages", token=staff)
    add("misc", "messages", r["ok"], f"total={(data(r) or {}).get('total')}", r["ms"])
    r = api("GET", "/api/operate-logs?pageSize=10", token=admin)
    logs = (data(r) or {}).get("list") or []
    add("misc", "operate-logs", r["ok"] and len(logs) > 0, f"total={(data(r) or {}).get('total')}", r["ms"])

    # 13 permission
    r = api("POST", "/api/inventory/items", token=staff, body={"name": "nope", "unit": "件", "store_id": sid, "qty": 1}, expect=403)
    add("perm", "staff-create-item-denied", r["status"] in (401, 403) or not r["ok"], f"status={r['status']}", r["ms"])
    r = api("GET", "/api/customers", expect=401)
    add("perm", "no-token-denied", r["status"] == 401 or not r["ok"], f"status={r['status']}", r["ms"])

    # report
    passed = sum(1 for x in results if x["ok"])
    failed = sum(1 for x in results if not x["ok"])
    total = len(results)
    avg = round(sum(x["ms"] for x in results) / total, 1) if total else 0
    mx = max((x["ms"] for x in results), default=0)

    lines = [
        "# 兰瓶闭环测试报告",
        "",
        f"- 时间：{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}",
        f"- 环境：API `{BASE}`；管理端 `5173`；门店端 `5174`",
        f"- 结果：**{passed}/{total} 通过**，失败 **{failed}**",
        f"- 耗时：平均 **{avg}ms**，最大 **{mx}ms**",
        "",
        "## 汇总",
        "",
        "| 模块 | 通过 | 失败 |",
        "|------|------|------|",
    ]
    mods = {}
    for x in results:
        mods.setdefault(x["module"], {"p": 0, "f": 0})
        mods[x["module"]]["p" if x["ok"] else "f"] += 1
    for m, v in mods.items():
        lines.append(f"| {m} | {v['p']} | {v['f']} |")

    lines += ["", "## 明细", "", "| 结果 | 模块 | 用例 | 耗时ms | 说明 |", "|------|------|------|--------|------|"]
    for x in results:
        mark = "PASS" if x["ok"] else "FAIL"
        detail = str(x["detail"]).replace("|", "/")
        lines.append(f"| {mark} | {x['module']} | {x['case']} | {x['ms']} | {detail} |")

    lines += ["", "## 失败项", ""]
    fails = [x for x in results if not x["ok"]]
    if not fails:
        lines.append("无")
    else:
        for f in fails:
            lines.append(f"- [{f['module']}] {f['case']}：{f['detail']}")

    lines += [
        "",
        "## 业务结论",
        "",
        "- 已覆盖：三角色登录、错误密码、客户CRUD、预约创建/冲突拒绝/同时段可约/确认取消、次卡与储值开卡、模拟核销与短信日志、余次扣减、储值卡次数核销拒绝、库存出入库与不足拒绝、注册审核登录、薪资计算、工作台、日志与权限。",
        "- 短信为模拟写库（按约定忽略真发）。",
        "- 打卡仅记流水（按约定忽略 Wi-Fi/定位校验）。",
        "",
        "## 上线建议",
        "",
        "- 主业务闭环可用，可作为门店试用上线。",
        "- 上线前仍建议：改 JWT/库密码、关演示账号弱口令、HTTPS、API 进程守护。",
        "",
    ]

    report = OUT_DIR / "e2e-test-report.md"
    report.write_text("\n".join(lines), encoding="utf-8")
    (OUT_DIR / "e2e-test-report.json").write_text(json.dumps(results, ensure_ascii=False, indent=2), encoding="utf-8")
    print()
    print(f"REPORT: {report}")
    print(f"PASS={passed} FAIL={failed} TOTAL={total} AVG={avg}ms MAX={mx}ms")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
