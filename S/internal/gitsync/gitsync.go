// Package gitsync tự động commit + push thay đổi trong ~/ten-reepo
// lên remote. Chạy khi ś được gọi với cờ -g. Có retry queue: nếu push
// fail, ghi flag vào tmp/S/push_queue/ để lần chạy sau thử lại.
// Timezone đã được main.init() set thành ICT, nên time.Now() ở đây
// trả giờ VN, không cần hàm localTime riêng.
package gitsync

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"ten-reepo/S/internal/werr"
)

// Sync chạy chuỗi: add -A → (nếu có thay đổi) commit → pull --rebase
// → push trong repoPath. Trả về chuỗi trạng thái ngắn cho report.
func Sync(repoPath, logPath string) string {
	logFile, err := os.Create(logPath)
	if err != nil {
		return werr.Wrap(werr.KindSyncFailed, err, "không mở được log").Error()
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

	queueDir := filepath.Join(repoPath, "tmp", "S", "push_queue")
	retryCount := countFlags(queueDir)

	if code, _ := run("add", "-A"); code != 0 {
		writeFailFlag(queueDir, "add")
		return fmt.Sprintf("failed: git add exit=%d", code)
	}

	checkCmd := exec.Command("git", "-C", repoPath, "diff", "--cached", "--quiet")
	if checkCmd.Run() == nil {
		return "skipped: nothing to commit"
	}

	msg := fmt.Sprintf("ś auto-sync %s", time.Now().Format("2006-01-02 15:04:05"))
	if code, _ := run("commit", "-m", msg); code != 0 {
		writeFailFlag(queueDir, "commit")
		return fmt.Sprintf("failed: git commit exit=%d", code)
	}

	run("pull", "--rebase", "origin", "HEAD")

	if code, _ := run("push", "origin", "HEAD"); code != 0 {
		writeFailFlag(queueDir, "push")
		return fmt.Sprintf("failed: git push exit=%d", code)
	}

	clearFlags(queueDir)

	if retryCount > 0 {
		return fmt.Sprintf("pushed (retried %d)", retryCount)
	}
	return "pushed"
}

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

func writeFailFlag(dir, reason string) {
	_ = os.MkdirAll(dir, 0o755)
	name := fmt.Sprintf("%d_%s.fail", time.Now().UnixNano(), reason)
	_ = os.WriteFile(filepath.Join(dir, name), []byte(reason), 0o600)
}

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
