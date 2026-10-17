// Package paths chốt các đường dẫn tuyệt đối nằm trong scope cho phép:
// ~/ten-reepo/S (main) và ~/ten-reepo/tmp/S (scratch).
package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// home trả về thư mục home của user hiện tại.
func home() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("không lấy được home: %w", err)
	}
	return h, nil
}

// ScriptPath tạo đường dẫn tuyệt đối cho file script mới trong
// ~/ten-reepo/S/scripts/clip_<timestamp>.<ext>. Thư mục scripts
// được tạo nếu chưa có.
func ScriptPath(ext string) (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "S", "scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được thư mục scripts: %w", err)
	}
	name := fmt.Sprintf("clip_%d.%s", time.Now().UnixNano(), ext)
	return filepath.Join(dir, name), nil
}

// ClipboardBackupPath tạo đường dẫn tuyệt đối cho file backup clipboard
// trong ~/ten-reepo/tmp/S/clipboard_backups/clipboard_<timestamp>.txt.
// Mục đích: lưu clipboard gốc trước khi ś ghi đè bằng report.
func ClipboardBackupPath() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "tmp", "S", "clipboard_backups")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được thư mục clipboard_backups: %w", err)
	}
	name := fmt.Sprintf("clipboard_%d.txt", time.Now().UnixNano())
	return filepath.Join(dir, name), nil
}
