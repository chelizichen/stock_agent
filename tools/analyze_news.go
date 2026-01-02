package tools

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

// AnalyzeNewsInput 分析新闻的输入参数
type AnalyzeNewsInput struct {
	Keyword   string     `json:"keyword" jsonschema_description:"股票关键词"`
	NewsItems []NewsItem `json:"newsItems" jsonschema_description:"要分析的新闻列表"`
}

// AnalyzeStockNews 分析股票新闻（Genkit Tool）
func AnalyzeStockNews(ctx *ai.ToolContext, input AnalyzeNewsInput) (string, error) {
	log.Printf("开始分析新闻: %s", input.Keyword)
	if len(input.NewsItems) == 0 {
		return "未找到相关新闻，建议手动查询。", nil
	}

	g := getGenkitInstance()
	if g == nil {
		return "", fmt.Errorf("genkit实例未初始化")
	}

	// 收集所有新闻内容，构建提示词
	var newsContent strings.Builder
	fmt.Fprintf(&newsContent, "请分析以下关于 %s 股票的新闻，并生成一份专业的分析报告。\n\n", input.Keyword)
	fmt.Fprintf(&newsContent, "共收集到 %d 条相关新闻：\n\n", len(input.NewsItems))

	// 限制每条新闻的内容长度，避免超出token限制
	// 只处理前20条新闻，避免内容过多导致超时
	maxNewsItems := 20
	if len(input.NewsItems) > maxNewsItems {
		input.NewsItems = input.NewsItems[:maxNewsItems]
	}

	for i, item := range input.NewsItems {
		fmt.Fprintf(&newsContent, "新闻 %d:\n", i+1)
		fmt.Fprintf(&newsContent, "标题: %s\n", item.Title)
		fmt.Fprintf(&newsContent, "URL: %s\n", item.URL)
		if item.Content != "" {
			// 限制每条新闻内容长度（保留前1000字符，减少token使用）
			content := item.Content
			if len(content) > 1000 {
				content = content[:1000] + "..."
			}
			fmt.Fprintf(&newsContent, "内容摘要: %s\n", content)
		}
		fmt.Fprintf(&newsContent, "\n")
	}

	// 构建AI提示词
	prompt := fmt.Sprintf(`你是一位专业的股票分析师。请基于以下新闻内容，输出一份详细的股票分析报告，采用markdown格式。

要求：
1. 分析市场情绪（正面、负面、中性）
2. 总结关键信息点，并列出相关新闻的URL和段落摘要
3. 评估潜在风险和机会
4. 给出投资建议（仅供参考）
5. 使用中文回答，格式清晰易读

新闻内容：
%s

请输出分析报告：`, newsContent.String())

	// 创建带超时的context（5分钟超时）
	genkitCtx, cancel := context.WithTimeout(ctx.Context, 5*time.Minute)
	defer cancel()

	// 调用AI生成分析
	resp, err := genkit.Generate(genkitCtx, g,
		ai.WithModelName("xiaomimimo/mimo-v2-flash"),
		ai.WithMessages(ai.NewUserMessage(ai.NewTextPart(prompt))),
		ai.WithMaxTurns(1),
	)
	if err != nil {
		return "", fmt.Errorf("AI分析失败: %v", err)
	}

	analysis := resp.Text()
	log.Printf("AI分析结果: %s\n", analysis)
	return analysis, nil
}

var globalGenkit *genkit.Genkit

func SetGenkitInstance(g *genkit.Genkit) {
	globalGenkit = g
}

func getGenkitInstance() *genkit.Genkit {
	return globalGenkit
}
