// Command ś biến nội dung clipboard thành file script, chạy script đó,
// rồi set kết quả (report markdown) trở lại clipboard. Nếu có cờ -g,
// tự động commit + push thay đổi trong ~/ten-reepo lên git remote.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"ten-reepo/S/internal/clipboard"
	"ten-reepo/S/internal/detect"
	"ten-reepo/S/internal/gitsync"
	"ten-reepo/S/internal/paths"
	"ten-reepo/S/internal/report"
	"ten-reepo/S/internal/runner"
)

func main() {
	syncFlag := flag.Bool("g", false, "tự động commit + push thay đổi lên git sau khi chạy")
	flag.Parse()

	start := time.Now()

	// 1. Đọc clipboard.
	text, err := clipboard.Get()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ś: không đọc được clipboard: %v\n", err)
		os.Exit(1)
	}
	if text == "" {
		fmt.Fprintln(os.Stderr, "ś: clipboard rỗng.")
		os.Exit(1)
	}

	// 2. Lưu clipboard gốc vào tmp/S để không mất dữ liệu khi ghi đè.
	if backup, err := paths.ClipboardBackupPath(); err == nil {
		_ = os.WriteFile(backup, []byte(text), 0o600)
	}

	// 3. Đoán ngôn ngữ + đuôi file.
	lang, ext := detect.Guess(text)

	// 4. Ghi file script vào S/scripts/.
	filePath, err := paths.ScriptPath(ext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ś: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filePath, []byte(text), 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "ś: %v\n", err)
		os.Exit(1)
	}

	// 5. Chạy file, capture stdout/stderr song song terminal.
	res := runner.Run(lang, filePath)
	res.Lang = lang
	res.ScriptPath = filePath
	res.Duration = time.Since(start).Seconds()
	res.CodeContent = text
	res.Changes = []string{"CREATE " + filePath}

	// 6. Nếu có cờ -g, sync git.
	if *syncFlag {
		repoRoot, err1 := paths.RepoRoot()
		tmpDir, err2 := paths.TmpDir()
		if err1 != nil || err2 != nil {
			res.Sync = "failed: không lấy được path"
		} else {
			logPath := gitsync.LogPath(tmpDir)
			res.Sync = gitsync.Sync(repoRoot, logPath)
		}
	}

	// 7. Build report markdown + set clipboard.
	md := report.Build(res)
	if err := clipboard.Set(md); err != nil {
		fmt.Fprintf(os.Stderr, "ś: không set được clipboard: %v\n", err)
	}

	// 8. In trạng thái ngắn ra terminal.
	if res.Sync != "" {
		fmt.Printf("ś: %s exit=%d %.3fs sync=%s %s\n", res.Status, res.ExitCode, res.Duration, res.Sync, filePath)
	} else {
		fmt.Printf("ś: %s exit=%d %.3fs %s\n", res.Status, res.ExitCode, res.Duration, filePath)
	}
}
