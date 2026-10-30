#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
企业级 Go 并发编程课程 PPT 生成器
使用 python-pptx 库生成专业级别的教学演示文档
"""

from pptx import Presentation
from pptx.util import Inches, Pt
from pptx.enum.text import PP_ALIGN, MSO_ANCHOR
from pptx.enum.shapes import MSO_SHAPE
from pptx.dml.color import RGBColor
from pptx.enum.dml import MSO_LINE_DASH_STYLE

# 配色方案 - 专业深蓝商务风格
COLOR_PRIMARY = RGBColor(25, 55, 109)      # 深蓝色
COLOR_SECONDARY = RGBColor(52, 152, 219)   # 亮蓝色
COLOR_ACCENT = RGBColor(41, 128, 185)      # 中蓝色
COLOR_TEXT = RGBColor(44, 62, 80)          # 深灰色文本
COLOR_BG_LIGHT = RGBColor(236, 240, 241)   # 浅灰背景
COLOR_SUCCESS = RGBColor(39, 174, 96)      # 绿色
COLOR_WARNING = RGBColor(230, 126, 34)     # 橙色
COLOR_DANGER = RGBColor(231, 76, 60)       # 红色

# 代码语法高亮颜色
CODE_KEYWORD = RGBColor(215, 58, 73)       # 关键字-红色
CODE_STRING = RGBColor(106, 135, 89)       # 字符串-绿色
CODE_COMMENT = RGBColor(128, 128, 128)     # 注释-灰色
CODE_FUNCTION = RGBColor(121, 93, 163)     # 函数-紫色
CODE_NUMBER = RGBColor(104, 151, 187)      # 数字-蓝色

def create_presentation():
    """创建演示文稿"""
    prs = Presentation()
    prs.slide_width = Inches(10)
    prs.slide_height = Inches(7.5)
    
    return prs

def add_title_slide(prs):
    """添加封面页"""
    slide = prs.slides.add_slide(prs.slide_layouts[6])  # 空白布局
    
    # 深蓝色背景矩形
    bg = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE, 0, 0, 
        prs.slide_width, prs.slide_height
    )
    bg.fill.solid()
    bg.fill.fore_color.rgb = COLOR_PRIMARY
    bg.line.fill.background()
    
    # 装饰条
    accent_bar = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE, 
        Inches(0), Inches(3), 
        prs.slide_width, Inches(0.3)
    )
    accent_bar.fill.solid()
    accent_bar.fill.fore_color.rgb = COLOR_SECONDARY
    accent_bar.line.fill.background()
    
    # 主标题
    title_box = slide.shapes.add_textbox(
        Inches(1), Inches(2), Inches(8), Inches(0.8)
    )
    title_frame = title_box.text_frame
    title_frame.text = "Go 并发编程"
    p = title_frame.paragraphs[0]
    p.font.size = Pt(54)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    
    # 副标题
    subtitle_box = slide.shapes.add_textbox(
        Inches(1), Inches(3.5), Inches(8), Inches(0.6)
    )
    subtitle_frame = subtitle_box.text_frame
    subtitle_frame.text = "从基础到实战：掌握 Goroutines、Channels 与并发模式"
    p = subtitle_frame.paragraphs[0]
    p.font.size = Pt(20)
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    
    # 讲师信息区
    info_box = slide.shapes.add_textbox(
        Inches(1), Inches(5.5), Inches(8), Inches(1)
    )
    info_frame = info_box.text_frame
    info_frame.text = "企业级 Go 开发课程\n适用于中高级开发者"
    for paragraph in info_frame.paragraphs:
        paragraph.font.size = Pt(16)
        paragraph.font.color.rgb = RGBColor(189, 195, 199)
        paragraph.alignment = PP_ALIGN.CENTER
    
    return slide

def add_section_header(prs, title_text, subtitle_text=""):
    """添加章节标题页"""
    slide = prs.slides.add_slide(prs.slide_layouts[6])
    
    # 渐变背景
    bg = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE, 0, 0, 
        prs.slide_width, prs.slide_height
    )
    bg.fill.gradient()
    bg.fill.gradient_angle = 90
    bg.fill.gradient_stops[0].color.rgb = COLOR_PRIMARY
    bg.fill.gradient_stops[1].color.rgb = COLOR_ACCENT
    bg.line.fill.background()
    
    # 主标题
    title_box = slide.shapes.add_textbox(
        Inches(1), Inches(2.5), Inches(8), Inches(1)
    )
    title_frame = title_box.text_frame
    title_frame.text = title_text
    p = title_frame.paragraphs[0]
    p.font.size = Pt(48)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    
    # 副标题
    if subtitle_text:
        subtitle_box = slide.shapes.add_textbox(
            Inches(1), Inches(3.8), Inches(8), Inches(0.6)
        )
        subtitle_frame = subtitle_box.text_frame
        subtitle_frame.text = subtitle_text
        p = subtitle_frame.paragraphs[0]
        p.font.size = Pt(20)
        p.font.color.rgb = RGBColor(236, 240, 241)
        p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_content_slide(prs, title_text):
    """添加内容页（带标题栏）"""
    slide = prs.slides.add_slide(prs.slide_layouts[6])
    
    # 白色背景
    bg = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE, 0, 0, 
        prs.slide_width, prs.slide_height
    )
    bg.fill.solid()
    bg.fill.fore_color.rgb = RGBColor(255, 255, 255)
    bg.line.fill.background()
    
    # 标题栏背景
    header = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE, 
        0, 0, prs.slide_width, Inches(0.9)
    )
    header.fill.solid()
    header.fill.fore_color.rgb = COLOR_PRIMARY
    header.line.fill.background()
    
    # 标题文字
    title_box = slide.shapes.add_textbox(
        Inches(0.5), Inches(0.15), Inches(9), Inches(0.6)
    )
    title_frame = title_box.text_frame
    title_frame.text = title_text
    p = title_frame.paragraphs[0]
    p.font.size = Pt(28)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    
    return slide

def add_table_of_contents(prs):
    """添加目录页"""
    slide = add_content_slide(prs, "课程大纲")
    
    chapters = [
        ("第一章", "Go 并发基础", "并发模型、调度器、Goroutine"),
        ("第二章", "Goroutines 深入", "创建、同步、常见陷阱"),
        ("第三章", "Channels 通信", "无缓冲、有缓冲、关闭与遍历"),
        ("第四章", "Select 多路复用", "多通道操作、超时、非阻塞"),
        ("第五章", "Context 上下文", "取消传播、超时控制、值传递"),
        ("第六章", "同步原语", "Mutex、RWMutex、Once、Pool、Atomic"),
        ("第七章", "并发模式", "Worker Pool、Pipeline、Fan-out/in"),
        ("第八章", "常见陷阱", "竞态条件、Goroutine 泄漏、死锁"),
        ("第九章", "最佳实践", "设计原则、性能优化、调试工具"),
    ]
    
    left_margin = Inches(1)
    top_start = Inches(1.3)
    line_height = Inches(0.55)
    
    for i, (chapter_num, chapter_title, chapter_desc) in enumerate(chapters):
        y_pos = top_start + i * line_height
        
        # 章节编号背景圆角矩形
        num_box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            left_margin, y_pos,
            Inches(0.8), Inches(0.4)
        )
        num_box.fill.solid()
        num_box.fill.fore_color.rgb = COLOR_SECONDARY
        num_box.line.fill.background()
        
        # 章节编号
        num_text = num_box.text_frame
        num_text.text = chapter_num
        p = num_text.paragraphs[0]
        p.font.size = Pt(12)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        num_text.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 章节标题
        title_box = slide.shapes.add_textbox(
            left_margin + Inches(1), y_pos,
            Inches(3), Inches(0.4)
        )
        title_frame = title_box.text_frame
        title_frame.text = chapter_title
        p = title_frame.paragraphs[0]
        p.font.size = Pt(16)
        p.font.bold = True
        p.font.color.rgb = COLOR_TEXT
        title_frame.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 章节描述
        desc_box = slide.shapes.add_textbox(
            left_margin + Inches(4.2), y_pos,
            Inches(4.5), Inches(0.4)
        )
        desc_frame = desc_box.text_frame
        desc_frame.text = chapter_desc
        p = desc_frame.paragraphs[0]
        p.font.size = Pt(13)
        p.font.color.rgb = RGBColor(127, 140, 141)
        desc_frame.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    return slide

def add_concurrency_vs_parallelism(prs):
    """并发 vs 并行概念图"""
    slide = add_content_slide(prs, "并发 vs 并行")
    
    # 并发部分（左侧）
    left_x = Inches(0.8)
    top_y = Inches(1.5)
    
    # 并发标题
    concurrent_title = slide.shapes.add_textbox(
        left_x, top_y, Inches(4), Inches(0.4)
    )
    tf = concurrent_title.text_frame
    tf.text = "并发 (Concurrency)"
    p = tf.paragraphs[0]
    p.font.size = Pt(20)
    p.font.bold = True
    p.font.color.rgb = COLOR_SECONDARY
    p.alignment = PP_ALIGN.CENTER
    
    # 并发描述
    desc1 = slide.shapes.add_textbox(
        left_x, top_y + Inches(0.5), Inches(4), Inches(0.6)
    )
    tf = desc1.text_frame
    tf.text = "单核 CPU 交替执行多个任务\n任务在时间片上切换"
    tf.word_wrap = True
    for paragraph in tf.paragraphs:
        paragraph.font.size = Pt(12)
        paragraph.font.color.rgb = COLOR_TEXT
        paragraph.alignment = PP_ALIGN.CENTER
    
    # 并发示意图 - 单个处理器
    cpu_box = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        left_x + Inches(1.5), top_y + Inches(1.3),
        Inches(1), Inches(0.5)
    )
    cpu_box.fill.solid()
    cpu_box.fill.fore_color.rgb = COLOR_PRIMARY
    cpu_box.line.color.rgb = COLOR_PRIMARY
    cpu_box.line.width = Pt(2)
    tf = cpu_box.text_frame
    tf.text = "CPU"
    p = tf.paragraphs[0]
    p.font.size = Pt(14)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 任务1、2、3 交替
    task_colors = [COLOR_SUCCESS, COLOR_WARNING, COLOR_DANGER]
    task_names = ["任务1", "任务2", "任务3"]
    for i in range(3):
        task = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            left_x + Inches(0.3 + i * 1.2), top_y + Inches(2.1),
            Inches(0.8), Inches(0.35)
        )
        task.fill.solid()
        task.fill.fore_color.rgb = task_colors[i]
        task.line.fill.background()
        tf = task.text_frame
        tf.text = task_names[i]
        p = tf.paragraphs[0]
        p.font.size = Pt(11)
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 并行部分（右侧）
    right_x = Inches(5.2)
    
    # 并行标题
    parallel_title = slide.shapes.add_textbox(
        right_x, top_y, Inches(4), Inches(0.4)
    )
    tf = parallel_title.text_frame
    tf.text = "并行 (Parallelism)"
    p = tf.paragraphs[0]
    p.font.size = Pt(20)
    p.font.bold = True
    p.font.color.rgb = COLOR_ACCENT
    p.alignment = PP_ALIGN.CENTER
    
    # 并行描述
    desc2 = slide.shapes.add_textbox(
        right_x, top_y + Inches(0.5), Inches(4), Inches(0.6)
    )
    tf = desc2.text_frame
    tf.text = "多核 CPU 同时执行多个任务\n真正的同时运行"
    tf.word_wrap = True
    for paragraph in tf.paragraphs:
        paragraph.font.size = Pt(12)
        paragraph.font.color.rgb = COLOR_TEXT
        paragraph.alignment = PP_ALIGN.CENTER
    
    # 并行示意图 - 多个处理器
    for i in range(3):
        cpu_box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            right_x + Inches(0.3 + i * 1.2), top_y + Inches(1.3),
            Inches(1), Inches(0.5)
        )
        cpu_box.fill.solid()
        cpu_box.fill.fore_color.rgb = COLOR_PRIMARY
        cpu_box.line.color.rgb = COLOR_PRIMARY
        cpu_box.line.width = Pt(2)
        tf = cpu_box.text_frame
        tf.text = f"CPU{i+1}"
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 每个CPU对应一个任务
        task = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            right_x + Inches(0.3 + i * 1.2), top_y + Inches(2.1),
            Inches(1), Inches(0.35)
        )
        task.fill.solid()
        task.fill.fore_color.rgb = task_colors[i]
        task.line.fill.background()
        tf = task.text_frame
        tf.text = task_names[i]
        p = tf.paragraphs[0]
        p.font.size = Pt(11)
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 底部关键点
    key_points = slide.shapes.add_textbox(
        Inches(1), Inches(5.8), Inches(8), Inches(1)
    )
    tf = key_points.text_frame
    tf.text = "💡 Go 的并发模型：用 Goroutines 实现并发，运行时自动在多核上实现并行"
    p = tf.paragraphs[0]
    p.font.size = Pt(14)
    p.font.color.rgb = COLOR_ACCENT
    p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_gmp_model(prs):
    """G-M-P 调度器模型图"""
    slide = add_content_slide(prs, "Go 调度器：G-M-P 模型")
    
    # 模型组件说明
    components = [
        ("G - Goroutine", "轻量级协程，用户代码的执行单元", COLOR_SUCCESS),
        ("M - Machine", "操作系统线程，真正执行代码的载体", COLOR_WARNING),
        ("P - Processor", "逻辑处理器，调度上下文（默认 = CPU 核数）", COLOR_ACCENT),
    ]
    
    left_x = Inches(0.8)
    top_y = Inches(1.2)
    
    for i, (title, desc, color) in enumerate(components):
        y_pos = top_y + i * Inches(0.6)
        
        # 组件标题框
        box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            left_x, y_pos, Inches(1.5), Inches(0.4)
        )
        box.fill.solid()
        box.fill.fore_color.rgb = color
        box.line.fill.background()
        tf = box.text_frame
        tf.text = title
        p = tf.paragraphs[0]
        p.font.size = Pt(13)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 组件描述
        desc_box = slide.shapes.add_textbox(
            left_x + Inches(1.7), y_pos, Inches(6), Inches(0.4)
        )
        tf = desc_box.text_frame
        tf.text = desc
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = COLOR_TEXT
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # GMP 关系图
    diagram_top = Inches(3.2)
    
    # 绘制 P (Processor)
    p_box = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(4), diagram_top, Inches(2), Inches(1.2)
    )
    p_box.fill.solid()
    p_box.fill.fore_color.rgb = COLOR_BG_LIGHT
    p_box.line.color.rgb = COLOR_ACCENT
    p_box.line.width = Pt(3)
    tf = p_box.text_frame
    tf.text = "P\n本地队列"
    for paragraph in tf.paragraphs:
        paragraph.font.size = Pt(14)
        paragraph.font.bold = True
        paragraph.font.color.rgb = COLOR_ACCENT
        paragraph.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 绘制 M (Machine)
    m_box = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(4), diagram_top + Inches(1.5), Inches(2), Inches(0.7)
    )
    m_box.fill.solid()
    m_box.fill.fore_color.rgb = COLOR_WARNING
    m_box.line.fill.background()
    tf = m_box.text_frame
    tf.text = "M (OS Thread)"
    p = tf.paragraphs[0]
    p.font.size = Pt(13)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # M 到 P 的箭头
    arrow1 = slide.shapes.add_connector(
        1,  # STRAIGHT
        Inches(5), diagram_top + Inches(1.5),
        Inches(5), diagram_top + Inches(1.2)
    )
    arrow1.line.color.rgb = COLOR_TEXT
    arrow1.line.width = Pt(2)
    arrow1.line.end_arrow_type = 2
    
    # 绘制 G (Goroutines) 在 P 的本地队列
    g_positions = [(4.2, 3.5), (4.6, 3.5), (5.0, 3.5), (5.4, 3.5)]
    for i, (x, y) in enumerate(g_positions):
        g_circle = slide.shapes.add_shape(
            MSO_SHAPE.OVAL,
            Inches(x), Inches(y), Inches(0.35), Inches(0.35)
        )
        g_circle.fill.solid()
        g_circle.fill.fore_color.rgb = COLOR_SUCCESS
        g_circle.line.fill.background()
        tf = g_circle.text_frame
        tf.text = "G"
        p = tf.paragraphs[0]
        p.font.size = Pt(10)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 全局队列
    global_queue = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(1), diagram_top + Inches(0.3), Inches(2), Inches(0.6)
    )
    global_queue.fill.solid()
    global_queue.fill.fore_color.rgb = RGBColor(245, 245, 245)
    global_queue.line.color.rgb = COLOR_TEXT
    global_queue.line.width = Pt(2)
    tf = global_queue.text_frame
    tf.text = "全局 G 队列"
    p = tf.paragraphs[0]
    p.font.size = Pt(12)
    p.font.color.rgb = COLOR_TEXT
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 全局队列到 P 的箭头
    arrow2 = slide.shapes.add_connector(
        1,
        Inches(3), diagram_top + Inches(0.6),
        Inches(4), diagram_top + Inches(0.6)
    )
    arrow2.line.color.rgb = COLOR_TEXT
    arrow2.line.width = Pt(2)
    arrow2.line.dash_style = MSO_LINE_DASH_STYLE.DASH
    arrow2.line.end_arrow_type = 2
    
    # 窃取工作箭头说明
    steal_label = slide.shapes.add_textbox(
        Inches(7), diagram_top + Inches(0.5), Inches(2), Inches(0.5)
    )
    tf = steal_label.text_frame
    tf.text = "工作窃取\n(Work Stealing)"
    for paragraph in tf.paragraphs:
        paragraph.font.size = Pt(11)
        paragraph.font.color.rgb = COLOR_DANGER
        paragraph.alignment = PP_ALIGN.LEFT
    
    # 关键特性
    features = slide.shapes.add_textbox(
        Inches(0.8), Inches(5.8), Inches(8.4), Inches(1)
    )
    tf = features.text_frame
    tf.text = "⚡ 特性：M:N 调度（M 个 Goroutine 映射到 N 个 OS 线程）• 抢占式调度 • 工作窃取负载均衡"
    p = tf.paragraphs[0]
    p.font.size = Pt(12)
    p.font.color.rgb = COLOR_PRIMARY
    p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_goroutine_lifecycle(prs):
    """Goroutine 生命周期流程图"""
    slide = add_content_slide(prs, "Goroutine 生命周期")
    
    # 流程图节点
    states = [
        ("创建", Inches(2), Inches(1.5), COLOR_SUCCESS),
        ("就绪", Inches(2), Inches(2.7), COLOR_SECONDARY),
        ("运行", Inches(5), Inches(2.7), COLOR_WARNING),
        ("阻塞", Inches(8), Inches(2.7), COLOR_DANGER),
        ("完成", Inches(5), Inches(4.5), COLOR_TEXT),
    ]
    
    boxes = {}
    for state_name, x, y, color in states:
        if state_name == "完成":
            shape_type = MSO_SHAPE.ROUNDED_RECTANGLE
            width, height = Inches(1.5), Inches(0.6)
        else:
            shape_type = MSO_SHAPE.ROUNDED_RECTANGLE
            width, height = Inches(1.5), Inches(0.7)
        
        box = slide.shapes.add_shape(shape_type, x, y, width, height)
        box.fill.solid()
        box.fill.fore_color.rgb = color
        box.line.color.rgb = color
        box.line.width = Pt(2)
        
        tf = box.text_frame
        tf.text = state_name
        p = tf.paragraphs[0]
        p.font.size = Pt(16)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        boxes[state_name] = (x + width/2, y + height/2)
    
    # 箭头和标签
    arrows = [
        ("创建", "就绪", "go func()"),
        ("就绪", "运行", "调度器选中"),
        ("运行", "阻塞", "I/O、chan、锁"),
        ("阻塞", "就绪", "事件到达"),
        ("运行", "完成", "return"),
    ]
    
    # 创建 -> 就绪（向下）
    arrow = slide.shapes.add_connector(
        1, Inches(2.75), Inches(2.2), Inches(2.75), Inches(2.7)
    )
    arrow.line.color.rgb = COLOR_TEXT
    arrow.line.width = Pt(2.5)
    arrow.line.end_arrow_type = 2
    label = slide.shapes.add_textbox(Inches(3), Inches(2.3), Inches(1.2), Inches(0.3))
    tf = label.text_frame
    tf.text = "go func()"
    p = tf.paragraphs[0]
    p.font.size = Pt(11)
    p.font.color.rgb = COLOR_PRIMARY
    
    # 就绪 -> 运行（向右）
    arrow = slide.shapes.add_connector(
        1, Inches(3.5), Inches(3.05), Inches(5), Inches(3.05)
    )
    arrow.line.color.rgb = COLOR_TEXT
    arrow.line.width = Pt(2.5)
    arrow.line.end_arrow_type = 2
    label = slide.shapes.add_textbox(Inches(3.7), Inches(2.7), Inches(1.2), Inches(0.3))
    tf = label.text_frame
    tf.text = "调度器选中"
    p = tf.paragraphs[0]
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_PRIMARY
    
    # 运行 -> 阻塞（向右）
    arrow = slide.shapes.add_connector(
        1, Inches(6.5), Inches(3.05), Inches(8), Inches(3.05)
    )
    arrow.line.color.rgb = COLOR_TEXT
    arrow.line.width = Pt(2.5)
    arrow.line.end_arrow_type = 2
    label = slide.shapes.add_textbox(Inches(6.7), Inches(2.7), Inches(1.2), Inches(0.3))
    tf = label.text_frame
    tf.text = "I/O、chan"
    p = tf.paragraphs[0]
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_PRIMARY
    
    # 阻塞 -> 就绪（弧形返回）
    arrow = slide.shapes.add_connector(
        1, Inches(8.75), Inches(3.4), Inches(2.75), Inches(3.4)
    )
    arrow.line.color.rgb = COLOR_SUCCESS
    arrow.line.width = Pt(2.5)
    arrow.line.end_arrow_type = 2
    arrow.line.dash_style = MSO_LINE_DASH_STYLE.DASH
    label = slide.shapes.add_textbox(Inches(5), Inches(3.5), Inches(1.5), Inches(0.3))
    tf = label.text_frame
    tf.text = "事件就绪"
    p = tf.paragraphs[0]
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_SUCCESS
    p.alignment = PP_ALIGN.CENTER
    
    # 运行 -> 完成（向下）
    arrow = slide.shapes.add_connector(
        1, Inches(5.75), Inches(3.4), Inches(5.75), Inches(4.5)
    )
    arrow.line.color.rgb = COLOR_TEXT
    arrow.line.width = Pt(2.5)
    arrow.line.end_arrow_type = 2
    label = slide.shapes.add_textbox(Inches(6), Inches(3.8), Inches(0.8), Inches(0.3))
    tf = label.text_frame
    tf.text = "return"
    p = tf.paragraphs[0]
    p.font.size = Pt(11)
    p.font.color.rgb = COLOR_PRIMARY
    
    # 底部说明
    note = slide.shapes.add_textbox(
        Inches(1), Inches(5.8), Inches(8), Inches(1)
    )
    tf = note.text_frame
    tf.text = "💡 Goroutine 创建成本极低（约 2KB 栈空间），可轻松创建数万个并发协程"
    p = tf.paragraphs[0]
    p.font.size = Pt(13)
    p.font.color.rgb = COLOR_ACCENT
    p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_code_slide(prs, title, code_text, explanation=""):
    """添加代码示例页"""
    slide = add_content_slide(prs, title)
    
    # 代码框背景
    code_bg = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(0.8), Inches(1.3), Inches(8.4), Inches(4)
    )
    code_bg.fill.solid()
    code_bg.fill.fore_color.rgb = RGBColor(40, 44, 52)  # VS Code 暗色背景
    code_bg.line.color.rgb = COLOR_ACCENT
    code_bg.line.width = Pt(1)
    
    # 代码文本
    code_box = slide.shapes.add_textbox(
        Inches(1), Inches(1.5), Inches(8), Inches(3.6)
    )
    tf = code_box.text_frame
    tf.word_wrap = True
    
    # 简单的语法高亮处理
    lines = code_text.split('\n')
    for line in lines:
        p = tf.add_paragraph() if tf.paragraphs[0].text else tf.paragraphs[0]
        p.font.name = 'Consolas'
        p.font.size = Pt(11)
        p.font.color.rgb = RGBColor(171, 178, 191)  # 默认文本颜色
        p.text = line
        p.level = 0
    
    # 说明文字
    if explanation:
        exp_box = slide.shapes.add_textbox(
            Inches(1), Inches(5.5), Inches(8), Inches(1)
        )
        tf = exp_box.text_frame
        tf.text = explanation
        tf.word_wrap = True
        p = tf.paragraphs[0]
        p.font.size = Pt(13)
        p.font.color.rgb = COLOR_PRIMARY
        p.alignment = PP_ALIGN.LEFT
    
    return slide

def add_waitgroup_slide(prs):
    """WaitGroup 使用场景"""
    slide = add_content_slide(prs, "WaitGroup：等待所有 Goroutine 完成")
    
    code = """package main

