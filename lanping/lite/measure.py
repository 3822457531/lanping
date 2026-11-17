# -*- coding: utf-8 -*-
"""粗估页数 + 可视化预览：按 run 级别字号渲染，支持强制分页"""
import asyncio
import html
import os

from docx import Document
from docx.oxml.ns import qn

HERE = os.path.dirname(os.path.abspath(__file__))
DOCX = os.path.join(HERE, "兰瓶门店管理系统_建设方案与报价_精简版.docx")
HTML_OUT = os.path.join(HERE, "_preview.html")
SHOT_DIR = os.path.join(HERE, "_shots")


def outline_level(p):
    pPr = p._p.pPr
    if pPr is None:
        return None
    o = pPr.find(qn("w:outlineLvl"))
    if o is None:
        return None
    try:
        return int(o.get(qn("w:val")))
    except Exception:
        return None


def has_page_break(p):
    for br in p._p.iter(qn("w:br")):
        if br.get(qn("w:type")) == "page":
            return True
    return False


def cell_bg(cell):
    tcPr = cell._tc.tcPr
    if tcPr is None:
        return ""
    shd = tcPr.find(qn("w:shd"))
    if shd is None:
        return ""
    return (shd.get(qn("w:fill")) or "").upper()


def runs_html(p):
    out = []
    for r in p.runs:
        t = html.escape(r.text)
        if not t:
            continue
        st = []
        sz = r.font.size
        if sz is not None:
            st.append(f"font-size:{sz.pt}pt")
        col = r.font.color
        if col is not None and col.rgb is not None:
            st.append(f"color:#{col.rgb}")
        if r.font.bold:
            st.append("font-weight:700")
        if st:
            out.append(f'<span style="{";".join(st)}">{t}</span>')
        else:
            out.append(t)
    return "".join(out) or html.escape(p.text)


def table_html(t, body):
    grid = t._tbl.find(qn("w:tblGrid"))
    widths = []
    if grid is not None:
        for gc in grid.findall(qn("w:gridCol")):
            w = gc.get(qn("w:w"))
            widths.append(int(w) if w else 0)
    total = sum(widths) or 1
    tblPr = t._tbl.find(qn("w:tblPr"))
    layout = tblPr.find(qn("w:tblLayout")) if tblPr is not None else None
    fixed = layout is not None and layout.get(qn("w:type")) == "fixed"
    border = "border:0.4pt solid #B9A9CE;" if fixed else "border:none;"
    body.append('<table style="width:100%;border-collapse:collapse;table-layout:fixed;'
                'margin:5px 0 10px 0;">')
    if widths:
        body.append("<colgroup>")
        for w in widths:
            body.append(f'<col style="width:{w/total*100:.2f}%">')
        body.append("</colgroup>")
    for row in t.rows:
        body.append("<tr>")
        for cell in row.cells:
            bg = cell_bg(cell)
            style = f"{border}padding:2px 5px;font-size:9pt;vertical-align:top;line-height:1.35;"
            color = "#2F3542"
            if bg and bg not in ("AUTO", "FFFFFF", ""):
                style += f"background:#{bg};"
                if bg in ("6B4E8E", "8A6FA8"):
                    color = "#FFFFFF"
            body.append(f'<td style="{style}color:{color};">{html.escape(cell.text)}</td>')
        body.append("</tr>")
    body.append("</table>")


def to_html(path):
    d = Document(path)
    pmap = {p._p: p for p in d.paragraphs}
    tmap = {t._tbl: t for t in d.tables}
    body = ['<html><head><meta charset="utf-8"><style>'
            'body{font-family:"Microsoft YaHei",serif;font-size:10.5pt;line-height:1.45;'
            'color:#2F3542;margin:0;padding:2.4cm 2.6cm;}'
            'p{margin:0 0 4pt 0;}'
            '.cap{font-size:9pt;color:#8C7BA6;text-align:center;margin:1pt 0 9pt 0;}'
            '.pb{height:0;page-break-after:always;}'
            '</style></head><body>']
    for child in d.element.body.iterchildren():
        if child.tag == qn("w:p"):
            para = pmap.get(child)
            if para is None:
                continue
            if has_page_break(para):
                body.append('<div class="pb"></div>')
            txt = para.text.strip()
            if not txt:
                continue
            lvl = outline_level(para)
            inner = runs_html(para)
            if lvl == 0:
                body.append(f'<p style="margin:14pt 0 10pt 0;border-bottom:1.1pt solid #6B4E8E;'
                            f'padding-bottom:4pt;">{inner}</p>')
            elif lvl == 1:
                body.append(f'<p style="margin:11pt 0 5pt 0;">{inner}</p>')
            elif txt.startswith("▍"):
                body.append(f'<p style="margin:8pt 0 3pt 0;">{inner}</p>')
            elif txt.startswith("表 ") or txt.startswith("图 "):
                body.append(f'<p class="cap">{inner}</p>')
            elif txt.startswith("●") or txt.startswith("·"):
                body.append(f'<p style="margin-left:16pt;text-indent:-8pt;">{inner}</p>')
            else:
                body.append(f'<p style="text-indent:2em;">{inner}</p>')
        elif child.tag == qn("w:tbl"):
            t = tmap.get(child)
            if t is not None:
                table_html(t, body)
    body.append("</body></html>")
    open(HTML_OUT, "w", encoding="utf-8").write("\n".join(body))
    return HTML_OUT


JS_PAGINATE = """() => {
  const pageH = (29.7 - 2.4 - 2.2) * 96 / 2.54;
  let pages = 1, y = 0;
  const tops = [];
  for (const el of document.body.children) {
    if (el.classList.contains('pb')) { pages++; y = 0; tops.push(pages); continue; }
    const h = el.getBoundingClientRect().height;
    if (y + h > pageH) {
      pages++;
      y = h % pageH;
    } else {
      y += h;
    }
  }
  return {pages, pageH: Math.round(pageH)};
}"""


async def measure_and_shoot():
    from playwright.async_api import async_playwright
    os.makedirs(SHOT_DIR, exist_ok=True)
    async with async_playwright() as pw:
        b = await pw.chromium.launch(channel="chrome")
        pg = await b.new_page(viewport={"width": 1000, "height": 1350},
                              device_scale_factor=1.4)
        await pg.goto("file:///" + HTML_OUT.replace("\\", "/"), wait_until="load")
        r = await pg.evaluate(JS_PAGINATE)
        print(f"估算页数：{r['pages']} 页（每页可用 {r['pageH']}px）")
        h = await pg.evaluate("document.body.scrollHeight")
        n = 5
        for i in range(n):
            y = int(i * (h - 1300) / (n - 1))
            await pg.evaluate(f"window.scrollTo(0,{y})")
            await pg.wait_for_timeout(120)
            await pg.screenshot(path=os.path.join(SHOT_DIR, f"p{i+1}.png"))
        print("shots ->", SHOT_DIR)
        await b.close()


if __name__ == "__main__":
    to_html(DOCX)
    asyncio.run(measure_and_shoot())
