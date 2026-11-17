# -*- coding: utf-8 -*-
"""Word 渲染工具箱：兰瓶方案文档专用"""
from docx import Document
from docx.shared import Pt, Cm, RGBColor, Emu
from docx.enum.text import WD_ALIGN_PARAGRAPH, WD_LINE_SPACING, WD_BREAK
from docx.enum.table import WD_TABLE_ALIGNMENT, WD_ALIGN_VERTICAL
from docx.enum.section import WD_SECTION
from docx.oxml.ns import qn
from docx.oxml import OxmlElement
import copy

# ---------- 品牌色 ----------
C_MAIN = RGBColor(0x6B, 0x4E, 0x8E)      # 兰紫 主色
C_GOLD = RGBColor(0xB0, 0x8D, 0x57)      # 香槟金 辅色
C_DARK = RGBColor(0x2F, 0x35, 0x42)      # 正文深灰
C_GREY = RGBColor(0x6E, 0x74, 0x80)      # 次要灰
C_WHITE = RGBColor(0xFF, 0xFF, 0xFF)
HEX_MAIN = "6B4E8E"
HEX_LIGHT = "F2EEF7"
HEX_LIGHT2 = "FAF8FC"
HEX_GOLD = "B08D57"

FONT_CN = "微软雅黑"
FONT_BODY = "宋体"


def _set_run_font(run, cn=FONT_BODY, size=10.5, bold=False, color=C_DARK, italic=False, latin=None):
    run.font.name = latin or cn
    run.font.size = Pt(size)
    run.bold = bold
    run.italic = italic
    run.font.color.rgb = color
    rpr = run._element.get_or_add_rPr()
    rf = rpr.find(qn('w:rFonts'))
    if rf is None:
        rf = OxmlElement('w:rFonts')
        rpr.append(rf)
    rf.set(qn('w:ascii'), latin or cn)
    rf.set(qn('w:hAnsi'), latin or cn)
    rf.set(qn('w:eastAsia'), cn)
    rf.set(qn('w:cs'), latin or cn)


def shade(cell, hex_color):
    tcPr = cell._tc.get_or_add_tcPr()
    for old in tcPr.findall(qn('w:shd')):
        tcPr.remove(old)
    shd = OxmlElement('w:shd')
    shd.set(qn('w:val'), 'clear')
    shd.set(qn('w:color'), 'auto')
    shd.set(qn('w:fill'), hex_color)
    tcPr.append(shd)


def cell_margins(table, top=60, bottom=60, left=90, right=90):
    tblPr = table._tbl.tblPr
    mar = OxmlElement('w:tblCellMar')
    for k, v in (('top', top), ('left', left), ('bottom', bottom), ('right', right)):
        e = OxmlElement('w:' + k)
        e.set(qn('w:w'), str(v))
        e.set(qn('w:type'), 'dxa')
        mar.append(e)
    tblPr.append(mar)


def set_repeat_header(row):
    trPr = row._tr.get_or_add_trPr()
    e = OxmlElement('w:tblHeader')
    e.set(qn('w:val'), 'true')
    trPr.append(e)


def set_borders(table, color="C9C2D6", sz=4):
    tblPr = table._tbl.tblPr
    borders = OxmlElement('w:tblBorders')
    for edge in ('top', 'left', 'bottom', 'right', 'insideH', 'insideV'):
        e = OxmlElement('w:' + edge)
        e.set(qn('w:val'), 'single')
        e.set(qn('w:sz'), str(sz))
        e.set(qn('w:space'), '0')
        e.set(qn('w:color'), color)
        borders.append(e)
    tblPr.append(borders)


def no_borders(table):
    tblPr = table._tbl.tblPr
    borders = OxmlElement('w:tblBorders')
    for edge in ('top', 'left', 'bottom', 'right', 'insideH', 'insideV'):
        e = OxmlElement('w:' + edge)
        e.set(qn('w:val'), 'nil')
        borders.append(e)
    tblPr.append(borders)


