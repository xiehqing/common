package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BatchConvert 批量转换文档
func BatchConvert(inputDir, outputDir string, opts *ConvertOptions) error {
	if opts == nil {
		opts = DefaultConvertOptions()
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 遍历输入目录
	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查是否是支持的文档类型
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".pdf" && ext != ".docx" && ext != ".doc" {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(inputDir, path)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %w", err)
		}

		// 生成输出文件路径
		outputPath := filepath.Join(outputDir, strings.TrimSuffix(relPath, ext)+".md")

		// 确保输出文件的目录存在
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("创建输出子目录失败: %w", err)
		}

		// 转换文档
		fmt.Printf("正在转换: %s -> %s\n", path, outputPath)
		if err := Convert(path, outputPath, opts); err != nil {
			fmt.Printf("转换失败: %s, 错误: %v\n", path, err)
			return nil // 继续处理其他文件
		}

		return nil
	})
}

// ValidateFile 验证文件是否可以转换
func ValidateFile(filePath string) error {
	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("文件不存在: %s", filePath)
		}
		return fmt.Errorf("读取文件信息失败: %w", err)
	}

	// 检查是否是文件
	if info.IsDir() {
		return fmt.Errorf("路径是目录，不是文件: %s", filePath)
	}

	// 检查文件大小
	if info.Size() == 0 {
		return fmt.Errorf("文件为空: %s", filePath)
	}

	// 检查文件类型
	_, err = DetectDocumentType(filePath)
	if err != nil {
		return err
	}

	return nil
}

// GetDocumentInfo 获取文档信息
func GetDocumentInfo(filePath string) (map[string]interface{}, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件信息失败: %w", err)
	}

	docType, err := DetectDocumentType(filePath)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"path":     filePath,
		"name":     info.Name(),
		"size":     info.Size(),
		"type":     string(docType),
		"modified": info.ModTime(),
	}

	return result, nil
}

// CleanMarkdown 清理 Markdown 文本
func CleanMarkdown(markdown string) string {
	lines := strings.Split(markdown, "\n")
	var cleaned strings.Builder

	prevEmpty := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 移除连续的空行
		if trimmed == "" {
			if !prevEmpty {
				cleaned.WriteString("\n")
				prevEmpty = true
			}
			continue
		}

		prevEmpty = false
		cleaned.WriteString(line)
		cleaned.WriteString("\n")
	}

	return cleaned.String()
}

// SplitMarkdownByHeadings 按标题分割 Markdown
func SplitMarkdownByHeadings(markdown string) []map[string]string {
	lines := strings.Split(markdown, "\n")
	var sections []map[string]string
	var currentSection map[string]string
	var contentBuilder strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 检测标题
		if strings.HasPrefix(trimmed, "#") {
			// 保存之前的部分
			if currentSection != nil {
				currentSection["content"] = strings.TrimSpace(contentBuilder.String())
				sections = append(sections, currentSection)
				contentBuilder.Reset()
			}

			// 解析标题
			level := 0
			for i := 0; i < len(trimmed) && trimmed[i] == '#'; i++ {
				level++
			}

			title := strings.TrimSpace(trimmed[level:])
			currentSection = map[string]string{
				"level": fmt.Sprintf("%d", level),
				"title": title,
			}
		} else {
			contentBuilder.WriteString(line)
			contentBuilder.WriteString("\n")
		}
	}

	// 保存最后一个部分
	if currentSection != nil {
		currentSection["content"] = strings.TrimSpace(contentBuilder.String())
		sections = append(sections, currentSection)
	}

	return sections
}

// EstimateConversionTime 估算转换时间（毫秒）
func EstimateConversionTime(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	docType, err := DetectDocumentType(filePath)
	if err != nil {
		return 0, err
	}

	// 基于文件大小和类型估算
	sizeMB := float64(info.Size()) / (1024 * 1024)
	var timePerMB int64

	switch docType {
	case TypePDF:
		timePerMB = 2000 // PDF 处理较慢，约 2 秒/MB
	case TypeWord:
		timePerMB = 1000 // Word 处理较快，约 1 秒/MB
	default:
		timePerMB = 1500
	}

	estimatedTime := int64(sizeMB * float64(timePerMB))
	if estimatedTime < 100 {
		estimatedTime = 100 // 最少 100ms
	}

	return estimatedTime, nil
}
