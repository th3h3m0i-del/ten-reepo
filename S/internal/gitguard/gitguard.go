// Package gitguard scan nội dung trước khi git sync, chặn nếu phát hiện
// secret (API key, token, private key). Hai lớp:
//   - Scan: quét một chuỗi (clipboard content).
//   - ScanDir: quét toàn bộ file trong một thư mục (S/scripts/).
// Nếu bất kỳ lớp nào phát hiện, sync bị abort. File chứa secret được
// lưu vào tmp/S/blocked_scripts/ thay vì S/scripts/ để không bao giờ
// bị git add -A mang theo.
package gitguard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Finding mô tả một secret bị phát hiện.
type Finding struct {
	File    string // tên file (rỗng nếu scan chuỗi trực tiếp)
	Pattern string // pattern khớp (không in full secret)
	Line    int    // dòng trong nội dung (1-based)
}

// Scan kiểm tra content với danh sách patterns. File = "" (không phải
// từ file cụ thể). Không trả full secret trong Finding để tránh vô
// tình log ra ngoài.
func Scan(content string, patterns []string) []Finding {
	return scanInternal(content, "", patterns)
}

// ScanDir đọc toàn bộ file trong dir và scan từng file. Skip file
// binary (NUL byte đầu tiên) và thư mục con. Trả về mọi Finding kèm
// tên file. Nếu không đọc được dir, trả error; file lỗi riêng lẻ bị
// skip (không nên block sync chỉ vì 1 file unreadable).
func ScanDir(dir string, patterns []string) ([]Finding, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var all []Finding
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		// Skip binary: có NUL byte trong 512 byte đầu.
		head := data
		if len(head) > 512 {
			head = head[:512]
		}
		if strings.ContainsRune(string(head), 0) {
			continue
		}
		all = append(all, scanInternal(string(data), e.Name(), patterns)...)
	}
	return all, nil
}

// scanInternal là phần lõi dùng chung cho Scan và ScanDir.
func scanInternal(content, file string, patterns []string) []Finding {
	var findings []Finding
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		for _, p := range patterns {
			if p == "" {
				continue
			}
			if strings.Contains(line, p) {
				findings = append(findings, Finding{
					File:    file,
					Pattern: p,
					Line:    i + 1,
				})
			}
		}
	}
	return findings
}

// Blocked trả về true nếu có ít nhất một Finding.
func Blocked(findings []Finding) bool { return len(findings) > 0 }

// Summary trả về chuỗi ngắn gọn mô tả findings. Có file thì prefix
// "file=<file>"; không in full secret, chỉ in pattern + số dòng. Tối
// đa 5 finding để report không bị phình.
func Summary(findings []Finding) string {
	if len(findings) == 0 {
		return ""
	}
	var parts []string
	seen := map[string]bool{}
	for _, f := range findings {
		key := fmt.Sprintf("%s|%s|%d", f.File, f.Pattern, f.Line)
		if seen[key] {
			continue
		}
		seen[key] = true
		if f.File != "" {
			parts = append(parts, fmt.Sprintf("file=%q pattern=%q line=%d", f.File, f.Pattern, f.Line))
		} else {
			parts = append(parts, fmt.Sprintf("pattern=%q line=%d", f.Pattern, f.Line))
		}
		if len(parts) >= 5 {
			parts = append(parts, "...")
			break
		}
	}
	return strings.Join(parts, "; ")
}
