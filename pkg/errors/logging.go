package errors
import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
}
type zapLogger struct {
	logger *zap.SugaredLogger
}
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debugw(msg, fieldsToArgs(fields)...)
}
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Infow(msg, fieldsToArgs(fields)...)
}
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warnw(msg, fieldsToArgs(fields)...)
}
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Errorw(msg, fieldsToArgs(fields)...)
}
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatalw(msg, fieldsToArgs(fields)...)
}
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fieldsToArgs(fields)...)}
}
func fieldsToArgs(fields []zapcore.Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, field := range fields {
		args = append(args, field.Key, field.Interface)
	}
	return args
}
type Field = zapcore.Field
type loggerKey struct{}
func GetLogger(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerKey{}).(Logger); ok {
		return l
	}
	return defaultLogger
}
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}
type ErrorResponse struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	RequestID  string      `json:"request_id,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	StackTrace string      `json:"stack_trace,omitempty"`
}
func LogError(ctx context.Context, err error) *ErrorResponse {
	logger := GetLogger(ctx)
	reqID := GetRequestID(ctx)
	stack := make([]byte, 4096)
	runtime.Stack(stack, false)
	errResp := &ErrorResponse{
		Code:       getErrorCode(err),
		Message:    err.Error(),
		RequestID:  reqID,
		Timestamp:  time.Now(),
		StackTrace: string(stack),
	}
	if detailed, ok := err.(interface{ Details() interface{} }); ok {
		errResp.Details = detailed.Details()
	}
	logger.Error("Error occurred",
		zap.String("error_code", errResp.Code),
		zap.String("message", errResp.Message),
		zap.String("request_id", reqID),
		zap.Any("details", errResp.Details),
		zap.String("stack_trace", errResp.StackTrace),
	)
	return errResp
}
var defaultLogger Logger
func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defaultLogger = &zapLogger{logger: logger.Sugar()}
}
type TemporaryError struct {
	Err error
}
func (e *TemporaryError) Error() string {
	return fmt.Sprintf("temporary error: %v", e.Err)
}
type ConnectionError struct {
	Err error
}
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %v", e.Err)
}
type TimeoutError struct {
	Err error
}
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout error: %v", e.Err)
}
type CircuitOpenError struct {
	message string
}
func (e *CircuitOpenError) Error() string {
	return e.message
}
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return "unknown"
}
func getErrorCode(err error) string {
	switch err.(type) {
	case *TemporaryError:
		return "TEMPORARY_ERROR"
	case *ConnectionError:
		return "CONNECTION_ERROR"
	case *TimeoutError:
		return "TIMEOUT_ERROR"
	case *CircuitOpenError:
		return "CIRCUIT_OPEN"
	default:
		return "INTERNAL_ERROR"
	}
}
func (e *ErrorResponse) JSON() string {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Sprintf(`{"code":"MARSHAL_ERROR","message":"Failed to marshal error: %v"}`, err)
	}
	return string(data)
}