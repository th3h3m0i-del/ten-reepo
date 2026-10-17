// Package report dựng chuỗi markdown mô tả kết quả một lần chạy của ś,
// theo format tương thích harness log (RUN / code run / stdout / stderr / changes).
package report

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"ten-reepo/S/internal/runner"
)

const (
	maxCodeDump = 4096
	maxOutput   = 10240
)

// Build trả về markdown hoàn chỉnh cho một Result.
func Build(r runner.Result) string {
	truncated := len(r.Stdout) > maxOutput || len(r.Stderr) > maxOutput

	var b strings.Builder
	b.WriteString("## RUN\n\n")

	meta := map[string]interface{}{
		"status":             r.Status,
		"status_reason":      r.StatusReason,
		"exit":               r.ExitCode,
		"duration":           roundTo(r.Duration, 3),
		"lang":               r.Lang,
		"script_path":        r.ScriptPath,
		"stdout_total_lines": r.StdoutTotalLines,
		"stderr_total_lines": r.StderrTotalLines,
		"error_hits":         r.ErrorHits,
		"output_truncated":   truncated,
	}
	if r.Sync != "" {
		meta["sync"] = r.Sync
	}
	j, _ := json.MarshalIndent(meta, "", "  ")
	b.WriteString("```json\n")
	b.Write(j)
	b.WriteString("\n```\n\n")

	b.WriteString("### code run\n\n")
	if len(r.CodeContent) <= maxCodeDump {
		fence := "```"
		if strings.Contains(r.CodeContent, "```") {
			fence = "````"
		}
		b.WriteString(fence + langTag(r.Lang) + "\n")
		b.WriteString(r.CodeContent)
		if !strings.HasSuffix(r.CodeContent, "\n") {
			b.WriteString("\n")
		}
		b.WriteString(fence + "\n\n")
	} else {
		fmt.Fprintf(&b, "(file dài %d bytes, vượt ngưỡng %d, xem tại `%s`)\n\n",
			len(r.CodeContent), maxCodeDump, r.ScriptPath)
	}

	b.WriteString("### stdout\n\n```\n")
	b.WriteString(truncate(r.Stdout, maxOutput))
	b.WriteString("```\n\n")

	b.WriteString("### stderr\n\n```\n")
	b.WriteString(truncate(r.Stderr, maxOutput))
	b.WriteString("```\n\n")

	b.WriteString("### changes\n\n```\n")
	for _, c := range r.Changes {
		b.WriteString(c + "\n")
	}
	b.WriteString("```\n")

	return b.String()
}

// truncate cắt chuỗi nếu dài hơn max, thêm "... (truncated)".
func truncate(s string, max int) string {
	if s == "" {
		return "(không có)\n"
	}
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	if len(s) <= max {
		return s
	}
	cut := s[:max]
	if lastNL := strings.LastIndex(cut, "\n"); lastNL > 0 {
		cut = cut[:lastNL+1]
	}
	return cut + "... (truncated)\n"
}

// langTag trả về nhãn ngôn ngữ cho code fence trong markdown.
func langTag(lang string) string {
	switch lang {
	case "python":
		return "python"
	case "javascript":
		return "javascript"
	case "typescript", "deno":
		return "typescript"
	case "go":
		return "go"
	case "ruby":
		return "ruby"
	case "perl":
		return "perl"
	case "php":
		return "php"
	case "lua", "luajit":
		return "lua"
	case "bash", "zsh", "sh", "dash", "fish":
		return "bash"
	case "java":
		return "java"
	case "c":
		return "c"
	case "cpp":
		return "cpp"
	case "rust":
		return "rust"
	case "kotlin":
		return "kotlin"
	case "swift":
		return "swift"
	case "csharp":
		return "csharp"
	case "sql":
		return "sql"
	case "r":
		return "r"
	case "julia":
		return "julia"
	case "tcl":
		return "tcl"
	case "groovy":
		return "groovy"
	case "scala":
		return "scala"
	case "erlang":
		return "erlang"
	case "crystal":
		return "crystal"
	case "nim":
		return "nim"
	case "zig":
		return "zig"
	case "awk":
		return "awk"
	case "bun":
		return "javascript"
	default:
		return ""
	}
}

// roundTo làm tròn f đến digits chữ số thập phân.
func roundTo(f float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(f*pow) / pow
}