import (
    "fmt"
    "sync"
    "time"
)

func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done()  // 完成时调用 Done()
    fmt.Printf("Worker %d starting\\n", id)
    time.Sleep(time.Second)
    fmt.Printf("Worker %d done\\n", id)
}

func main() {
    var wg sync.WaitGroup
    
    for i := 1; i <= 3; i++ {
        wg.Add(1)  // 启动前 Add(1)
        go worker(i, &wg)
    }
    
    wg.Wait()  // 等待所有 goroutine 完成
    fmt.Println("All workers done")
}"""
    
    add_code_slide(prs, "WaitGroup 使用模式", code, 
                   "⚠️ 注意：Add() 必须在 go 语句之前调用，否则可能产生竞态条件")
    
    return slide

def add_channel_types(prs):
    """Channel 类型对比"""
    slide = add_content_slide(prs, "Channel 类型对比")
    
    # 表格数据
    headers = ["特性", "无缓冲 Channel", "有缓冲 Channel"]
    rows = [
        ["创建方式", "make(chan T)", "make(chan T, capacity)"],
        ["发送行为", "阻塞直到接收方就绪", "缓冲区未满时立即返回"],
        ["接收行为", "阻塞直到发送方就绪", "缓冲区非空时立即返回"],
        ["同步性", "强同步（会合点）", "异步（解耦生产消费）"],
        ["适用场景", "事件通知、协程同步", "消息队列、削峰填谷"],
        ["容量", "0", "固定大小"],
    ]
    
    # 创建表格
    left = Inches(1)
    top = Inches(1.3)
    width = Inches(8)
    height = Inches(3.5)
    
    table = slide.shapes.add_table(
        len(rows) + 1, len(headers), left, top, width, height
    ).table
    
    # 设置表头
    for col_idx, header in enumerate(headers):
        cell = table.cell(0, col_idx)
        cell.text = header
        cell.fill.solid()
        cell.fill.fore_color.rgb = COLOR_PRIMARY
        
        for paragraph in cell.text_frame.paragraphs:
            paragraph.font.size = Pt(13)
            paragraph.font.bold = True
            paragraph.font.color.rgb = RGBColor(255, 255, 255)
            paragraph.alignment = PP_ALIGN.CENTER
    
    # 填充数据
    for row_idx, row_data in enumerate(rows, start=1):
        for col_idx, cell_text in enumerate(row_data):
            cell = table.cell(row_idx, col_idx)
            cell.text = cell_text
            
            # 交替行背景色
            if row_idx % 2 == 0:
                cell.fill.solid()
                cell.fill.fore_color.rgb = RGBColor(248, 249, 250)
            
            for paragraph in cell.text_frame.paragraphs:
                paragraph.font.size = Pt(11)
                paragraph.font.color.rgb = COLOR_TEXT
                if col_idx == 0:
                    paragraph.font.bold = True
                    paragraph.alignment = PP_ALIGN.CENTER
                else:
                    paragraph.alignment = PP_ALIGN.LEFT
    
    # 底部提示
    tip = slide.shapes.add_textbox(
        Inches(1), Inches(5.2), Inches(8), Inches(1)
    )
    tf = tip.text_frame
    tf.text = "💡 选择原则：需要严格同步用无缓冲；需要削峰填谷用有缓冲"
    p = tf.paragraphs[0]
    p.font.size = Pt(13)
    p.font.color.rgb = COLOR_ACCENT
    p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_channel_flow_diagram(prs):
    """无缓冲 vs 有缓冲流程图"""
    slide = add_content_slide(prs, "Channel 通信机制")
    
    # 左侧：无缓冲
    left_x = Inches(0.6)
    top_y = Inches(1.3)
    
    # 标题
    title1 = slide.shapes.add_textbox(left_x, top_y, Inches(4), Inches(0.4))
    tf = title1.text_frame
    tf.text = "无缓冲 Channel（同步）"
    p = tf.paragraphs[0]
    p.font.size = Pt(16)
    p.font.bold = True
    p.font.color.rgb = COLOR_SECONDARY
    p.alignment = PP_ALIGN.CENTER
    
    # 发送者
    sender1 = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        left_x + Inches(0.3), top_y + Inches(0.8),
        Inches(1.2), Inches(0.6)
    )
    sender1.fill.solid()
    sender1.fill.fore_color.rgb = COLOR_WARNING
    sender1.line.fill.background()
    tf = sender1.text_frame
    tf.text = "发送者\n阻塞等待"
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = RGBColor(255, 255, 255)
        para.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # Channel
    chan1 = slide.shapes.add_shape(
        MSO_SHAPE.RECTANGLE,
        left_x + Inches(1.7), top_y + Inches(0.9),
        Inches(0.6), Inches(0.4)
    )
    chan1.fill.solid()
    chan1.fill.fore_color.rgb = COLOR_BG_LIGHT
    chan1.line.color.rgb = COLOR_TEXT
    chan1.line.width = Pt(2)
    tf = chan1.text_frame
    tf.text = "ch"
    p = tf.paragraphs[0]
    p.font.size = Pt(12)
    p.font.bold = True
    p.font.color.rgb = COLOR_TEXT
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 接收者
    receiver1 = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        left_x + Inches(2.5), top_y + Inches(0.8),
        Inches(1.2), Inches(0.6)
    )
    receiver1.fill.solid()
    receiver1.fill.fore_color.rgb = COLOR_SUCCESS
    receiver1.line.fill.background()
    tf = receiver1.text_frame
    tf.text = "接收者\n同步接收"
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = RGBColor(255, 255, 255)
        para.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 箭头
    arrow1 = slide.shapes.add_connector(
        1, left_x + Inches(1.5), top_y + Inches(1.1),
        left_x + Inches(1.7), top_y + Inches(1.1)
    )
    arrow1.line.color.rgb = COLOR_DANGER
    arrow1.line.width = Pt(3)
    arrow1.line.end_arrow_type = 2
    
    arrow2 = slide.shapes.add_connector(
        1, left_x + Inches(2.3), top_y + Inches(1.1),
        left_x + Inches(2.5), top_y + Inches(1.1)
    )
    arrow2.line.color.rgb = COLOR_SUCCESS
    arrow2.line.width = Pt(3)
    arrow2.line.end_arrow_type = 2
    
    # 说明
    note1 = slide.shapes.add_textbox(
        left_x, top_y + Inches(1.8), Inches(4), Inches(0.6)
    )
    tf = note1.text_frame
    tf.text = "必须同时有发送和接收\n否则发生死锁"
    tf.word_wrap = True
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = COLOR_DANGER
        para.alignment = PP_ALIGN.CENTER
    
    # 右侧：有缓冲
    right_x = Inches(5.2)
    
    # 标题
    title2 = slide.shapes.add_textbox(right_x, top_y, Inches(4), Inches(0.4))
    tf = title2.text_frame
    tf.text = "有缓冲 Channel（异步）"
    p = tf.paragraphs[0]
    p.font.size = Pt(16)
    p.font.bold = True
    p.font.color.rgb = COLOR_ACCENT
    p.alignment = PP_ALIGN.CENTER
    
    # 发送者
    sender2 = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        right_x + Inches(0.3), top_y + Inches(0.8),
        Inches(1.2), Inches(0.6)
    )
    sender2.fill.solid()
    sender2.fill.fore_color.rgb = COLOR_WARNING
    sender2.line.fill.background()
    tf = sender2.text_frame
    tf.text = "发送者\n非阻塞"
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = RGBColor(255, 255, 255)
        para.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 缓冲区（多个格子）
    for i in range(3):
        buf_cell = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            right_x + Inches(1.7 + i * 0.35), top_y + Inches(0.9),
            Inches(0.3), Inches(0.4)
        )
        buf_cell.fill.solid()
        if i == 0:
            buf_cell.fill.fore_color.rgb = COLOR_SECONDARY
        else:
            buf_cell.fill.fore_color.rgb = RGBColor(240, 240, 240)
        buf_cell.line.color.rgb = COLOR_TEXT
        buf_cell.line.width = Pt(1)
    
    # 容量标签
    cap_label = slide.shapes.add_textbox(
        right_x + Inches(1.7), top_y + Inches(1.4), Inches(1), Inches(0.3)
    )
    tf = cap_label.text_frame
    tf.text = "cap = 3"
    p = tf.paragraphs[0]
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_TEXT
    p.alignment = PP_ALIGN.CENTER
    
    # 接收者
    receiver2 = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        right_x + Inches(2.5), top_y + Inches(0.8),
        Inches(1.2), Inches(0.6)
    )
    receiver2.fill.solid()
    receiver2.fill.fore_color.rgb = COLOR_SUCCESS
    receiver2.line.fill.background()
    tf = receiver2.text_frame
    tf.text = "接收者\n独立接收"
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = RGBColor(255, 255, 255)
        para.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 箭头
    arrow3 = slide.shapes.add_connector(
        1, right_x + Inches(1.5), top_y + Inches(1.1),
        right_x + Inches(1.7), top_y + Inches(1.1)
    )
    arrow3.line.color.rgb = COLOR_DANGER
    arrow3.line.width = Pt(3)
    arrow3.line.end_arrow_type = 2
    
    arrow4 = slide.shapes.add_connector(
        1, right_x + Inches(2.75), top_y + Inches(1.1),
        right_x + Inches(2.5), top_y + Inches(1.1)
    )
    arrow4.line.color.rgb = COLOR_SUCCESS
    arrow4.line.width = Pt(3)
    arrow4.line.end_arrow_type = 2
    
    # 说明
    note2 = slide.shapes.add_textbox(
        right_x, top_y + Inches(1.8), Inches(4), Inches(0.6)
    )
    tf = note2.text_frame
    tf.text = "缓冲区未满可继续发送\n解耦生产者和消费者"
    tf.word_wrap = True
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = COLOR_SUCCESS
        para.alignment = PP_ALIGN.CENTER
    
    # 底部代码示例
    code_box = slide.shapes.add_textbox(
        Inches(1), Inches(3.3), Inches(8), Inches(2.5)
    )
    tf = code_box.text_frame
    code = """// 无缓冲
