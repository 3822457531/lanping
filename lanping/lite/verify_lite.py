# -*- coding: utf-8 -*-
"""精简版校验：表格宽度、字数、旧数字残留"""
import os
import re

from docx import Document
from docx.oxml.ns import qn

HERE = os.path.dirname(os.path.abspath(__file__))
DOCX = os.path.join(HERE, "兰瓶门店管理系统_建设方案与报价_精简版.docx")

d = Document(DOCX)

# 1) 表格宽度
bad = []
for i, t in enumerate(d.tables):
    grid = t._tbl.find(qn("w:tblGrid"))
    if grid is None:
        continue
    total = 0
    for gc in grid.findall(qn("w:gridCol")):
        w = gc.get(qn("w:w"))
        if w:
            total += int(w)
    cm = total / 567.0
    if cm > 15.85:
        bad.append((i + 1, round(cm, 2)))
print(f"表格 {len(d.tables)} 张，超宽 {len(bad)} 张", bad if bad else "")

# 2) 正文字数
text = "\n".join(p.text for p in d.paragraphs)
tbl_text = "\n".join(c.text for t in d.tables for r in t.rows for c in r.cells)
print(f"段落 {len(d.paragraphs)}，正文 {len(text)} 字符，表格 {len(tbl_text)} 字符")

# 3) 旧数字残留
allt = text + tbl_text
for kw in ["574", "797,100", "1,389", "753,700", "764,900", "551", "22 周", "26.1 人月"]:
    if kw in allt:
        print("!! 残留:", kw)
print("旧数字检查完成")

# 4) 关键数字是否在
for kw in ["50,000", "15,000", "45", "6 周", "8,343.07"]:
    print(f"  含 {kw}: {kw in allt}")

# 5) 页眉
hp = d.sections[0].header.paragraphs[0].text
print("页眉:", hp)