def _fix_table_layout(table, total_width):
    """强制固定表格布局并显式写入总宽，避免 Word 重算列宽导致溢出页面"""
    tblPr = table._tbl.tblPr
    for tag in ('w:tblLayout', 'w:tblW'):
        old = tblPr.find(qn(tag))
        if old is not None:
            tblPr.remove(old)
    lay = OxmlElement('w:tblLayout')
    lay.set(qn('w:type'), 'fixed')
    tblPr.append(lay)
    tw = OxmlElement('w:tblW')
    tw.set(qn('w:type'), 'dxa')
    tw.set(qn('w:w'), str(int(total_width.twips)))
    tblPr.append(tw)


def para_border_bottom(p, color=HEX_MAIN, sz=8):
    pPr = p._p.get_or_add_pPr()
    pbdr = OxmlElement('w:pBdr')
    b = OxmlElement('w:bottom')
    b.set(qn('w:val'), 'single')
    b.set(qn('w:sz'), str(sz))
    b.set(qn('w:space'), '4')
    b.set(qn('w:color'), color)
    pbdr.append(b)
    pPr.append(pbdr)


def set_outline(p, lvl):
    pPr = p._p.get_or_add_pPr()
    ol = OxmlElement('w:outlineLvl')
    ol.set(qn('w:val'), str(lvl))
    pPr.append(ol)


def shade_para(p, hex_color):
    pPr = p._p.get_or_add_pPr()
    shd = OxmlElement('w:shd')
    shd.set(qn('w:val'), 'clear')
    shd.set(qn('w:color'), 'auto')
    shd.set(qn('w:fill'), hex_color)
    pPr.append(shd)


