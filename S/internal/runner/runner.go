// Package runner chạy file script bằng interpreter tương ứng, capture
// stdout/stderr song song với in ra terminal, và trả về Result có cấu
// trúc để package report dựng markdown. Với C/C++/Rust phải compile
// trước rồi chạy binary, binary tạm nằm trong tmp/S/build/.
package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Result chứa toàn bộ thông tin của một lần chạy script.
type Result struct {
	Status           string   // "ok" | "error"
	StatusReason     string   // "completed" | "nonzero_exit" | "interpreter_not_found" | "unsupported_lang" | "compile_failed"
	ExitCode         int
	Duration         float64  // giây
	Lang             string
	ScriptPath       string
	CodeContent      string
	Stdout           string
	Stderr           string
	StdoutTotalLines int
	StderrTotalLines int
	ErrorHits        int
	Changes          []string
	Sync             string   // kết quả git sync; rỗng nếu không dùng -g
}

// Run chọn cách chạy theo lang. Python/bash/js chạy trực tiếp qua
// interpreter. C/C++/Rust compile rồi chạy binary. Ngôn ngữ không hỗ
// trợ (kotlin, swift, sql) trả lỗi unsupported_lang nhưng KHÔNG xoá
// file script — nội dung clipboard luôn được giữ lại.
func Run(lang, filePath string) Result {
	res := Result{Status: "ok", StatusReason: "completed", ExitCode: 0}

	// Compile-run cho C/C++/Rust.
	if lang == "c" || lang == "cpp" || lang == "rust" {
		return runCompiled(lang, filePath, res)
	}

	cmd, err := commandFor(lang, filePath)
	if err != nil {
		res.Status = "error"
		res.StatusReason = "unsupported_lang"
		res.Stderr = err.Error()
		res.ExitCode = 1
		res.ErrorHits = 1
		return res
	}
	if _, err := exec.LookPath(cmd[0]); err != nil {
		res.Status = "error"
		res.StatusReason = "interpreter_not_found"
		res.Stderr = fmt.Sprintf("không tìm thấy interpreter %q trong PATH", cmd[0])
		res.ExitCode = 127
		res.ErrorHits = 1
		return res
	}

	return execCapture(cmd, res)
}

// runCompiled xử lý riêng C/C++/Rust: compile → chạy binary.
// Binary build nằm trong tmp/S/build/, không đụng ngoài scope.
func runCompiled(lang, filePath string, res Result) Result {
	home, err := os.UserHomeDir()
	if err != nil {
		res.Status = "error"
		res.StatusReason = "compile_failed"
		res.Stderr = "không lấy được home"
		res.ExitCode = 1
		res.ErrorHits = 1
		return res
	}
	buildDir := filepath.Join(home, "ten-reepo", "tmp", "S", "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		res.Status = "error"
		res.StatusReason = "compile_failed"
		res.Stderr = "không tạo được build dir"
		res.ExitCode = 1
		res.ErrorHits = 1
		return res
	}
	binPath := filepath.Join(buildDir, "clip_bin_"+fmt.Sprintf("%d", os.Getpid()))

	var compiler []string
	switch lang {
	case "c":
		compiler = []string{"gcc", filePath, "-o", binPath}
	case "cpp":
		compiler = []string{"g++", filePath, "-o", binPath}
	case "rust":
		compiler = []string{"rustc", filePath, "-o", binPath}
	default:
		res.Status = "error"
		res.StatusReason = "unsupported_lang"
		res.Stderr = "ngôn ngữ compile không hỗ trợ: " + lang
		res.ExitCode = 1
		res.ErrorHits = 1
		return res
	}

	if _, err := exec.LookPath(compiler[0]); err != nil {
		res.Status = "error"
		res.StatusReason = "interpreter_not_found"
		res.Stderr = fmt.Sprintf("không tìm thấy compiler %q", compiler[0])
		res.ExitCode = 127
		res.ErrorHits = 1
		return res
	}

	// Compile.
	compileCmd := exec.Command(compiler[0], compiler[1:]...)
	var compileOut bytes.Buffer
	compileCmd.Stdout = &compileOut
	compileCmd.Stderr = &compileOut
	if err := compileCmd.Run(); err != nil {
		res.Status = "error"
		res.StatusReason = "compile_failed"
		res.Stderr = compileOut.String()
		res.ExitCode = 1
		res.ErrorHits = 1
		return res
	}

	// Chạy binary vừa compile.
	runRes := execCapture([]string{binPath}, res)

	// Xoá binary sau khi chạy (dọn scratch).
	_ = os.Remove(binPath)

	return runRes
}

// execCapture chạy cmd, capture stdout/stderr song song terminal, trả Result.
func execCapture(cmd []string, res Result) Result {
	var stdoutBuf, stderrBuf bytes.Buffer
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Stdin = os.Stdin
	c.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	c.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)

	runErr := c.Run()
	res.Stdout = stdoutBuf.String()
	res.Stderr = stderrBuf.String()
	res.StdoutTotalLines = countLines(res.Stdout)
	res.StderrTotalLines = countLines(res.Stderr)

	if runErr != nil {
		res.Status = "error"
		res.ErrorHits = 1
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
			res.StatusReason = "nonzero_exit"
		} else {
			res.ExitCode = 1
			res.StatusReason = runErr.Error()
		}
	}
	return res
}

// countLines đếm số dòng của chuỗi, tính cả dòng cuối không có "\n".
func countLines(s string) int {
	if s == "" {
		return 0
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Count(s, "\n") + 1
}

// commandFor trả về mảng lệnh (interpreter + tham số) cho các ngôn ngữ
// chạy trực tiếp. C/C++/Rust không đi qua đây (xử lý ở runCompiled).
// Kotlin/Swift/SQL trả unsupported_lang — không có interpreter trên Termux.
func commandFor(lang, filePath string) ([]string, error) {
	switch lang {
	case "go":
		return []string{"go", "run", filePath}, nil
	case "python":
		return []string{"python3", filePath}, nil
	case "javascript":
		return []string{"node", filePath}, nil
	case "typescript":
		return []string{"deno", "run", filePath}, nil
	case "ruby":
		return []string{"ruby", filePath}, nil
	case "perl":
		return []string{"perl", filePath}, nil
	case "php":
		return []string{"php", filePath}, nil
	case "lua":
		return []string{"lua", filePath}, nil
	case "bash":
		return []string{"bash", filePath}, nil
	case "zsh":
		return []string{"zsh", filePath}, nil
	case "sh":
		return []string{"sh", filePath}, nil
	case "deno":
		return []string{"deno", "run", filePath}, nil
	case "bun":
		return []string{"bun", "run", filePath}, nil
	case "java":
		return []string{"java", filePath}, nil
	case "kotlin", "swift", "sql":
		return nil, fmt.Errorf("ngôn ngữ %q chưa có interpreter trên Termux; file vẫn được lưu", lang)
	default:
		return nil, fmt.Errorf("ngôn ngữ %q chưa được hỗ trợ", lang)
	}
}
