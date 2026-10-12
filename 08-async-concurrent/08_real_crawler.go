//go:build level

package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chichengyu/async"
)

/*
🎯 核心概念：真实爬虫系统 - async库综合实战

这个文件展示一个完整的生产级爬虫系统架构，综合运用：
- Pipeline: 多阶段数据流（URL → HTML → 解析 → 存储）
- BoundedRunner: 百万级URL并发控制
- RateLimiter: API保护，防止被封
- Retry: 失败重试，指数退避
- Map: 批量数据转换
- Group: 任务分组管理

真实场景：
电商数据采集系统，需要：
1. 从种子URL开始，递归爬取商品页
2. 限制并发（避免封IP）
3. 失败重试（网络抖动）
4. 结构化存储（JSON/数据库）
5. 实时监控（进度、错误率）
*/

// ============================================================
// 领域模型
// ============================================================

// URL 原始URL
type URL struct {
	Address string
	Depth   int // 爬取深度
}

// HTMLPage 下载的HTML
type HTMLPage struct {
	URL     string
	Content string
	Status  int
}

// Product 解析后的商品信息
type Product struct {
	ID          string
	Name        string
	Price       float64
	SourceURL   string
	ExtractedAt time.Time
}

// CrawlResult 最终存储结果
type CrawlResult struct {
	Product Product
	Success bool
	Error   error
}

// ============================================================
// 爬虫统计
// ============================================================

type CrawlerStats struct {
	totalURLs      atomic.Int64
	successCount   atomic.Int64
	failureCount   atomic.Int64
	retryCount     atomic.Int64
	rateLimitHits  atomic.Int64
	totalProducts  atomic.Int64
	startTime      time.Time
}

func NewCrawlerStats() *CrawlerStats {
	return &CrawlerStats{
		startTime: time.Now(),
	}
}

func (s *CrawlerStats) Report() {
	elapsed := time.Since(s.startTime)
	total := s.totalURLs.Load()
	success := s.successCount.Load()
	failure := s.failureCount.Load()
	retry := s.retryCount.Load()
	rateLimit := s.rateLimitHits.Load()
	products := s.totalProducts.Load()

	fmt.Printf("\n📊 爬虫统计报告\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("运行时长:     %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("总URL数:      %d\n", total)
	fmt.Printf("成功数:       %d (%.1f%%)\n", success, float64(success)/float64(total)*100)
	fmt.Printf("失败数:       %d\n", failure)
	fmt.Printf("重试次数:     %d\n", retry)
	fmt.Printf("限流命中:     %d\n", rateLimit)
	fmt.Printf("采集商品:     %d\n", products)
	fmt.Printf("平均速度:     %.1f URL/s\n", float64(total)/elapsed.Seconds())
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// ============================================================
// 第一阶段：URL下载器（带限流 + 重试）
// ============================================================

type Downloader struct {
	client      *http.Client
	rateLimiter async.RateLimiter
	stats       *CrawlerStats
}

func NewDownloader(qps int, stats *CrawlerStats) *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		rateLimiter: async.NewTokenBucketRateLimiter(qps, qps*2), // QPS限制
		stats:       stats,
	}
}

func (d *Downloader) Download(ctx context.Context, url URL) (HTMLPage, error) {
	// 限流：等待令牌
	if !d.rateLimiter.Allow() {
		d.stats.rateLimitHits.Add(1)
		time.Sleep(50 * time.Millisecond) // 简单延迟
		if !d.rateLimiter.Allow() {
			return HTMLPage{}, fmt.Errorf("rate limit exceeded")
		}
	}

	// 带重试的下载
	var page HTMLPage
	err := async.Retry(ctx, func() error {
		d.stats.totalURLs.Add(1)

		// 模拟HTTP请求
		resp, err := d.client.Get(url.Address)
		if err != nil {
			d.stats.retryCount.Add(1)
			return err // 重试
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// 5xx错误重试，4xx不重试
			if resp.StatusCode >= 500 {
				d.stats.retryCount.Add(1)
				return fmt.Errorf("server error: %d", resp.StatusCode)
			}
			return async.StopRetrying(fmt.Errorf("client error: %d", resp.StatusCode))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			d.stats.retryCount.Add(1)
			return err
		}

		page = HTMLPage{
			URL:     url.Address,
			Content: string(body),
			Status:  resp.StatusCode,
		}
		d.stats.successCount.Add(1)
		return nil
	},
		async.MaxRetries(3),
		async.ExponentialBackoff(100*time.Millisecond, 2.0),
	)

	if err != nil {
		d.stats.failureCount.Add(1)
	}

	return page, err
}

// ============================================================
// 第二阶段：HTML解析器（并行解析）
// ============================================================

type Parser struct {
	stats *CrawlerStats
}

func NewParser(stats *CrawlerStats) *Parser {
	return &Parser{stats: stats}
}

