// Package runner chạy file script bằng interpreter tương ứng, capture
// stdout/stderr song song với in ra terminal, và trả về Result có cấu
// trúc để package report dựng markdown.
package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Result chứa toàn bộ thông tin của một lần chạy script.
type Result struct {
	Status           string   // "ok" | "error"
	StatusReason     string   // "completed" | "nonzero_exit" | "interpreter_not_found" | "unsupported_lang"
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
}

// Run chọn interpreter theo lang, chạy file, capture output và trả Result.
// Stdout/stderr được in ra terminal song song với capture qua io.MultiWriter.
func Run(lang, filePath string) Result {
	res := Result{Status: "ok", StatusReason: "completed", ExitCode: 0}

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

// commandFor trả về mảng lệnh (interpreter + tham số) cho từng ngôn ngữ.
func commandFor(lang, filePath string) ([]string, error) {
	switch lang {
	case "go":
		return []string{"go", "run", filePath}, nil
	case "python":
		return []string{"python3", filePath}, nil
	case "javascript":
		return []string{"node", filePath}, nil
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
	default:
		return nil, fmt.Errorf("ngôn ngữ %q chưa được hỗ trợ", lang)
	}
}
