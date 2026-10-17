// Package detect đoán ngôn ngữ lập trình của một đoạn text dựa trên
// shebang (nếu có) và token đặc trưng. Đây là heuristic, không phải
// parser. Mục tiêu: phủ hết các ngôn ngữ có interpreter/compiler trên
// Termux của Bosz (recon 2026-10-17), fallback bash khi không đoán được.
package detect

import (
	"strings"
)

// Guess trả về tên ngôn ngữ và đuôi file gợi ý cho text.
// Thứ tự: shebang (chính xác nhất) → token đặc trưng → bash fallback.
func Guess(text string) (lang string, ext string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "bash", "sh"
	}

	// 1. Shebang: tin tuyệt đối vì người viết khai báo rõ interpreter.
	firstLine := strings.SplitN(trimmed, "\n", 2)[0]
	if strings.HasPrefix(firstLine, "#!") {
		if lang, ext, ok := fromShebang(strings.ToLower(firstLine)); ok {
			return lang, ext
		}
	}

	// 2. Token đặc trưng. Thứ tự quan trọng: cái đặc hiệu hơn đứng trước
	// để tránh nhầm (vd scala trước kotlin vì cùng có val/def).
	if lang, ext, ok := fromTokens(trimmed); ok {
		return lang, ext
	}

	// 3. Fallback.
	return "bash", "sh"
}

// fromShebang nhận dòng shebang đã lowercase, trả về lang/ext nếu khớp.
// Thứ tự check: các shell chứa "sh" (bash/zsh/dash/fish/tclsh) phải
// đứng trước "sh" trơn để không bị nhầm.
func fromShebang(line string) (string, string, bool) {
	switch {
	case strings.Contains(line, "python"):
		return "python", "py", true
	case strings.Contains(line, "node"):
		return "javascript", "js", true
	case strings.Contains(line, "deno"):
		return "deno", "ts", true
	case strings.Contains(line, "bun"):
		return "bun", "js", true
	case strings.Contains(line, "ruby"):
		return "ruby", "rb", true
	case strings.Contains(line, "perl"):
		return "perl", "pl", true
	case strings.Contains(line, "php"):
		return "php", "php", true
	case strings.Contains(line, "luajit"):
		return "luajit", "lua", true
	case strings.Contains(line, "lua"):
		return "lua", "lua", true
	case strings.Contains(line, "rscript"):
		return "r", "R", true
	case strings.Contains(line, "julia"):
		return "julia", "jl", true
	case strings.Contains(line, "tclsh") || strings.Contains(line, "wish"):
		return "tcl", "tcl", true
	case strings.Contains(line, "groovy"):
		return "groovy", "groovy", true
	case strings.Contains(line, "scala"):
		return "scala", "scala", true
	case strings.Contains(line, "escript"):
		return "erlang", "erl", true
	case strings.Contains(line, "gawk") || strings.Contains(line, "awk"):
		return "awk", "awk", true
	case strings.Contains(line, "fish"):
		return "fish", "fish", true
	case strings.Contains(line, "crystal"):
		return "crystal", "cr", true
	case strings.Contains(line, "nim"):
		return "nim", "nim", true
	case strings.Contains(line, "zig"):
		return "zig", "zig", true
	case strings.Contains(line, "kotlin"):
		return "kotlin", "kt", true
	case strings.Contains(line, "swift"):
		return "swift", "swift", true
	case strings.Contains(line, "dotnet"):
		return "csharp", "cs", true
	case strings.Contains(line, "bash"):
		return "bash", "sh", true
	case strings.Contains(line, "zsh"):
		return "zsh", "sh", true
	case strings.Contains(line, "dash"):
		return "dash", "sh", true
	case strings.Contains(line, "sh"):
		return "sh", "sh", true
	}
	return "", "", false
}

