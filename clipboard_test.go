package main

import "testing"

func TestFormatClipboardText(t *testing.T) {
	tests := []struct {
		name, title, content, want string
	}{
		{"title and body", "测试标题", "第一行\n第二行", "测试标题\r\n\r\n第一行\n第二行"},
		{"title only", "  标题  ", "", "标题"},
		{"body only", "", "正文", "正文"},
		{"empty", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatClipboardText(tt.title, tt.content); got != tt.want {
				t.Fatalf("formatClipboardText() = %q, want %q", got, tt.want)
			}
		})
	}
}
