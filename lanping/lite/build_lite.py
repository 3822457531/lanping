# -*- coding: utf-8 -*-
"""兰瓶门店管理系统 · 精简版（20 页以内）构建入口"""
import os
import sys
import zipfile

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(HERE)
sys.path.insert(0, ROOT)
sys.path.insert(0, HERE)

from docx.shared import Pt
from docx_kit import Doc, _set_run_font, C_GREY, FONT_CN
import part_a
import part_b

OUT = os.path.join(HERE, "兰瓶门店管理系统_建设方案与报价_精简版.docx")


def relax_spacing(d: Doc):
    """精简版正文更疏朗：行距 1.55、段后 6pt"""
    st = d.d.styles["Normal"]
    st.paragraph_format.line_spacing = 1.55
    st.paragraph_format.space_after = Pt(6)


def fix_header(d: Doc):
    hp = d.d.sections[0].header.paragraphs[0]
    for r in list(hp.runs):
        r._r.getparent().remove(r._r)
    r = hp.add_run("兰瓶美业 · 门店管理系统建设方案")
    _set_run_font(r, FONT_CN, 8.5, False, C_GREY)


def post_process(path):
    tmp = path + ".tmp"
    with zipfile.ZipFile(path, "r") as zin:
        items = zin.infolist()
        data = {i.filename: zin.read(i.filename) for i in items}
    s = data["word/settings.xml"].decode("utf-8")
    if "updateFields" not in s:
        tag = '<w:updateFields w:val="true"/>'
        if "<w:compat>" in s:
            s = s.replace("<w:compat>", tag + "<w:compat>", 1)
        else:
            s = s.replace("</w:settings>", tag + "</w:settings>", 1)
    s = s.replace('w:eastAsia="ja-JP"', 'w:eastAsia="zh-CN"')
    data["word/settings.xml"] = s.encode("utf-8")
    with zipfile.ZipFile(tmp, "w", zipfile.ZIP_DEFLATED) as zout:
        for i in items:
            zout.writestr(i, data[i.filename])
    os.replace(tmp, path)


def build():
    d = Doc()
    fix_header(d)
    relax_spacing(d)
    part_a.cover(d)
    part_a.chapter1(d)
    part_a.chapter2(d)
    part_a.chapter3(d)
    part_a.chapter4(d)
    part_a.chapter5(d)
    part_a.chapter6(d)
    part_b.chapter7(d)
    part_b.chapter8(d)
    part_b.chapter9(d)
    part_b.chapter10(d)
    part_b.chapter11(d)
    part_b.chapter13(d)
    part_b.chapter12(d)

    cp = d.d.core_properties
    cp.title = "兰瓶门店管理系统 · 建设方案与报价（精简版）"
    cp.author = "兰瓶美业项目组"
    cp.subject = "客户、开卡、核销、库存、考勤、工资一体化管理"

    d.save(OUT)
    post_process(OUT)
    print("OK ->", OUT)
    print("size:", os.path.getsize(OUT), "bytes")
    return OUT


if __name__ == "__main__":
    build()