ch1 := make(chan int)
ch1 <- 42  // 阻塞，直到有人接收

// 有缓冲
ch2 := make(chan int, 3)
ch2 <- 1  // 立即返回
ch2 <- 2  // 立即返回
ch2 <- 3  // 立即返回
ch2 <- 4  // 阻塞，缓冲区已满"""
    
    p = tf.paragraphs[0]
    p.text = code
    p.font.name = 'Consolas'
    p.font.size = Pt(11)
    p.font.color.rgb = COLOR_TEXT
    
    return slide

def add_select_diagram(prs):
    """Select 多路复用流程图"""
    slide = add_content_slide(prs, "Select 多路复用")
    
    # 顶部说明
    desc = slide.shapes.add_textbox(
        Inches(1), Inches(1.2), Inches(8), Inches(0.4)
    )
    tf = desc.text_frame
    tf.text = "同时监听多个 Channel，哪个就绪就执行哪个（随机选择）"
    p = tf.paragraphs[0]
    p.font.size = Pt(14)
    p.font.color.rgb = COLOR_TEXT
    p.alignment = PP_ALIGN.CENTER
    
    # 中央 Select 框
    select_box = slide.shapes.add_shape(
        MSO_SHAPE.FLOWCHART_DECISION,
        Inches(4), Inches(2),
        Inches(2), Inches(1.2)
    )
    select_box.fill.solid()
    select_box.fill.fore_color.rgb = COLOR_SECONDARY
    select_box.line.color.rgb = COLOR_PRIMARY
    select_box.line.width = Pt(3)
    tf = select_box.text_frame
    tf.text = "select"
    p = tf.paragraphs[0]
    p.font.size = Pt(20)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 三个 case 分支
    cases = [
        ("case <-ch1:", Inches(1.5), COLOR_SUCCESS),
        ("case <-ch2:", Inches(5), COLOR_WARNING),
        ("case <-ch3:", Inches(8.5), COLOR_DANGER),
    ]
    
    for case_text, x_pos, color in cases:
        # case 框
        case_box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            x_pos - Inches(0.6), Inches(3.8),
            Inches(1.2), Inches(0.5)
        )
        case_box.fill.solid()
        case_box.fill.fore_color.rgb = color
        case_box.line.fill.background()
        tf = case_box.text_frame
        tf.text = case_text
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 从 select 到 case 的箭头
        arrow = slide.shapes.add_connector(
            1, Inches(5), Inches(3.2),
            x_pos, Inches(3.8)
        )
        arrow.line.color.rgb = color
        arrow.line.width = Pt(2.5)
        arrow.line.end_arrow_type = 2
    
    # default 分支
    default_box = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(4.4), Inches(3.8),
        Inches(1.2), Inches(0.5)
    )
    default_box.fill.solid()
    default_box.fill.fore_color.rgb = COLOR_TEXT
    default_box.line.fill.background()
    tf = default_box.text_frame
    tf.text = "default:"
    p = tf.paragraphs[0]
    p.font.size = Pt(12)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    arrow_def = slide.shapes.add_connector(
        1, Inches(5), Inches(3.2),
        Inches(5), Inches(3.8)
    )
    arrow_def.line.color.rgb = COLOR_TEXT
    arrow_def.line.width = Pt(2.5)
    arrow_def.line.end_arrow_type = 2
    
    # 代码示例
    code_box = slide.shapes.add_textbox(
        Inches(1.5), Inches(4.8), Inches(7), Inches(1.5)
    )
    tf = code_box.text_frame
    code = """select {
case msg1 := <-ch1:
    fmt.Println("从 ch1 接收:", msg1)
case msg2 := <-ch2:
    fmt.Println("从 ch2 接收:", msg2)
case <-time.After(time.Second):
    fmt.Println("超时")
default:
    fmt.Println("没有就绪的 channel")
}"""
    p = tf.paragraphs[0]
    p.text = code
    p.font.name = 'Consolas'
    p.font.size = Pt(11)
    p.font.color.rgb = COLOR_TEXT
    
    return slide

