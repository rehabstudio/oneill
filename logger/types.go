package logger

type LoggerInterface interface {
	Debug(string) error
	Info(string) error
	Warning(string) error
	Error(string) error
}
