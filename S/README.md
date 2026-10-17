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