func (p *Parser) Parse(ctx context.Context, pages []HTMLPage) ([]Product, error) {
	// 使用 async.Map 并行解析
	products, err := async.Map(ctx, pages, async.CPU(), func(ctx context.Context, page HTMLPage) (Product, error) {
		// 模拟解析HTML（实际应该用 goquery 等库）
		product := Product{
			ID:          fmt.Sprintf("PROD-%d", rand.Intn(100000)),
			Name:        extractProductName(page.Content),
			Price:       extractPrice(page.Content),
			SourceURL:   page.URL,
			ExtractedAt: time.Now(),
		}

		p.stats.totalProducts.Add(1)
		return product, nil
	})

	return products, err
}

// 模拟提取商品名称
func extractProductName(html string) string {
	// 实际应该用正则或 DOM 解析
	if strings.Contains(html, "<title>") {
		return "iPhone 15 Pro Max"
	}
	return "Unknown Product"
}

// 模拟提取价格
func extractPrice(html string) float64 {
	return 9999.0 + rand.Float64()*1000.0
}

// ============================================================
// 第三阶段：数据存储器（批量写入）
// ============================================================

type Storage struct {
	stats *CrawlerStats
}

func NewStorage(stats *CrawlerStats) *Storage {
	return &Storage{stats: stats}
}

func (s *Storage) Save(ctx context.Context, products []Product) error {
	// 模拟批量写入数据库
	fmt.Printf("💾 保存 %d 个商品到数据库\n", len(products))

	// 使用 async.ForEach 并发写入
	return async.ForEach(ctx, products, async.IO(), func(ctx context.Context, p Product) error {
		// 模拟数据库写入延迟
		time.Sleep(5 * time.Millisecond)
		// fmt.Printf("  ✓ 已保存: %s - ¥%.2f\n", p.Name, p.Price)
		return nil
	})
}

// ============================================================
// 爬虫引擎：Pipeline 编排
// ============================================================

type CrawlerEngine struct {
	downloader *Downloader
	parser     *Parser
	storage    *Storage
	stats      *CrawlerStats
	pipeline   *async.Pipeline[URL, HTMLPage, []Product, error]
}

func NewCrawlerEngine(qps int) *CrawlerEngine {
	stats := NewCrawlerStats()
	downloader := NewDownloader(qps, stats)
	parser := NewParser(stats)
	storage := NewStorage(stats)

	// 构建三阶段 Pipeline：URL → HTML → Product → 存储
	pipeline := async.NewPipeline[URL, HTMLPage, []Product, error]().
		Stage("下载", 10, func(ctx context.Context, url URL) (HTMLPage, error) {
			return downloader.Download(ctx, url)
		}).
		Stage("解析", 5, func(ctx context.Context, pages []HTMLPage) ([]Product, error) {
			return parser.Parse(ctx, pages)
		}).
		Stage("存储", 2, func(ctx context.Context, products []Product) (error, error) {
			return storage.Save(ctx, products), nil
		})

	return &CrawlerEngine{
		downloader: downloader,
		parser:     parser,
		storage:    storage,
		stats:      stats,
		pipeline:   pipeline,
	}
}

func (e *CrawlerEngine) Crawl(ctx context.Context, seedURLs []URL) error {
	fmt.Printf("🚀 启动爬虫引擎...\n")
	fmt.Printf("📌 种子URL数: %d\n", len(seedURLs))
	fmt.Printf("📌 限流QPS: 10\n")
	fmt.Printf("📌 最大并发: 100\n\n")

	// 使用 BoundedRunner 控制整体并发
	runner := async.NewBoundedRunner(100) // 最多100个并发任务

	group := async.NewGroup[CrawlResult]()

	// 提交所有URL到 BoundedRunner
	for _, url := range seedURLs {
		url := url // 捕获循环变量
		group.Go(func(ctx context.Context) (CrawlResult, error) {
			// 限制并发
			return runner.Run(ctx, func(ctx context.Context) (CrawlResult, error) {
				// 下载
				page, err := e.downloader.Download(ctx, url)
				if err != nil {
					return CrawlResult{Success: false, Error: err}, nil
				}

				// 解析
				products, err := e.parser.Parse(ctx, []HTMLPage{page})
				if err != nil {
					return CrawlResult{Success: false, Error: err}, nil
				}

				// 存储
				if len(products) > 0 {
					err = e.storage.Save(ctx, products)
					if err != nil {
						return CrawlResult{Success: false, Error: err}, nil
					}
					return CrawlResult{Product: products[0], Success: true}, nil
				}

				return CrawlResult{Success: true}, nil
			})
		})
	}

	// 等待所有任务完成
	results, err := group.Wait(ctx)
	if err != nil {
		return err
	}

	// 统计结果
	for _, result := range results {
		if !result.Success && result.Error != nil {
			fmt.Printf("❌ 错误: %v\n", result.Error)
		}
	}

	e.stats.Report()
	return nil
}

// ============================================================
// 主程序：真实场景演示
// ============================================================

