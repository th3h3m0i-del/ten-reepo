// Package detect đoán ngôn ngữ lập trình của một đoạn text dựa trên
// shebang (nếu có) và một vài dấu hiệu đặc trưng. Nếu không đoán được,
// mặc định là bash. Đây là heuristic, không phải parser.
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
			return "typescript", "ts"
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
	// Go: package + func main.
	case strings.Contains(trimmed, "package main") && strings.Contains(trimmed, "func main"):
		return "go", "go"

	// TypeScript: import type / interface / type annotation phổ biến.
	case strings.Contains(trimmed, "interface ") && strings.Contains(trimmed, ": ") && strings.Contains(trimmed, "export ") && !strings.Contains(trimmed, "package main"):
		return "typescript", "ts"

	// C++: #include <iostream> hoặc using namespace std.
	case strings.Contains(trimmed, "#include <iostream>") || strings.Contains(trimmed, "using namespace std"):
		return "cpp", "cpp"

	// C: #include <stdio.h> hoặc int main(void) không có class/iostream.
	case strings.Contains(trimmed, "#include <stdio.h>") || strings.Contains(trimmed, "#include <stdlib.h>"):
		return "c", "c"

	// Rust: fn main + let mut / use std.
	case strings.Contains(trimmed, "fn main") && (strings.Contains(trimmed, "let mut") || strings.Contains(trimmed, "use std")):
		return "rust", "rs"

	// Kotlin: fun main + val/var.
	case strings.Contains(trimmed, "fun main") && (strings.Contains(trimmed, "val ") || strings.Contains(trimmed, "var ")):
		return "kotlin", "kt"

	// Swift: import Foundation / import SwiftUI + func/let/var.
	case (strings.Contains(trimmed, "import Foundation") || strings.Contains(trimmed, "import SwiftUI")) && (strings.Contains(trimmed, "func ") || strings.Contains(trimmed, "let ")):
		return "swift", "swift"

	// SQL: SELECT ... FROM, CREATE TABLE, INSERT INTO.
	case strings.Contains(strings.ToUpper(trimmed), "SELECT ") && strings.Contains(strings.ToUpper(trimmed), " FROM ") && !strings.Contains(trimmed, "func "):
		return "sql", "sql"
	case strings.Contains(strings.ToUpper(trimmed), "CREATE TABLE ") || strings.Contains(strings.ToUpper(trimmed), "INSERT INTO "):
		return "sql", "sql"

	// Python fallback (không shebang).
	case strings.Contains(trimmed, "def ") && (strings.Contains(trimmed, "import ") || strings.Contains(trimmed, "print(")):
		return "python", "py"

	// PHP.
	case strings.Contains(trimmed, "<?php"):
		return "php", "php"

	// JavaScript.
	case (strings.Contains(trimmed, "function") || strings.Contains(trimmed, "=>")) && strings.Contains(trimmed, "console.log"):
		return "javascript", "js"

	// Ruby.
	case strings.Contains(trimmed, "puts") && strings.Contains(trimmed, "end"):
		return "ruby", "rb"
	}

	return "bash", "sh"
}
