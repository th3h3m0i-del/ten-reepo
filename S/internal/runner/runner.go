// Package runner chạy file script bằng interpreter/compiler tương ứng,
// capture stdout/stderr song song với in ra terminal, và trả về Result.
// Hỗ trợ 3 kiểu chạy:
//   - trực tiếp qua interpreter (python, node, ruby, bash, java, R, ...)
//   - compile rồi chạy binary (C, C++, Rust, Kotlin, C#, Zig, Nim, Crystal)
//   - đặc biệt (awk cần stdin rỗng, sqlite3 cần :memory:, bun báo lỗi nếu hỏng)
// Ngôn ngữ không có interpreter/compiler: trả unsupported_lang nhưng file
// script vẫn được giữ, không mất nội dung clipboard.
package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ten-reepo/S/internal/werr"
)

// Result chứa toàn bộ thông tin một lần chạy script.
type Result struct {
	Status           string   // "ok" | "error"
	StatusReason     string   // "completed" | "nonzero_exit" | "interpreter_not_found" | "unsupported_lang" | "compile_failed"
	ExitCode         int
	Duration         float64
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
	ErrorKind        string   // loại lỗi theo werr.Kind; rỗng nếu status ok
}

// runKind phân loại cách chạy.
type runKind int

const (
	kindDirect    runKind = iota // chạy thẳng interpreter
	kindCompile                  // compile rồi chạy binary/jar/exe
	kindSpecial                  // awk, sqlite cần stdin đặc biệt
)

// langSpec mô tả cách chạy một ngôn ngữ.
type langSpec struct {
	kind runKind
	// interpreter/compiler đầu tiên (dùng để LookPath)
	primary string
	// hàm build commandline (trả về argv) khi chạy hoặc compile
	build func(filePath, buildDir string) ([]string, error)
	// hàm build commandline chạy binary sau khi compile
	runBuilt func(filePath, buildDir string) ([]string, error)
}

// registry map lang → spec. Ngôn ngữ không có trong map = unsupported.
var registry = map[string]langSpec{
	// === Direct interpreters ===
	"python":     {kindDirect, "python3", single("python3"), nil},
	"javascript": {kindDirect, "node", single("node"), nil},
	"typescript": {kindDirect, "deno", argCmd("deno", "run"), nil},
	"deno":       {kindDirect, "deno", argCmd("deno", "run"), nil},
	"ruby":       {kindDirect, "ruby", single("ruby"), nil},
	"perl":       {kindDirect, "perl", single("perl"), nil},
	"php":        {kindDirect, "php", single("php"), nil},
	"lua":        {kindDirect, "lua", single("lua"), nil},
	"luajit":     {kindDirect, "luajit", single("luajit"), nil},
	"bash":       {kindDirect, "bash", single("bash"), nil},
	"zsh":        {kindDirect, "zsh", single("zsh"), nil},
	"sh":         {kindDirect, "sh", single("sh"), nil},
	"dash":       {kindDirect, "dash", single("dash"), nil},
	"fish":       {kindDirect, "fish", single("fish"), nil},
	"java":       {kindDirect, "java", single("java"), nil},
	"r":          {kindDirect, "Rscript", single("Rscript"), nil},
	"julia":      {kindDirect, "julia", single("julia"), nil},
	"tcl":        {kindDirect, "tclsh", single("tclsh"), nil},
	"groovy":     {kindDirect, "groovy", single("groovy"), nil},
	"scala":      {kindDirect, "scala", single("scala"), nil},
	"swift":      {kindDirect, "swift", single("swift"), nil},
	"erlang":     {kindDirect, "escript", single("escript"), nil},
	"bun":        {kindDirect, "bun", argCmd("bun", "run"), nil},

	// === Special (cần stdin đặc biệt) ===
	"awk":    {kindSpecial, "awk", awkCmd, nil},
	"sql":    {kindSpecial, "sqlite3", sqliteCmd, nil},

	// === Compile rồi chạy ===
	"c":     {kindCompile, "gcc", compileSimple("gcc"), runBinary},
	"cpp":   {kindCompile, "g++", compileSimple("g++"), runBinary},
	"rust":  {kindCompile, "rustc", compileSimple("rustc"), runBinary},
	"zig":   {kindCompile, "zig", compileSimple("zig"), runBinary},
	"kotlin": {kindCompile, "kotlinc", compileKotlin, runJar},
	"csharp": {kindCompile, "csc", compileCSharp, runMono},
	"crystal": {kindCompile, "crystal", compileCrystal, runBinary},
	"nim":     {kindCompile, "nim", compileNim, runBinary},
}

