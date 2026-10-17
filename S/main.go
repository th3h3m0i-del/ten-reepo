// Command ś biến nội dung clipboard thành file script, chạy script đó,
// rồi set kết quả (report markdown) trở lại clipboard. Nếu có cờ -g,
// tự động commit + push thay đổi trong ~/ten-reepo lên git remote.
// Có gitguard chặn secret trước khi push; có atomicwrite cho mọi file
// ghi ra; có trace profiling khi WW_TRACE=1.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"ten-reepo/S/internal/atomicwrite"
	"ten-reepo/S/internal/clipboard"
	"ten-reepo/S/internal/config"
	"ten-reepo/S/internal/detect"
	"ten-reepo/S/internal/gitguard"
	"ten-reepo/S/internal/gitsync"
	"ten-reepo/S/internal/paths"
	"ten-reepo/S/internal/report"
	"ten-reepo/S/internal/runner"
	"ten-reepo/S/internal/trace"
)

// init set timezone global ICT (UTC+7) một lần, để mọi package dùng
// time.Now() đều trả giờ Việt Nam, không cần hàm localTime riêng.
func init() {
	time.Local = time.FixedZone("ICT", 7*3600)
}

func main() {
	syncFlag := flag.Bool("g", false, "tự động commit + push thay đổi lên git sau khi chạy")
	flag.Parse()

	start := time.Now()
	cfg := config.Load()

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

	// 2. Gitguard: scan secret trước khi làm gì khác. Nếu phát hiện,
	// vẫn lưu file và chạy (script có thể vô hại), nhưng KHÔNG sync
	// và report ghi rõ lý do.
	guardBlocked := false
	guardMsg := ""
	if cfg.EnableGitGuard {
		findings := gitguard.Scan(text, cfg.SecretPatterns)
		if gitguard.Blocked(findings) {
			guardBlocked = true
			guardMsg = gitguard.Summary(findings)
		}
	}

	// 3. Lưu clipboard gốc vào tmp/S (atomic).
	if backup, err := paths.ClipboardBackupPath(); err == nil {
		_ = atomicwrite.WriteFile(backup, []byte(text), 0o600)
	}

	// 4. Đoán ngôn ngữ + đuôi file.
	lang, ext := detect.Guess(text)

	// 5. Ghi file script vào S/scripts/ (atomic).
	filePath, err := paths.ScriptPath(ext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ś: %v\n", err)
		os.Exit(1)
	}
	if err := atomicwrite.WriteFile(filePath, []byte(text), 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "ś: %v\n", err)
		os.Exit(1)
	}

	// 6. Chạy file, capture stdout/stderr song song terminal.
	traceStart := time.Now()
	res := runner.Run(lang, filePath)
	if trace.Enabled {
		fmt.Fprintf(os.Stderr, "[trace] runner.Run: %s\n", time.Since(traceStart))
	}
	res.Lang = lang
	res.ScriptPath = filePath
	res.Duration = time.Since(start).Seconds()
	res.CodeContent = text
	res.Changes = []string{"CREATE " + filePath}

	// 7. Sync git nếu có cờ -g. Nếu gitguard blocked, không sync.
	if *syncFlag {
		if guardBlocked {
			res.Sync = "blocked: secret detected (" + guardMsg + ")"
		} else {
			repoRoot, err1 := paths.RepoRoot()
			tmpDir, err2 := paths.TmpDir()
			if err1 != nil || err2 != nil {
				res.Sync = "failed: không lấy được path"
			} else {
				logPath := gitsync.LogPath(tmpDir)
				res.Sync = gitsync.Sync(repoRoot, logPath)
			}
		}
	}

	// 8. Build report markdown + set clipboard.
	traceStart = time.Now()
	md := report.Build(res, cfg)
	if trace.Enabled {
		fmt.Fprintf(os.Stderr, "[trace] report.Build: %s\n", time.Since(traceStart))
	}
	if err := clipboard.Set(md); err != nil {
		fmt.Fprintf(os.Stderr, "ś: không set được clipboard: %v\n", err)
	}

	// 9. In trạng thái ngắn ra terminal.
	if res.Sync != "" {
		fmt.Printf("ś: %s exit=%d %.3fs sync=%s %s\n", res.Status, res.ExitCode, res.Duration, res.Sync, filePath)
	} else {
		fmt.Printf("ś: %s exit=%d %.3fs %s\n", res.Status, res.ExitCode, res.Duration, filePath)
	}
	if guardBlocked {
		fmt.Fprintf(os.Stderr, "ś: gitguard phát hiện secret — không sync: %s\n", guardMsg)
	}
}
