// Command ś biến nội dung clipboard thành file script, chạy script đó,
// rồi set kết quả (report markdown) trở lại clipboard. Nếu có cờ -g,
// tự động commit + push thay đổi trong ~/ten-reepo lên git remote.
//
// Gitguard hai lớp:
//   - Lớp 1: scan clipboard trước khi lưu. Nếu secret, file được lưu
//     vào tmp/S/blocked_scripts/ (không push được), không vào S/scripts/.
//   - Lớp 2: trước khi sync, scan toàn bộ S/scripts/. Nếu bất kỳ file
//     nào có secret, block sync (không add, không push).
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

// init set timezone global ICT (UTC+7).
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

	// 2. Gitguard lớp 1: scan clipboard. Nếu secret, đánh dấu để lưu
	// file vào blocked_scripts/ thay vì scripts/.
	clipboardBlocked := false
	clipboardMsg := ""
	if cfg.EnableGitGuard {
		findings := gitguard.Scan(text, cfg.SecretPatterns)
		if gitguard.Blocked(findings) {
			clipboardBlocked = true
			clipboardMsg = gitguard.Summary(findings)
		}
	}

	// 3. Lưu clipboard gốc (atomic).
	if backup, err := paths.ClipboardBackupPath(); err == nil {
		_ = atomicwrite.WriteFile(backup, []byte(text), 0o600)
	}

	// 4. Đoán ngôn ngữ + đuôi file.
	lang, ext := detect.Guess(text)

	// 5. Ghi file script vào đúng thư mục tuỳ lớp 1.
	var filePath string
	if clipboardBlocked {
		filePath, err = paths.BlockedScriptPath(ext)
	} else {
		filePath, err = paths.ScriptPath(ext)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ś: %v\n", err)
		os.Exit(1)
	}
	if err := atomicwrite.WriteFile(filePath, []byte(text), 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "ś: %v\n", err)
		os.Exit(1)
	}

	// 6. Chạy file, capture output song song terminal.
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

	// 7. Sync nếu có cờ -g. Kiểm tra cả hai lớp trước khi gọi gitsync.
	if *syncFlag {
		switch {
		case clipboardBlocked:
			res.Sync = "blocked: secret in clipboard (" + clipboardMsg + ")"
		default:
			// Lớp 2: scan toàn bộ S/scripts/.
			scriptsDir, err := paths.ScriptsDir()
			if err != nil {
				res.Sync = "failed: không lấy được scripts dir"
				break
			}
			findings, err := gitguard.ScanDir(scriptsDir, cfg.SecretPatterns)
			if err != nil {
				res.Sync = "failed: không scan được scripts dir"
				break
			}
			if gitguard.Blocked(findings) {
				res.Sync = "blocked: secret in scripts/ (" + gitguard.Summary(findings) + ")"
				break
			}
			// Cả hai lớp sạch → gọi gitsync.
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

	// 8. Build report + set clipboard.
	traceStart = time.Now()
	md := report.Build(res, cfg)
	if trace.Enabled {
		fmt.Fprintf(os.Stderr, "[trace] report.Build: %s\n", time.Since(traceStart))
	}
	if err := clipboard.Set(md); err != nil {
		fmt.Fprintf(os.Stderr, "ś: không set được clipboard: %v\n", err)
	}

	// 9. In trạng thái ngắn.
	if res.Sync != "" {
		fmt.Printf("ś: %s exit=%d %.3fs sync=%s %s\n", res.Status, res.ExitCode, res.Duration, res.Sync, filePath)
	} else {
		fmt.Printf("ś: %s exit=%d %.3fs %s\n", res.Status, res.ExitCode, res.Duration, filePath)
	}
	if clipboardBlocked {
		fmt.Fprintf(os.Stderr, "ś: gitguard lớp 1 — clipboard có secret, file lưu vào blocked_scripts: %s\n", clipboardMsg)
	}
}
