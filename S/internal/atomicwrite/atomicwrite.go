// Package atomicwrite ghi file an toàn: ghi vào file tạm cùng thư mục
// rồi rename đè lên file đích. Nếu crash giữa chừng, file đích vẫn
// nguyên vẹn (không bị truncate hay ghi dở). Dùng cho mọi chỗ ś ghi
// file quan trọng: script, report, clipboard backup.
package atomicwrite

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile ghi data vào path một cách atomic. File tạm được tạo cùng
// thư mục với path để đảm bảo rename cùng filesystem (rename跨
// filesystem sẽ fail). File tạm có tên `.tmp_<pid>_<basename>`.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, fmt.Sprintf(".tmp_%d_%s", os.Getpid(), filepath.Base(path)))

	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("atomicwrite: ghi temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("atomicwrite: rename: %w", err)
	}
	return nil
}
