package logger

// Logger is a simple interface describing services that emit messages to gather
// certain runtime information.
type Logger interface {
	// Decorate takes a sequence of alternating key/value pairs which are used to
	// decorate the log message.
	Decorate(v ...interface{})
	// Log takes a sequence of alternating key/value pairs which are used to
	// create the log message structure.
	Log(v ...interface{}) error
}
