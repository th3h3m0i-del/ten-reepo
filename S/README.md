# ś — clipboard runner

Đọc clipboard → đoán ngôn ngữ → lưu file vào `scripts/` → chạy
→ set report markdown trở lại clipboard.

## Build

```bash
cd ~/ten-reepo/S
export GOCACHE="$HOME/ten-reepo/tmp/S/gocache"
export GOPATH="$HOME/ten-reepo/tmp/S/gopath"
export GOMODCACHE="$HOME/ten-reepo/tmp/S/gomodcache"
export GOTMPDIR="$HOME/ten-reepo/tmp/S/gotmp"
go build -o ś .
```

Dùng

Chạy thường (không sync):

```bash
~/ten-reepo/S/ś
```

Chạy kèm sync git (tự commit + push sau khi chạy):

```bash
~/ten-reepo/S/ś -g
```

Cờ -g chạy chuỗi git add -A → commit → pull --rebase → push
trong ~/ten-reepo. Log git ghi vào tmp/S/git_sync_*.log.
Report trong clipboard có thêm field sync với một trong các giá trị:
pushed, skipped: nothing to commit, failed: <lý do>.

Output

File script: S/scripts/clip_<timestamp>.<ext>.
Clipboard gốc: tmp/S/clipboard_backups/clipboard_<timestamp>.txt.
