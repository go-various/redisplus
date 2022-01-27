package redisplus


// Logger describes the interface that must be implemeted by all loggers.
type Logger interface {

	//Trace Emit a message and key/value pairs at the TRACE level
	Trace(msg string, args ...interface{})

	//Debug Emit a message and key/value pairs at the DEBUG level
	Debug(msg string, args ...interface{})

	//Info Emit a message and key/value pairs at the INFO level
	Info(msg string, args ...interface{})

	//Warn Emit a message and key/value pairs at the WARN level
	Warn(msg string, args ...interface{})

	//Error Emit a message and key/value pairs at the ERROR level
	Error(msg string, args ...interface{})

	// IsTrace Indicate if TRACE logs would be emitted. This and the other Is* guards
	// are used to elide expensive logging code based on the current level.
	IsTrace() bool

	// IsDebug Indicate if DEBUG logs would be emitted. This and the other Is* guards
	IsDebug() bool

	// IsInfo Indicate if INFO logs would be emitted. This and the other Is* guards
	IsInfo() bool

	// IsWarn Indicate if WARN logs would be emitted. This and the other Is* guards
	IsWarn() bool

	// IsError Indicate if ERROR logs would be emitted. This and the other Is* guards
	IsError() bool

	// Name Returns the Name of the logger
	Name() string

}
