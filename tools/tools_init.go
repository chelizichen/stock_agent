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
		"使用AI分析股票相关新闻，生成专业的分析报告，包括市场情绪分析、关键信息总结、风险评估和投资建议。",
		AnalyzeStockNews,
	)

	markdownExportTool := genkit.DefineTool[MarkdownExportInput, string](
		g,
		"markdownExport",
		"将AI分析结果导出为Markdown文件。",
		MarkdownExport,
	)

	toolList := []ai.ToolRef{searchNewsTool, xqSearchStockTool, analyzeNewsTool, markdownExportTool}
	return toolList
}
