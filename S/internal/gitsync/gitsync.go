// Package gitsync tự động commit + push thay đổi trong ~/ten-reepo
// lên remote. Chạy khi ś được gọi với cờ -g. Có retry queue: nếu push
// fail, ghi flag vào tmp/S/push_queue/ để lần chạy sau thử lại.
package gitsync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Sync chạy chuỗi: add -A → (nếu có thay đổi) commit → pull --rebase
// → push trong repoPath. Trả về chuỗi trạng thái ngắn để đưa vào report.
// Commit message dùng giờ local, không phải UTC.
func Sync(repoPath, logPath string) string {
	logFile, err := os.Create(logPath)
	if err != nil {
		return "failed: không mở được log"
	}
	defer logFile.Close()

	run := func(args ...string) (int, error) {
		cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.Env = append(os.Environ(),
			"GIT_TERMINAL_PROMPT=0",
			"GIT_ASKPASS=/bin/true",
		)
		err := cmd.Run()
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), err
		}
		if err != nil {
			return -1, err
		}
		return 0, nil
	}

	// Đếm số flag retry trước khi push.
	queueDir := filepath.Join(repoPath, "tmp", "S", "push_queue")
	retryCount := countFlags(queueDir)

	// Bước 1: stage mọi thay đổi.
	if code, _ := run("add", "-A"); code != 0 {
		writeFailFlag(queueDir, "add")
		return fmt.Sprintf("failed: git add exit=%d", code)
	}

	// Bước 2: kiểm tra có gì staged để commit không.
	// "git diff --cached --quiet" trả 0 = không có thay đổi staged.
	checkCmd := exec.Command("git", "-C", repoPath, "diff", "--cached", "--quiet")
	if checkCmd.Run() == nil {
		return "skipped: nothing to commit"
	}

	// Bước 3: commit với giờ local.
	msg := fmt.Sprintf("ś auto-sync %s", time.Now().Local().Format("2006-01-02 15:04:05"))
	if code, _ := run("commit", "-m", msg); code != 0 {
		writeFailFlag(queueDir, "commit")
		return fmt.Sprintf("failed: git commit exit=%d", code)
	}

	// Bước 4: pull --rebase để tránh reject. Fail không chặn push.
	run("pull", "--rebase", "origin", "HEAD")

	// Bước 5: push.
	if code, _ := run("push", "origin", "HEAD"); code != 0 {
		writeFailFlag(queueDir, "push")
		return fmt.Sprintf("failed: git push exit=%d", code)
	}

	// Push thành công: xoá hết flag retry.
	clearFlags(queueDir)

	if retryCount > 0 {
		return fmt.Sprintf("pushed (retried %d)", retryCount)
	}
	return "pushed"
}

// countFlags đếm số file flag trong queueDir.
func countFlags(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".fail" {
			n++
		}
	}
	return n
}

// writeFailFlag ghi một file flag đánh dấu lần fail, để lần sau retry.
func writeFailFlag(dir, reason string) {
	_ = os.MkdirAll(dir, 0o755)
	name := fmt.Sprintf("%d_%s.fail", time.Now().UnixNano(), reason)
	_ = os.WriteFile(filepath.Join(dir, name), []byte(reason), 0o600)
}

// clearFlags xoá toàn bộ file flag sau khi push thành công.
func clearFlags(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".fail" {
			_ = os.Remove(filepath.Join(dir, e.Name()))
		}
	}
}

// LogPath sinh đường dẫn log cho một lần sync trong tmpDir.
func LogPath(tmpDir string) string {
	name := fmt.Sprintf("git_sync_%d.log", time.Now().UnixNano())
	return filepath.Join(tmpDir, name)
}