func main() {
	fmt.Println("========================================")
	fmt.Println("真实爬虫系统 - async 库综合实战")
	fmt.Println("========================================\n")

	// 场景1：小规模测试（10个URL）
	fmt.Println("📋 场景1：小规模测试（10个URL）")
	fmt.Println("----------------------------------------")
	testSmallScale()

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// 场景2：中等规模（100个URL）
	fmt.Println("📋 场景2：中等规模（100个URL + 限流）")
	fmt.Println("----------------------------------------")
	testMediumScale()

	fmt.Println("\n" + strings.Repeat("=", 80) + "\n")

	// 场景3：大规模模拟（1000个URL）
	fmt.Println("📋 场景3：大规模模拟（1000个URL + 并发控制）")
	fmt.Println("----------------------------------------")
	testLargeScale()

	fmt.Println("\n✅ 所有场景测试完成！")
	fmt.Println("\n💡 关键要点总结：")
	fmt.Println("   1. BoundedRunner 控制总并发数，防止资源耗尽")
	fmt.Println("   2. RateLimiter 保护目标服务器，避免被封")
	fmt.Println("   3. Retry 处理网络抖动，提高成功率")
	fmt.Println("   4. Pipeline 清晰分离各阶段逻辑")
	fmt.Println("   5. Group 管理所有任务，统一等待")
	fmt.Println("   6. Map 并行处理数据转换")
}

// ============================================================
// 测试场景
// ============================================================

func testSmallScale() {
	ctx := context.Background()
	engine := NewCrawlerEngine(10) // 10 QPS

	// 生成10个测试URL
	urls := generateTestURLs(10)

	start := time.Now()
	err := engine.Crawl(ctx, urls)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 爬取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 爬取完成，耗时: %v\n", elapsed.Round(time.Millisecond))
	}
}

func testMediumScale() {
	ctx := context.Background()
	engine := NewCrawlerEngine(20) // 20 QPS

	urls := generateTestURLs(100)

	start := time.Now()
	err := engine.Crawl(ctx, urls)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 爬取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 爬取完成，耗时: %v\n", elapsed.Round(time.Millisecond))
	}
}

func testLargeScale() {
	ctx := context.Background()
	engine := NewCrawlerEngine(50) // 50 QPS

	urls := generateTestURLs(1000)

	start := time.Now()
	err := engine.Crawl(ctx, urls)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 爬取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 爬取完成，耗时: %v\n", elapsed.Round(time.Millisecond))
	}
}

// 生成测试URL
func generateTestURLs(count int) []URL {
	urls := make([]URL, count)
	domains := []string{
		"https://example.com/product/",
		"https://shop.example.com/item/",
		"https://store.example.com/goods/",
	}

	for i := 0; i < count; i++ {
		urls[i] = URL{
			Address: fmt.Sprintf("%s%d", domains[i%len(domains)], i),
			Depth:   rand.Intn(3) + 1,
		}
	}
	return urls
}

/*
🤔 思考题：
1. 为什么要用 BoundedRunner 包裹整个任务？
   → 控制总并发数，防止创建过多 goroutine 导致 OOM

2. Pipeline 和直接用 Group 有什么区别？
   → Pipeline 强制按阶段流转，清晰分离关注点
   → Group 只是并发执行，不关心数据流向

3. 限流器为什么放在 Downloader 里而不是 Engine？
   → 下载是唯一的外部请求点，集中限流最有效
   → Engine 层次太高，无法精准控制

4. 重试策略为什么区分 5xx 和 4xx？
   → 5xx 是服务器错误，可能临时，值得重试
   → 4xx 是客户端错误（如404），重试无意义

5. 如何应对反爬虫？
   → 添加 User-Agent 和 Cookie
   → 使用代理池（IP轮换）
   → 降低 QPS，模拟人工行为
   → 添加随机延迟

6. 如何实现增量爬取（去重）？
   → 使用 Bloom Filter 或 Redis Set 记录已爬URL
   → 在 Downloader 中检查，跳过已爬页面

7. 如何处理动态渲染页面（JS）？
   → 使用无头浏览器（如 chromedp）
   → 或分析 AJAX 接口，直接请求 JSON 数据

8. 如何做分布式爬虫？
   → URL队列用 Kafka/Redis
   → BoundedRunner 改为分布式任务队列（如 Asynq）
   → 用数据库协调各节点状态

🎯 通关标准：
✅ 理解爬虫的四个核心组件（下载、解析、存储、调度）
✅ 掌握 async 库的综合运用（6个核心API）
✅ 能够设计限流 + 重试 + 并发控制的容错系统
✅ 能够根据业务需求调整并发度和 QPS
✅ 理解反爬虫对抗和分布式扩展方案

下一步：
→ 阅读项目 README.md，系统学习所有场景
→ 修改参数（QPS、并发数），观察性能变化
→ 添加真实的 HTTP 请求，爬取实际网站
→ 集成 goquery 解析 HTML，提取真实数据
→ 添加持久化存储（MySQL/MongoDB）
*/