// single trả về builder cho lệnh dạng "<interpreter> <file>".
func single(bin string) func(string, string) ([]string, error) {
	return func(fp, _ string) ([]string, error) { return []string{bin, fp}, nil }
}

// argCmd trả về builder cho lệnh dạng "<bin> <sub> <file>".
func argCmd(bin, sub string) func(string, string) ([]string, error) {
	return func(fp, _ string) ([]string, error) { return []string{bin, sub, fp}, nil }
}

// awkCmd: awk -f <file> < /dev/null.
func awkCmd(fp, _ string) ([]string, error) { return []string{"awk", "-f", fp}, nil }

// sqliteCmd: sqlite3 :memory: (stdin từ file .sql).
func sqliteCmd(fp, _ string) ([]string, error) { return []string{"sqlite3", ":memory:"}, nil }

// compileSimple trả về builder compile binary cơ bản: "<cc> <file> -o <bin>".
func compileSimple(cc string) func(string, string) ([]string, error) {
	return func(fp, bd string) ([]string, error) {
		bin := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid()))
		return []string{cc, fp, "-o", bin}, nil
	}
}

// compileKotlin: kotlinc -include-runtime -d out.jar.
func compileKotlin(fp, bd string) ([]string, error) {
	out := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid())+".jar")
	return []string{"kotlinc", fp, "-include-runtime", "-d", out}, nil
}

// compileCSharp: csc -out:out.exe.
func compileCSharp(fp, bd string) ([]string, error) {
	out := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid())+".exe")
	return []string{"csc", "-out:" + out, fp}, nil
}

// compileCrystal: crystal build <file> -o <bin>.
func compileCrystal(fp, bd string) ([]string, error) {
	bin := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid()))
	return []string{"crystal", "build", fp, "-o", bin}, nil
}

// compileNim: nim c -o:<bin> --hints:off <file>.
func compileNim(fp, bd string) ([]string, error) {
	bin := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid()))
	return []string{"nim", "c", "--hints:off", "-o:" + bin, fp}, nil
}

// runBinary: chạy binary đã compile (path suy ra từ cùng quy tắc).
func runBinary(_ string, bd string) ([]string, error) {
	bin := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid()))
	return []string{bin}, nil
}

// runJar: java -jar out.jar.
func runJar(_ string, bd string) ([]string, error) {
	out := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid())+".jar")
	return []string{"java", "-jar", out}, nil
}

// runMono: mono out.exe.
func runMono(_ string, bd string) ([]string, error) {
	out := filepath.Join(bd, "out_"+fmt.Sprintf("%d", os.Getpid())+".exe")
	return []string{"mono", out}, nil
}

