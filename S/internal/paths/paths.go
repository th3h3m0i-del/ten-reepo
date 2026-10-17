// Package paths chốt các đường dẫn tuyệt đối trong repo ~/ten-reepo:
// S/ (main), tmp/S (scratch), backup/S (snapshot). Mọi hàm đều đảm
// bảo thư mục đích tồn tại trước khi trả về.
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

// RepoRoot trả về ~/ten-reepo.
func RepoRoot() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, "ten-reepo"), nil
}

// TmpDir trả về ~/ten-reepo/tmp/S, tạo nếu chưa có.
func TmpDir() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "tmp", "S")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được tmp/S: %w", err)
	}
	return dir, nil
}

// ScriptPath tạo đường dẫn tuyệt đối cho file script mới trong
// ~/ten-reepo/S/scripts/clip_<timestamp>.<ext>.
func ScriptPath(ext string) (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "S", "scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được scripts: %w", err)
	}
	name := fmt.Sprintf("clip_%d.%s", time.Now().UnixNano(), ext)
	return filepath.Join(dir, name), nil
}

// ClipboardBackupPath tạo đường dẫn file backup clipboard gốc trong
// ~/ten-reepo/tmp/S/clipboard_backups/clipboard_<timestamp>.txt.
func ClipboardBackupPath() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "tmp", "S", "clipboard_backups")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được clipboard_backups: %w", err)
	}
	name := fmt.Sprintf("clipboard_%d.txt", time.Now().UnixNano())
	return filepath.Join(dir, name), nil
}
