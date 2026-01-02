package tools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/firebase/genkit/go/ai"
)

// https://xueqiu.com/k?q=
// search__stock__bd search__stock__ai__table tr td

type XqSearchStockInput struct {
	Keyword string `json:"keyword" jsonschema_description:"要查询的股票关键词，例如：腾讯、阿里巴巴、AAPL等"`
}

func XqSearchStock(ctx *ai.ToolContext, input XqSearchStockInput) ([]NewsItem, error) {
	log.Printf("雪球搜索股票: %s", input.Keyword)
	searchCtx, cancel := context.WithTimeout(ctx.Context, 20*time.Minute)
	defer cancel()
	browser, err := getBrowser()
	if err != nil {
		return nil, fmt.Errorf("获取浏览器失败: %v", err)
	}

	defer browser.Close()
	newsItems := make([]NewsItem, 0, 1)
	xqURL := getXqChannel(input.Keyword)

	// 检查context是否已取消
	select {
	case <-searchCtx.Done():
		return nil, fmt.Errorf("搜索超时: %v", searchCtx.Err())
	default:
	}

	// 使用公共方法安全创建页面
	page, err := createPageSafely(browser, 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("创建页面失败: %v", err)
	}
	defer page.Close()

	// 使用公共方法导航并等待页面加载
	if err := navigateAndWait(page, xqURL, true); err != nil {
		return nil, fmt.Errorf("导航失败: %v", err)
	}

	// 使用公共方法查找元素
	selectors := []string{
		".search__stock__bd .search__stock__ai__table tr td a",
		".search__stock__bd.search__stock__ai__table tr td a",
		".search__stock__ai__table tr td a",
		"table.search__stock__ai__table tr td a",
		"tr td a[href*='/S/']",
		"td a[href*='/S/']",
	}
	links, err := findElementsWithSelectors(page, selectors)
	if err != nil {
		log.Printf("警告：未找到任何链接，可能选择器需要调整: %v", err)
		return newsItems, nil
	}
	for _, link := range links {
		href, err := link.Attribute("href")
		if err != nil {
			continue
		}
		stockURL := getXqStock(*href)

		item, err := fetchNewsContent(context.Background(), browser, stockURL, input.Keyword+"-雪球股票", 60*time.Second)
		if err != nil {
			log.Printf("爬取雪球股票失败: %v", err)
			continue
		}
		newsItems = append(newsItems, item)
		log.Printf("爬取雪球股票成功: %s", stockURL)
	}
	return newsItems, nil
}

func getXqChannel(keyword string) string {
	return fmt.Sprintf("https://xueqiu.com/k?q=%s", keyword)
}

func getXqStock(s string) string {
	return fmt.Sprintf("https://xueqiu.com%s", s)
}
