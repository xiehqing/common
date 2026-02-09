package document

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"baliance.com/gooxml/document"
)

// WordConverter Word 文档转换器
type WordConverter struct {
	opts *ConvertOptions
}

// NewWordConverter 创建 Word 转换器
func NewWordConverter(opts *ConvertOptions) *WordConverter {
	if opts == nil {
		opts = DefaultConvertOptions()
	}
	return &WordConverter{
		opts: opts,
	}
}

// ToMarkdown 将 Word 文档转换为 Markdown
func (c *WordConverter) ToMarkdown(input io.Reader) (string, error) {
	// gooxml 需要文件路径，所以需要创建临时文件
	tmpFile, err := os.CreateTemp("", "word-*.docx")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 复制内容到临时文件
	if _, err := io.Copy(tmpFile, input); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	tmpFile.Close()

	return c.extractTextFromWord(tmpFile.Name())
}

// ToMarkdownFile 将 Word 文件转换为 Markdown 文件
func (c *WordConverter) ToMarkdownFile(inputPath, outputPath string) error {
	markdown, err := c.extractTextFromWord(inputPath)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(markdown), 0o644)
}

// SupportedTypes 返回支持的文档类型
func (c *WordConverter) SupportedTypes() []DocumentType {
	return []DocumentType{TypeWord, TypeDOC}
}

// extractTextFromWord 从 Word 文档提取文本并转换为 Markdown
func (c *WordConverter) extractTextFromWord(docPath string) (string, error) {
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(docPath))
	if ext != ".docx" {
		return "", fmt.Errorf("仅支持 .docx 格式，不支持旧的 .doc 格式")
	}

	doc, err := document.Open(docPath)
	if err != nil {
		return "", fmt.Errorf("打开 Word 文档失败: %w", err)
	}
	//defer doc.Close()

	var markdown strings.Builder
	markdown.WriteString("# Word 文档\n\n")

	// 提取段落
	for _, para := range doc.Paragraphs() {
		text := c.extractParagraphText(para)
		if strings.TrimSpace(text) == "" {
			continue
		}

		// 根据样式判断是否是标题
		style := para.Style()
		if style != "" && strings.Contains(strings.ToLower(style), "heading") {
			// 提取标题级别
			level := c.extractHeadingLevel(style)
			markdown.WriteString(strings.Repeat("#", level))
			markdown.WriteString(" ")
			markdown.WriteString(text)
			markdown.WriteString("\n\n")
		} else {
			markdown.WriteString(text)
			markdown.WriteString("\n\n")
		}
	}

	// 提取表格
	if c.opts.ExtractTables {
		tables := doc.Tables()
		if len(tables) > 0 {
			markdown.WriteString("\n## 表格\n\n")
			for i, table := range tables {
				markdown.WriteString(fmt.Sprintf("### 表格 %d\n\n", i+1))
				markdown.WriteString(c.extractTable(table))
				markdown.WriteString("\n\n")
			}
		}
	}

	return markdown.String(), nil
}

// extractParagraphText 提取段落文本，保留格式
func (c *WordConverter) extractParagraphText(para document.Paragraph) string {
	var text strings.Builder

	for _, run := range para.Runs() {
		runText := run.Text()
		if runText == "" {
			continue
		}

		if c.opts.PreserveFormatting {
			// 检查格式
			props := run.Properties()
			isBold := props.IsBold()
			isItalic := props.IsItalic()
			// 应用 Markdown 格式
			if isBold && isItalic {
				text.WriteString("***")
				text.WriteString(runText)
				text.WriteString("***")
			} else if isBold {
				text.WriteString("**")
				text.WriteString(runText)
				text.WriteString("**")
			} else if isItalic {
				text.WriteString("*")
				text.WriteString(runText)
				text.WriteString("*")
			} else {
				text.WriteString(runText)
			}
		} else {
			text.WriteString(runText)
		}
	}

	return text.String()
}

// extractHeadingLevel 提取标题级别
func (c *WordConverter) extractHeadingLevel(styleName string) int {
	styleName = strings.ToLower(styleName)

	// 常见的标题样式名称
	headingMap := map[string]int{
		"heading 1": 1,
		"heading 2": 2,
		"heading 3": 3,
		"heading 4": 4,
		"heading 5": 5,
		"heading 6": 6,
		"heading1":  1,
		"heading2":  2,
		"heading3":  3,
		"heading4":  4,
		"heading5":  5,
		"heading6":  6,
		"标题 1":      1,
		"标题 2":      2,
		"标题 3":      3,
		"标题 4":      4,
		"标题 5":      5,
		"标题 6":      6,
	}

	if level, ok := headingMap[styleName]; ok {
		return level
	}

	// 尝试从样式名称中提取数字
	for i := 1; i <= 6; i++ {
		if strings.Contains(styleName, fmt.Sprintf("%d", i)) {
			return i
		}
	}

	return 2 // 默认为二级标题
}

// extractTable 提取表格为 Markdown
func (c *WordConverter) extractTable(table document.Table) string {
	rows := table.Rows()
	if len(rows) == 0 {
		return ""
	}

	var markdown strings.Builder

	// 提取第一行作为表头
	headerRow := rows[0]
	cells := headerRow.Cells()

	// 表头
	markdown.WriteString("|")
	for _, cell := range cells {
		cellText := c.extractCellText(cell)
		markdown.WriteString(" ")
		markdown.WriteString(cellText)
		markdown.WriteString(" |")
	}
	markdown.WriteString("\n")

	// 分隔符
	markdown.WriteString("|")
	for range cells {
		markdown.WriteString(" --- |")
	}
	markdown.WriteString("\n")

	// 数据行
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		cells := row.Cells()

		markdown.WriteString("|")
		for _, cell := range cells {
			cellText := c.extractCellText(cell)
			markdown.WriteString(" ")
			markdown.WriteString(cellText)
			markdown.WriteString(" |")
		}
		markdown.WriteString("\n")
	}

	return markdown.String()
}

// extractCellText 提取单元格文本
func (c *WordConverter) extractCellText(cell document.Cell) string {
	var text strings.Builder

	for _, para := range cell.Paragraphs() {
		paraText := c.extractParagraphText(para)
		if strings.TrimSpace(paraText) != "" {
			if text.Len() > 0 {
				text.WriteString(" ")
			}
			text.WriteString(strings.TrimSpace(paraText))
		}
	}

	return text.String()
}
