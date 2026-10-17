// Package gitguard scan nội dung trước khi git sync, chặn nếu phát hiện
// secret (API key, token, private key). Mục đích: ś tự push lên GitHub
// public, không được để lọt credential. Nếu phát hiện, sync bị abort và
// report ghi sync: blocked: secret detected.
package gitguard

import (
	"fmt"
	"strings"
)

// Finding mô tả một secret bị phát hiện.
type Finding struct {
	Pattern string // pattern khớp (không in full secret)
	Line    int    // dòng trong nội dung (1-based)
}

// Scan kiểm tra content với danh sách patterns. Trả về danh sách Finding.
// Không trả full secret trong Finding để tránh vô tình log ra ngoài.
func Scan(content string, patterns []string) []Finding {
	var findings []Finding
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		for _, p := range patterns {
			if p == "" {
				continue
			}
			if strings.Contains(line, p) {
				findings = append(findings, Finding{Pattern: p, Line: i + 1})
			}
		}
	}
	return findings
}

// Blocked trả về true nếu có ít nhất một Finding.
func Blocked(findings []Finding) bool { return len(findings) > 0 }

// Summary trả về chuỗi ngắn gọn mô tả findings để đưa vào report.
// Không in full secret, chỉ in pattern + số dòng.
func Summary(findings []Finding) string {
	if len(findings) == 0 {
		return ""
	}
	var parts []string
	seen := map[string]bool{}
	for _, f := range findings {
		key := fmt.Sprintf("%s@L%d", f.Pattern, f.Line)
		if seen[key] {
			continue
		}
		seen[key] = true
		parts = append(parts, fmt.Sprintf("pattern=%q line=%d", f.Pattern, f.Line))
	}
	return strings.Join(parts, "; ")
}
