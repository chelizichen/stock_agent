package tools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/firebase/genkit/go/ai"
)

// SearchNewsInput 搜索新闻的输入参数
type SearchNewsInput struct {
	Keyword string `json:"keyword" jsonschema_description:"要查询的股票关键词，例如：腾讯、阿里巴巴、AAPL等"`
}

// NewsItem 新闻项结构
type NewsItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
	Time    string `json:"time"`
}

// SearchStockNews 搜索股票相关新闻（Genkit Tool）
func SearchStockNews(ctx *ai.ToolContext, input SearchNewsInput) ([]NewsItem, error) {
	// 创建带超时的context（20分钟超时，给爬取足够时间）
	log.Printf("搜索财联社新闻: %s", input.Keyword)
	searchCtx, cancel := context.WithTimeout(ctx.Context, 20*time.Minute)
	defer cancel()

	// 使用公共方法获取浏览器
	browser, err := getBrowser()
	if err != nil {
		return nil, fmt.Errorf("财联社电报频道新闻获取浏览器失败: %v", err)
	}
	defer browser.Close()

	// 检查context是否已取消
	select {
	case <-searchCtx.Done():
		return nil, fmt.Errorf("财联社电报频道新闻搜索超时: %v", searchCtx.Err())
	default:
	}

	// 使用公共方法安全创建页面
	page, err := createPageSafely(browser, 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("财联社电报频道新闻创建页面失败: %v", err)
	}
	defer page.Close()

	newsItems := make([]NewsItem, 0, 1)
	channelURL := getClsChannel(input.Keyword)
	channelNewsItem, err := fetchNewsContent(context.Background(), browser, channelURL, input.Keyword+"-电报频道", 60*time.Second)
	if err != nil {
		log.Printf("财联社电报频道新闻爬取失败: %v", err)
	}
	newsItems = append(newsItems, channelNewsItem)
	log.Printf("财联社电报频道新闻爬取成功，共获取 %d 条新闻", len(newsItems))
	return newsItems, nil
}

func getClsChannel(keyword string) string {
	return fmt.Sprintf(`https://www.cls.cn/searchPage?keyword=%s&type=telegram`, keyword)
}

func getRealtimeInfo(keyword string) string {
	return fmt.Sprintf(`https://www.cls.cn/searchPage?keyword=%s&type=depth`, keyword)
}
