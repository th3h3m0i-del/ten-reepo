# ś — clipboard runner

Đọc clipboard → đoán ngôn ngữ → lưu file vào `S/scripts/` → chạy
→ set report markdown trở lại clipboard. Có cờ `-g` để tự commit + push.

## Build

```bash
cd ~/ten-reepo/S
export GOCACHE="$HOME/ten-reepo/tmp/S/gocache"
export GOPATH="$HOME/ten-reepo/tmp/S/gopath"
export GOMODCACHE="$HOME/ten-reepo/tmp/S/gomodcache"
export GOTMPDIR="$HOME/ten-reepo/tmp/S/gotmp"
go build -o ś .
```

Chạy

Chạy thường (không sync):

```bash
~/ten-reepo/S/ś
```

Chạy kèm sync git (tự commit + push sau khi chạy):

```bash
~/ten-reepo/S/ś -g
```

Cờ -g chạy chuỗi git add -A → commit → pull --rebase → push
trong ~/ten-reepo. Commit message dùng giờ local (ICT, UTC+7).
Log git ghi vào tmp/S/git_sync_*.log.

Ngôn ngữ hỗ trợ (33)

Máy Termux hiện có interpreter/compiler cho các ngôn ngữ sau.
Detector dựa trên shebang trước, sau đó token đặc trưng, cuối cùng
fallback bash.

Chạy trực tiếp (interpreter)

Ngôn ngữ Interpreter Đuôi
Python python3 .py
JavaScript node .js
TypeScript / Deno deno run .ts
Ruby ruby .rb
Perl perl .pl
PHP php .php
Lua lua .lua
LuaJIT luajit .lua
Bash bash .sh
Zsh zsh .sh
Sh sh .sh
Dash dash .sh
Fish fish .fish
Java java .java
R Rscript .R
Julia julia .jl
Tcl tclsh .tcl
Groovy groovy .groovy
Scala scala .scala
Swift swift .swift
Erlang escript .erl
Bun (nếu không hỏng) bun run .js

Compile rồi chạy (binary/jar/exe trong tmp/S/build/)

Ngôn ngữ Compiler Runtime Đuôi
C gcc binary .c
C++ g++ binary .cpp
Rust rustc binary .rs
Zig zig binary .zig
Kotlin kotlinc java -jar .kt
C# csc mono .cs
Crystal crystal build binary .cr
Nim nim c binary .nim

Đặc biệt (stdin)

Ngôn ngữ Cách chạy Đuôi
AWK awk -f <file> .awk
SQL sqlite3 :memory: (stdin) .sql

Lưu file nhưng không chạy

Ngôn ngữ không có interpreter/compiler trên máy: file vẫn được lưu
đúng đuôi, report ghi status: error, status_reason: unsupported_lang.
Nội dung clipboard không bao giờ mất.

Cơ chế retry khi push fail

Khi -g push thất bại (offline, credential, conflict), ś ghi flag
vào tmp/S/push_queue/<timestamp>_<reason>.fail. Lần chạy -g kế
tiếp sẽ thử push lại; thành công thì xoá hết flag, report ghi
sync: pushed (retried N). Flag sống trong tmp/ nên không đẩy
lên git.

Output

· File script: S/scripts/clip_<timestamp>.<ext>
· Clipboard gốc (trước khi ś ghi đè): tmp/S/clipboard_backups/clipboard_<timestamp>.txt
· Log git sync: tmp/S/git_sync_<timestamp>.log
· Binary build tạm: tmp/S/build/ (tự xoá sau khi chạy)
· Report markdown trong clipboard có các field:
  status, status_reason, exit, duration, lang, script_path,
  stdout_total_lines, stderr_total_lines, error_hits,
  output_truncated, và sync (nếu dùng -g).

Test

```bash
cd ~/ten-reepo/S
export GOCACHE="$HOME/ten-reepo/tmp/S/gocache"
export GOPATH="$HOME/ten-reepo/tmp/S/gopath"
export GOMODCACHE="$HOME/ten-reepo/tmp/S/gomodcache"
export GOTMPDIR="$HOME/ten-reepo/tmp/S/gotmp"
go test ./internal/detect/
```

52 test case cover shebang + token + fallback cho toàn bộ ngôn ngữ.
