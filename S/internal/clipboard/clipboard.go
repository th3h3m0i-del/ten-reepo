// Package clipboard đọc và ghi nội dung clipboard trên Termux thông qua
// termux-clipboard-get và termux-clipboard-set. Không dùng thư viện ngoài
// để tránh kéo thêm dependency và cache ngoài scope.
package clipboard

import (
	"os/exec"
	"strings"
)

// Get chạy termux-clipboard-get và trả về nội dung đã trim khoảng trắng.
func Get() (string, error) {
	out, err := exec.Command("termux-clipboard-get").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Set ghi nội dung vào clipboard qua stdin của termux-clipboard-set.
// Dùng stdin thay vì argv để tránh giới hạn độ dài tham số và tránh
// shell escape phức tạp.
func Set(text string) error {
	cmd := exec.Command("termux-clipboard-set")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
