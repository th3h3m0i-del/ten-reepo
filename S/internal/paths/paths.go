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

// PushQueueDir trả về ~/ten-reepo/tmp/S/push_queue, tạo nếu chưa có.
func PushQueueDir() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "tmp", "S", "push_queue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được push_queue: %w", err)
	}
	return dir, nil
}

// ScriptsDir trả về ~/ten-reepo/S/scripts, tạo nếu chưa có.
func ScriptsDir() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "S", "scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được scripts: %w", err)
	}
	return dir, nil
}

// BlockedScriptsDir trả về ~/ten-reepo/tmp/S/blocked_scripts, tạo nếu
// chưa có. Nơi lưu script có secret — không bao giờ được push.
func BlockedScriptsDir() (string, error) {
	h, err := home()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(h, "ten-reepo", "tmp", "S", "blocked_scripts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("không tạo được blocked_scripts: %w", err)
	}
	return dir, nil
}

// ScriptPath tạo đường dẫn tuyệt đối cho file script mới trong
// ~/ten-reepo/S/scripts/clip_<timestamp>.<ext>.
func ScriptPath(ext string) (string, error) {
	dir, err := ScriptsDir()
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("clip_%d.%s", time.Now().UnixNano(), ext)
	return filepath.Join(dir, name), nil
}

// BlockedScriptPath tạo đường dẫn tuyệt đối cho file script bị block
// trong ~/ten-reepo/tmp/S/blocked_scripts/clip_<timestamp>.<ext>.
func BlockedScriptPath(ext string) (string, error) {
	dir, err := BlockedScriptsDir()
	if err != nil {
		return "", err
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
