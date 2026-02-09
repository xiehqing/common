package document

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

// PDFConverter PDF 转换器
type PDFConverter struct {
	opts *ConvertOptions
}

// NewPDFConverter 创建 PDF 转换器
func NewPDFConverter(opts *ConvertOptions) *PDFConverter {
	if opts == nil {
		opts = DefaultConvertOptions()
	}
	return &PDFConverter{
		opts: opts,
	}
}

var defaultFirstLineFlag = []string{
	"一、", "二、", "三、", "四、", "五、", "六、", "七、", "八、", "九、", "十、",
	"十一、", "十二、", "十三、", "十四、", "十五、", "十六、", "十七、", "十八、", "十九、", "二十、",
	"二十一、", "二十二、", "二十三、", "二十四、", "二十五、", "二十六、",
	"一. ", "二. ", "三. ", "四. ", "五. ", "六. ", "七. ", "八. ", "九. ", "十. ",
	"十一. ", "十二. ", "十三. ", "十四. ", "十五. ", "十六. ", "十七. ", "十八. ", "十九. ", "二十. ",
	"二十一. ", "二十二. ", "二十三. ", "二十四. ", "二十五. ", "二十六. ",
}

var defaultSubFirstLineFlag = []string{
	"（一）", "（二）", "（三）", "（四）", "（五）", "（六）", "（七）", "（八）", "（九）", "（十）",
	"（十一）", "（十二）", "（十三）", "（十四）", "（十五）", "（十六）", "（十七）", "（十八）", "（十九）", "（二十）",
	"（二十一）", "（二十二）", "（二十三）", "（二十四）", "（二十五）", "（二十六）",
	"(一)", "(二)", "(三)", "(四)", "(五)", "(六)", "(七)", "(八)", "(九)", "(十)",
	"(十一)", "(十二)", "(十三)", "(十四)", "(十五)", "(十六)", "(十七)", "(十八)", "(十九)", "(二十)",
	"(二十一)", "(二十二)", "(二十三)", "(二十四)", "(二十五)", "(二十六)",
}

var defaultSecondLineFlag = []string{
	"1、", "2、", "3、", "4、", "5、", "6、", "7、", "8、", "9、", "10、",
	"11、", "12、", "13、", "14、", "15、", "16、", "17、", "18、", "19、", "20、",
	"21、", "22、", "23、", "24、", "25、", "26、", "27、", "28、", "29、", "30、",
	"1. ", "2. ", "3. ", "4. ", "5. ", "6. ", "7. ", "8. ", "9. ", "10. ",
	"11. ", "12. ", "13. ", "14. ", "15. ", "16. ", "17. ", "18. ", "19. ", "20. ",
	"21. ", "22. ", "23. ", "24. ", "25. ", "26. ", "27. ", "28. ", "29. ", "30. ",
}

var defaultThirdLineFlag = []string{
	"A、", "B、", "C、", "D、", "E、", "F、", "G、", "H、", "I、", "J、",
	"K、", "L、", "M、", "N、", "O、", "P、", "Q、", "R、", "S、", "T、",
	"U、", "V、", "W、", "X、", "Y、", "Z、",
	"A.", "B.", "C.", "D.", "E.", "F.", "G.", "H.", "I.", "J.",
	"K.", "L.", "M.", "N.", "O.", "P.", "Q.", "R.", "S.", "T.",
	"U.", "V.", "W.", "X.", "Y.", "Z.",
}

// ToMarkdown 将 PDF 转换为 Markdown
func (c *PDFConverter) ToMarkdown(input io.Reader) (string, error) {
	// 由于 PDF 库通常需要文件路径或 ReadSeeker，这里需要特殊处理
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 复制内容到临时文件
	if _, err := io.Copy(tmpFile, input); err != nil {
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	// 重置文件指针
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return "", fmt.Errorf("重置文件指针失败: %w", err)
	}

	return c.extractTextFromPDF(tmpFile.Name())
}

// ToMarkdownFile 将 PDF 文件转换为 Markdown 文件
func (c *PDFConverter) ToMarkdownFile(inputPath, outputPath string) error {
	markdown, err := c.extractTextFromPDF(inputPath)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(markdown), 0o644)
}

// SupportedTypes 返回支持的文档类型
func (c *PDFConverter) SupportedTypes() []DocumentType {
	return []DocumentType{TypePDF}
}

// extractTextFromPDF 从 PDF 文件提取文本并转换为 Markdown
func (c *PDFConverter) extractTextFromPDF(pdfPath string) (string, error) {
	// 优先使用 unipdf，功能更强大
	markdown, err := c.extractWithUnipdf(pdfPath)
	if err == nil {
		return markdown, nil
	}

	// 如果 unipdf 失败，使用 ledongthuc/pdf 作为备用
	return c.extractWithSimplePDF(pdfPath)
}

