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
	totalURLs     atomic.Int64
	successCount  atomic.Int64
	failureCount  atomic.Int64
	retryCount    atomic.Int64
	rateLimitHits atomic.Int64
	totalProducts atomic.Int64
	startTime     time.Time
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

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 爬虫统计报告")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏱️  运行时长:     %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("📋 总URL数:      %d\n", total)
	if total > 0 {
		fmt.Printf("✅ 成功数:       %d (%.1f%%)\n", success, float64(success)/float64(total)*100)
	} else {
		fmt.Printf("✅ 成功数:       %d\n", success)
	}
	fmt.Printf("❌ 失败数:       %d\n", failure)
	fmt.Printf("🔁 重试次数:     %d\n", retry)
	fmt.Printf("🚦 限流命中:     %d\n", rateLimit)
	fmt.Printf("📦 采集商品:     %d\n", products)
	if elapsed.Seconds() > 0 {
		fmt.Printf("⚡ 平均速度:     %.1f URL/s\n", float64(total)/elapsed.Seconds())
	}
	fmt.Println(strings.Repeat("=", 60))
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
	ctx := context.Background()
	return &Downloader{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		rateLimiter: async.Ratelimit(ctx).
			TokenBucket().
			Rate(qps).
			Capacity(qps * 2).
			Build(),
		stats: stats,
	}
}

