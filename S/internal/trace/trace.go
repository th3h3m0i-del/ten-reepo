// Package trace cung cấp profiling nhẹ: đo thời gian một hàm chạy và
// in ra stderr nếu biến môi trường WW_TRACE=1. Dùng defer trace.Enter()()
// ở đầu mỗi hàm export quan trọng. Không dependency ngoài.
package trace

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// Enabled đọc env WW_TRACE một lần. Nếu false, Enter trả no-op.
var Enabled = os.Getenv("WW_TRACE") == "1"

// Enter đo thời gian từ lúc gọi đến lúc defer chạy. Trả về hàm defer.
// Dùng:
//   defer trace.Enter()()
// hoặc có tên:
//   defer trace.EnterNamed("pkg.Func")()
func Enter() func() {
	if !Enabled {
		return func() {}
	}
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	name := "unknown"
	if fn != nil {
		name = fn.Name()
	}
	return EnterNamed(name)
}

// EnterNamed đo thời gian và in ra stderr khi kết thúc.
func EnterNamed(name string) func() {
	if !Enabled {
		return func() {}
	}
	start := time.Now()
	return func() {
		fmt.Fprintf(os.Stderr, "[trace] %s: %s\n", name, time.Since(start))
	}
}