// extractWithUnipdf 使用 unipdf 提取文本
func (c *PDFConverter) extractWithUnipdf(pdfPath string) (string, error) {
	file, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("打开 PDF 文件失败: %w", err)
	}
	defer file.Close()

	pdfReader, err := model.NewPdfReader(file)
	if err != nil {
		return "", fmt.Errorf("读取 PDF 文件失败: %w", err)
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", fmt.Errorf("获取页数失败: %w", err)
	}

	var markdown strings.Builder
	markdown.WriteString(fmt.Sprintf("# PDF 文档\n\n总页数: %d\n\n", numPages))

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return "", fmt.Errorf("获取第 %d 页失败: %w", pageNum, err)
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", fmt.Errorf("创建提取器失败: %w", err)
		}

		text, err := ex.ExtractText()
		if err != nil {
			return "", fmt.Errorf("提取文本失败: %w", err)
		}
		if strings.TrimSpace(text) != "" {
			//markdown.WriteString(fmt.Sprintf("## 第 %d 页\n\n", pageNum))
			markdown.WriteString(c.formatTextToMarkdown(text))
			markdown.WriteString("\n\n---\n\n")
		}
	}

	return markdown.String(), nil
}

// extractWithSimplePDF 使用 ledongthuc/pdf 提取文本（备用方法）
func (c *PDFConverter) extractWithSimplePDF(pdfPath string) (string, error) {
	file, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("打开 PDF 文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	reader, err := pdf.NewReader(file, fileInfo.Size())
	if err != nil {
		return "", fmt.Errorf("创建 PDF 读取器失败: %w", err)
	}

	numPages := reader.NumPage()
	var markdown strings.Builder
	//markdown.WriteString(fmt.Sprintf("# PDF 文档\n\n总页数: %d\n\n", numPages))

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := reader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		all := strings.ReplaceAll(text, "\n", "")
		all = strings.ReplaceAll(all, "\r", "")
		all = strings.ReplaceAll(all, "\t", "")
		all = strings.ReplaceAll(all, "\\n", "")
		if strings.TrimSpace(all) != "" {
			//markdown.WriteString(fmt.Sprintf("## 第 %d 页\n\n", pageNum))
			markdown.WriteString(c.formatTextToMarkdownNew(all))
			//markdown.WriteString("\n\n---\n\n")
		}
	}

	return markdown.String(), nil
}

// formatTextToMarkdown 格式化文本为 Markdown
func (c *PDFConverter) formatTextToMarkdownNew(text string) string {
	for _, dlf := range defaultFirstLineFlag {
		if strings.Contains(text, dlf) {
			text = strings.ReplaceAll(text, dlf, "\n# "+dlf)
		}
	}
	for _, dlf := range defaultSubFirstLineFlag {
		if strings.Contains(text, dlf) {
			text = strings.ReplaceAll(text, dlf, "\n## "+dlf)
		}
	}
	for _, dlf := range defaultSecondLineFlag {
		if strings.Contains(text, dlf) {
			text = strings.ReplaceAll(text, dlf, "\n### "+dlf)
		}
	}
	for _, dlf := range defaultThirdLineFlag {
		if strings.Contains(text, dlf) {
			text = strings.ReplaceAll(text, dlf, "\n#### "+dlf)
		}
	}
	return text
}

// formatTextToMarkdown 格式化文本为 Markdown
func (c *PDFConverter) formatTextToMarkdown(text string) string {
	// 清理多余的空行
	lines := strings.Split(text, "\n")
	var result strings.Builder

	prevEmpty := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if !prevEmpty {
				result.WriteString("\n")
				prevEmpty = true
			}
			continue
		}

		prevEmpty = false

		// 检测可能的标题（全大写或以数字开头的短行）
		if c.opts.PreserveFormatting {
			if len(trimmed) < 100 {
				// 可能是标题
				if isAllUpper(trimmed) || isNumberedLine(trimmed) {
					result.WriteString("### ")
				}
			}
		}

		result.WriteString(trimmed)
		result.WriteString("\n")
	}

	return result.String()
}

// isAllUpper 检查字符串是否全部大写
func isAllUpper(s string) bool {
	hasLetter := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		}
	}
	return hasLetter
}

// isNumberedLine 检查是否是编号行（如 "1. "，"1.1 " 等）
func isNumberedLine(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}

	// 检查是否以数字开头
	firstChar := s[0]
	if firstChar < '0' || firstChar > '9' {
		return false
	}

	// 查找第一个空格或句点
	for i := 1; i < len(s) && i < 10; i++ {
		if s[i] == '.' || s[i] == ' ' {
			return true
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return false
}
