# -*- coding: utf-8 -*-
"""主入口：组装兰瓶方案 Word 文档"""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from docx_kit import Doc
import part1_base as p1
import part2_modules as p2
import part3_tech as p3
import part4_plan as p4

OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                   "兰瓶美业连锁门店数字化管理系统_整体方案与报价_V1.0.docx")


def main():
    d = Doc()
    p1.cover(d)
    p1.front_matter(d)
    p1.toc(d)
    p1.chapter1(d)
    p1.chapter2(d)
    p1.chapter3(d)
    p2.chapter4(d)
    p2.chapter5(d)
    p2.chapter6(d)
    p2.chapter7(d)
    p3.chapter8(d)
    p3.chapter9(d)
    p3.chapter10(d)
    p3.chapter11(d)
    p4.chapter12(d)
    p4.chapter13(d)
    p4.chapter14(d)
    p4.chapter15(d)
    p4.chapter16(d)
    p4.appendix(d)

    # 文档属性
    cp = d.d.core_properties
    cp.title = "兰瓶美业连锁门店数字化管理系统 · 整体解决方案与项目报价"
    cp.subject = "员工管理 / 出入库 / 薪资考核 一体化管理平台"
    cp.author = "项目组"
    cp.comments = "V1.0 终版，含人天核算与商务报价"

    d.save(OUT)

    # 打开文档时提示更新域（目录、页码），并预置 updateFields
    import zipfile, shutil, re, os as _os
    tmp = OUT + ".tmp"
    shutil.move(OUT, tmp)
    with zipfile.ZipFile(tmp) as zin, zipfile.ZipFile(OUT, "w", zipfile.ZIP_DEFLATED) as zout:
        for item in zin.infolist():
            data = zin.read(item.filename)
            if item.filename == "word/settings.xml":
                s = data.decode("utf-8")
                if "updateFields" not in s:
                    # 必须置于 w:compat 之前，符合 CT_Settings 元素顺序
                    if "<w:compat>" in s:
                        s = s.replace("<w:compat>", '<w:updateFields w:val="true"/><w:compat>', 1)
                    else:
                        s = s.replace("</w:settings>", '<w:updateFields w:val="true"/></w:settings>')
                s = s.replace('w:eastAsia="ja-JP"', 'w:eastAsia="zh-CN"')
                data = s.encode("utf-8")
            zout.writestr(item, data)
    _os.remove(tmp)

    print("OK ->", OUT)
    print("size:", os.path.getsize(OUT), "bytes")


if __name__ == "__main__":
    main()
