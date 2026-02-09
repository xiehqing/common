package document

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DocumentType 文档类型
type DocumentType string

const (
	TypePDF  DocumentType = "pdf"
	TypeWord DocumentType = "docx"
	TypeDOC  DocumentType = "doc"
)

// Converter 文档转换器接口
type Converter interface {
	// ToMarkdown 将文档转换为 Markdown 格式
	ToMarkdown(input io.Reader) (string, error)

	// ToMarkdownFile 将文档文件转换为 Markdown 格式并保存到文件
	ToMarkdownFile(inputPath, outputPath string) error

	// SupportedTypes 返回支持的文档类型
	SupportedTypes() []DocumentType
}

// ConvertOptions 转换选项
type ConvertOptions struct {
	// PreserveImages 是否保留图片
	PreserveImages bool

	// ImageOutputDir 图片输出目录
	ImageOutputDir string

	// ExtractTables 是否提取表格
	ExtractTables bool

	// PreserveFormatting 是否保留格式（加粗、斜体等）
	PreserveFormatting bool
}

// DefaultConvertOptions 默认转换选项
func DefaultConvertOptions() *ConvertOptions {
	return &ConvertOptions{
		PreserveImages:     true,
		ExtractTables:      true,
		PreserveFormatting: true,
	}
}

// DetectDocumentType 根据文件扩展名检测文档类型
func DetectDocumentType(filename string) (DocumentType, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return TypePDF, nil
	case ".docx":
		return TypeWord, nil
	case ".doc":
		return TypeDOC, nil
	default:
		return "", fmt.Errorf("不支持的文档类型: %s", ext)
	}
}

// DefaultConvert 转换文档到 Markdown
func DefaultConvert(inputPath, outputPath string) error {
	opts := DefaultConvertOptions()
	opts.PreserveFormatting = true
	opts.ExtractTables = true
	return Convert(inputPath, outputPath, opts)
}

// Convert 转换文档到 Markdown
func Convert(inputPath, outputPath string, opts *ConvertOptions) error {
	if opts == nil {
		opts = DefaultConvertOptions()
	}

	docType, err := DetectDocumentType(inputPath)
	if err != nil {
		return err
	}

	var converter Converter
	switch docType {
	case TypePDF:
		converter = NewPDFConverter(opts)
	case TypeWord, TypeDOC:
		converter = NewWordConverter(opts)
	default:
		return fmt.Errorf("不支持的文档类型: %s", docType)
	}

	return converter.ToMarkdownFile(inputPath, outputPath)
}

// ConvertToString 转换文档到 Markdown 字符串
func ConvertToString(inputPath string, opts *ConvertOptions) (string, error) {
	if opts == nil {
		opts = DefaultConvertOptions()
	}

	docType, err := DetectDocumentType(inputPath)
	if err != nil {
		return "", err
	}

	var converter Converter
	switch docType {
	case TypePDF:
		converter = NewPDFConverter(opts)
	case TypeWord, TypeDOC:
		converter = NewWordConverter(opts)
	default:
		return "", fmt.Errorf("不支持的文档类型: %s", docType)
	}

	file, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	return converter.ToMarkdown(file)
}
