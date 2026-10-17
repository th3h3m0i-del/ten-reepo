// Package detect đoán ngôn ngữ lập trình của một đoạn text dựa trên
// shebang (nếu có) và một vài dấu hiệu đặc trưng. Mục tiêu là đủ dùng
// cho các ngôn ngữ phổ biến có mặt trên Termux, không cần chính xác
// tuyệt đối. Nếu không đoán được, mặc định là bash.
package detect

import (
	"strings"
)

// Guess nhận text và trả về tên ngôn ngữ cùng đuôi file gợi ý.
// Thứ tự ưu tiên: shebang trước, rồi token đặc trưng, cuối cùng là bash.
func Guess(text string) (lang string, ext string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "bash", "sh"
	}

	// 1. Shebang.
	firstLine := strings.SplitN(trimmed, "\n", 2)[0]
	if strings.HasPrefix(firstLine, "#!") {
		lower := strings.ToLower(firstLine)
		switch {
		case strings.Contains(lower, "python"):
			return "python", "py"
		case strings.Contains(lower, "node"):
			return "javascript", "js"
		case strings.Contains(lower, "ruby"):
			return "ruby", "rb"
		case strings.Contains(lower, "perl"):
			return "perl", "pl"
		case strings.Contains(lower, "php"):
			return "php", "php"
		case strings.Contains(lower, "lua"):
			return "lua", "lua"
		case strings.Contains(lower, "deno"):
			return "deno", "ts"
		case strings.Contains(lower, "bun"):
			return "bun", "js"
		case strings.Contains(lower, "bash"):
			return "bash", "sh"
		case strings.Contains(lower, "zsh"):
			return "zsh", "sh"
		case strings.Contains(lower, "sh"):
			return "sh", "sh"
		}
	}

	// 2. Token đặc trưng.
	switch {
	case strings.Contains(trimmed, "package main") && strings.Contains(trimmed, "func main"):
		return "go", "go"
	case strings.Contains(trimmed, "def ") && (strings.Contains(trimmed, "import ") || strings.Contains(trimmed, "print(")):
		return "python", "py"
	case strings.Contains(trimmed, "<?php"):
		return "php", "php"
	case (strings.Contains(trimmed, "function") || strings.Contains(trimmed, "=>")) && strings.Contains(trimmed, "console.log"):
		return "javascript", "js"
	case strings.Contains(trimmed, "puts") && strings.Contains(trimmed, "end"):
		return "ruby", "rb"
	}

	// 3. Mặc định.
	return "bash", "sh"
}
