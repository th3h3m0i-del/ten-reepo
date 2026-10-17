// Package report dựng chuỗi markdown mô tả kết quả một lần chạy của ś,
// theo format tương thích với harness log (RUN / code run / stdout /
// stderr / changes).
package report

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"ten-reepo/S/internal/runner"
)

const (
	// maxCodeDump: nếu nội dung script ngắn hơn hoặc bằng ngưỡng này thì
	// dump nguyên vào mục "code run"; dài hơn chỉ ghi path + kích thước.
	maxCodeDump = 4096
	// maxOutput: nếu stdout hoặc stderr dài hơn ngưỡng này thì cắt và
	// set output_truncated = true.
	maxOutput = 10240
)

// Build trả về chuỗi markdown hoàn chỉnh cho một Result.
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

// truncate cắt chuỗi nếu dài hơn max, thêm ghi chú "... (truncated)".
// Nếu chuỗi rỗng thì trả về "(không có)".
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
	case "javascript", "bun":
		return "javascript"
	case "go":
		return "go"
	case "ruby":
		return "ruby"
	case "perl":
		return "perl"
	case "php":
		return "php"
	case "lua":
		return "lua"
	case "bash", "zsh", "sh":
		return "bash"
	case "deno":
		return "typescript"
	case "java":
		return "java"
	default:
		return ""
	}
}

// roundTo làm tròn f đến digits chữ số thập phân.
func roundTo(f float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(f*pow) / pow
}
