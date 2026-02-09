package document

import "testing"

func TestPdfConverter(t *testing.T) {
	opts := DefaultConvertOptions()
	opts.PreserveFormatting = true
	opts.ExtractTables = true
	//converter := NewPDFConverter(opts)
	converter := NewWordConverter(opts)
	//converter.ToMarkdownFile("F:\\project\\jaime\\go_project\\ai-apps\\bidding\\resources\\国泰海通基于大模型的智能运维平台全链路诊断项目采购需求.pdf", "F:\\project\\jaime\\go_project\\ai-apps\\bidding\\resources\\markdown\\国泰海通基于大模型的智能运维平台全链路诊断项目采购.md")
	converter.ToMarkdownFile("F:\\文档中心\\中畅科技\\标书\\全链路诊断\\国泰海通基于大模型的智能运维平台全链路诊断项目采购回标分析-20251118.docx", "F:\\project\\jaime\\go_project\\ai-apps\\bidding\\resources\\markdown\\国泰海通基于大模型的智能运维平台全链路诊断项目采购回标分析.md")
}