// fromTokens nhận text gốc, dò token đặc trưng. Trả về lang/ext nếu khớp.
// Dùng hasWord cho các keyword ngắn dễ nhầm (end/def/fun/set/puts/echo/
// proc) thay vì strings.Contains để tránh false positive ("end" trong
// "append", "puts" trong "outputs"...).
func fromTokens(t string) (string, string, bool) {
	upper := strings.ToUpper(t)

	switch {
	// Go: package + func main.
	case strings.Contains(t, "package main") && strings.Contains(t, "func main"):
		return "go", "go", true

	// Java: signature "public static void main" là đủ độc nhất, không
	// cần "public class" vì class tên Hello vẫn có thể không public.
	case strings.Contains(t, "public static void main"):
		return "java", "java", true

	// Scala: object + def main (check trước kotlin vì cùng có val/def).
	case strings.Contains(t, "object ") && hasWord(t, "def") && strings.Contains(t, "def main"):
		return "scala", "scala", true

	// Kotlin: "fun main" là signature độc nhất (scala dùng "def main").
	// Không cần val/var vì snippet ngắn có thể chỉ có println.
	case strings.Contains(t, "fun main"):
		return "kotlin", "kt", true

	// Swift: import Foundation/SwiftUI + func.
	case (strings.Contains(t, "import Foundation") || strings.Contains(t, "import SwiftUI")) && strings.Contains(t, "func "):
		return "swift", "swift", true

	// C#: using System + namespace/class Program.
	case strings.Contains(t, "using System") && (strings.Contains(t, "namespace ") || strings.Contains(t, "class Program")):
		return "csharp", "cs", true

	// C++: iostream hoặc using namespace std.
	case strings.Contains(t, "#include <iostream>") || strings.Contains(t, "using namespace std"):
		return "cpp", "cpp", true

	// C: stdio.h hoặc stdlib.h.
	case strings.Contains(t, "#include <stdio.h>") || strings.Contains(t, "#include <stdlib.h>"):
		return "c", "c", true

	// Rust: fn main + let mut/use std.
	case strings.Contains(t, "fn main") && (strings.Contains(t, "let mut") || strings.Contains(t, "use std")):
		return "rust", "rs", true

	// Zig: @import + pub fn main.
	case strings.Contains(t, "@import") && strings.Contains(t, "pub fn main"):
		return "zig", "zig", true

	// TypeScript: interface + export + type annotation.
	case strings.Contains(t, "interface ") && strings.Contains(t, "export ") && strings.Contains(t, ": "):
		return "typescript", "ts", true

	// Julia: function + end + println (không có def/fun).
	case strings.Contains(t, "function ") && hasWord(t, "end") && strings.Contains(t, "println(") && !hasWord(t, "def") && !strings.Contains(t, "fun "):
		return "julia", "jl", true

	// R: gán <- bắt buộc, kèm function( hoặc cat( — cả hai đều đặc
	// trưng của R, ít lẫn với ngôn ngữ khác.
	case strings.Contains(t, "<-") && (strings.Contains(t, "function(") || strings.Contains(t, "cat(")):
		return "r", "R", true

	// Tcl: proc + set + puts.
	case hasWord(t, "proc") && hasWord(t, "set") && hasWord(t, "puts"):
		return "tcl", "tcl", true

	// Groovy: def + println + class/@Grab.
	case hasWord(t, "def") && strings.Contains(t, "println") && (strings.Contains(t, "class ") || strings.Contains(t, "@Grab")):
		return "groovy", "groovy", true

	// Nim: proc + echo.
	case hasWord(t, "proc") && hasWord(t, "echo"):
		return "nim", "nim", true

	// SQL: SELECT ... FROM, CREATE TABLE, INSERT INTO.
	case strings.Contains(upper, "SELECT ") && strings.Contains(upper, " FROM "):
		return "sql", "sql", true
	case strings.Contains(upper, "CREATE TABLE ") || strings.Contains(upper, "INSERT INTO "):
		return "sql", "sql", true

	// Python fallback: def + import/print.
	case hasWord(t, "def") && (strings.Contains(t, "import ") || strings.Contains(t, "print(")):
		return "python", "py", true

	// PHP.
	case strings.Contains(t, "<?php"):
		return "php", "php", true

	// JavaScript: function/=> + console.log.
	case (strings.Contains(t, "function") || strings.Contains(t, "=>")) && strings.Contains(t, "console.log"):
		return "javascript", "js", true

	// Ruby: puts + end (không có def main kiểu scala).
	case hasWord(t, "puts") && hasWord(t, "end") && !strings.Contains(t, "def main"):
		return "ruby", "rb", true

	// AWK: BEGIN { hoặc END {.
	case strings.Contains(t, "BEGIN {") || strings.Contains(t, "END {"):
		return "awk", "awk", true
	}

	return "", "", false
}

// hasWord kiểm tra t có chứa word như một token tách biệt (ranh giới
// là ký tự không phải chữ/số/gạch dưới, hoặc đầu/cuối chuỗi).
func hasWord(t, word string) bool {
	idx := 0
	for {
		i := strings.Index(t[idx:], word)
		if i < 0 {
			return false
		}
		start := idx + i
		endPos := start + len(word)
		beforeOK := start == 0 || !isIdentChar(t[start-1])
		afterOK := endPos == len(t) || !isIdentChar(t[endPos])
		if beforeOK && afterOK {
			return true
		}
		idx = start + 1
	}
}

// isIdentChar trả về true nếu b là ký tự định danh (chữ, số, gạch dưới).
func isIdentChar(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
