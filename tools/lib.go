package tools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// getBrowser 获取浏览器实例（公共方法）
func getBrowser() (*rod.Browser, error) {
	// 创建带超时的context（20分钟超时，给爬取足够时间）
	searchCtx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	// 启动浏览器（使用无头模式，添加更多配置避免被检测）
	launcher := launcher.New().
		Headless(true).
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-dev-shm-usage").
		Set("no-sandbox").
		UserDataDir("").
		Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	browserURL, err := launcher.Launch()
	if err != nil {
		return nil, fmt.Errorf("启动浏览器失败: %v", err)
	}
	log.Printf("浏览器URL: %s", browserURL)
	browser := rod.New().ControlURL(browserURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("连接浏览器失败: %v", err)
	}

	// 设置浏览器超时
	browser = browser.Timeout(20 * time.Minute)

	// 检查context是否已取消
	select {
	case <-searchCtx.Done():
		return nil, fmt.Errorf("搜索超时: %v", searchCtx.Err())
	default:
	}
	return browser, nil
}

// createPageSafely 安全创建页面（捕获panic，避免程序崩溃）
func createPageSafely(browser *rod.Browser, timeout time.Duration) (*rod.Page, error) {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	var page *rod.Page
	var pageErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				pageErr = fmt.Errorf("创建页面时发生错误: %v", r)
			}
		}()
		page = browser.Timeout(timeout).MustPage()
	}()

	if pageErr != nil {
		return nil, pageErr
	}
	if page == nil {
		return nil, fmt.Errorf("创建页面失败: 页面为nil")
	}

	// 设置页面超时
	page = page.Timeout(timeout)
	return page, nil
}

// navigateAndWait 导航到URL并等待页面加载完成
func navigateAndWait(page *rod.Page, url string, waitStable bool) error {
	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("导航失败: %v", err)
	}

	// 等待页面加载
	page.WaitLoad()
	time.Sleep(1 * time.Second)

	// 等待页面稳定（可选）
	if waitStable {
		page.Timeout(20 * time.Second).WaitStable(2 * time.Second)
		time.Sleep(2 * time.Second)
	}

	// 验证页面是否成功加载
	pageInfo, err := page.Info()
	if err != nil {
		return fmt.Errorf("获取页面信息失败: %v", err)
	}
	if pageInfo == nil || pageInfo.Title == "" {
		return fmt.Errorf("页面加载失败：标题为空")
	}

	return nil
}

