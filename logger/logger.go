package logger

// configured log level for the application
var L LoggerInterface

// correctly initialise logging level from config
func InitLogger(logLevel int) {
	L = NewStdOutLogger(logLevel)
}
