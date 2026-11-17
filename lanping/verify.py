# -*- coding: utf-8 -*-
"""结构化自检：表格完整性、宽度溢出、数字一致性、章节完整性"""
import re, sys, glob
from docx import Document
from docx.shared import Cm

PATH = glob.glob(r"D:\workSpace\lanping\兰瓶*.docx")[0]
d = Document(PATH)

problems = []

# ---------- 1. 表格结构检查 ----------
page_w = Cm(15.8)
tbl_w_ok = 0
for ti, t in enumerate(d.tables, 1):
    ncols = len(t.columns)
    # 列宽合计
    try:
        tot = sum((c.width or 0) for c in t.columns)
    except Exception:
        tot = 0
    if tot > page_w * 1.06:
        problems.append(f"表{ti} 列宽合计 {tot/360000:.2f}cm 超出页面 {page_w/360000:.2f}cm")
    else:
        tbl_w_ok += 1
    # 每行列数
    for ri, row in enumerate(t.rows, 1):
        if len(row.cells) != ncols:
            problems.append(f"表{ti} 第{ri}行列数 {len(row.cells)} != {ncols}")
    # 空表（流程图由单行无边框表格模拟，不计入）
    if len(t.rows) < 2 and len(t.columns) > 3 and not t.rows[0].cells[0].text.strip():
        problems.append(f"表{ti} 只有表头无数据")

# ---------- 2. 全文数字一致性 ----------
def all_text(doc):
    out = []
    for p in doc.paragraphs:
        out.append(p.text)
    for t in doc.tables:
        for row in t.rows:
            for c in row.cells:
                out.append(c.text)
    return "\n".join(out)

txt = all_text(d)

checks = {
    "旧总人天 543": r"(?<!\d)543(?!\d)",
    "旧总价 753,700": r"753,700",
    "旧均价 1,341": r"1,341",
    "旧人月 24.7": r"24\.7 人月",
    "旧支撑 142": r"142 人天",
    "旧模块数 26 个": r"26 个功能模块",
    "旧表数 38 张": r"38 张",
    "旧功能域 10 个": r"10 个功能域",
    "旧团队 10.5": r"10\.5 人",
}
for name, pat in checks.items():
    hits = re.findall(pat, txt)
    if hits:
        problems.append(f"残留旧值 [{name}] x{len(hits)}")

# 新值必备
must = ["574 人天", "797,100", "柒拾玖万柒仟壹佰元整", "173 人天", "56 张", "30 个功能模块", "12 个功能域"]
for m in must:
    if m not in txt:
        problems.append(f"缺失新值 [{m}]")

# 付款合计
pay = [239130, 159420, 159420, 119565, 79710, 39855]
if sum(pay) != 797100:
    problems.append(f"付款合计 {sum(pay)} != 797100")
for v in pay:
    f = f"{v:,}"
    if f not in txt:
        problems.append(f"付款金额缺失 [{f}]")

# 分项报价合计
items = [33800, 30800, 337500, 246400, 57200, 77000, 14400]
if sum(items) != 797100:
    problems.append(f"分项合计 {sum(items)} != 797100")

# 人天合计
days = [26, 28, 225, 176, 52, 55, 12]
if sum(days) != 574:
    problems.append(f"人天合计 {sum(days)} != 574")
if 401 + 173 != 574:
    problems.append("开发401+支撑173 != 574")

# 占比
pct = [69.9, 9.4, 9.1, 9.5, 2.1]
if abs(sum(pct) - 100) > 0.15:
    problems.append(f"占比合计 {sum(pct)} != 100")

# ---------- 3. 章节完整性 ----------
heads = [p.text.strip() for p in d.paragraphs]
need = ["第 1 章", "第 2 章", "第 3 章", "第 4 章", "第 5 章", "第 6 章", "第 7 章",
        "第 8 章", "第 9 章", "第 10 章", "第 11 章", "第 12 章", "第 13 章",
        "第 14 章", "第 15 章", "第 16 章", "附　录"]
for n in need:
    if not any(n in h for h in heads):
        problems.append(f"缺章节标题 [{n}]")

# 空段落 / None 文本
for p in d.paragraphs:
    if "None" in p.text:
        problems.append(f"出现 None 文本: {p.text[:40]}")

# ---------- 4. 输出 ----------
nchars = len(txt.replace("\n", ""))
npara = len(d.paragraphs)
print(f"文件: {PATH}")
print(f"段落 {npara} / 表格 {len(d.tables)} / 正文字符 {nchars}")
print(f"表格宽度合规 {tbl_w_ok}/{len(d.tables)}")
print(f"问题数: {len(problems)}")
for p in problems:
    print("  ! " + p)
if not problems:
    print("PASS 全部检查通过")

# 导出纯文本供人工通读
with open(r"D:\workSpace\lanping\_dump.txt", "w", encoding="utf-8") as f:
    f.write(txt)
print("dump -> _dump.txt")
