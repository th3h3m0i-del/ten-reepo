# ś — clipboard runner

Đọc clipboard → đoán ngôn ngữ → lưu file vào `S/scripts/` → chạy
→ set report markdown trở lại clipboard. Có cờ `-g` để tự commit + push.
Có gitguard 2 lớp chặn secret trước khi push.

## Build

```bash
cd ~/ten-reepo/S
export GOCACHE="$HOME/ten-reepo/tmp/S/gocache"
export GOPATH="$HOME/ten-reepo/tmp/S/gopath"
export GOMODCACHE="$HOME/ten-reepo/tmp/S/gomodcache"
export GOTMPDIR="$HOME/ten-reepo/tmp/S/gotmp"
go build -o ś .
```

## Chạy

```bash
~/ten-reepo/S/ś          # chạy thường
~/ten-reepo/S/ś -g       # chạy + tự commit + push
```

Cờ -g chạy chuỗi git add -A → commit → pull --rebase → push
trong ~/ten-reepo. Commit message dùng giờ local ICT (UTC+7),
set global qua main.init(). Log git ghi vào tmp/S/git_sync_*.log.

## Ngôn ngữ hỗ trợ (33)

Detector dựa trên shebang trước, sau đó token đặc trưng, cuối cùng
fallback bash. Máy Termux hiện có interpreter/compiler cho:

### Chạy trực tiếp (interpreter)

- Python → python3 → .py
- JavaScript → node → .js
- TypeScript / Deno → deno run → .ts
- Ruby → ruby → .rb
- Perl → perl → .pl
- PHP → php → .php
- Lua → lua → .lua
- LuaJIT → luajit → .lua
- Bash → bash → .sh
- Zsh → zsh → .sh
- Sh → sh → .sh
- Dash → dash → .sh
- Fish → fish → .fish
- Java → java → .java
- R → Rscript → .R
- Julia → julia → .jl
- Tcl → tclsh → .tcl
- Groovy → groovy → .groovy
- Scala → scala → .scala
- Swift → swift → .swift
- Erlang → escript → .erl
- Bun (nếu không hỏng) → bun run → .js

### Compile rồi chạy (binary/jar/exe trong tmp/S/build/)

- C → gcc → binary → .c
- C++ → g++ → binary → .cpp
- Rust → rustc → binary → .rs
- Zig → zig → binary → .zig
- Kotlin → kotlinc → java -jar → .kt
- C# → csc → mono → .cs
- Crystal → crystal build → binary → .cr
- Nim → nim c → binary → .nim

### Đặc biệt (stdin)

- AWK → awk -f <file> → .awk
- SQL → sqlite3 :memory: (stdin) → .sql

### Lưu file nhưng không chạy

Ngôn ngữ không có interpreter/compiler trên máy: file vẫn được lưu
đúng đuôi, report ghi status: error, status_reason: unsupported_lang,
error_kind: missing_interp. Nội dung clipboard không bao giờ mất.

## Gitguard 2 lớp — chặn rò rỉ secret

ś tự push lên GitHub public, nên phải chặn secret (API key, token,
private key) trước khi push. Hai lớp:

### Lớp 1 — scan clipboard trước khi lưu

Nếu clipboard chứa secret (sk-, ghp_, AKIA, -----BEGIN PRIVATE KEY-----,
xoxb-, AIza, ya29., ...), file được lưu vào
tmp/S/blocked_scripts/ thay vì S/scripts/. Thư mục tmp/ đã
gitignore nên không bao giờ bị push. Report ghi
sync: blocked: secret in clipboard (...).

### Lớp 2 — scan toàn bộ S/scripts/ trước khi sync

Trước khi gọi gitsync, ś scan toàn bộ file trong S/scripts/
(kể cả file từ lần chạy trước). Nếu bất kỳ file nào có secret,
sync bị block, không add, không push. Report ghi
sync: blocked: secret in scripts/ (file="..." pattern="..." line=N).

