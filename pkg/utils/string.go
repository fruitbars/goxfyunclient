package utils

import (
	"strings"
	"unicode/utf8"
)

func SafeSnippet(body []byte, maxLen int) string {
	if len(body) == 0 {
		return ""
	}

	if len(body) <= maxLen {
		// 数据不超过 maxLen，直接返回
		if utf8.Valid(body) {
			return string(body)
		}
		// 如果有非法 UTF-8，逐字节修正
		return string([]rune(string(body)))
	}

	// 截取前 maxLen
	snippet := body[:maxLen]

	// 确保结尾是合法的 UTF-8
	for !utf8.Valid(snippet) && len(snippet) > 0 {
		snippet = snippet[:len(snippet)-1]
	}

	return string(snippet) + "...(truncated)"
}

// 一个按“字符”截断的新版本
func SafeSnippetByRunes(body []byte, maxRunes int) string {
	if !utf8.Valid(body) {
		// 对于非法utf8，先修正
		body = []byte(strings.ToValidUTF8(string(body), ""))
	}

	runes := []rune(string(body))
	if len(runes) <= maxRunes {
		return string(runes)
	}

	return string(runes[:maxRunes]) + "...(truncated)"
}