def add_context_tree(prs):
    """Context 树状结构图"""
    slide = add_content_slide(prs, "Context 树状结构与取消传播")
    
    # 根 Context
    root = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(4), Inches(1.3),
        Inches(2), Inches(0.5)
    )
    root.fill.solid()
    root.fill.fore_color.rgb = COLOR_PRIMARY
    root.line.fill.background()
    tf = root.text_frame
    tf.text = "context.Background()"
    p = tf.paragraphs[0]
    p.font.size = Pt(12)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # 第一层子 Context
    children_l1 = [
        ("WithCancel", Inches(2), COLOR_SECONDARY),
        ("WithTimeout", Inches(5), COLOR_WARNING),
        ("WithValue", Inches(8), COLOR_ACCENT),
    ]
    
    for label, x_pos, color in children_l1:
        box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            x_pos - Inches(0.75), Inches(2.3),
            Inches(1.5), Inches(0.5)
        )
        box.fill.solid()
        box.fill.fore_color.rgb = color
        box.line.fill.background()
        tf = box.text_frame
        tf.text = label
        p = tf.paragraphs[0]
        p.font.size = Pt(11)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 连线
        arrow = slide.shapes.add_connector(
            1, Inches(5), Inches(1.8),
            x_pos, Inches(2.3)
        )
        arrow.line.color.rgb = COLOR_TEXT
        arrow.line.width = Pt(2)
        arrow.line.end_arrow_type = 2
    
    # 第二层子 Context（从 WithCancel 派生）
    child_l2 = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(1.25), Inches(3.3),
        Inches(1.5), Inches(0.5)
    )
    child_l2.fill.solid()
    child_l2.fill.fore_color.rgb = COLOR_SUCCESS
    child_l2.line.fill.background()
    tf = child_l2.text_frame
    tf.text = "子 Context"
    p = tf.paragraphs[0]
    p.font.size = Pt(11)
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    arrow_l2 = slide.shapes.add_connector(
        1, Inches(2), Inches(2.8),
        Inches(2), Inches(3.3)
    )
    arrow_l2.line.color.rgb = COLOR_TEXT
    arrow_l2.line.width = Pt(2)
    arrow_l2.line.end_arrow_type = 2
    
    # 取消传播说明
    cancel_note = slide.shapes.add_shape(
        MSO_SHAPE.FLOWCHART_DOCUMENT,
        Inches(6.5), Inches(1.5),
        Inches(2.5), Inches(1.5)
    )
    cancel_note.fill.solid()
    cancel_note.fill.fore_color.rgb = RGBColor(255, 243, 224)
    cancel_note.line.color.rgb = COLOR_WARNING
    cancel_note.line.width = Pt(2)
    tf = cancel_note.text_frame
    tf.text = "取消传播规则：\n父 Context 取消\n↓\n所有子孙 Context\n自动取消"
    tf.word_wrap = True
    for para in tf.paragraphs:
        para.font.size = Pt(11)
        para.font.color.rgb = COLOR_TEXT
        para.alignment = PP_ALIGN.CENTER
    
    # 代码示例
    code_box = slide.shapes.add_textbox(
        Inches(1), Inches(4.3), Inches(8), Inches(2)
    )
    tf = code_box.text_frame
    code = """// 创建可取消的 Context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // 确保资源释放

// 超时 Context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 监听取消信号
select {
case <-ctx.Done():
    fmt.Println("Context 被取消:", ctx.Err())
}"""
    p = tf.paragraphs[0]
    p.text = code
    p.font.name = 'Consolas'
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_TEXT
    
    return slide

