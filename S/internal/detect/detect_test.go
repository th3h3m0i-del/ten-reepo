// Package detect test cho Guess. Mỗi case kiểm tra lang+ext trả về
// khi đưa vào một snippet đặc trưng. Cover toàn bộ nhánh shebang và
// token của các ngôn ngữ có trên Termux (recon 2026-10-17).
package detect

import "testing"

func TestGuess(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantLang string
		wantExt  string
	}{
		// === Shebang ===
		{"shebang_python", "#!/usr/bin/env python3\nprint('hi')", "python", "py"},
		{"shebang_node", "#!/usr/bin/env node\nconsole.log('hi')", "javascript", "js"},
		{"shebang_bash", "#!/bin/bash\necho hi", "bash", "sh"},
		{"shebang_zsh", "#!/bin/zsh\necho hi", "zsh", "sh"},
		{"shebang_sh", "#!/bin/sh\necho hi", "sh", "sh"},
		{"shebang_dash", "#!/bin/dash\necho hi", "dash", "sh"},
		{"shebang_fish", "#!/usr/bin/env fish\necho hi", "fish", "fish"},
		{"shebang_ruby", "#!/usr/bin/env ruby\nputs 'hi'", "ruby", "rb"},
		{"shebang_php", "#!/usr/bin/env php\n<?php echo 'hi';", "php", "php"},
		{"shebang_lua", "#!/usr/bin/env lua\nprint('hi')", "lua", "lua"},
		{"shebang_luajit", "#!/usr/bin/env luajit\nprint('hi')", "luajit", "lua"},
		{"shebang_rscript", "#!/usr/bin/env Rscript\nprint('hi')", "r", "R"},
		{"shebang_julia", "#!/usr/bin/env julia\nprintln('hi')", "julia", "jl"},
		{"shebang_tclsh", "#!/usr/bin/env tclsh\nputs hi", "tcl", "tcl"},
		{"shebang_groovy", "#!/usr/bin/env groovy\nprintln 'hi'", "groovy", "groovy"},
		{"shebang_scala", "#!/usr/bin/env scala\nobject M", "scala", "scala"},
		{"shebang_escript", "#!/usr/bin/env escript\nmain(_) -> ok.", "erlang", "erl"},
		{"shebang_gawk", "#!/usr/bin/gawk -f\nBEGIN { print 1 }", "awk", "awk"},
		{"shebang_deno", "#!/usr/bin/env deno\nconsole.log(1)", "deno", "ts"},
		{"shebang_bun", "#!/usr/bin/env bun\nconsole.log(1)", "bun", "js"},
		{"shebang_crystal", "#!/usr/bin/env crystal\nputs 1", "crystal", "cr"},
		{"shebang_nim", "#!/usr/bin/env nim\necho 1", "nim", "nim"},
		{"shebang_zig", "#!/usr/bin/env zig\npub fn main() void {}", "zig", "zig"},
		{"shebang_kotlin", "#!/usr/bin/env kotlin\nfun main() {}", "kotlin", "kt"},
		{"shebang_swift", "#!/usr/bin/env swift\nprint(1)", "swift", "swift"},

		// === Token only (không shebang) ===
		{"token_go", "package main\n\nfunc main() {\n}", "go", "go"},
		{"token_java", "public class Main {\n  public static void main(String[] a) {}\n}", "java", "java"},
		{"token_scala", "object Main {\n  def main(args: Array[String]) = println(1)\n}", "scala", "scala"},
		{"token_kotlin", "fun main() {\n  val x = 1\n  println(x)\n}", "kotlin", "kt"},
		{"token_swift", "import Foundation\nfunc greet() { print(1) }", "swift", "swift"},
		{"token_csharp", "using System;\nnamespace App {\n  class Program { static void Main() {} }\n}", "csharp", "cs"},
		{"token_cpp", "#include <iostream>\nint main() { std::cout << 1; }", "cpp", "cpp"},
		{"token_c", "#include <stdio.h>\nint main(void) { printf(\"hi\"); return 0; }", "c", "c"},
		{"token_rust", "fn main() {\n    let mut x = 1;\n    println!(\"{}\", x);\n}", "rust", "rs"},
		{"token_zig", "const std = @import(\"std\");\npub fn main() void {}", "zig", "zig"},
		{"token_typescript", "interface User { name: string }\nexport class A {}", "typescript", "ts"},
		{"token_julia", "function hello()\n  println(\"hi\")\nend", "julia", "jl"},
		{"token_r", "x <- function(a) a + 1\nprint(x(1))", "r", "R"},
		{"token_tcl", "proc hello {} {\n  set x 1\n  puts $x\n}", "tcl", "tcl"},
		{"token_groovy", "class Hello {\n  def greet() { println 'hi' }\n}", "groovy", "groovy"},
		{"token_nim", "proc hello() =\n  echo \"hi\"", "nim", "nim"},
		{"token_sql_select", "SELECT id, name FROM users WHERE id = 1;", "sql", "sql"},
		{"token_sql_create", "CREATE TABLE users (id INT, name TEXT);", "sql", "sql"},
		{"token_python", "def greet():\n    print(\"hi\")", "python", "py"},
		{"token_php", "<?php\necho 'hi';", "php", "php"},
		{"token_javascript", "function hi() { console.log('hi'); }", "javascript", "js"},
		{"token_ruby", "def hi\n  puts 'hi'\nend", "ruby", "rb"},
		{"token_awk", "BEGIN { print \"hi\" }", "awk", "awk"},

		// === Fallback ===
		{"empty", "", "bash", "sh"},
		{"whitespace", "   \n\t  ", "bash", "sh"},
		{"random_text", "hello world this is not code", "bash", "sh"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotLang, gotExt := Guess(c.input)
			if gotLang != c.wantLang || gotExt != c.wantExt {
				t.Errorf("Guess() = (%q, %q), want (%q, %q)", gotLang, gotExt, c.wantLang, c.wantExt)
			}
		})
	}
}
