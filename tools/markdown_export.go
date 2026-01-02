package tools

import (
	"fmt"
	"os"

	"github.com/firebase/genkit/go/ai"
)

// MarkdownExportInput 导出Markdown的输入参数

type MarkdownExportInput struct {
	Analysis string `json:"analysis" jsonschema_description:"AI分析结果"`
	Keyword  string `json:"keyword" jsonschema_description:"股票关键词"`
}

// MarkdownExport 导出Markdown（Genkit Tool）
func MarkdownExport(ctx *ai.ToolContext, input MarkdownExportInput) (string, error) {
	// 创建Markdown文件
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前目录失败: %v", err)
	}
	markdownFile, err := os.Create(fmt.Sprintf("%s/markdown/analysis_%s.md", dir, input.Keyword))
	if err != nil {
		return "", fmt.Errorf("创建Markdown文件失败: %v", err)
	}
	defer markdownFile.Close()

	// 写入分析结果
	_, err = markdownFile.WriteString(input.Analysis)
	if err != nil {
		return "", fmt.Errorf("写入Markdown文件失败: %v", err)
	}
	return "Markdown文件已创建", nil
}