def add_mutex_comparison(prs):
    """Mutex vs RWMutex 对比"""
    slide = add_content_slide(prs, "Mutex vs RWMutex 对比")
    
    # 左侧：Mutex
    left_x = Inches(0.8)
    top_y = Inches(1.3)
    
    mutex_box = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        left_x, top_y, Inches(4), Inches(0.6)
    )
    mutex_box.fill.solid()
    mutex_box.fill.fore_color.rgb = COLOR_SECONDARY
    mutex_box.line.fill.background()
    tf = mutex_box.text_frame
    tf.text = "sync.Mutex（互斥锁）"
    p = tf.paragraphs[0]
    p.font.size = Pt(16)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    mutex_features = [
        "• 同一时刻只允许一个 goroutine 访问",
        "• Lock() / Unlock() 成对使用",
        "• 读写都需要加锁",
        "• 适用：读写频率相当",
        "• 性能：写多场景最优"
    ]
    
    y_offset = top_y + Inches(0.8)
    for feature in mutex_features:
        feat_box = slide.shapes.add_textbox(
            left_x + Inches(0.2), y_offset, Inches(3.6), Inches(0.35)
        )
        tf = feat_box.text_frame
        tf.text = feature
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = COLOR_TEXT
        y_offset += Inches(0.4)
    
    # 右侧：RWMutex
    right_x = Inches(5.2)
    
    rwmutex_box = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        right_x, top_y, Inches(4), Inches(0.6)
    )
    rwmutex_box.fill.solid()
    rwmutex_box.fill.fore_color.rgb = COLOR_ACCENT
    rwmutex_box.line.fill.background()
    tf = rwmutex_box.text_frame
    tf.text = "sync.RWMutex（读写锁）"
    p = tf.paragraphs[0]
    p.font.size = Pt(16)
    p.font.bold = True
    p.font.color.rgb = RGBColor(255, 255, 255)
    p.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    rwmutex_features = [
        "• 允许多个 goroutine 同时读",
        "• RLock() / RUnlock() 读锁",
        "• Lock() / Unlock() 写锁（互斥）",
        "• 适用：读多写少场景",
        "• 性能：读密集场景最优"
    ]
    
    y_offset = top_y + Inches(0.8)
    for feature in rwmutex_features:
        feat_box = slide.shapes.add_textbox(
            right_x + Inches(0.2), y_offset, Inches(3.6), Inches(0.35)
        )
        tf = feat_box.text_frame
        tf.text = feature
        p = tf.paragraphs[0]
        p.font.size = Pt(12)
        p.font.color.rgb = COLOR_TEXT
        y_offset += Inches(0.4)
    
    # 代码对比
    code_box = slide.shapes.add_textbox(
        Inches(1), Inches(3.8), Inches(8), Inches(2.5)
    )
    tf = code_box.text_frame
    code = """// Mutex 示例
var mu sync.Mutex
mu.Lock()
counter++  // 只有一个 goroutine 可访问
mu.Unlock()

// RWMutex 示例
var rwmu sync.RWMutex

// 多个读者可同时读
rwmu.RLock()
value := data[key]
rwmu.RUnlock()

// 写操作独占
rwmu.Lock()
data[key] = newValue
rwmu.Unlock()"""
    p = tf.paragraphs[0]
    p.text = code
    p.font.name = 'Consolas'
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_TEXT
    
    return slide