// Run chọn cách chạy theo lang và thực thi.
func Run(lang, filePath string) Result {
	res := Result{Status: "ok", StatusReason: "completed", ExitCode: 0}

	spec, ok := registry[lang]
	if !ok {
		res.Status = "error"
		res.StatusReason = "unsupported_lang"
		res.Stderr = fmt.Sprintf("ngôn ngữ %q chưa được hỗ trợ; file vẫn được lưu", lang)
		res.ExitCode = 1
		res.ErrorHits = 1
		res.ErrorKind = string(werr.KindMissingInterp)
		return res
	}

	// Kiểm tra interpreter/compiler có trên máy.
	if _, err := exec.LookPath(spec.primary); err != nil {
		res.Status = "error"
		res.StatusReason = "interpreter_not_found"
		res.Stderr = fmt.Sprintf("không tìm thấy %q trong PATH; file vẫn được lưu", spec.primary)
		res.ExitCode = 127
		res.ErrorHits = 1
		res.ErrorKind = string(werr.KindMissingInterp)
		return res
	}

	// Bun đặc biệt: binary có nhưng có thể hỏng — verify bằng --version.
	if lang == "bun" {
		if err := exec.Command("bun", "--version").Run(); err != nil {
			res.Status = "error"
			res.StatusReason = "interpreter_not_found"
			res.Stderr = "bun binary tồn tại nhưng không chạy được (hỏng); file vẫn được lưu"
			res.ExitCode = 127
			res.ErrorHits = 1
			res.ErrorKind = string(werr.KindMissingInterp)
			return res
		}
	}

	// Build dir cho compile target.
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

	switch spec.kind {
	case kindDirect:
		cmd, err := spec.build(filePath, buildDir)
		if err != nil {
			return failCompile(res, err.Error())
		}
		return execCapture(cmd, res, "")

	case kindCompile:
		// Bước 1: compile.
		compileCmd, err := spec.build(filePath, buildDir)
		if err != nil {
			return failCompile(res, err.Error())
		}
		var compileOut bytes.Buffer
		cc := exec.Command(compileCmd[0], compileCmd[1:]...)
		cc.Stdout = &compileOut
		cc.Stderr = &compileOut
		if err := cc.Run(); err != nil {
			res.Status = "error"
			res.StatusReason = "compile_failed"
			res.Stderr = compileOut.String()
			res.ExitCode = 1
			res.ErrorHits = 1
			return res
		}
		// Bước 2: chạy binary/jar/exe.
		runCmd, err := spec.runBuilt(filePath, buildDir)
		if err != nil {
			return failCompile(res, err.Error())
		}
		out := execCapture(runCmd, res, "")
		// Dọn binary sau khi chạy.
		_ = os.Remove(runCmd[0])
		return out

	case kindSpecial:
		// awk và sqlite cần stdin đặc biệt.
		code := ""
		cmd, err := spec.build(filePath, buildDir)
		if err != nil {
			return failCompile(res, err.Error())
		}
		if lang == "sql" {
			// Đọc file .sql vào stdin.
			data, err := os.ReadFile(filePath)
			if err != nil {
				return failCompile(res, "không đọc được file .sql")
			}
			code = string(data)
		}
		return execCapture(cmd, res, code)
	}

	return res
}

// failCompile trả Result lỗi với reason=compile_failed.
func failCompile(res Result, msg string) Result {
	res.Status = "error"
	res.StatusReason = "compile_failed"
	res.Stderr = msg
	res.ExitCode = 1
	res.ErrorHits = 1
	res.ErrorKind = string(werr.KindSyntax)
	return res
}

// execCapture chạy cmd, capture stdout/stderr song song terminal.
// stdinData != "" thì feed vào stdin, ngược lại /dev/null (rỗng).
func execCapture(cmd []string, res Result, stdinData string) Result {
	var stdoutBuf, stderrBuf bytes.Buffer
	c := exec.Command(cmd[0], cmd[1:]...)
	if stdinData != "" {
		c.Stdin = strings.NewReader(stdinData)
	} else {
		c.Stdin = os.Stdin
	}
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
			res.ErrorKind = string(werr.KindRuntime)
		} else {
			res.ExitCode = 1
			res.StatusReason = runErr.Error()
			res.ErrorKind = string(werr.KindUnknown)
		}
	}
	return res
}

// countLines đếm số dòng (tính cả dòng cuối không có "\n").
func countLines(s string) int {
	if s == "" {
		return 0
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Count(s, "\n") + 1
}
