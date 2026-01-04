package tools

import (
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func InitTools(g *genkit.Genkit) []ai.ToolRef {
	searchNewsTool := genkit.DefineTool[SearchNewsInput, []NewsItem](
		g,
		"searchStockNews",
		"财联社股票财经新闻爬取，主要用于爬取个股的实时相关新闻",
		SearchStockNews,
	)

	xqSearchStockTool := genkit.DefineTool[XqSearchStockInput, []NewsItem](
		g,
		"xqSearchStock",
		"雪球股票搜索，主要用于搜索股票相关新闻",
		XqSearchStock,
	)

	analyzeNewsTool := genkit.DefineTool[AnalyzeNewsInput, string](
		g,
		"analyzeStockNews",
		"使用AI分析股票相关新闻，生成专业的分析报告，包括市场情绪分析、关键信息总结、风险评估和投资建议。注意：newsItems 参数必须是数组格式，每个元素是包含 title、content、url、time 字段的对象。通常应该先调用 searchStockNews 或 xqSearchStock 获取新闻列表，然后将结果传递给此工具。",
		AnalyzeStockNews,
	)

	markdownExportTool := genkit.DefineTool[MarkdownExportInput, string](
		g,
		"markdownExport",
		"将AI分析结果导出为Markdown文件。",
		MarkdownExport,
	)

	analyzeInputTool := genkit.DefineTool[AnalyzeInput, string](
		g,
		"analyzeInput",
		"在不确定用户提到哪个个股的情况下，分析用户输入，返回可能的个股列表，最多返回三个",
		Analyze,
	)

	toolList := []ai.ToolRef{analyzeInputTool,searchNewsTool, xqSearchStockTool, analyzeNewsTool, markdownExportTool}
	return toolList
}