def add_worker_pool_architecture(prs):
    """Worker Pool 架构图"""
    slide = add_content_slide(prs, "Worker Pool 并发模式")
    
    # 任务队列
    jobs_queue = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(1), Inches(1.5), Inches(2), Inches(4)
    )
    jobs_queue.fill.solid()
    jobs_queue.fill.fore_color.rgb = COLOR_BG_LIGHT
    jobs_queue.line.color.rgb = COLOR_SECONDARY
    jobs_queue.line.width = Pt(3)
    tf = jobs_queue.text_frame
    tf.text = "任务队列\n(jobs chan)"
    tf.word_wrap = True
    for para in tf.paragraphs:
        para.font.size = Pt(14)
        para.font.bold = True
        para.font.color.rgb = COLOR_SECONDARY
        para.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.TOP
    tf.margin_top = Inches(0.1)
    
    # 任务项
    for i in range(5):
        job_item = slide.shapes.add_shape(
            MSO_SHAPE.RECTANGLE,
            Inches(1.3), Inches(2 + i * 0.6), Inches(1.4), Inches(0.4)
        )
        job_item.fill.solid()
        job_item.fill.fore_color.rgb = COLOR_WARNING
        job_item.line.fill.background()
        tf = job_item.text_frame
        tf.text = f"Job {i+1}"
        p = tf.paragraphs[0]
        p.font.size = Pt(11)
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
    
    # Workers
    worker_x = Inches(4)
    for i in range(3):
        y_pos = Inches(2 + i * 1.2)
        
        worker = slide.shapes.add_shape(
            MSO_SHAPE.FLOWCHART_PROCESS,
            worker_x, y_pos, Inches(1.8), Inches(0.8)
        )
        worker.fill.solid()
        worker.fill.fore_color.rgb = COLOR_ACCENT
        worker.line.color.rgb = COLOR_PRIMARY
        worker.line.width = Pt(2)
        tf = worker.text_frame
        tf.text = f"Worker {i+1}\ngoroutine"
        for para in tf.paragraphs:
            para.font.size = Pt(12)
            para.font.bold = True
            para.font.color.rgb = RGBColor(255, 255, 255)
            para.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 从任务队列到 worker 的箭头
        arrow = slide.shapes.add_connector(
            1, Inches(3), y_pos + Inches(0.4),
            worker_x, y_pos + Inches(0.4)
        )
        arrow.line.color.rgb = COLOR_TEXT
        arrow.line.width = Pt(2.5)
        arrow.line.end_arrow_type = 2
    
    # 结果队列
    results_queue = slide.shapes.add_shape(
        MSO_SHAPE.ROUNDED_RECTANGLE,
        Inches(7), Inches(1.5), Inches(2), Inches(4)
    )
    results_queue.fill.solid()
    results_queue.fill.fore_color.rgb = COLOR_BG_LIGHT
    results_queue.line.color.rgb = COLOR_SUCCESS
    results_queue.line.width = Pt(3)
    tf = results_queue.text_frame
    tf.text = "结果队列\n(results chan)"
    tf.word_wrap = True
    for para in tf.paragraphs:
        para.font.size = Pt(14)
        para.font.bold = True
        para.font.color.rgb = COLOR_SUCCESS
        para.alignment = PP_ALIGN.CENTER
    tf.vertical_anchor = MSO_ANCHOR.TOP
    tf.margin_top = Inches(0.1)
    
    # workers 到结果队列的箭头
    for i in range(3):
        y_pos = Inches(2 + i * 1.2)
        arrow = slide.shapes.add_connector(
            1, worker_x + Inches(1.8), y_pos + Inches(0.4),
            Inches(7), y_pos + Inches(0.4)
        )
        arrow.line.color.rgb = COLOR_SUCCESS
        arrow.line.width = Pt(2.5)
        arrow.line.end_arrow_type = 2
    
    # 底部代码骨架
    code_box = slide.shapes.add_textbox(
        Inches(0.8), Inches(5.8), Inches(8.4), Inches(1.2)
    )
    tf = code_box.text_frame
    code = """jobs := make(chan Job, 100); results := make(chan Result, 100)
for w := 1; w <= numWorkers; w++ { go worker(jobs, results) }
for _, job := range allJobs { jobs <- job }
close(jobs); for r := range results { process(r) }"""
    p = tf.paragraphs[0]
    p.text = code
    p.font.name = 'Consolas'
    p.font.size = Pt(10)
    p.font.color.rgb = COLOR_TEXT
    
    return slide

