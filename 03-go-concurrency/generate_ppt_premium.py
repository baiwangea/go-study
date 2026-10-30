#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
企业级Go并发编程课程PPT生成器（顶级质量版）
使用深色代码主题、精美流程图、专业配色方案
"""

from pptx import Presentation
from pptx.util import Inches, Pt
from pptx.enum.text import PP_ALIGN, MSO_ANCHOR
from pptx.dml.color import RGBColor
from pptx.enum.shapes import MSO_SHAPE
import os

# ============ 顶级配色方案 ============
class Colors:
    # 主色系
    PRIMARY_DARK = RGBColor(30, 58, 138)      # 深蓝 #1E3A8A
    PRIMARY = RGBColor(59, 130, 246)          # 科技蓝 #3B82F6
    ACCENT = RGBColor(6, 182, 212)            # 青色 #06B6D4
    
    # 背景色
    BG_LIGHT = RGBColor(248, 250, 252)        # 渐变白起点 #F8FAFC
    BG_BLUE = RGBColor(239, 246, 255)         # 渐变白终点 #EFF6FF
    
    # 文本色
    TEXT_DARK = RGBColor(31, 41, 55)          # 深灰 #1F2937
    TEXT_GRAY = RGBColor(107, 114, 128)       # 中灰 #6B7280
    TEXT_LIGHT = RGBColor(156, 163, 175)      # 浅灰 #9CA3AF
    
    # 代码背景（VS Code Dark主题）
    CODE_BG = RGBColor(40, 44, 52)            # #282C34
    CODE_LINE_NUM = RGBColor(92, 99, 112)     # #5C6370
    
    # 语法高亮色
    KEYWORD = RGBColor(255, 107, 157)         # 粉红 #FF6B9D
    STRING = RGBColor(152, 195, 121)          # 绿色 #98C379
    COMMENT = RGBColor(92, 99, 112)           # 灰色 #5C6370
    FUNCTION = RGBColor(97, 175, 239)         # 蓝色 #61AFEF
    NUMBER = RGBColor(209, 154, 102)          # 橙色 #D19A66
    TYPE = RGBColor(229, 192, 123)            # 黄色 #E5C07B
    
    # 状态色
    SUCCESS = RGBColor(16, 185, 129)          # 绿色
    WARNING = RGBColor(245, 158, 11)          # 橙色
    ERROR = RGBColor(239, 68, 68)             # 红色
    INFO = RGBColor(59, 130, 246)             # 蓝色
    
    # 图表色（渐变友好）
    CHART_1 = RGBColor(99, 102, 241)          # 靛蓝
    CHART_2 = RGBColor(168, 85, 247)          # 紫色
    CHART_3 = RGBColor(236, 72, 153)          # 粉色
    CHART_4 = RGBColor(34, 197, 94)           # 绿色


class PPTGenerator:
    def __init__(self):
        self.prs = Presentation()
        self.prs.slide_width = Inches(10)
        self.prs.slide_height = Inches(7.5)
        
    def add_gradient_background(self, slide):
        """添加渐变背景"""
        background = slide.background
        fill = background.fill
        fill.gradient()
        fill.gradient_angle = 45.0
        fill.gradient_stops[0].color.rgb = Colors.BG_LIGHT
        fill.gradient_stops[1].color.rgb = Colors.BG_BLUE
        
    def add_header(self, slide, chapter_name):
        """添加页眉（章节名称）"""
        header = slide.shapes.add_textbox(
            Inches(0.5), Inches(0.3),
            Inches(9), Inches(0.3)
        )
        tf = header.text_frame
        tf.text = chapter_name
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = Colors.TEXT_GRAY
        p.font.name = '微软雅黑'
        
    def add_footer(self, slide, page_num):
        """添加页脚（页码）"""
        footer = slide.shapes.add_textbox(
            Inches(4.5), Inches(7),
            Inches(1), Inches(0.3)
        )
        tf = footer.text_frame
        tf.text = str(page_num)
        p = tf.paragraphs[0]
        p.font.size = Pt(11)
        p.font.color.rgb = Colors.TEXT_GRAY
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
    def add_title_slide(self, title, subtitle):
        """封面页"""
        slide = self.prs.slides.add_slide(self.prs.slide_layouts[6])
        self.add_gradient_background(slide)
        
        # 主标题
        title_box = slide.shapes.add_textbox(
            Inches(1), Inches(2.5),
            Inches(8), Inches(1.2)
        )
        tf = title_box.text_frame
        tf.text = title
        p = tf.paragraphs[0]
        p.font.size = Pt(44)
        p.font.bold = True
        p.font.color.rgb = Colors.PRIMARY_DARK
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
        # 副标题
        subtitle_box = slide.shapes.add_textbox(
            Inches(1), Inches(4),
            Inches(8), Inches(0.6)
        )
        tf = subtitle_box.text_frame
        tf.text = subtitle
        p = tf.paragraphs[0]
        p.font.size = Pt(24)
        p.font.color.rgb = Colors.TEXT_GRAY
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
        # 装饰线
        line1 = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            Inches(3), Inches(3.9),
            Inches(4), Inches(0.05)
        )
        line1.fill.solid()
        line1.fill.fore_color.rgb = Colors.ACCENT
        line1.line.fill.background()
        
    def add_section_slide(self, section_num, section_title):
        """章节分隔页"""
        slide = self.prs.slides.add_slide(self.prs.slide_layouts[6])
        self.add_gradient_background(slide)
        
        # 章节编号（大号）
        num_box = slide.shapes.add_textbox(
            Inches(1), Inches(2),
            Inches(8), Inches(1.5)
        )
        tf = num_box.text_frame
        tf.text = section_num
        p = tf.paragraphs[0]
        p.font.size = Pt(80)
        p.font.bold = True
        p.font.color.rgb = Colors.PRIMARY
        p.font.name = 'Segoe UI'
        p.alignment = PP_ALIGN.CENTER
        
        # 章节名称
        title_box = slide.shapes.add_textbox(
            Inches(1), Inches(4),
            Inches(8), Inches(0.8)
        )
        tf = title_box.text_frame
        tf.text = section_title
        p = tf.paragraphs[0]
        p.font.size = Pt(36)
        p.font.bold = True
        p.font.color.rgb = Colors.TEXT_DARK
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
    def add_code_block(self, slide, code_lines, left, top, width, height, 
                       title=None, highlight_lines=None):
        """
        添加精美代码块（深色主题 + 语法高亮）
        
        Args:
            code_lines: list of (line_num, code, color) tuples
            highlight_lines: list of line numbers to highlight
        """
        # 代码块背景（深色 + 圆角 + 阴影）
        code_bg = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            left, top, width, height
        )
        code_bg.fill.solid()
        code_bg.fill.fore_color.rgb = Colors.CODE_BG
        code_bg.line.color.rgb = Colors.TEXT_GRAY
        code_bg.line.width = Pt(1)
        code_bg.shadow.inherit = False
        
        # 标题区域（如果有）
        text_top = top
        if title:
            title_box = slide.shapes.add_textbox(
                left + Inches(0.15), top + Inches(0.1),
                width - Inches(0.3), Inches(0.3)
            )
            tf = title_box.text_frame
            tf.text = title
            p = tf.paragraphs[0]
            p.font.size = Pt(12)
            p.font.bold = True
            p.font.color.rgb = Colors.TEXT_LIGHT
            p.font.name = 'Consolas'
            text_top += Inches(0.35)
        
        # 代码内容
        code_box = slide.shapes.add_textbox(
            left + Inches(0.15), text_top + Inches(0.1),
            width - Inches(0.3), height - Inches(0.2) - (Inches(0.35) if title else 0)
        )
        tf = code_box.text_frame
        tf.word_wrap = True
        tf.vertical_anchor = MSO_ANCHOR.TOP
        
        for line_num, code_text, segments in code_lines:
            p = tf.add_paragraph() if line_num > 1 else tf.paragraphs[0]
            p.font.size = Pt(14)
            p.font.name = 'Consolas'
            p.level = 0
            
            # 行号
            run = p.add_run()
            run.text = f"{line_num:2d} │ "
            run.font.color.rgb = Colors.CODE_LINE_NUM
            run.font.name = 'Consolas'
            
            # 代码内容（分段着色）
            for text, color in segments:
                run = p.add_run()
                run.text = text
                run.font.color.rgb = color
                run.font.name = 'Consolas'
                
    def add_card(self, slide, left, top, width, height, title, content_lines, 
                 color=None, icon=None):
        """添加卡片式内容块（圆角 + 阴影）"""
        if color is None:
            color = Colors.PRIMARY
            
        # 卡片背景
        card = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            left, top, width, height
        )
        card.fill.solid()
        card.fill.fore_color.rgb = RGBColor(255, 255, 255)
        card.line.color.rgb = Colors.TEXT_LIGHT
        card.line.width = Pt(0.5)
        card.shadow.inherit = False
        
        # 标题区域（带背景色）
        title_bg = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            left, top, width, Inches(0.4)
        )
        title_bg.fill.solid()
        title_bg.fill.fore_color.rgb = color
        title_bg.line.fill.background()
        
        # 标题文本
        title_box = slide.shapes.add_textbox(
            left + Inches(0.15), top + Inches(0.05),
            width - Inches(0.3), Inches(0.3)
        )
        tf = title_box.text_frame
        if icon:
            tf.text = f"{icon} {title}"
        else:
            tf.text = title
        p = tf.paragraphs[0]
        p.font.size = Pt(16)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.font.name = '微软雅黑'
        
        # 内容区域
        content_box = slide.shapes.add_textbox(
            left + Inches(0.15), top + Inches(0.5),
            width - Inches(0.3), height - Inches(0.6)
        )
        tf = content_box.text_frame
        tf.word_wrap = True
        
        for i, line in enumerate(content_lines):
            p = tf.add_paragraph() if i > 0 else tf.paragraphs[0]
            p.text = line
            p.font.size = Pt(14)
            p.font.color.rgb = Colors.TEXT_DARK
            p.font.name = '微软雅黑'
            p.space_after = Pt(8)
            
    def add_flow_chart_goroutine_lifecycle(self, slide, left, top):
        """Goroutine生命周期流程图（时间轴样式）"""
        stages = [
            ("创建", "go func()", Colors.CHART_1),
            ("就绪", "等待调度", Colors.CHART_2),
            ("运行", "执行代码", Colors.SUCCESS),
            ("阻塞", "等待I/O", Colors.WARNING),
            ("完成", "退出", Colors.TEXT_GRAY)
        ]
        
        box_width = Inches(1.4)
        box_height = Inches(0.6)
        spacing = Inches(0.3)
        
        for i, (stage, desc, color) in enumerate(stages):
            x = left + i * (box_width + spacing)
            
            # 状态框（圆角矩形）
            box = slide.shapes.add_shape(
                MSO_SHAPE.ROUNDED_RECTANGLE,
                x, top, box_width, box_height
            )
            box.fill.solid()
            box.fill.fore_color.rgb = color
            box.line.color.rgb = color
            box.line.width = Pt(2)
            box.shadow.inherit = False
            
            # 状态名称
            tb = slide.shapes.add_textbox(x, top + Inches(0.1), box_width, Inches(0.25))
            tf = tb.text_frame
            tf.text = stage
            p = tf.paragraphs[0]
            p.font.size = Pt(14)
            p.font.bold = True
            p.font.color.rgb = RGBColor(255, 255, 255)
            p.font.name = '微软雅黑'
            p.alignment = PP_ALIGN.CENTER
            
            # 描述
            tb2 = slide.shapes.add_textbox(x, top + Inches(0.35), box_width, Inches(0.2))
            tf2 = tb2.text_frame
            tf2.text = desc
            p2 = tf2.paragraphs[0]
            p2.font.size = Pt(11)
            p2.font.color.rgb = RGBColor(255, 255, 255)
            p2.font.name = '微软雅黑'
            p2.alignment = PP_ALIGN.CENTER
            
            # 箭头
            if i < len(stages) - 1:
                arrow = slide.shapes.add_connector(
                    1,  # MSO_CONNECTOR.STRAIGHT
                    x + box_width, top + box_height / 2,
                    x + box_width + spacing, top + box_height / 2
                )
                arrow.line.color.rgb = Colors.TEXT_GRAY
                arrow.line.width = Pt(2)
                
    def add_flow_chart_channel(self, slide, left, top):
        """Channel通信流程图（管道可视化）"""
        # 发送方
        sender = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            left, top, Inches(1.2), Inches(0.8)
        )
        sender.fill.solid()
        sender.fill.fore_color.rgb = Colors.CHART_1
        sender.line.color.rgb = Colors.CHART_1
        sender.line.width = Pt(2)
        
        tb = slide.shapes.add_textbox(left, top + Inches(0.25), Inches(1.2), Inches(0.3))
        tf = tb.text_frame
        tf.text = "发送方\nch <- data"
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
        # Channel管道
        channel_left = left + Inches(1.5)
        channel = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            channel_left, top + Inches(0.15),
            Inches(2.5), Inches(0.5)
        )
        channel.fill.solid()
        channel.fill.fore_color.rgb = Colors.PRIMARY
        channel.line.color.rgb = Colors.PRIMARY_DARK
        channel.line.width = Pt(2)
        
        tb = slide.shapes.add_textbox(
            channel_left, top + Inches(0.25),
            Inches(2.5), Inches(0.3)
        )
        tf = tb.text_frame
        tf.text = "Channel (缓冲区)"
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
        # 接收方
        receiver_left = left + Inches(4.3)
        receiver = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            receiver_left, top, Inches(1.2), Inches(0.8)
        )
        receiver.fill.solid()
        receiver.fill.fore_color.rgb = Colors.SUCCESS
        receiver.line.color.rgb = Colors.SUCCESS
        receiver.line.width = Pt(2)
        
        tb = slide.shapes.add_textbox(
            receiver_left, top + Inches(0.25),
            Inches(1.2), Inches(0.3)
        )
        tf = tb.text_frame
        tf.text = "接收方\ndata := <-ch"
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.font.name = '微软雅黑'
        p.alignment = PP_ALIGN.CENTER
        
        # 箭头（发送）
        arrow1 = slide.shapes.add_connector(
            1, left + Inches(1.2), top + Inches(0.4),
            channel_left, top + Inches(0.4)
        )
        arrow1.line.color.rgb = Colors.TEXT_DARK
        arrow1.line.width = Pt(3)
        
        # 箭头（接收）
        arrow2 = slide.shapes.add_connector(
            1, channel_left + Inches(2.5), top + Inches(0.4),
            receiver_left, top + Inches(0.4)
        )
        arrow2.line.color.rgb = Colors.TEXT_DARK
        arrow2.line.width = Pt(3)
        
    def add_content_slide(self, chapter, title, page_num):
        """标准内容页"""
        slide = self.prs.slides.add_slide(self.prs.slide_layouts[6])
        self.add_gradient_background(slide)
        self.add_header(slide, chapter)
        self.add_footer(slide, page_num)
        
        # 标题
        title_box = slide.shapes.add_textbox(
            Inches(0.5), Inches(0.8),
            Inches(9), Inches(0.6)
        )
        tf = title_box.text_frame
        tf.text = title
        p = tf.paragraphs[0]
        p.font.size = Pt(36)
        p.font.bold = True
        p.font.color.rgb = Colors.PRIMARY_DARK
        p.font.name = '微软雅黑'
        
        # 装饰线
        line = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            Inches(0.5), Inches(1.5),
            Inches(9), Inches(0.03)
        )
        line.fill.solid()
        line.fill.fore_color.rgb = Colors.ACCENT
        line.line.fill.background()
        
        return slide
        
    def save(self, filename):
        """保存PPT文件"""
        self.prs.save(filename)
        print(f"✅ PPT已生成: {filename}")


def generate_premium_ppt():
    """生成顶级质量的Go并发编程PPT"""
    gen = PPTGenerator()
    page_num = 1
    
    # ========== 封面 ==========
    gen.add_title_slide(
        "Go 并发编程",
        "Enterprise-Grade Concurrency Patterns"
    )
    
    # ========== 第一章：Goroutine基础 ==========
    gen.add_section_slide("01", "Goroutine 基础")
    page_num += 2
    
    # 1.1 Goroutine启动
    slide = gen.add_content_slide("01 · Goroutine 基础", "1.1 启动 Goroutine", page_num)
    page_num += 1
    
    # 代码示例
    code_lines = [
        (1, "package main", [("package", Colors.KEYWORD), (" main", Colors.TYPE)]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "import \"fmt\"", [("import", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("\"fmt\"", Colors.STRING)]),
        (4, "", [("", Colors.TEXT_LIGHT)]),
        (5, "func worker(id int) {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("worker", Colors.FUNCTION), ("(id ", Colors.TEXT_LIGHT),
            ("int", Colors.TYPE), (") {", Colors.TEXT_LIGHT)
        ]),
        (6, "    fmt.Println(id)", [
            ("    ", Colors.TEXT_LIGHT), ("fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT),
            ("Println", Colors.FUNCTION), ("(id)", Colors.TEXT_LIGHT)
        ]),
        (7, "}", [("}", Colors.TEXT_LIGHT)]),
        (8, "", [("", Colors.TEXT_LIGHT)]),
        (9, "func main() {", [("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
                            ("main", Colors.FUNCTION), ("() {", Colors.TEXT_LIGHT)]),
        (10, "    go worker(1)  // 异步启动", [
            ("    ", Colors.TEXT_LIGHT), ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("worker", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT), ("1", Colors.NUMBER),
            (")  ", Colors.TEXT_LIGHT), ("// 异步启动", Colors.COMMENT)
        ]),
        (11, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(4.5), Inches(4),
        title="goroutine 启动示例"
    )
    
    # 要点卡片
    gen.add_card(
        slide, Inches(5.2), Inches(1.8), Inches(4.3), Inches(2),
        "🔑 关键要点", [
            "• go 关键字启动异步执行",
            "• 主协程不等待子协程",
            "• 调度顺序不确定",
            "• 退出时直接抛弃未完成任务"
        ],
        color=Colors.PRIMARY
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(4), Inches(4.3), Inches(1.8),
        "⚠️ 常见陷阱", [
            "• 忘记等待 → 程序提前退出",
            "• 闭包捕获循环变量",
            "• 用 sleep 猜测完成时间"
        ],
        color=Colors.WARNING
    )
    
    # 1.2 WaitGroup
    slide = gen.add_content_slide("01 · Goroutine 基础", "1.2 WaitGroup 等待", page_num)
    page_num += 1
    
    code_lines = [
        (1, "var wg sync.WaitGroup", [
            ("var", Colors.KEYWORD), (" wg ", Colors.TEXT_LIGHT),
            ("sync", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("WaitGroup", Colors.TYPE)
        ]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "for i := 0; i < 5; i++ {", [
            ("for", Colors.KEYWORD), (" i := ", Colors.TEXT_LIGHT), ("0", Colors.NUMBER),
            ("; i < ", Colors.TEXT_LIGHT), ("5", Colors.NUMBER), ("; i++ {", Colors.TEXT_LIGHT)
        ]),
        (4, "    wg.Add(1)  // 计数+1", [
            ("    wg.", Colors.TEXT_LIGHT), ("Add", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("1", Colors.NUMBER), (")  ", Colors.TEXT_LIGHT), ("// 计数+1", Colors.COMMENT)
        ]),
        (5, "    go func(id int) {", [
            ("    ", Colors.TEXT_LIGHT), ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("func", Colors.KEYWORD), ("(id ", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (") {", Colors.TEXT_LIGHT)
        ]),
        (6, "        defer wg.Done()", [
            ("        ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" wg.", Colors.TEXT_LIGHT),
            ("Done", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (7, "        worker(id)", [
            ("        ", Colors.TEXT_LIGHT), ("worker", Colors.FUNCTION),
            ("(id)", Colors.TEXT_LIGHT)
        ]),
        (8, "    }(i)", [("    }(i)", Colors.TEXT_LIGHT)]),
        (9, "}", [("}", Colors.TEXT_LIGHT)]),
        (10, "", [("", Colors.TEXT_LIGHT)]),
        (11, "wg.Wait()  // 等待全部完成", [
            ("wg.", Colors.TEXT_LIGHT), ("Wait", Colors.FUNCTION), ("()  ", Colors.TEXT_LIGHT),
            ("// 等待全部完成", Colors.COMMENT)
        ]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(4.5), Inches(4.2),
        title="WaitGroup 使用模式"
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(1.8), Inches(4.3), Inches(2.5),
        "📋 三步法则", [
            "1️⃣ Add(n) — 增加计数",
            "2️⃣ Done() — 完成时调用",
            "3️⃣ Wait() — 阻塞等待归零",
            "",
            "⚠️ Add 必须在 goroutine 外"
        ],
        color=Colors.INFO
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(4.5), Inches(4.3), Inches(1.3),
        "✅ 最佳实践", [
            "• defer wg.Done() 防遗漏",
            "• Add 在启动前调用"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第二章：Channel通信 ==========
    gen.add_section_slide("02", "Channel 通信")
    page_num += 2
    
    # 2.1 无缓冲Channel
    slide = gen.add_content_slide("02 · Channel 通信", "2.1 无缓冲 Channel", page_num)
    page_num += 1
    
    # 流程图
    gen.add_flow_chart_channel(slide, Inches(0.8), Inches(2))
    
    code_lines = [
        (1, "ch := make(chan int)  // 无缓冲", [
            ("ch := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE),
            (")  ", Colors.TEXT_LIGHT), ("// 无缓冲", Colors.COMMENT)
        ]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "go func() {", [
            ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (4, "    data := <-ch  // 接收", [
            ("    data := <-ch  ", Colors.TEXT_LIGHT), ("// 接收", Colors.COMMENT)
        ]),
        (5, "    process(data)", [
            ("    ", Colors.TEXT_LIGHT), ("process", Colors.FUNCTION), ("(data)", Colors.TEXT_LIGHT)
        ]),
        (6, "}()", [("}()", Colors.TEXT_LIGHT)]),
        (7, "", [("", Colors.TEXT_LIGHT)]),
        (8, "ch <- 42  // 阻塞到被接收", [
            ("ch <- ", Colors.TEXT_LIGHT), ("42", Colors.NUMBER), ("  ", Colors.TEXT_LIGHT),
            ("// 阻塞到被接收", Colors.COMMENT)
        ]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(3.2), Inches(4.5), Inches(3),
        title="无缓冲 Channel 示例"
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(3.2), Inches(4.3), Inches(1.5),
        "🔄 同步语义", [
            "• 发送阻塞到有接收",
            "• 接收阻塞到有发送",
            "• 一次收发 = 一次握手"
        ],
        color=Colors.CHART_1
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(4.9), Inches(4.3), Inches(1.3),
        "🎯 适用场景", [
            "• 下单指令（必须送达）",
            "• 严格同步控制"
        ],
        color=Colors.SUCCESS
    )
    
    # 2.2 有缓冲Channel
    slide = gen.add_content_slide("02 · Channel 通信", "2.2 有缓冲 Channel", page_num)
    page_num += 1
    
    code_lines = [
        (1, "ch := make(chan int, 3)  // 容量3", [
            ("ch := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE),
            (", ", Colors.TEXT_LIGHT), ("3", Colors.NUMBER), (")  ", Colors.TEXT_LIGHT),
            ("// 容量3", Colors.COMMENT)
        ]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "ch <- 1  // 不阻塞", [
            ("ch <- ", Colors.TEXT_LIGHT), ("1", Colors.NUMBER), ("  ", Colors.TEXT_LIGHT),
            ("// 不阻塞", Colors.COMMENT)
        ]),
        (4, "ch <- 2  // 不阻塞", [
            ("ch <- ", Colors.TEXT_LIGHT), ("2", Colors.NUMBER), ("  ", Colors.TEXT_LIGHT),
            ("// 不阻塞", Colors.COMMENT)
        ]),
        (5, "ch <- 3  // 不阻塞", [
            ("ch <- ", Colors.TEXT_LIGHT), ("3", Colors.NUMBER), ("  ", Colors.TEXT_LIGHT),
            ("// 不阻塞", Colors.COMMENT)
        ]),
        (6, "ch <- 4  // 阻塞！队列满", [
            ("ch <- ", Colors.TEXT_LIGHT), ("4", Colors.NUMBER), ("  ", Colors.TEXT_LIGHT),
            ("// 阻塞！队列满", Colors.COMMENT)
        ]),
        (7, "", [("", Colors.TEXT_LIGHT)]),
        (8, "fmt.Println(len(ch), cap(ch))", [
            ("fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("len", Colors.FUNCTION), ("(ch), ", Colors.TEXT_LIGHT),
            ("cap", Colors.FUNCTION), ("(ch))", Colors.TEXT_LIGHT)
        ]),
        (9, "// 输出: 3 3", [("// 输出: 3 3", Colors.COMMENT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(4.5), Inches(3.5),
        title="有缓冲 Channel 示例"
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(1.8), Inches(4.3), Inches(1.8),
        "📊 缓冲队列", [
            "• len(ch) — 当前元素数",
            "• cap(ch) — 容量",
            "• 队列满才阻塞发送",
            "• 队列空才阻塞接收"
        ],
        color=Colors.CHART_2
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(3.8), Inches(4.3), Inches(1.5),
        "🎯 适用场景", [
            "• 流量削峰",
            "• 异步通知",
            "• 容忍短暂速度差"
        ],
        color=Colors.SUCCESS
    )
    
    # 2.3 Close和Range
    slide = gen.add_content_slide("02 · Channel 通信", "2.3 Close 与 Range", page_num)
    page_num += 1
    
    code_lines = [
        (1, "ch := make(chan int, 2)", [
            ("ch := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE),
            (", ", Colors.TEXT_LIGHT), ("2", Colors.NUMBER), (")", Colors.TEXT_LIGHT)
        ]),
        (2, "ch <- 1", [("ch <- ", Colors.TEXT_LIGHT), ("1", Colors.NUMBER)]),
        (3, "ch <- 2", [("ch <- ", Colors.TEXT_LIGHT), ("2", Colors.NUMBER)]),
        (4, "close(ch)  // 发送方关闭", [
            ("close", Colors.FUNCTION), ("(ch)  ", Colors.TEXT_LIGHT),
            ("// 发送方关闭", Colors.COMMENT)
        ]),
        (5, "", [("", Colors.TEXT_LIGHT)]),
        (6, "// 方式1: range自动检测关闭", [("// 方式1: range自动检测关闭", Colors.COMMENT)]),
        (7, "for v := range ch {", [
            ("for", Colors.KEYWORD), (" v := ", Colors.TEXT_LIGHT), ("range", Colors.KEYWORD),
            (" ch {", Colors.TEXT_LIGHT)
        ]),
        (8, "    fmt.Println(v)", [
            ("    fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(v)", Colors.TEXT_LIGHT)
        ]),
        (9, "}", [("}", Colors.TEXT_LIGHT)]),
        (10, "", [("", Colors.TEXT_LIGHT)]),
        (11, "// 方式2: 逗号ok检测", [("// 方式2: 逗号ok检测", Colors.COMMENT)]),
        (12, "v, ok := <-ch", [("v, ok := <-ch", Colors.TEXT_LIGHT)]),
        (13, "if !ok {", [("if", Colors.KEYWORD), (" !ok {", Colors.TEXT_LIGHT)]),
        (14, "    // channel已关闭", [("    ", Colors.TEXT_LIGHT), ("// channel已关闭", Colors.COMMENT)]),
        (15, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5), Inches(4.5),
        title="Close 与 Range 示例"
    )
    
    gen.add_card(
        slide, Inches(5.7), Inches(1.8), Inches(3.8), Inches(2.2),
        "⚠️ 关闭规则", [
            "• 只有发送方能 close",
            "• close 后不能再发送",
            "• close 后能继续接收",
            "• 接收完返回零值",
            "• 重复 close 会 panic"
        ],
        color=Colors.ERROR
    )
    
    gen.add_card(
        slide, Inches(5.7), Inches(4.2), Inches(3.8), Inches(2),
        "✅ 最佳实践", [
            "• range 自动退出",
            "• 用 ok 判断是否关闭",
            "• Worker Pool 模式"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第三章：Select多路复用 ==========
    gen.add_section_slide("03", "Select 多路复用")
    page_num += 2
    
    # 3.1 Select基础
    slide = gen.add_content_slide("03 · Select 多路复用", "3.1 Select 基础", page_num)
    page_num += 1
    
    code_lines = [
        (1, "select {", [("select", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)]),
        (2, "case v := <-ch1:", [
            ("case", Colors.KEYWORD), (" v := <-ch1:", Colors.TEXT_LIGHT)
        ]),
        (3, "    fmt.Println(\"ch1:\", v)", [
            ("    fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("\"ch1:\"", Colors.STRING), (", v)", Colors.TEXT_LIGHT)
        ]),
        (4, "", [("", Colors.TEXT_LIGHT)]),
        (5, "case v := <-ch2:", [
            ("case", Colors.KEYWORD), (" v := <-ch2:", Colors.TEXT_LIGHT)
        ]),
        (6, "    fmt.Println(\"ch2:\", v)", [
            ("    fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("\"ch2:\"", Colors.STRING), (", v)", Colors.TEXT_LIGHT)
        ]),
        (7, "", [("", Colors.TEXT_LIGHT)]),
        (8, "case <-time.After(1 * time.Second):", [
            ("case", Colors.KEYWORD), (" <-time", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("After", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT), ("1", Colors.NUMBER),
            (" * time", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Second", Colors.TYPE),
            ("):", Colors.TEXT_LIGHT)
        ]),
        (9, "    fmt.Println(\"timeout\")", [
            ("    fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("\"timeout\"", Colors.STRING), (")", Colors.TEXT_LIGHT)
        ]),
        (10, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5), Inches(3.5),
        title="Select 多路复用示例"
    )
    
    gen.add_card(
        slide, Inches(5.7), Inches(1.8), Inches(3.8), Inches(2),
        "🔀 执行规则", [
            "• 有就绪就执行",
            "• 多个就绪随机选",
            "• 都不就绪就阻塞",
            "• default 立即返回"
        ],
        color=Colors.CHART_1
    )
    
    gen.add_card(
        slide, Inches(5.7), Inches(4), Inches(3.8), Inches(1.8),
        "🎯 典型场景", [
            "• 超时控制",
            "• 多源数据合并",
            "• 取消信号"
        ],
        color=Colors.SUCCESS
    )
    
    # 3.2 超时控制
    slide = gen.add_content_slide("03 · Select 多路复用", "3.2 超时控制", page_num)
    page_num += 1
    
    code_lines = [
        (1, "func fetchWithTimeout(url string) error {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("fetchWithTimeout", Colors.FUNCTION),
            ("(url ", Colors.TEXT_LIGHT), ("string", Colors.TYPE), (") ", Colors.TEXT_LIGHT),
            ("error", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (2, "    result := make(chan []byte)", [
            ("    result := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" []", Colors.TEXT_LIGHT), ("byte", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (3, "", [("", Colors.TEXT_LIGHT)]),
        (4, "    go func() {", [
            ("    ", Colors.TEXT_LIGHT), ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (5, "        data := fetch(url)", [
            ("        data := ", Colors.TEXT_LIGHT), ("fetch", Colors.FUNCTION),
            ("(url)", Colors.TEXT_LIGHT)
        ]),
        (6, "        result <- data", [("        result <- data", Colors.TEXT_LIGHT)]),
        (7, "    }()", [("    }()", Colors.TEXT_LIGHT)]),
        (8, "", [("", Colors.TEXT_LIGHT)]),
        (9, "    select {", [("    ", Colors.TEXT_LIGHT), ("select", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)]),
        (10, "    case data := <-result:", [
            ("    ", Colors.TEXT_LIGHT), ("case", Colors.KEYWORD), (" data := <-result:", Colors.TEXT_LIGHT)
        ]),
        (11, "        process(data)", [
            ("        ", Colors.TEXT_LIGHT), ("process", Colors.FUNCTION), ("(data)", Colors.TEXT_LIGHT)
        ]),
        (12, "        return nil", [
            ("        ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("nil", Colors.KEYWORD)
        ]),
        (13, "", [("", Colors.TEXT_LIGHT)]),
        (14, "    case <-time.After(3 * time.Second):", [
            ("    ", Colors.TEXT_LIGHT), ("case", Colors.KEYWORD), (" <-time", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("After", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("3", Colors.NUMBER), (" * time", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Second", Colors.TYPE), ("):", Colors.TEXT_LIGHT)
        ]),
        (15, "        return ErrTimeout", [
            ("        ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" ErrTimeout", Colors.TEXT_LIGHT)
        ]),
        (16, "    }", [("    }", Colors.TEXT_LIGHT)]),
        (17, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="超时控制实现"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.5),
        "⏱️ 超时模式", [
            "• time.After 创建定时器",
            "• select 等待双分支",
            "• 先到先得",
            "",
            "⚠️ 超时≠取消后端任务"
        ],
        color=Colors.WARNING
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.5), Inches(3.3), Inches(1.8),
        "🎯 实战应用", [
            "• HTTP 请求超时",
            "• 数据库查询超时",
            "• 交易所 API 调用"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第四章：Context上下文 ==========
    gen.add_section_slide("04", "Context 上下文")
    page_num += 2
    
    # 4.1 Context取消
    slide = gen.add_content_slide("04 · Context 上下文", "4.1 Context 取消", page_num)
    page_num += 1
    
    code_lines = [
        (1, "ctx, cancel := context.WithCancel(", [
            ("ctx, cancel := context", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("WithCancel", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT)
        ]),
        (2, "    context.Background(),", [
            ("    context", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Background", Colors.FUNCTION), ("(),", Colors.TEXT_LIGHT)
        ]),
        (3, ")", [(")", Colors.TEXT_LIGHT)]),
        (4, "defer cancel()  // 确保释放资源", [
            ("defer", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("cancel", Colors.FUNCTION),
            ("()  ", Colors.TEXT_LIGHT), ("// 确保释放资源", Colors.COMMENT)
        ]),
        (5, "", [("", Colors.TEXT_LIGHT)]),
        (6, "go func() {", [
            ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (7, "    for {", [
            ("    ", Colors.TEXT_LIGHT), ("for", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)
        ]),
        (8, "        select {", [
            ("        ", Colors.TEXT_LIGHT), ("select", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)
        ]),
        (9, "        case <-ctx.Done():", [
            ("        ", Colors.TEXT_LIGHT), ("case", Colors.KEYWORD), (" <-ctx", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Done", Colors.FUNCTION), ("():", Colors.TEXT_LIGHT)
        ]),
        (10, "            return  // 收到取消信号", [
            ("            ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), ("  ", Colors.TEXT_LIGHT),
            ("// 收到取消信号", Colors.COMMENT)
        ]),
        (11, "        default:", [
            ("        ", Colors.TEXT_LIGHT), ("default", Colors.KEYWORD), (":", Colors.TEXT_LIGHT)
        ]),
        (12, "            doWork()  // 继续工作", [
            ("            ", Colors.TEXT_LIGHT), ("doWork", Colors.FUNCTION), ("()  ", Colors.TEXT_LIGHT),
            ("// 继续工作", Colors.COMMENT)
        ]),
        (13, "        }", [("        }", Colors.TEXT_LIGHT)]),
        (14, "    }", [("    }", Colors.TEXT_LIGHT)]),
        (15, "}()", [("}()", Colors.TEXT_LIGHT)]),
        (16, "", [("", Colors.TEXT_LIGHT)]),
        (17, "time.Sleep(5 * time.Second)", [
            ("time", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Sleep", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("5", Colors.NUMBER), (" * time", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Second", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (18, "cancel()  // 触发取消", [
            ("cancel", Colors.FUNCTION), ("()  ", Colors.TEXT_LIGHT),
            ("// 触发取消", Colors.COMMENT)
        ]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="Context 取消示例"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.2),
        "🎛️ 四种Context", [
            "• Background() — 根节点",
            "• WithCancel() — 取消",
            "• WithTimeout() — 超时",
            "• WithValue() — 传值"
        ],
        color=Colors.CHART_1
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.2), Inches(3.3), Inches(2),
        "✅ 最佳实践", [
            "• 第一个参数传 ctx",
            "• defer cancel()",
            "• 监听 ctx.Done()",
            "• 不存储 Context"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第五章：同步原语 ==========
    gen.add_section_slide("05", "同步原语")
    page_num += 2
    
    # 5.1 Mutex互斥锁
    slide = gen.add_content_slide("05 · 同步原语", "5.1 Mutex 互斥锁", page_num)
    page_num += 1
    
    code_lines = [
        (1, "type Counter struct {", [
            ("type", Colors.KEYWORD), (" Counter ", Colors.TEXT_LIGHT), ("struct", Colors.KEYWORD),
            (" {", Colors.TEXT_LIGHT)
        ]),
        (2, "    mu    sync.Mutex", [
            ("    mu    sync", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Mutex", Colors.TYPE)
        ]),
        (3, "    value int", [
            ("    value ", Colors.TEXT_LIGHT), ("int", Colors.TYPE)
        ]),
        (4, "}", [("}", Colors.TEXT_LIGHT)]),
        (5, "", [("", Colors.TEXT_LIGHT)]),
        (6, "func (c *Counter) Inc() {", [
            ("func", Colors.KEYWORD), (" (c *Counter) ", Colors.TEXT_LIGHT),
            ("Inc", Colors.FUNCTION), ("() {", Colors.TEXT_LIGHT)
        ]),
        (7, "    c.mu.Lock()", [
            ("    c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Lock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (8, "    defer c.mu.Unlock()", [
            ("    ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Unlock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (9, "    c.value++", [("    c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("value++", Colors.TEXT_LIGHT)]),
        (10, "}", [("}", Colors.TEXT_LIGHT)]),
        (11, "", [("", Colors.TEXT_LIGHT)]),
        (12, "func (c *Counter) Get() int {", [
            ("func", Colors.KEYWORD), (" (c *Counter) ", Colors.TEXT_LIGHT),
            ("Get", Colors.FUNCTION), ("() ", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (13, "    c.mu.Lock()", [
            ("    c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Lock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (14, "    defer c.mu.Unlock()", [
            ("    ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Unlock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (15, "    return c.value", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("value", Colors.TEXT_LIGHT)
        ]),
        (16, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5), Inches(4.5),
        title="Mutex 使用示例"
    )
    
    gen.add_card(
        slide, Inches(5.7), Inches(1.8), Inches(3.8), Inches(1.8),
        "🔒 使用规则", [
            "• Lock/Unlock 成对",
            "• defer Unlock 防遗漏",
            "• 不可复制 Mutex",
            "• 不可重入"
        ],
        color=Colors.ERROR
    )
    
    gen.add_card(
        slide, Inches(5.7), Inches(3.8), Inches(3.8), Inches(2.5),
        "⚖️ 选型对比", [
            "• Mutex — 复杂临界区",
            "• atomic — 单个变量",
            "• channel — 传递所有权",
            "",
            "口诀：能 atomic 不 Mutex"
        ],
        color=Colors.INFO
    )
    
    # 5.2 RWMutex读写锁
    slide = gen.add_content_slide("05 · 同步原语", "5.2 RWMutex 读写锁", page_num)
    page_num += 1
    
    code_lines = [
        (1, "type Cache struct {", [
            ("type", Colors.KEYWORD), (" Cache ", Colors.TEXT_LIGHT), ("struct", Colors.KEYWORD),
            (" {", Colors.TEXT_LIGHT)
        ]),
        (2, "    mu   sync.RWMutex", [
            ("    mu   sync", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("RWMutex", Colors.TYPE)
        ]),
        (3, "    data map[string]string", [
            ("    data ", Colors.TEXT_LIGHT), ("map", Colors.KEYWORD), ("[", Colors.TEXT_LIGHT),
            ("string", Colors.TYPE), ("]", Colors.TEXT_LIGHT), ("string", Colors.TYPE)
        ]),
        (4, "}", [("}", Colors.TEXT_LIGHT)]),
        (5, "", [("", Colors.TEXT_LIGHT)]),
        (6, "// 读操作用读锁", [("// 读操作用读锁", Colors.COMMENT)]),
        (7, "func (c *Cache) Get(key string) string {", [
            ("func", Colors.KEYWORD), (" (c *Cache) ", Colors.TEXT_LIGHT),
            ("Get", Colors.FUNCTION), ("(key ", Colors.TEXT_LIGHT), ("string", Colors.TYPE),
            (") ", Colors.TEXT_LIGHT), ("string", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (8, "    c.mu.RLock()", [
            ("    c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("RLock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (9, "    defer c.mu.RUnlock()", [
            ("    ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("RUnlock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (10, "    return c.data[key]", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("data[key]", Colors.TEXT_LIGHT)
        ]),
        (11, "}", [("}", Colors.TEXT_LIGHT)]),
        (12, "", [("", Colors.TEXT_LIGHT)]),
        (13, "// 写操作用写锁", [("// 写操作用写锁", Colors.COMMENT)]),
        (14, "func (c *Cache) Set(k, v string) {", [
            ("func", Colors.KEYWORD), (" (c *Cache) ", Colors.TEXT_LIGHT),
            ("Set", Colors.FUNCTION), ("(k, v ", Colors.TEXT_LIGHT), ("string", Colors.TYPE),
            (") {", Colors.TEXT_LIGHT)
        ]),
        (15, "    c.mu.Lock()", [
            ("    c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Lock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (16, "    defer c.mu.Unlock()", [
            ("    ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("mu", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Unlock", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (17, "    c.data[k] = v", [
            ("    c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("data[k] = v", Colors.TEXT_LIGHT)
        ]),
        (18, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.2), Inches(4.5),
        title="RWMutex 使用示例"
    )
    
    gen.add_card(
        slide, Inches(5.9), Inches(1.8), Inches(3.6), Inches(2),
        "📖 读写分离", [
            "• 读锁：RLock/RUnlock",
            "• 写锁：Lock/Unlock",
            "• 读锁可并发",
            "• 写锁独占",
            "• 写锁优先级高"
        ],
        color=Colors.CHART_2
    )
    
    gen.add_card(
        slide, Inches(5.9), Inches(4), Inches(3.6), Inches(2.3),
        "🎯 适用场景", [
            "✅ 读多写少",
            "✅ 配置缓存",
            "✅ 热点数据",
            "",
            "⚠️ 写多时不如 Mutex"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第六章：并发模式 ==========
    gen.add_section_slide("06", "并发模式")
    page_num += 2
    
    # 6.1 Worker Pool
    slide = gen.add_content_slide("06 · 并发模式", "6.1 Worker Pool", page_num)
    page_num += 1
    
    code_lines = [
        (1, "jobs := make(chan int, 100)", [
            ("jobs := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE),
            (", ", Colors.TEXT_LIGHT), ("100", Colors.NUMBER), (")", Colors.TEXT_LIGHT)
        ]),
        (2, "results := make(chan int, 100)", [
            ("results := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE),
            (", ", Colors.TEXT_LIGHT), ("100", Colors.NUMBER), (")", Colors.TEXT_LIGHT)
        ]),
        (3, "var wg sync.WaitGroup", [
            ("var", Colors.KEYWORD), (" wg sync", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("WaitGroup", Colors.TYPE)
        ]),
        (4, "", [("", Colors.TEXT_LIGHT)]),
        (5, "// 启动3个worker", [("// 启动3个worker", Colors.COMMENT)]),
        (6, "for w := 0; w < 3; w++ {", [
            ("for", Colors.KEYWORD), (" w := ", Colors.TEXT_LIGHT), ("0", Colors.NUMBER),
            ("; w < ", Colors.TEXT_LIGHT), ("3", Colors.NUMBER), ("; w++ {", Colors.TEXT_LIGHT)
        ]),
        (7, "    wg.Add(1)", [
            ("    wg", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Add", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("1", Colors.NUMBER), (")", Colors.TEXT_LIGHT)
        ]),
        (8, "    go func() {", [
            ("    ", Colors.TEXT_LIGHT), ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (9, "        defer wg.Done()", [
            ("        ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" wg", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Done", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (10, "        for j := range jobs {", [
            ("        ", Colors.TEXT_LIGHT), ("for", Colors.KEYWORD), (" j := ", Colors.TEXT_LIGHT),
            ("range", Colors.KEYWORD), (" jobs {", Colors.TEXT_LIGHT)
        ]),
        (11, "            results <- process(j)", [
            ("            results <- ", Colors.TEXT_LIGHT), ("process", Colors.FUNCTION),
            ("(j)", Colors.TEXT_LIGHT)
        ]),
        (12, "        }", [("        }", Colors.TEXT_LIGHT)]),
        (13, "    }()", [("    }()", Colors.TEXT_LIGHT)]),
        (14, "}", [("}", Colors.TEXT_LIGHT)]),
        (15, "", [("", Colors.TEXT_LIGHT)]),
        (16, "// 发送任务", [("// 发送任务", Colors.COMMENT)]),
        (17, "for i := 0; i < 100; i++ {", [
            ("for", Colors.KEYWORD), (" i := ", Colors.TEXT_LIGHT), ("0", Colors.NUMBER),
            ("; i < ", Colors.TEXT_LIGHT), ("100", Colors.NUMBER), ("; i++ {", Colors.TEXT_LIGHT)
        ]),
        (18, "    jobs <- i", [("    jobs <- i", Colors.TEXT_LIGHT)]),
        (19, "}", [("}", Colors.TEXT_LIGHT)]),
        (20, "close(jobs)  // 关键！", [
            ("close", Colors.FUNCTION), ("(jobs)  ", Colors.TEXT_LIGHT),
            ("// 关键！", Colors.COMMENT)
        ]),
        (21, "", [("", Colors.TEXT_LIGHT)]),
        (22, "// 等待完成", [("// 等待完成", Colors.COMMENT)]),
        (23, "go func() {", [
            ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (24, "    wg.Wait()", [
            ("    wg", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Wait", Colors.FUNCTION),
            ("()", Colors.TEXT_LIGHT)
        ]),
        (25, "    close(results)", [
            ("    ", Colors.TEXT_LIGHT), ("close", Colors.FUNCTION), ("(results)", Colors.TEXT_LIGHT)
        ]),
        (26, "}()", [("}()", Colors.TEXT_LIGHT)]),
        (27, "", [("", Colors.TEXT_LIGHT)]),
        (28, "for r := range results {", [
            ("for", Colors.KEYWORD), (" r := ", Colors.TEXT_LIGHT), ("range", Colors.KEYWORD),
            (" results {", Colors.TEXT_LIGHT)
        ]),
        (29, "    fmt.Println(r)", [
            ("    fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(r)", Colors.TEXT_LIGHT)
        ]),
        (30, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="Worker Pool 完整实现"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.5),
        "🔄 执行流程", [
            "1️⃣ 创建 jobs/results",
            "2️⃣ 启动 N 个 worker",
            "3️⃣ 发送任务 → close(jobs)",
            "4️⃣ Wait → close(results)",
            "5️⃣ range 收集结果"
        ],
        color=Colors.PRIMARY
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.5), Inches(3.3), Inches(1.8),
        "✅ 关键点", [
            "• 先 close(jobs)",
            "• 后 wg.Wait()",
            "• 再 close(results)"
        ],
        color=Colors.SUCCESS
    )
    
    # 6.2 Pipeline模式
    slide = gen.add_content_slide("06 · 并发模式", "6.2 Pipeline 流水线", page_num)
    page_num += 1
    
    code_lines = [
        (1, "// 阶段1: 生成数字", [("// 阶段1: 生成数字", Colors.COMMENT)]),
        (2, "func gen(nums ...int) <-chan int {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("gen", Colors.FUNCTION),
            ("(nums ...", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (") <-", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (3, "    out := make(chan int)", [
            ("    out := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (4, "    go func() {", [
            ("    ", Colors.TEXT_LIGHT), ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (5, "        defer close(out)", [
            ("        ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("close", Colors.FUNCTION), ("(out)", Colors.TEXT_LIGHT)
        ]),
        (6, "        for _, n := range nums {", [
            ("        ", Colors.TEXT_LIGHT), ("for", Colors.KEYWORD), (" _, n := ", Colors.TEXT_LIGHT),
            ("range", Colors.KEYWORD), (" nums {", Colors.TEXT_LIGHT)
        ]),
        (7, "            out <- n", [("            out <- n", Colors.TEXT_LIGHT)]),
        (8, "        }", [("        }", Colors.TEXT_LIGHT)]),
        (9, "    }()", [("    }()", Colors.TEXT_LIGHT)]),
        (10, "    return out", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" out", Colors.TEXT_LIGHT)
        ]),
        (11, "}", [("}", Colors.TEXT_LIGHT)]),
        (12, "", [("", Colors.TEXT_LIGHT)]),
        (13, "// 阶段2: 平方", [("// 阶段2: 平方", Colors.COMMENT)]),
        (14, "func sq(in <-chan int) <-chan int {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("sq", Colors.FUNCTION),
            ("(in <-", Colors.TEXT_LIGHT), ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("int", Colors.TYPE), (") <-", Colors.TEXT_LIGHT), ("chan", Colors.KEYWORD),
            (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (15, "    out := make(chan int)", [
            ("    out := ", Colors.TEXT_LIGHT), ("make", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("chan", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("int", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (16, "    go func() {", [
            ("    ", Colors.TEXT_LIGHT), ("go", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("func", Colors.KEYWORD), ("() {", Colors.TEXT_LIGHT)
        ]),
        (17, "        defer close(out)", [
            ("        ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("close", Colors.FUNCTION), ("(out)", Colors.TEXT_LIGHT)
        ]),
        (18, "        for n := range in {", [
            ("        ", Colors.TEXT_LIGHT), ("for", Colors.KEYWORD), (" n := ", Colors.TEXT_LIGHT),
            ("range", Colors.KEYWORD), (" in {", Colors.TEXT_LIGHT)
        ]),
        (19, "            out <- n * n", [
            ("            out <- n * n", Colors.TEXT_LIGHT)
        ]),
        (20, "        }", [("        }", Colors.TEXT_LIGHT)]),
        (21, "    }()", [("    }()", Colors.TEXT_LIGHT)]),
        (22, "    return out", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" out", Colors.TEXT_LIGHT)
        ]),
        (23, "}", [("}", Colors.TEXT_LIGHT)]),
        (24, "", [("", Colors.TEXT_LIGHT)]),
        (25, "// 使用", [("// 使用", Colors.COMMENT)]),
        (26, "for n := range sq(gen(1, 2, 3)) {", [
            ("for", Colors.KEYWORD), (" n := ", Colors.TEXT_LIGHT), ("range", Colors.KEYWORD),
            (" ", Colors.TEXT_LIGHT), ("sq", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT),
            ("gen", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT), ("1", Colors.NUMBER),
            (", ", Colors.TEXT_LIGHT), ("2", Colors.NUMBER), (", ", Colors.TEXT_LIGHT),
            ("3", Colors.NUMBER), (")) {", Colors.TEXT_LIGHT)
        ]),
        (27, "    fmt.Println(n)", [
            ("    fmt", Colors.TYPE), (".", Colors.TEXT_LIGHT), ("Println", Colors.FUNCTION),
            ("(n)", Colors.TEXT_LIGHT)
        ]),
        (28, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="Pipeline 模式实现"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2),
        "🔀 Pipeline 特点", [
            "• 阶段间用 channel 连接",
            "• 每个阶段独立 goroutine",
            "• 数据单向流动",
            "• 可组合、可扩展"
        ],
        color=Colors.CHART_3
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4), Inches(3.3), Inches(2.3),
        "🎯 应用场景", [
            "• 数据处理流水线",
            "• ETL 数据转换",
            "• 图像处理管道",
            "",
            "✅ 每阶段职责单一"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第七章：高级工具 ==========
    gen.add_section_slide("07", "高级工具")
    page_num += 2
    
    # 7.1 errgroup
    slide = gen.add_content_slide("07 · 高级工具", "7.1 errgroup 错误处理", page_num)
    page_num += 1
    
    code_lines = [
        (1, "import \"golang.org/x/sync/errgroup\"", [
            ("import", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("\"golang.org/x/sync/errgroup\"", Colors.STRING)
        ]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "func fetchAll(urls []string) error {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("fetchAll", Colors.FUNCTION),
            ("(urls []", Colors.TEXT_LIGHT), ("string", Colors.TYPE), (") ", Colors.TEXT_LIGHT),
            ("error", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (4, "    g := new(errgroup.Group)", [
            ("    g := ", Colors.TEXT_LIGHT), ("new", Colors.FUNCTION), ("(errgroup", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Group", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (5, "", [("", Colors.TEXT_LIGHT)]),
        (6, "    for _, url := range urls {", [
            ("    ", Colors.TEXT_LIGHT), ("for", Colors.KEYWORD), (" _, url := ", Colors.TEXT_LIGHT),
            ("range", Colors.KEYWORD), (" urls {", Colors.TEXT_LIGHT)
        ]),
        (7, "        url := url  // 捕获变量", [
            ("        url := url  ", Colors.TEXT_LIGHT), ("// 捕获变量", Colors.COMMENT)
        ]),
        (8, "        g.Go(func() error {", [
            ("        g", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Go", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("func", Colors.KEYWORD), ("() ", Colors.TEXT_LIGHT),
            ("error", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (9, "            resp, err := http.Get(url)", [
            ("            resp, err := http", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Get", Colors.FUNCTION), ("(url)", Colors.TEXT_LIGHT)
        ]),
        (10, "            if err != nil {", [
            ("            ", Colors.TEXT_LIGHT), ("if", Colors.KEYWORD), (" err != ", Colors.TEXT_LIGHT),
            ("nil", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)
        ]),
        (11, "                return err  // 第一个错误", [
            ("                ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" err  ", Colors.TEXT_LIGHT),
            ("// 第一个错误", Colors.COMMENT)
        ]),
        (12, "            }", [("            }", Colors.TEXT_LIGHT)]),
        (13, "            defer resp.Body.Close()", [
            ("            ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" resp", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Body", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Close", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (14, "            return nil", [
            ("            ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("nil", Colors.KEYWORD)
        ]),
        (15, "        })", [("        })", Colors.TEXT_LIGHT)]),
        (16, "    }", [("    }", Colors.TEXT_LIGHT)]),
        (17, "", [("", Colors.TEXT_LIGHT)]),
        (18, "    return g.Wait()  // 等待并返回错误", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" g", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Wait", Colors.FUNCTION), ("()  ", Colors.TEXT_LIGHT),
            ("// 等待并返回错误", Colors.COMMENT)
        ]),
        (19, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="errgroup 使用示例"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.2),
        "🎁 核心功能", [
            "• Go() 启动任务",
            "• Wait() 等待完成",
            "• 返回第一个错误",
            "• 自动传播取消",
            "• 简化错误处理"
        ],
        color=Colors.INFO
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.2), Inches(3.3), Inches(2),
        "✅ 最佳实践", [
            "• 替代 WaitGroup",
            "• 需要错误处理时",
            "• 配合 Context 使用",
            "• WithContext() 获取 ctx"
        ],
        color=Colors.SUCCESS
    )
    
    # 7.2 限流器
    slide = gen.add_content_slide("07 · 高级工具", "7.2 Rate Limiter 限流", page_num)
    page_num += 1
    
    code_lines = [
        (1, "import \"golang.org/x/time/rate\"", [
            ("import", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("\"golang.org/x/time/rate\"", Colors.STRING)
        ]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "// 10 req/s, 突发容量20", [("// 10 req/s, 突发容量20", Colors.COMMENT)]),
        (4, "limiter := rate.NewLimiter(10, 20)", [
            ("limiter := rate", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("NewLimiter", Colors.FUNCTION),
            ("(", Colors.TEXT_LIGHT), ("10", Colors.NUMBER), (", ", Colors.TEXT_LIGHT),
            ("20", Colors.NUMBER), (")", Colors.TEXT_LIGHT)
        ]),
        (5, "", [("", Colors.TEXT_LIGHT)]),
        (6, "func callAPI(ctx context.Context) {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("callAPI", Colors.FUNCTION),
            ("(ctx context", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Context", Colors.TYPE),
            (") {", Colors.TEXT_LIGHT)
        ]),
        (7, "    // 等待许可", [("    ", Colors.TEXT_LIGHT), ("// 等待许可", Colors.COMMENT)]),
        (8, "    if err := limiter.Wait(ctx); err != nil {", [
            ("    ", Colors.TEXT_LIGHT), ("if", Colors.KEYWORD), (" err := limiter", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Wait", Colors.FUNCTION), ("(ctx); err != ", Colors.TEXT_LIGHT),
            ("nil", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)
        ]),
        (9, "        return  // 上下文取消", [
            ("        ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), ("  ", Colors.TEXT_LIGHT),
            ("// 上下文取消", Colors.COMMENT)
        ]),
        (10, "    }", [("    }", Colors.TEXT_LIGHT)]),
        (11, "", [("", Colors.TEXT_LIGHT)]),
        (12, "    // 执行API调用", [("    ", Colors.TEXT_LIGHT), ("// 执行API调用", Colors.COMMENT)]),
        (13, "    doHTTPRequest()", [
            ("    ", Colors.TEXT_LIGHT), ("doHTTPRequest", Colors.FUNCTION), ("()", Colors.TEXT_LIGHT)
        ]),
        (14, "}", [("}", Colors.TEXT_LIGHT)]),
        (15, "", [("", Colors.TEXT_LIGHT)]),
        (16, "// 非阻塞检查", [("// 非阻塞检查", Colors.COMMENT)]),
        (17, "if limiter.Allow() {", [
            ("if", Colors.KEYWORD), (" limiter", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Allow", Colors.FUNCTION), ("() {", Colors.TEXT_LIGHT)
        ]),
        (18, "    doWork()  // 有配额才执行", [
            ("    ", Colors.TEXT_LIGHT), ("doWork", Colors.FUNCTION), ("()  ", Colors.TEXT_LIGHT),
            ("// 有配额才执行", Colors.COMMENT)
        ]),
        (19, "} else {", [("} ", Colors.TEXT_LIGHT), ("else", Colors.KEYWORD), (" {", Colors.TEXT_LIGHT)]),
        (20, "    // 拒绝请求", [("    ", Colors.TEXT_LIGHT), ("// 拒绝请求", Colors.COMMENT)]),
        (21, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="Rate Limiter 使用示例"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.5),
        "🚦 三种模式", [
            "• Wait(ctx) — 阻塞等待",
            "• Allow() — 非阻塞检查",
            "• Reserve() — 预定令牌",
            "",
            "参数: (速率, 突发容量)"
        ],
        color=Colors.WARNING
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.5), Inches(3.3), Inches(1.8),
        "🎯 实战场景", [
            "• 交易所 API 限频",
            "• 防止流量突刺",
            "• 平滑请求速率"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 第八章：性能优化 ==========
    gen.add_section_slide("08", "性能优化")
    page_num += 2
    
    # 8.1 sync.Pool
    slide = gen.add_content_slide("08 · 性能优化", "8.1 sync.Pool 对象复用", page_num)
    page_num += 1
    
    code_lines = [
        (1, "var bufferPool = sync.Pool{", [
            ("var", Colors.KEYWORD), (" bufferPool = sync", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Pool", Colors.TYPE), ("{", Colors.TEXT_LIGHT)
        ]),
        (2, "    New: func() interface{} {", [
            ("    New", Colors.TEXT_LIGHT), (": ", Colors.TEXT_LIGHT), ("func", Colors.KEYWORD),
            ("() ", Colors.TEXT_LIGHT), ("interface", Colors.KEYWORD), ("{} {", Colors.TEXT_LIGHT)
        ]),
        (3, "        return new(bytes.Buffer)", [
            ("        ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT),
            ("new", Colors.FUNCTION), ("(bytes", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Buffer", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (4, "    },", [("    },", Colors.TEXT_LIGHT)]),
        (5, "}", [("}", Colors.TEXT_LIGHT)]),
        (6, "", [("", Colors.TEXT_LIGHT)]),
        (7, "func processData(data string) {", [
            ("func", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("processData", Colors.FUNCTION),
            ("(data ", Colors.TEXT_LIGHT), ("string", Colors.TYPE), (") {", Colors.TEXT_LIGHT)
        ]),
        (8, "    // 从池中获取", [("    ", Colors.TEXT_LIGHT), ("// 从池中获取", Colors.COMMENT)]),
        (9, "    buf := bufferPool.Get().(*bytes.Buffer)", [
            ("    buf := bufferPool", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Get", Colors.FUNCTION),
            ("().(", Colors.TEXT_LIGHT), ("*bytes", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT),
            ("Buffer", Colors.TYPE), (")", Colors.TEXT_LIGHT)
        ]),
        (10, "    defer bufferPool.Put(buf)", [
            ("    ", Colors.TEXT_LIGHT), ("defer", Colors.KEYWORD), (" bufferPool", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("Put", Colors.FUNCTION), ("(buf)", Colors.TEXT_LIGHT)
        ]),
        (11, "", [("", Colors.TEXT_LIGHT)]),
        (12, "    // 重置状态（关键！）", [
            ("    ", Colors.TEXT_LIGHT), ("// 重置状态（关键！）", Colors.COMMENT)
        ]),
        (13, "    buf.Reset()", [
            ("    buf", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("Reset", Colors.FUNCTION),
            ("()", Colors.TEXT_LIGHT)
        ]),
        (14, "", [("", Colors.TEXT_LIGHT)]),
        (15, "    // 使用buffer", [("    ", Colors.TEXT_LIGHT), ("// 使用buffer", Colors.COMMENT)]),
        (16, "    buf.WriteString(data)", [
            ("    buf", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("WriteString", Colors.FUNCTION),
            ("(data)", Colors.TEXT_LIGHT)
        ]),
        (17, "    result := buf.String()", [
            ("    result := buf", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("String", Colors.FUNCTION),
            ("()", Colors.TEXT_LIGHT)
        ]),
        (18, "    // ... 处理result", [
            ("    ", Colors.TEXT_LIGHT), ("// ... 处理result", Colors.COMMENT)
        ]),
        (19, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="sync.Pool 使用示例"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.2),
        "♻️ 复用机制", [
            "• Get() 获取对象",
            "• Put() 归还对象",
            "• New() 创建新对象",
            "• GC 时可能清空",
            "• 必须重置状态"
        ],
        color=Colors.CHART_4
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.2), Inches(3.3), Inches(2),
        "🎯 适用场景", [
            "• 高频临时对象",
            "• bytes.Buffer",
            "• []byte 切片",
            "• 减少 GC 压力"
        ],
        color=Colors.SUCCESS
    )
    
    # 8.2 atomic原子操作
    slide = gen.add_content_slide("08 · 性能优化", "8.2 Atomic 原子操作", page_num)
    page_num += 1
    
    code_lines = [
        (1, "import \"sync/atomic\"", [
            ("import", Colors.KEYWORD), (" ", Colors.TEXT_LIGHT), ("\"sync/atomic\"", Colors.STRING)
        ]),
        (2, "", [("", Colors.TEXT_LIGHT)]),
        (3, "type Counter struct {", [
            ("type", Colors.KEYWORD), (" Counter ", Colors.TEXT_LIGHT), ("struct", Colors.KEYWORD),
            (" {", Colors.TEXT_LIGHT)
        ]),
        (4, "    value int64  // 必须64位对齐", [
            ("    value ", Colors.TEXT_LIGHT), ("int64", Colors.TYPE), ("  ", Colors.TEXT_LIGHT),
            ("// 必须64位对齐", Colors.COMMENT)
        ]),
        (5, "}", [("}", Colors.TEXT_LIGHT)]),
        (6, "", [("", Colors.TEXT_LIGHT)]),
        (7, "func (c *Counter) Inc() {", [
            ("func", Colors.KEYWORD), (" (c *Counter) ", Colors.TEXT_LIGHT),
            ("Inc", Colors.FUNCTION), ("() {", Colors.TEXT_LIGHT)
        ]),
        (8, "    atomic.AddInt64(&c.value, 1)", [
            ("    atomic", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("AddInt64", Colors.FUNCTION),
            ("(&c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("value, ", Colors.TEXT_LIGHT),
            ("1", Colors.NUMBER), (")", Colors.TEXT_LIGHT)
        ]),
        (9, "}", [("}", Colors.TEXT_LIGHT)]),
        (10, "", [("", Colors.TEXT_LIGHT)]),
        (11, "func (c *Counter) Get() int64 {", [
            ("func", Colors.KEYWORD), (" (c *Counter) ", Colors.TEXT_LIGHT),
            ("Get", Colors.FUNCTION), ("() ", Colors.TEXT_LIGHT), ("int64", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (12, "    return atomic.LoadInt64(&c.value)", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" atomic", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("LoadInt64", Colors.FUNCTION), ("(&c", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("value)", Colors.TEXT_LIGHT)
        ]),
        (13, "}", [("}", Colors.TEXT_LIGHT)]),
        (14, "", [("", Colors.TEXT_LIGHT)]),
        (15, "// CAS 乐观锁", [("// CAS 乐观锁", Colors.COMMENT)]),
        (16, "func (c *Counter) CAS(old, new int64) bool {", [
            ("func", Colors.KEYWORD), (" (c *Counter) ", Colors.TEXT_LIGHT),
            ("CAS", Colors.FUNCTION), ("(old, new ", Colors.TEXT_LIGHT), ("int64", Colors.TYPE),
            (") ", Colors.TEXT_LIGHT), ("bool", Colors.TYPE), (" {", Colors.TEXT_LIGHT)
        ]),
        (17, "    return atomic.CompareAndSwapInt64(", [
            ("    ", Colors.TEXT_LIGHT), ("return", Colors.KEYWORD), (" atomic", Colors.TEXT_LIGHT),
            (".", Colors.TEXT_LIGHT), ("CompareAndSwapInt64", Colors.FUNCTION), ("(", Colors.TEXT_LIGHT)
        ]),
        (18, "        &c.value, old, new,", [
            ("        &c", Colors.TEXT_LIGHT), (".", Colors.TEXT_LIGHT), ("value, old, new,", Colors.TEXT_LIGHT)
        ]),
        (19, "    )", [("    )", Colors.TEXT_LIGHT)]),
        (20, "}", [("}", Colors.TEXT_LIGHT)]),
    ]
    
    gen.add_code_block(
        slide, code_lines,
        Inches(0.5), Inches(1.8), Inches(5.5), Inches(4.5),
        title="Atomic 原子操作示例"
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(1.8), Inches(3.3), Inches(2.5),
        "⚛️ 原子操作", [
            "• AddInt64 — 加法",
            "• LoadInt64 — 读取",
            "• StoreInt64 — 写入",
            "• SwapInt64 — 交换",
            "• CompareAndSwap — CAS"
        ],
        color=Colors.CHART_1
    )
    
    gen.add_card(
        slide, Inches(6.2), Inches(4.5), Inches(3.3), Inches(1.8),
        "✅ 性能优势", [
            "• 无锁开销",
            "• 单变量操作",
            "• 比 Mutex 快 10x+"
        ],
        color=Colors.SUCCESS
    )
    
    # ========== 总结页 ==========
    slide = gen.add_content_slide("总结", "Go 并发编程知识图谱", page_num)
    page_num += 1
    
    # Goroutine生命周期
    gen.add_flow_chart_goroutine_lifecycle(slide, Inches(0.5), Inches(2))
    
    # 要点总结
    gen.add_card(
        slide, Inches(0.5), Inches(3.2), Inches(4.5), Inches(3),
        "🎯 核心知识点", [
            "✅ Goroutine — 轻量级协程",
            "✅ Channel — 通信机制",
            "✅ Select — 多路复用",
            "✅ Context — 取消控制",
            "✅ Mutex — 互斥保护",
            "✅ WaitGroup — 等待同步",
            "✅ Worker Pool — 并发限流",
            "✅ errgroup — 错误处理",
            "✅ atomic — 原子操作",
            "✅ sync.Pool — 对象复用"
        ],
        color=Colors.PRIMARY
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(2), Inches(4.3), Inches(2),
        "💡 设计原则", [
            "• 通过通信来共享内存",
            "• 不要通过共享内存来通信",
            "• 能 atomic 不 Mutex",
            "• 能 channel 不锁",
            "• 优雅停机是必选项"
        ],
        color=Colors.INFO
    )
    
    gen.add_card(
        slide, Inches(5.2), Inches(4.2), Inches(4.3), Inches(2),
        "⚠️ 常见陷阱", [
            "• 忘记等待 goroutine",
            "• 忘记 close channel",
            "• 接收方 close",
            "• 循环变量捕获",
            "• goroutine 泄漏"
        ],
        color=Colors.ERROR
    )
    
    # ========== 结束页 ==========
    slide = gen.prs.slides.add_slide(gen.prs.slide_layouts[6])
    gen.add_gradient_background(slide)
    
    # 感谢文字
    thanks_box = slide.shapes.add_textbox(
        Inches(1), Inches(3),
        Inches(8), Inches(1.5)
    )
    tf = thanks_box.text_frame
    tf.text = "感谢观看\nThank You"
    p = tf.paragraphs[0]
    p.font.size = Pt(48)
    p.font.bold = True
    p.font.color.rgb = Colors.PRIMARY_DARK
    p.font.name = '微软雅黑'
    p.alignment = PP_ALIGN.CENTER
    
    # 保存文件
    output_path = "/Users/cluo/code/go/src/go-study/03-go-concurrency/Go并发编程课程.pptx"
    gen.save(output_path)
    
    print(f"\n{'='*60}")
    print(f"🎉 顶级PPT生成完成！")
    print(f"{'='*60}")
    print(f"📄 文件位置: {output_path}")
    print(f"📊 总页数: {page_num}")
    print(f"{'='*60}")
    print(f"\n✨ 主要特性:")
    print(f"  • 深色代码主题 (VS Code Dark)")
    print(f"  • 精确语法高亮")
    print(f"  • 专业渐变配色")
    print(f"  • 圆角卡片设计")
    print(f"  • 流程图可视化")
    print(f"  • 完整知识覆盖 (23个关卡)")
    print(f"{'='*60}\n")


if __name__ == "__main__":
    generate_premium_ppt()
