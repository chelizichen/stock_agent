package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"stock_agent/config"
	"stock_agent/tools"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai"
)

func main() {
	ctx := context.Background()

	// 加载配置文件
	config, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化 Genkit + OpenAI
	g := genkit.Init(ctx, genkit.WithPlugins(
		&compat_oai.OpenAICompatible{
			Provider: config.AI.Provider,
			APIKey:   config.AI.APIKey,
			BaseURL:  config.AI.BaseURL,
		},
	)) 

	// 设置全局genkit实例（供tools使用）
	tools.SetGenkitInstance(g)
	// 查询农业银行相关股票信息，爬取30条新闻，并生成分析报告，并生成Markdown报告
	// 定义工具
	toolList := tools.InitTools(g)

	// 多轮对话历史
	var history []*ai.Message

	fmt.Println("🤖 股票行情查询Agent（支持搜索新闻、AI分析和Markdown生成）")
	fmt.Println("输入 'exit' 退出，例如：")
	fmt.Println("  - 帮我查询腾讯的股票新闻并生成分析报告")
	fmt.Println("  - 搜索阿里巴巴的最新30条新闻")
	fmt.Println("  - 分析AAPL的股票新闻并导出Markdown文件")

	history = append(history, ai.NewMessage(ai.RoleSystem, map[string]any{}, ai.NewTextPart(`
		你是一位专业的股票分析师,当用户输入股票关键词时，请先搜索相关新闻，然后基于新闻内容，输出一份详细的股票分析报告，采用markdown格式
		并将输入的内容导出为Markdown文件。
		注： 生成文档的最新日期按照 爬取内容的最新日期为准。
		`,
	)))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n你: ")
		if !scanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}
		if strings.ToLower(userInput) == "exit" {
			fmt.Println("👋 再见！")
			break
		}

		// 添加用户消息
		history = append(history, ai.NewUserMessage(ai.NewTextPart(userInput)))

		// 调用模型（自动处理工具调用循环）
		fmt.Print("AI: ")
		resp, err := genkit.Generate(ctx, g,
			ai.WithModelName(config.AI.ModelName),
			ai.WithMessages(history...),
			ai.WithTools(toolList...),
			ai.WithMaxTurns(10), // 最多10轮工具调用循环
		)
		if err != nil {
			fmt.Printf("\n❌ 错误: %v\n", err)
			log.Printf("详细错误: %+v", err)
			continue
		}

		// 打印最终文本（如果有工具调用，会自动执行后给出完整回答）
		text := resp.Text()
		fmt.Println(text)

		// 将 AI 回复加入历史
		history = append(history, ai.NewModelMessage(ai.NewTextPart(text)))
	}
}