Lớp 2 bắt được kịch bản: file secret cũ còn sót lại trong S/scripts/
từ lần trước, lần này clipboard sạch nhưng git add -A sẽ mang file
cũ lên GitHub. Lớp 2 chặn đúng chỗ đó. User phải tự xóa file secret
cũ thì sync mới đi tiếp.

### Tắt gitguard (không khuyến nghị)

```bash
WW_NO_GITGUARD=1 ~/ten-reepo/S/ś -g
```

## Config

ś đọc ~/ten-reepo/S/config.toml (nếu có) và override bằng env.
Không có file thì dùng default an toàn. Các key hỗ trợ:

- max_output — cắt stdout/stderr ở ngưỡng này (bytes). Default 10240.
- max_code_dump — dump code vào report nếu ≤ ngưỡng này. Default 4096.
- enable_git_guard — bật/tắt gitguard. Default true.
- secret_patterns — mảng string pattern secret cần chặn.

Env override: WW_MAX_OUTPUT, WW_MAX_CODE_DUMP, WW_NO_GITGUARD=1.

Ví dụ config.toml:

```toml
max_output = 20000
max_code_dump = 8192
enable_git_guard = true
secret_patterns = ["sk-", "ghp_", "AKIA", "xoxb-"]
```

## Trace profiling

Bật WW_TRACE=1 để đo thời gian runner.Run và report.Build.
Output ra stderr dạng [trace] runner.Run: 1.234s. Hữu ích debug
khi kotlin/swift chậm.

```bash
WW_TRACE=1 ~/ten-reepo/S/ś
```

## Atomic write

Mọi file ś ghi ra (S/scripts/clip_*, tmp/S/clipboard_backups/*,
report markdown) đều qua temp file + rename. Nếu crash giữa chừng,
file cũ nguyên vẹn, không bị truncate.

## Cơ chế retry khi push fail

Khi -g push thất bại (offline, credential, conflict), ś ghi flag
vào tmp/S/push_queue/<timestamp>_<reason>.fail. Lần chạy -g kế
tiếp sẽ thử push lại; thành công thì xóa hết flag, report ghi
sync: pushed (retried N).

## Error kind trong report

Report markdown có field error_kind (khi status = error), một trong:
missing_interp, syntax, runtime, secret_blocked, sync_failed,
io, unknown. AI đọc report biết ngay loại lỗi, không đoán.

## Output

- File script (sạch): S/scripts/clip_<timestamp>.<ext>
- File script (bị block vì secret): tmp/S/blocked_scripts/clip_<timestamp>.<ext>
- Clipboard gốc: tmp/S/clipboard_backups/clipboard_<timestamp>.txt
- Log git sync: tmp/S/git_sync_<timestamp>.log
- Binary build tạm: tmp/S/build/ (tự xóa sau khi chạy)
- Report markdown trong clipboard có các field: status, status_reason,
error_kind (nếu lỗi), exit, duration, lang, script_path,
stdout_total_lines, stderr_total_lines, error_hits,
output_truncated, sync (nếu dùng -g).

## Test

```bash
cd ~/ten-reepo/S
export GOCACHE="$HOME/ten-reepo/tmp/S/gocache"
export GOPATH="$HOME/ten-reepo/tmp/S/gopath"
export GOMODCACHE="$HOME/ten-reepo/tmp/S/gomodcache"
export GOTMPDIR="$HOME/ten-reepo/tmp/S/gotmp"
go test ./internal/detect/ ./internal/gitguard/ ./internal/config/ \
        ./internal/werr/ ./internal/atomicwrite/
```

Test coverage:
- detect: 52 case shebang + token + fallback cho 33 ngôn ngữ.
- gitguard: 8 case scan clean/OpenAI/GitHub/AWS/private key/multi-line.
- config: 5 case default/env override/parse.
- werr: 3 case New/Wrap/KindOf.
- atomicwrite: 3 case write/overwrite/no-temp-leftover.
