package tools

import (
	"fmt"
	"log"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type AnalyzeInput struct {
	Keyword string `json:"keyword" jsonschema_description:"用户想要分析的板块关键词，通过先查询板块，再查出板块内的股票，再分析股票"`
}

func Analyze(ctx *ai.ToolContext, input AnalyzeInput) (string, error) {
	fmt.Printf("分析用户输入: %s\n", input.Keyword)
	g := getGenkitInstance()
	if g == nil {
		return "", fmt.Errorf("genkit实例未初始化")
	}

	prompt := fmt.Sprintf(`从用户的输入中，分析用户想要了解哪些股票,并返回股票列表,最多返回三个,不要返回任何其他内容。
	用户输入: %s`, input.Keyword)
	resp, err := genkit.Generate(ctx, g,
		ai.WithModelName("xiaomimimo/mimo-v2-flash"),
		ai.WithMessages(ai.NewUserMessage(ai.NewTextPart(prompt))),
		ai.WithMaxTurns(1),
	)
	if err != nil {
		return "", fmt.Errorf("分析失败: %v", err)
	}
	log.Printf("分析用户输入结果: %s", resp.Text())
	return resp.Text(), nil
}