// findElementsWithSelectors 使用多个选择器尝试查找元素（返回第一个成功的）
func findElementsWithSelectors(page *rod.Page, selectors []string) (rod.Elements, error) {
	var elements rod.Elements
	var lastErr error

	for _, selector := range selectors {
		elements, lastErr = page.Elements(selector)
		if lastErr == nil && len(elements) > 0 {
			log.Printf("使用选择器 '%s' 找到 %d 个元素", selector, len(elements))
			return elements, nil
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("所有选择器都失败，最后一个错误: %v", lastErr)
	}
	return nil, fmt.Errorf("所有选择器都未找到元素")
}

// fetchNewsContent 获取单个新闻页面的内容
func fetchNewsContent(ctx context.Context, browser *rod.Browser, url, title string, timeout time.Duration) (NewsItem, error) {
	// 检查context是否已取消
	select {
	case <-ctx.Done():
		return NewsItem{}, fmt.Errorf("操作已取消: %v", ctx.Err())
	default:
	}

	// 为每个页面创建独立的超时context，避免影响其他爬取
	// 使用独立的context，不依赖父context，避免一个失败影响其他
	pageCtx, pageCancel := context.WithTimeout(context.Background(), timeout)
	defer pageCancel()

	// 使用非panic版本，避免超时panic
	// 使用 MustPage 但捕获可能的 panic
	var newPage *rod.Page
	var pageErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				pageErr = fmt.Errorf("创建页面时发生错误: %v", r)
			}
		}()
		// 检查context是否已取消
		select {
		case <-pageCtx.Done():
			pageErr = fmt.Errorf("操作已取消: %v", pageCtx.Err())
			return
		default:
		}
		// 使用独立的超时创建页面，避免影响浏览器状态
		newPage = browser.Timeout(timeout).MustPage()
	}()
	if pageErr != nil {
		return NewsItem{}, pageErr
	}
	if newPage == nil {
		return NewsItem{}, fmt.Errorf("创建页面失败: 页面为nil")
	}

	// 确保页面一定会被关闭，即使发生错误
	pageClosed := false
	defer func() {
		if !pageClosed && newPage != nil {
			// 直接关闭，不异步，确保资源释放
			if closeErr := newPage.Close(); closeErr != nil {
				// 忽略关闭错误，避免影响其他爬取
				log.Printf("关闭页面失败（已忽略）: %v", closeErr)
			}
		}
	}()

	// 增加页面超时时间
	newPage = newPage.Timeout(timeout * 2) // 给页面操作更多时间

	// 注意：用户代理已在 launcher 中设置，无需在页面中再次设置

	// 重试机制：最多重试2次
	maxRetries := 2
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 检查context是否已取消
		select {
		case <-pageCtx.Done():
			pageClosed = true
			return NewsItem{}, fmt.Errorf("操作已取消: %v", pageCtx.Err())
		default:
		}

		if attempt > 1 {
			log.Printf("  重试连接 %s (第 %d 次)...", url, attempt)
			time.Sleep(2 * time.Second)
		}

		// 导航到新闻页面
		err := newPage.Navigate(url)
		if err == nil {
			// 导航成功，检查页面是否加载
			newPage.WaitLoad()
			time.Sleep(1 * time.Second)
			// 使用非panic版本，避免超时panic
			pageInfo, err := newPage.Info()
			if err == nil && pageInfo != nil && pageInfo.Title != "" {
				lastErr = nil
				break
			}
			if err != nil {
				lastErr = err
			}
		} else {
			lastErr = err
		}
	}

	if lastErr != nil {
		return NewsItem{}, fmt.Errorf("导航失败（已重试%d次）: %v", maxRetries, lastErr)
	}

	// 等待页面加载，使用更宽松的超时
	newPage.WaitLoad()
	done := make(chan error, 1)
	go func() {
		// 增加等待稳定时间
		newPage.Timeout(30 * time.Second).WaitStable(3 * time.Second)
		done <- nil
	}()

	select {
	case <-pageCtx.Done():
		// context已取消，返回错误
		pageClosed = true
		return NewsItem{}, fmt.Errorf("操作已取消: %v", pageCtx.Err())
	case <-time.After(timeout):
		// 即使超时也继续，可能页面已经部分加载
		log.Printf("  页面加载超时，但继续尝试获取内容")
	case <-done:
	}

	time.Sleep(2 * time.Second) // 等待时间

	// 获取页面纯文本内容（移除所有HTML标签）
	var content string
	jsCode := `
		(function() {
			var scripts = document.querySelectorAll('script, style, noscript');
			scripts.forEach(function(el) { el.remove(); });
			var body = document.body;
			if (!body) return '';
			var walker = document.createTreeWalker(body, NodeFilter.SHOW_TEXT, null, false);
			var textNodes = [];
			var node;
			while (node = walker.nextNode()) {
				var text = node.textContent.trim();
				if (text.length > 0) {
					textNodes.push(text);
				}
			}
			return textNodes.join('\n');
		})();
	`

	jsResult, err := newPage.Eval(jsCode)
	if err == nil && jsResult != nil {
		text := jsResult.Value.String()
		if len(text) > 2 && text[0] == '"' && text[len(text)-1] == '"' {
			text = text[1 : len(text)-1]
		}
		text = strings.ReplaceAll(text, "\\n", "\n")
		text = strings.ReplaceAll(text, "\\\"", "\"")
		if len(text) > 100 {
			content = text
		}
	}

	// 如果JavaScript获取失败，尝试传统方法
	if content == "" {
		selectors := []string{"article", ".article-content", ".content", "#content", ".post-content", ".news-content", "main", ".main-content"}
		for _, selector := range selectors {
			elements, err := newPage.Elements(selector)
			if err == nil && len(elements) > 0 {
				text, err := elements[0].Text()
				if err == nil && len(text) > 100 {
					content = text
					break
				}
			}
		}

		if content == "" {
			body, err := newPage.Element("body")
			if err == nil {
				content, _ = body.Text()
			}
		}
	}

	// 清理内容
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n\n\n", "\n\n")

	// 限制内容长度
	if len(content) > 5000 {
		content = content[:5000] + "..."
	}

	return NewsItem{
		Title:   title,
		Content: content,
		URL:     url,
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}
