package tools

import (
	"github.com/hatcher/common/pkg/logs"
	"testing"
	"unicode/utf8"
)

func TestView(t *testing.T) {
	textFile, linecount, err := readTextFile("F:\\project\\jaime\\go_project\\ai-apps\\bidding\\8b6d43df-30f2-4be0-872e-bb4ec557b741\\国泰海通基于大模型的智能运维平台全链路诊断项目采购需求.md", 100, 100)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(linecount)
	isValidUt8 := utf8.ValidString(textFile)
	logs.Infof("内容：is valid utf8: %v", isValidUt8)
	if !isValidUt8 {
		t.Log("File content is not valid UTF-8")
	}
	t.Log(textFile)
}