def add_common_pitfalls(prs):
    """常见陷阱"""
    slide = add_content_slide(prs, "并发常见陷阱")
    
    pitfalls = [
        {
            "title": "1. Goroutine 泄漏",
            "desc": "goroutine 启动后永远无法退出，占用内存",
            "example": "for { select { case <-ch: } }  // 没有退出条件",
            "solution": "✓ 添加 context 或 done channel"
        },
        {
            "title": "2. 竞态条件 (Data Race)",
            "desc": "多个 goroutine 同时读写同一变量",
            "example": "counter++  // 非原子操作，并发不安全",
            "solution": "✓ 使用 Mutex 或 atomic 包"
        },
        {
            "title": "3. 死锁 (Deadlock)",
            "desc": "所有 goroutine 都在等待，无法继续",
            "example": "ch := make(chan int); ch <- 1  // 无人接收",
            "solution": "✓ 确保 channel 有接收者或使用缓冲"
        },
        {
            "title": "4. 关闭已关闭的 Channel",
            "desc": "重复关闭 channel 会 panic",
            "example": "close(ch); close(ch)  // panic!",
            "solution": "✓ 只由发送方关闭，只关闭一次"
        },
    ]
    
    top_y = Inches(1.3)
    for i, pitfall in enumerate(pitfalls):
        y_pos = top_y + i * Inches(1.1)
        
        # 陷阱标题
        title_box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            Inches(0.8), y_pos, Inches(8.4), Inches(0.35)
        )
        title_box.fill.solid()
        title_box.fill.fore_color.rgb = COLOR_DANGER
        title_box.line.fill.background()
        tf = title_box.text_frame
        tf.text = pitfall["title"]
        p = tf.paragraphs[0]
        p.font.size = Pt(13)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 描述
        desc_box = slide.shapes.add_textbox(
            Inches(1), y_pos + Inches(0.4), Inches(8), Inches(0.2)
        )
        tf = desc_box.text_frame
        tf.text = f"问题：{pitfall['desc']}"
        p = tf.paragraphs[0]
        p.font.size = Pt(10)
        p.font.color.rgb = COLOR_TEXT
        
        # 示例和解决方案
        ex_sol_box = slide.shapes.add_textbox(
            Inches(1), y_pos + Inches(0.6), Inches(8), Inches(0.35)
        )
        tf = ex_sol_box.text_frame
        p = tf.add_paragraph() if tf.paragraphs[0].text else tf.paragraphs[0]
        p.text = f"❌ {pitfall['example']}"
        p.font.size = Pt(9)
        p.font.name = 'Consolas'
        p.font.color.rgb = COLOR_DANGER
        
        p2 = tf.add_paragraph()
        p2.text = pitfall['solution']
        p2.font.size = Pt(9)
        p2.font.color.rgb = COLOR_SUCCESS
    
    # 调试工具提示
    tool_box = slide.shapes.add_textbox(
        Inches(1), Inches(5.8), Inches(8), Inches(0.8)
    )
    tf = tool_box.text_frame
    tf.text = "🔧 调试工具：go run -race  |  pprof  |  runtime.NumGoroutine()"
    p = tf.paragraphs[0]
    p.font.size = Pt(13)
    p.font.bold = True
    p.font.color.rgb = COLOR_ACCENT
    p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_best_practices(prs):
    """最佳实践"""
    slide = add_content_slide(prs, "Go 并发最佳实践")
    
    practices = [
        ("设计原则", [
            "• 通过通信来共享内存，而不是通过共享内存来通信",
            "• 单一职责：每个 goroutine 做好一件事",
            "• 明确生命周期：有启动就要有退出机制",
            "• 优先使用 channel，必要时才用锁",
        ]),
        ("性能优化", [
            "• 控制并发数量：使用 Worker Pool 限制 goroutine 数量",
            "• 避免过度同步：减少锁的粒度和持有时间",
            "• 复用对象：使用 sync.Pool 减少 GC 压力",
            "• 使用有缓冲 channel 减少阻塞",
        ]),
        ("错误处理", [
            "• 使用 Context 传递取消信号和超时",
            "• errgroup.Group 收集并发任务的错误",
            "• defer 确保资源释放（Unlock、cancel()）",
            "• 监控 goroutine 数量，防止泄漏",
        ]),
        ("测试调试", [
            "• go test -race 检测竞态条件",
            "• pprof 分析 goroutine 泄漏和性能瓶颈",
            "• 使用 context.WithTimeout 避免永久阻塞",
            "• 单元测试中使用带超时的 channel",
        ]),
    ]
    
    left_x = Inches(0.8)
    top_y = Inches(1.3)
    col_width = Inches(4.5)
    
    for i, (category, items) in enumerate(practices):
        x_pos = left_x + (i % 2) * col_width
        y_pos = top_y + (i // 2) * Inches(2.3)
        
        # 类别标题
        cat_box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            x_pos, y_pos, Inches(4), Inches(0.4)
        )
        cat_box.fill.solid()
        colors = [COLOR_SECONDARY, COLOR_WARNING, COLOR_SUCCESS, COLOR_ACCENT]
        cat_box.fill.fore_color.rgb = colors[i]
        cat_box.line.fill.background()
        tf = cat_box.text_frame
        tf.text = category
        p = tf.paragraphs[0]
        p.font.size = Pt(14)
        p.font.bold = True
        p.font.color.rgb = RGBColor(255, 255, 255)
        p.alignment = PP_ALIGN.CENTER
        tf.vertical_anchor = MSO_ANCHOR.MIDDLE
        
        # 实践项
        items_box = slide.shapes.add_textbox(
            x_pos + Inches(0.1), y_pos + Inches(0.5),
            Inches(3.8), Inches(1.5)
        )
        tf = items_box.text_frame
        tf.text = '\n'.join(items)
        tf.word_wrap = True
        for para in tf.paragraphs:
            para.font.size = Pt(10)
            para.font.color.rgb = COLOR_TEXT
            para.line_spacing = 1.3
    
    # 底部核心原则
    principle_box = slide.shapes.add_textbox(
        Inches(1), Inches(5.8), Inches(8), Inches(0.8)
    )
    tf = principle_box.text_frame
    tf.text = "💡 核心原则：简单、清晰、可控 —— 能用同步就不用异步，能用一个 goroutine 就不用多个"
    p = tf.paragraphs[0]
    p.font.size = Pt(13)
    p.font.bold = True
    p.font.color.rgb = COLOR_PRIMARY
    p.alignment = PP_ALIGN.CENTER
    
    return slide

def add_summary_slide(prs):
    """总结页"""
    slide = add_content_slide(prs, "课程总结")
    
    summary_sections = [
        ("核心概念", [
            "Goroutine：轻量级协程",
            "Channel：通信管道",
            "Select：多路复用",
            "Context：上下文控制",
        ]),
        ("同步工具", [
            "WaitGroup：等待组",
            "Mutex/RWMutex：互斥锁",
            "Once：单次执行",
            "Pool：对象池",
        ]),
        ("并发模式", [
            "Worker Pool：工作池",
            "Pipeline：流水线",
            "Fan-out/in：扇出扇入",
            "Errgroup：错误组",
        ]),
    ]
    
    top_y = Inches(1.5)
    for i, (section, items) in enumerate(summary_sections):
        x_pos = Inches(1 + i * 2.8)
        
        # 区块
        section_box = slide.shapes.add_shape(
            MSO_SHAPE.ROUNDED_RECTANGLE,
            x_pos, top_y, Inches(2.5), Inches(3)
        )
        section_box.fill.solid()
        colors = [COLOR_SECONDARY, COLOR_WARNING, COLOR_SUCCESS]
        section_box.fill.fore_color.rgb = RGBColor(245, 247, 250)
        section_box.line.color.rgb = colors[i]
        section_box.line.width = Pt(3)
        
        # 标题
        title_box = slide.shapes.add_textbox(
            x_pos + Inches(0.2), top_y + Inches(0.2),
            Inches(2.1), Inches(0.4)
        )
        tf = title_box.text_frame
        tf.text = section
        p = tf.paragraphs[0]
        p.font.size = Pt(16)
        p.font.bold = True
        p.font.color.rgb = colors[i]
        p.alignment = PP_ALIGN.CENTER
        
        # 项目
        items_box = slide.shapes.add_textbox(
            x_pos + Inches(0.3), top_y + Inches(0.8),
            Inches(1.9), Inches(2)
        )
        tf = items_box.text_frame
        for item in items:
            p = tf.add_paragraph() if tf.paragraphs[0].text else tf.paragraphs[0]
            p.text = f"✓ {item}"
            p.font.size = Pt(11)
            p.font.color.rgb = COLOR_TEXT
            p.space_before = Pt(8)
    
    # 关键要点
    key_points = slide.shapes.add_textbox(
        Inches(1), Inches(5), Inches(8), Inches(1.5)
    )
    tf = key_points.text_frame
    
    points = [
        "🎯 掌握并发不是目标，写出正确、高效、可维护的并发代码才是",
        "⚡ 并发能力是 Go 的核心竞争力，也是构建高性能系统的基石",
        "🔧 持续实践：从简单的 Worker Pool 到复杂的微服务编排",
    ]
    
    for point in points:
        p = tf.add_paragraph() if tf.paragraphs[0].text else tf.paragraphs[0]
        p.text = point
        p.font.size = Pt(12)
        p.font.color.rgb = COLOR_PRIMARY
        p.space_after = Pt(10)
    
    return slide

def add_qa_slide(prs):
    """Q&A 页"""
    slide = add_content_slide(prs, "")
    
    # 大标题
    qa_title = slide.shapes.add_textbox(
        Inches(2), Inches(2.5), Inches(6), Inches(1.5)
    )
    tf = qa_title.text_frame
    tf.text = "Q & A"
    p = tf.paragraphs[0]
    p.font.size = Pt(72)
    p.font.bold = True
    p.font.color.rgb = COLOR_PRIMARY
    p.alignment = PP_ALIGN.CENTER
    
    # 副标题
    subtitle = slide.shapes.add_textbox(
        Inches(2), Inches(4.2), Inches(6), Inches(0.6)
    )
    tf = subtitle.text_frame
    tf.text = "感谢聆听 | 欢迎提问"
    p = tf.paragraphs[0]
    p.font.size = Pt(20)
    p.font.color.rgb = COLOR_TEXT
    p.alignment = PP_ALIGN.CENTER
    
    # 装饰圆圈
    for i, x_pos in enumerate([Inches(1.5), Inches(4.5), Inches(7.5)]):
        circle = slide.shapes.add_shape(
            MSO_SHAPE.OVAL,
            x_pos, Inches(5.5), Inches(0.8), Inches(0.8)
        )
        circle.fill.solid()
        colors = [COLOR_SECONDARY, COLOR_WARNING, COLOR_SUCCESS]
        circle.fill.fore_color.rgb = colors[i]
        circle.line.fill.background()
    
    return slide

def add_page_numbers(prs):
    """添加页码"""
    for i, slide in enumerate(prs.slides, start=1):
        if i == 1:  # 跳过封面
            continue
        
        page_num = slide.shapes.add_textbox(
            Inches(9.2), Inches(7), Inches(0.6), Inches(0.3)
        )
        tf = page_num.text_frame
        tf.text = str(i)
        p = tf.paragraphs[0]
        p.font.size = Pt(10)
        p.font.color.rgb = COLOR_TEXT
        p.alignment = PP_ALIGN.CENTER