class Doc:
    def __init__(self):
        self.d = Document()
        self._setup_page()
        self._setup_styles()
        self._setup_header_footer()
        self.fig_no = 0
        self.tab_no = 0

    # ---------- 基础设置 ----------
    def _setup_page(self):
        s = self.d.sections[0]
        s.page_width = Cm(21.0)
        s.page_height = Cm(29.7)
        s.left_margin = Cm(2.6)
        s.right_margin = Cm(2.6)
        s.top_margin = Cm(2.4)
        s.bottom_margin = Cm(2.2)
        s.header_distance = Cm(1.4)
        s.footer_distance = Cm(1.2)

    def _setup_styles(self):
        st = self.d.styles['Normal']
        st.font.name = FONT_BODY
        st.font.size = Pt(10.5)
        st.font.color.rgb = C_DARK
        st.element.rPr.rFonts.set(qn('w:eastAsia'), FONT_BODY)
        pf = st.paragraph_format
        pf.line_spacing_rule = WD_LINE_SPACING.MULTIPLE
        pf.line_spacing = 1.45
        pf.space_before = Pt(0)
        pf.space_after = Pt(3)

    def _field(self, paragraph, instr):
        r = paragraph.add_run()
        fld_begin = OxmlElement('w:fldChar')
        fld_begin.set(qn('w:fldCharType'), 'begin')
        instrText = OxmlElement('w:instrText')
        instrText.set(qn('xml:space'), 'preserve')
        instrText.text = instr
        fld_sep = OxmlElement('w:fldChar')
        fld_sep.set(qn('w:fldCharType'), 'separate')
        t = OxmlElement('w:t')
        t.text = "1"
        fld_end = OxmlElement('w:fldChar')
        fld_end.set(qn('w:fldCharType'), 'end')
        r._r.append(fld_begin)
        r._r.append(instrText)
        r._r.append(fld_sep)
        r._r.append(t)
        r._r.append(fld_end)
        return r

    def _setup_header_footer(self):
        s = self.d.sections[0]
        hp = s.header.paragraphs[0]
        hp.alignment = WD_ALIGN_PARAGRAPH.RIGHT
        r = hp.add_run("兰瓶美业连锁 · 门店数字化管理系统｜整体解决方案与项目报价")
        _set_run_font(r, FONT_CN, 8.5, False, C_GREY)
        para_border_bottom(hp, "D9D2E4", 6)

        fp = s.footer.paragraphs[0]
        fp.alignment = WD_ALIGN_PARAGRAPH.CENTER
        r1 = fp.add_run("第 ")
        _set_run_font(r1, FONT_CN, 8.5, False, C_GREY)
        self._field(fp, " PAGE ")
        r2 = fp.add_run(" 页 / 共 ")
        _set_run_font(r2, FONT_CN, 8.5, False, C_GREY)
        self._field(fp, " NUMPAGES ")
        r3 = fp.add_run(" 页    ·    商业保密文件")
        _set_run_font(r3, FONT_CN, 8.5, False, C_GREY)

    # ---------- 内容元件 ----------
    def h1(self, text, num=None, page_break=True):
        if page_break:
            self.page_break()
        p = self.d.add_paragraph()
        p.paragraph_format.space_before = Pt(4)
        p.paragraph_format.space_after = Pt(12)
        p.paragraph_format.keep_with_next = True
        if num is not None:
            r0 = p.add_run(f"{num}  ")
            _set_run_font(r0, FONT_CN, 17, True, C_GOLD)
        r = p.add_run(text)
        _set_run_font(r, FONT_CN, 17, True, C_MAIN)
        para_border_bottom(p, HEX_MAIN, 10)
        set_outline(p, 0)
        return p

    def h2(self, text):
        p = self.d.add_paragraph()
        p.paragraph_format.space_before = Pt(14)
        p.paragraph_format.space_after = Pt(6)
        p.paragraph_format.keep_with_next = True
        r = p.add_run(text)
        _set_run_font(r, FONT_CN, 13.5, True, C_MAIN)
        set_outline(p, 1)
        return p

    def h3(self, text):
        p = self.d.add_paragraph()
        p.paragraph_format.space_before = Pt(9)
        p.paragraph_format.space_after = Pt(4)
        p.paragraph_format.keep_with_next = True
        bar = p.add_run("▍")
        _set_run_font(bar, FONT_CN, 12, True, C_GOLD)
        r = p.add_run(text)
        _set_run_font(r, FONT_CN, 11.5, True, C_DARK)
        return p

    def p(self, text, indent=True, size=10.5, color=C_DARK, bold=False, align=None,
          space_after=4, first_line=True):
        para = self.d.add_paragraph()
        pf = para.paragraph_format
        pf.space_after = Pt(space_after)
        if indent:
            pf.left_indent = Cm(0.0)
        if first_line:
            pf.first_line_indent = Pt(size * 2)
        if align is not None:
            para.alignment = align
        r = para.add_run(text)
        _set_run_font(r, FONT_BODY, size, bold, color)
        return para

    def rich(self, parts, indent=True, size=10.5, first_line=True, space_after=4):
        """parts = [(text, bold, color), ...]"""
        para = self.d.add_paragraph()
        para.paragraph_format.space_after = Pt(space_after)
        if first_line:
            para.paragraph_format.first_line_indent = Pt(size * 2)
        for t, b, c in parts:
            r = para.add_run(t)
            _set_run_font(r, FONT_BODY, size, b, c or C_DARK)
        return para

    def bullet(self, text, level=0, size=10.5, color=C_DARK, bold=False):
        para = self.d.add_paragraph()
        pf = para.paragraph_format
        pf.left_indent = Cm(0.75 + level * 0.7)
        pf.first_line_indent = Cm(-0.4)
        pf.space_after = Pt(2)
        mark = "●  " if level == 0 else "○  "
        r0 = para.add_run(mark)
        _set_run_font(r0, FONT_CN, size - 2.5, False, C_MAIN if level == 0 else C_GOLD)
        r = para.add_run(text)
        _set_run_font(r, FONT_BODY, size, bold, color)
        return para

    def step(self, idx, title, body=""):
        para = self.d.add_paragraph()
        pf = para.paragraph_format
        pf.left_indent = Cm(1.3)
        pf.first_line_indent = Cm(-1.3)
        pf.space_after = Pt(3)
        r0 = para.add_run(f" STEP {idx}  ")
        _set_run_font(r0, FONT_CN, 8, True, C_WHITE)
        # 给序号加底纹
        rpr = r0._element.get_or_add_rPr()
        shd = OxmlElement('w:shd')
        shd.set(qn('w:val'), 'clear')
        shd.set(qn('w:fill'), HEX_MAIN)
        rpr.append(shd)
        r1 = para.add_run(" " + title)
        _set_run_font(r1, FONT_CN, 10.5, True, C_MAIN)
        if body:
            r2 = para.add_run("  " + body)
            _set_run_font(r2, FONT_BODY, 10.5, False, C_DARK)
        return para

    def note(self, text, kind="提示"):
        para = self.d.add_paragraph()
        pf = para.paragraph_format
        pf.left_indent = Cm(0.3)
        pf.right_indent = Cm(0.3)
        pf.space_before = Pt(6)
        pf.space_after = Pt(8)
        shade_para(para, HEX_LIGHT)
        r0 = para.add_run(f"【{kind}】")
        _set_run_font(r0, FONT_CN, 9.5, True, C_GOLD)
        r = para.add_run(text)
        _set_run_font(r, FONT_BODY, 9.5, False, C_DARK)
        return para

    def code(self, text, size=9):
        para = self.d.add_paragraph()
        pf = para.paragraph_format
        pf.left_indent = Cm(0.5)
        pf.space_before = Pt(3)
        pf.space_after = Pt(6)
        pf.line_spacing = 1.15
        shade_para(para, "F5F4F7")
        r = para.add_run(text)
        _set_run_font(r, FONT_CN, size, False, RGBColor(0x3A, 0x2F, 0x50), latin="Consolas")
        return para

    def caption(self, text, kind="table"):
        para = self.d.add_paragraph()
        para.alignment = WD_ALIGN_PARAGRAPH.LEFT
        para.paragraph_format.space_before = Pt(8)
        para.paragraph_format.space_after = Pt(3)
        para.paragraph_format.keep_with_next = True
        if kind == "table":
            self.tab_no += 1
            tag = f"表 {self.tab_no}  "
        else:
            self.fig_no += 1
            tag = f"图 {self.fig_no}  "
        r0 = para.add_run(tag)
        _set_run_font(r0, FONT_CN, 9, True, C_GOLD)
        r = para.add_run(text)
        _set_run_font(r, FONT_CN, 9, True, C_GREY)
        return para

    def table(self, headers, rows, widths=None, caption=None, size=9,
              header_fill=HEX_MAIN, first_col_bold=False, align_center_cols=None,
              zebra=True, total_row=False):
        if caption:
            self.caption(caption, "table")
        ncol = len(headers)
        t = self.d.add_table(rows=1, cols=ncol)
        t.alignment = WD_TABLE_ALIGNMENT.CENTER
        t.autofit = False
        set_borders(t)
        cell_margins(t)
        total_w = Cm(15.8)
        if widths is None or len(widths) != ncol:
            widths = [1.0] * ncol
        ssum = sum(widths)
        colw = [Emu(int(total_w * w / ssum)) for w in widths]
        _fix_table_layout(t, total_w)
        for i in range(ncol):
            t.columns[i].width = colw[i]

        hdr = t.rows[0]
        set_repeat_header(hdr)
        for i, h in enumerate(headers):
            c = hdr.cells[i]
            c.width = colw[i]
            c.text = ""
            p = c.paragraphs[0]
            p.alignment = WD_ALIGN_PARAGRAPH.CENTER
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            r = p.add_run(h)
            _set_run_font(r, FONT_CN, size, True, C_WHITE)
            shade(c, header_fill)

        for ri, row in enumerate(rows):
            cells = t.add_row().cells
            is_total = total_row and ri == len(rows) - 1
            for ci, val in enumerate(row):
                c = cells[ci]
                c.width = colw[ci]
                c.text = ""
                p = c.paragraphs[0]
                p.paragraph_format.space_before = Pt(1.5)
                p.paragraph_format.space_after = Pt(1.5)
                p.paragraph_format.line_spacing = 1.2
                if align_center_cols and ci in align_center_cols:
                    p.alignment = WD_ALIGN_PARAGRAPH.CENTER
                txt = str(val)
                bold = is_total or (first_col_bold and ci == 0)
                r = p.add_run(txt)
                _set_run_font(r, FONT_BODY, size, bold,
                              C_MAIN if is_total else C_DARK)
                if is_total:
                    shade(c, HEX_LIGHT)
                elif zebra and ri % 2 == 1:
                    shade(c, HEX_LIGHT2)
        # 段后间距
        sp = self.d.add_paragraph()
        sp.paragraph_format.space_after = Pt(4)
        sp.paragraph_format.line_spacing = 0.6
        return t

    def figure_flow(self, title, nodes, caption=None):
        """横向流程：nodes = ['A','B','C']，用表格模拟（宽度严格控制在版心内）"""
        if title:
            p = self.d.add_paragraph()
            p.alignment = WD_ALIGN_PARAGRAPH.CENTER
            p.paragraph_format.space_before = Pt(6)
            p.paragraph_format.space_after = Pt(3)
            r = p.add_run(title)
            _set_run_font(r, FONT_CN, 9, True, C_MAIN)
        n = len(nodes)
        cols = n * 2 - 1
        t = self.d.add_table(rows=1, cols=cols)
        t.alignment = WD_TABLE_ALIGNMENT.CENTER
        t.autofit = False
        _fix_table_layout(t, Cm(15.6))
        no_borders(t)
        cell_margins(t, 10, 20, 10, 20)
        ARROW = 0.45
        node_w = Cm((15.6 - ARROW * (n - 1)) / n)
        arrow_w = Cm(ARROW)
        fsize = 7.4 if n >= 6 else 8.2
        for j in range(cols):
            t.columns[j].width = node_w if j % 2 == 0 else arrow_w
        i = 0
        for k, nd in enumerate(nodes):
            c = t.rows[0].cells[i]
            c.width = node_w
            shade(c, HEX_MAIN if k % 2 == 0 else "8A6FA8")
            p = c.paragraphs[0]
            p.alignment = WD_ALIGN_PARAGRAPH.CENTER
            p.paragraph_format.space_before = Pt(6)
            p.paragraph_format.space_after = Pt(6)
            p.paragraph_format.line_spacing = 1.0
            r = p.add_run(nd)
            _set_run_font(r, FONT_CN, fsize, True, C_WHITE)
            if k < n - 1:
                c2 = t.rows[0].cells[i + 1]
                c2.width = arrow_w
                c2.vertical_alignment = WD_ALIGN_VERTICAL.CENTER
                p2 = c2.paragraphs[0]
                p2.alignment = WD_ALIGN_PARAGRAPH.CENTER
                p2.paragraph_format.space_before = Pt(0)
                p2.paragraph_format.space_after = Pt(0)
                r2 = p2.add_run("→")
                _set_run_font(r2, "Segoe UI Symbol", 9, True, C_GOLD)
            i += 2
        sp = self.d.add_paragraph()
        sp.paragraph_format.space_after = Pt(4)
        sp.paragraph_format.line_spacing = 0.6
        if caption:
            self.caption(caption, "figure")

    def page_break(self):
        p = self.d.add_paragraph()
        p.paragraph_format.space_after = Pt(0)
        p.add_run().add_break(WD_BREAK.PAGE)

    def save(self, path):
        self.d.save(path)
