// Package werr phân loại lỗi thành các Kind rõ ràng, để report của ś
// ghi được error_kind thay vì chỉ status_reason string. AI đọc report
// biết ngay lỗi thuộc loại nào: cú pháp, thiếu interpreter, runtime,
// secret bị chặn, hay lỗi hệ thống.
package werr

import "fmt"

// Kind là loại lỗi.
type Kind string

const (
	KindSyntax        Kind = "syntax"          // script sai cú pháp
	KindMissingInterp Kind = "missing_interp"  // thiếu interpreter/compiler
	KindRuntime       Kind = "runtime"         // chạy nhưng exit != 0
	KindSecretBlocked Kind = "secret_blocked"  // gitguard phát hiện secret
	KindSyncFailed    Kind = "sync_failed"     // git sync lỗi
	KindIO            Kind = "io"              // lỗi đọc/ghi file
	KindUnknown       Kind = "unknown"
)

// Error là lỗi có Kind + message.
type Error struct {
	Kind Kind
	Msg  string
}

// Error implement error interface.
func (e *Error) Error() string { return e.Msg }

// New tạo Error mới.
func New(kind Kind, format string, args ...interface{}) *Error {
	return &Error{Kind: kind, Msg: fmt.Sprintf(format, args...)}
}

// Wrap bọc một error thành Error với Kind cho trước.
func Wrap(kind Kind, err error, format string, args ...interface{}) *Error {
	msg := fmt.Sprintf(format, args...)
	if err != nil {
		msg = msg + ": " + err.Error()
	}
	return &Error{Kind: kind, Msg: msg}
}

// KindOf trả về Kind của err nếu là *Error, ngược lại KindUnknown.
func KindOf(err error) Kind {
	if e, ok := err.(*Error); ok {
		return e.Kind
	}
	return KindUnknown
}