def main():
    """主函数"""
    print("开始生成企业级 Go 并发编程课程 PPT...")
    
    prs = create_presentation()
    
    # 1. 封面
    print("  [1/30] 创建封面...")
    add_title_slide(prs)
    
    # 2. 目录
    print("  [2/30] 创建目录...")
    add_table_of_contents(prs)
    
    # 第一章：Go 并发基础
    print("  [3/30] 第一章开始...")
    add_section_header(prs, "第一章", "Go 并发基础")
    
    print("  [4/30] 并发 vs 并行...")
    add_concurrency_vs_parallelism(prs)
    
    print("  [5/30] G-M-P 模型...")
    add_gmp_model(prs)
    
    print("  [6/30] Goroutine 生命周期...")
    add_goroutine_lifecycle(prs)
    
    # 第二章：Goroutines 深入
    print("  [7/30] 第二章开始...")
    add_section_header(prs, "第二章", "Goroutines 深入")
    
    print("  [8/30] 创建 Goroutine...")
    add_code_slide(prs, "创建和启动 Goroutine", 
"""package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    fmt.Printf("Hello, %s!\\n", name)
}

func main() {
    // 同步调用
    sayHello("World")
    
    // 异步调用 - 启动新的 goroutine
    go sayHello("Alice")
    go sayHello("Bob")
    
    // 等待 goroutines 执行
    time.Sleep(time.Second)
}""",
    "💡 go 关键字启动新的 goroutine，立即返回不阻塞")
    
    print("  [9/30] WaitGroup 使用...")
    add_waitgroup_slide(prs)
    
    print("  [10/30] 闭包陷阱...")
    add_code_slide(prs, "闭包陷阱与解决方案",
"""// ❌ 错误：循环变量被所有 goroutine 共享
for i := 0; i < 3; i++ {
    go func() {
        fmt.Println(i)  // 可能全部输出 3
    }()
}

// ✅ 方案 1：传参
for i := 0; i < 3; i++ {
    go func(n int) {
        fmt.Println(n)  // 正确输出 0, 1, 2
    }(i)
}

// ✅ 方案 2：新变量
for i := 0; i < 3; i++ {
    i := i  // 创建新变量
    go func() {
        fmt.Println(i)
    }()
}""",
    "⚠️ 循环启动 goroutine 时，必须注意循环变量的作用域")
    
    # 第三章：Channels 通信
    print("  [11/30] 第三章开始...")
    add_section_header(prs, "第三章", "Channels 通信")
    
    print("  [12/30] Channel 类型对比...")
    add_channel_types(prs)
    
    print("  [13/30] Channel 流程图...")
    add_channel_flow_diagram(prs)
    
    print("  [14/30] Close 和 Range...")
    add_code_slide(prs, "关闭 Channel 与遍历",
"""func producer(ch chan int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // 发送方关闭 channel
}

func consumer(ch chan int) {
    // 方式 1：range 遍历（自动检测关闭）
    for val := range ch {
        fmt.Println(val)
    }
    
    // 方式 2：逗号 ok 模式
    for {
        val, ok := <-ch
        if !ok {  // channel 已关闭
            break
        }
        fmt.Println(val)
    }
}""",
    "⚠️ 规则：只由发送方关闭，接收方不关闭；不要重复关闭")
    
    print("  [15/30] 死锁场景...")
    add_code_slide(prs, "死锁场景与解决",
"""// ❌ 死锁 1：无缓冲 channel 无接收者
ch := make(chan int)
ch <- 42  // fatal error: all goroutines are asleep

// ✅ 解决：启动 goroutine 接收
go func() { <-ch }()
ch <- 42

// ❌ 死锁 2：循环等待
ch1, ch2 := make(chan int), make(chan int)
go func() { ch1 <- <-ch2 }()  // 等待 ch2
ch2 <- <-ch1  // 等待 ch1，死锁

// ✅ 解决：使用 select 或缓冲区
ch := make(chan int, 1)  // 有缓冲不会阻塞""",
    "🔍 使用 go run -race 检测死锁和竞态条件")
    
    # 第四章：Select 多路复用
    print("  [16/30] 第四章开始...")
    add_section_header(prs, "第四章", "Select 多路复用")
    
    print("  [17/30] Select 流程图...")
    add_select_diagram(prs)
    
    print("  [18/30] 超时控制...")
    add_code_slide(prs, "超时控制模式",
"""func fetchWithTimeout(url string) (string, error) {
    resultCh := make(chan string, 1)
    errCh := make(chan error, 1)
    
    go func() {
        data, err := fetch(url)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- data
    }()
    
    select {
    case data := <-resultCh:
        return data, nil
    case err := <-errCh:
        return "", err
    case <-time.After(5 * time.Second):
        return "", fmt.Errorf("timeout")
    }
}""",
    "💡 time.After() 创建超时 channel，避免永久阻塞")
    
    print("  [19/30] 非阻塞模式...")
    add_code_slide(prs, "非阻塞发送与接收",
"""// 非阻塞发送
select {
case ch <- value:
    fmt.Println("发送成功")
default:
    fmt.Println("channel 满了，丢弃")
}

// 非阻塞接收
select {
case val := <-ch:
    fmt.Println("接收到:", val)
default:
    fmt.Println("channel 空，不等待")
}

// 轮询多个 channel
for {
    select {
    case msg := <-ch1:
        process(msg)
    case msg := <-ch2:
        process(msg)
    default:
        time.Sleep(time.Millisecond * 100)
    }
}""",
    "⚠️ default 使 select 非阻塞，适合丢弃策略但不适合关键数据")
    
    # 第五章：Context 上下文
    print("  [20/30] 第五章开始...")
    add_section_header(prs, "第五章", "Context 上下文控制")
    
    print("  [21/30] Context 树...")
    add_context_tree(prs)
    
    print("  [22/30] Context 实战...")
    add_code_slide(prs, "Context 取消传播实战",
"""func worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d 收到取消信号\\n", id)
            return
        default:
            // 执行任务
            time.Sleep(time.Millisecond * 500)
            fmt.Printf("Worker %d 工作中...\\n", id)
        }
    }
}

func main() {
    ctx, cancel := context.WithTimeout(
        context.Background(), 
        3*time.Second,
    )
    defer cancel()
    
    for i := 1; i <= 3; i++ {
        go worker(ctx, i)
    }
    
    <-ctx.Done()
    fmt.Println("所有 worker 已停止")
}""",
    "💡 Context 是取消信号传播的标准方式，贯穿整个调用链")
    
    # 第六章：同步原语
    print("  [23/30] 第六章开始...")
    add_section_header(prs, "第六章", "同步原语")
    
    print("  [24/30] Mutex vs RWMutex...")
    add_mutex_comparison(prs)
    
    print("  [25/30] sync.Once...")
    add_code_slide(prs, "sync.Once：单次初始化",
"""var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        fmt.Println("初始化单例")
        instance = &Singleton{}
    })
    return instance
}

// 并发调用 GetInstance() 多次
// 只会初始化一次
func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            GetInstance()
        }()
    }
    wg.Wait()
}""",
    "💡 sync.Once 保证函数只执行一次，常用于单例、初始化配置")
    
    print("  [26/30] sync.Pool 和 atomic...")
    add_code_slide(prs, "sync.Pool 对象池 & atomic 原子操作",
"""// sync.Pool - 复用对象减少 GC
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process() {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()  // 重置后放回
    buf.WriteString("hello")
}

// atomic - 原子操作
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)  // 原子自增
}

func getCounter() int64 {
    return atomic.LoadInt64(&counter)  // 原子读取
}""",
    "💡 Pool 适合频繁创建销毁的对象；atomic 适合简单计数器")
    
    # 第七章：并发模式
    print("  [27/30] 第七章开始...")
    add_section_header(prs, "第七章", "并发模式")
    
    print("  [28/30] Worker Pool...")
    add_worker_pool_architecture(prs)
    
    print("  [29/30] Pipeline...")
    add_code_slide(prs, "Pipeline 模式",
"""func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func main() {
    // 流水线：生成 -> 平方 -> 输出
    c := generator(1, 2, 3, 4)
    out := square(c)
    for result := range out {
        fmt.Println(result)  // 1, 4, 9, 16
    }
}""",
    "💡 Pipeline 将数据处理拆分为多个阶段，每阶段独立并发")
    
    # 第八章：常见陷阱
    print("  [30/30] 第八章：常见陷阱...")
    add_section_header(prs, "第八章", "常见陷阱与调试")
    add_common_pitfalls(prs)
    
    # 第九章：最佳实践
    add_section_header(prs, "第九章", "最佳实践与性能优化")
    add_best_practices(prs)
    
    # 总结和 Q&A
    add_summary_slide(prs)
    add_qa_slide(prs)
    
    # 添加页码
    print("添加页码...")
    add_page_numbers(prs)
    
    # 保存文件
    output_path = "/Users/cluo/code/go/src/go-study/03-go-concurrency/Go并发编程课程.pptx"
    prs.save(output_path)
    print(f"\n✅ PPT 生成成功！")
    print(f"📄 文件路径：{output_path}")
    print(f"📊 总页数：{len(prs.slides)} 页")
    print("\n特性总结：")
    print("  • 企业级专业设计，深蓝商务配色")
    print("  • 完整的流程图和架构图")
    print("  • 代码示例带等宽字体")
    print("  • 包含 G-M-P 模型、Worker Pool、Context 树等核心图解")
    print("  • 涵盖从基础到进阶的全部并发知识点")

if __name__ == "__main__":
    main()
