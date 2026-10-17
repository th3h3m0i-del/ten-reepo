// Package gitsync tự động commit + push thay đổi trong ~/ten-reepo
// lên remote. Chạy khi ś được gọi với cờ -g. Mục đích: Bosz không phải
// chạy git tay trên Termux, nhưng AI vẫn recon được repo qua GitHub.
// Không dùng --force, không prompt, không đụng credential ngoài env.
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
// Output của git được ghi vào logPath (nằm trong tmp/S).
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

	// Bước 1: stage mọi thay đổi.
	if code, _ := run("add", "-A"); code != 0 {
		return fmt.Sprintf("failed: git add exit=%d", code)
	}

	// Bước 2: kiểm tra có gì staged để commit không.
	// "git diff --cached --quiet" trả 0 = không có thay đổi staged.
	checkCmd := exec.Command("git", "-C", repoPath, "diff", "--cached", "--quiet")
	if checkCmd.Run() == nil {
		return "skipped: nothing to commit"
	}

	// Bước 3: commit.
	msg := fmt.Sprintf("ś auto-sync %s", time.Now().Format("2006-01-02 15:04:05"))
	if code, _ := run("commit", "-m", msg); code != 0 {
		return fmt.Sprintf("failed: git commit exit=%d", code)
	}

	// Bước 4: pull --rebase để tránh reject nếu remote có commit mới.
	// Fail ở đây không chặn push (có thể đang offline nhưng push bằng
	// credential đã lưu vẫn đi được, hoặc local đã up-to-date).
	run("pull", "--rebase", "origin", "HEAD")

	// Bước 5: push.
	if code, _ := run("push", "origin", "HEAD"); code != 0 {
		return fmt.Sprintf("failed: git push exit=%d", code)
	}

	return "pushed"
}

// LogPath sinh đường dẫn log cho một lần sync trong tmpDir.
func LogPath(tmpDir string) string {
	name := fmt.Sprintf("git_sync_%d.log", time.Now().UnixNano())
	return filepath.Join(tmpDir, name)
}