func (d *Downloader) Download(ctx context.Context, url URL) (HTMLPage, error) {
	// 限流：等待令牌
	if err := d.rateLimiter.Allow(); err != nil {
		d.stats.rateLimitHits.Add(1)
		time.Sleep(50 * time.Millisecond) // 简单延迟
	}

	// 带重试的下载
	result, err := async.Retry[HTMLPage](ctx).
		MaxRetries(3).
		Exponential().
		Backoff(100*time.Millisecond, 5*time.Second).
		Execute(func(ctx context.Context) (HTMLPage, error) {
			d.stats.totalURLs.Add(1)

			// 模拟HTTP请求
			resp, err := d.client.Get(url.Address)
			if err != nil {
				d.stats.retryCount.Add(1)
				return HTMLPage{}, err // 重试
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				// 5xx错误重试，4xx不重试
				if resp.StatusCode >= 500 {
					d.stats.retryCount.Add(1)
					return HTMLPage{}, fmt.Errorf("server error: %d", resp.StatusCode)
				}
				d.stats.failureCount.Add(1)
				return HTMLPage{}, fmt.Errorf("client error: %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				d.stats.retryCount.Add(1)
				return HTMLPage{}, err
			}

			page := HTMLPage{
				URL:     url.Address,
				Content: string(body),
				Status:  resp.StatusCode,
			}
			d.stats.successCount.Add(1)
			return page, nil
		})

	if err != nil {
		d.stats.failureCount.Add(1)
		return HTMLPage{}, err
	}

	return result, nil
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
	result := async.Slice(ctx, pages).
		Worker(async.CPU()).
		Map(func(ctx context.Context, page HTMLPage) (Product, error) {
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

	return result.Values(), nil
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
	if len(products) > 0 {
		fmt.Printf("     💾 保存 %d 个商品到数据库\n", len(products))
	}

	// 使用 async.ForEach 并发写入
	result := async.Slice(ctx, products).
		Worker(async.IO()).
		ForEach(func(ctx context.Context, p Product) error {
			// 模拟数据库写入延迟
			time.Sleep(5 * time.Millisecond)
			return nil
		})

	if !result.Ok() {
		return result.Error()
	}
	return nil
}

// ============================================================
// 爬虫引擎：综合运用
// ============================================================

type CrawlerEngine struct {
	downloader *Downloader
	parser     *Parser
	storage    *Storage
	stats      *CrawlerStats
	runner     *async.BoundedRunner
}

func NewCrawlerEngine(qps int) *CrawlerEngine {
	stats := NewCrawlerStats()
	downloader := NewDownloader(qps, stats)
	parser := NewParser(stats)
	storage := NewStorage(stats)
	runner := async.NewBoundedRunnerBuilder().
		Max(100). // 最多100个并发任务
		Build()

	return &CrawlerEngine{
		downloader: downloader,
		parser:     parser,
		storage:    storage,
		stats:      stats,
		runner:     runner,
	}
}

func (e *CrawlerEngine) Crawl(ctx context.Context, seedURLs []URL) error {
	fmt.Println("🚀 启动爬虫引擎...")
	fmt.Printf("📌 种子URL数: %d\n", len(seedURLs))
	fmt.Printf("📌 限流QPS: 10\n")
	fmt.Printf("📌 最大并发: 100\n\n")

	group := async.Group[CrawlResult]().
		Context(ctx).
		Build()
	defer group.Close()

	// 提交所有URL到 BoundedRunner
	for _, url := range seedURLs {
		url := url // 捕获循环变量
		group.Go(ctx, func(ctx context.Context) (CrawlResult, error) {
			// 限制并发
			return async.BoundedGo(e.runner, ctx, func(ctx context.Context) (CrawlResult, error) {
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
	groupResults := group.Wait()
	var results []CrawlResult
	err := error(nil)
	for _, r := range groupResults {
		if r.Err != nil && err == nil {
			err = r.Err
		}
		results = append(results, r.Value)
	}

	if err != nil {
		// 不返回错误，继续统计
	}

	// 统计结果
	errorCount := 0
	for _, result := range results {
		if !result.Success && result.Error != nil {
			errorCount++
			if errorCount <= 3 {
				fmt.Printf("     ❌ 错误: %v\n", result.Error)
			}
		}
	}
	if errorCount > 3 {
		fmt.Printf("     ... 还有 %d 个错误\n", errorCount-3)
	}

	e.stats.Report()
	return nil
}

// ============================================================
// 主程序：真实场景演示
// ============================================================

func main() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🚀 真实爬虫系统 - async 库综合实战")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	// 场景1：小规模测试（10个URL）
	fmt.Println("📝 场景1：小规模测试（10个URL）")
	fmt.Println(strings.Repeat("-", 60))
	testSmallScale()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// 场景2：中等规模（100个URL）
	fmt.Println("📝 场景2：中等规模（100个URL + 限流）")
	fmt.Println(strings.Repeat("-", 60))
	testMediumScale()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// 场景3：大规模模拟（1000个URL）
	fmt.Println("📝 场景3：大规模模拟（1000个URL + 并发控制）")
	fmt.Println(strings.Repeat("-", 60))
	testLargeScale()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("💡 关键要点总结")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  ✅ BoundedRunner 控制总并发数，防止资源耗尽")
	fmt.Println("  ✅ RateLimiter 保护目标服务器，避免被封")
	fmt.Println("  ✅ Retry 处理网络抖动，提高成功率")
	fmt.Println("  ✅ Map 并行处理数据转换")
	fmt.Println("  ✅ Group 管理所有任务，统一等待")
	fmt.Println("  ✅ ForEach 批量存储数据")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
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
		fmt.Printf("✅ 爬取完成，⏱️  耗时: %v\n", elapsed.Round(time.Millisecond))
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
		fmt.Printf("✅ 爬取完成，⏱️  耗时: %v\n", elapsed.Round(time.Millisecond))
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
		fmt.Printf("✅ 爬取完成，⏱️  耗时: %v\n", elapsed.Round(time.Millisecond))
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

2. 限流器为什么放在 Downloader 里而不是 Engine？
   → 下载是唯一的外部请求点，集中限流最有效

3. 重试策略为什么区分 5xx 和 4xx？
   → 5xx 是服务器错误，可能临时，值得重试
   → 4xx 是客户端错误（如404），重试无意义

4. 如何应对反爬虫？
   → 添加 User-Agent 和 Cookie
   → 使用代理池（IP轮换）
   → 降低 QPS，模拟人工行为
   → 添加随机延迟

🎯 通关标准：
✅ 理解爬虫的四个核心组件（下载、解析、存储、调度）
✅ 掌握 async 库的综合运用（6个核心API）
✅ 能够设计限流 + 重试 + 并发控制的容错系统
✅ 能够根据业务需求调整并发度和 QPS
*/
