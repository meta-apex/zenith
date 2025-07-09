package zlog

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// defaultLogger is the global zlog.

var (
	defaultLogger = &ProductionLogger

	ProductionLogger = Logger{
		Level:      InfoLevel,
		Caller:     0,
		TimeField:  "",
		TimeFormat: "",
		Writer:     IOWriter{os.Stdout},
	}

	DevelopLogger = Logger{
		Level:      DebugLevel,
		Caller:     1,
		TimeField:  "",
		TimeFormat: "",
		Writer: &ConsoleWriter{
			ColorOutput:    true,
			QuoteString:    true,
			EndWithMessage: true,
			Writer:         IOWriter{os.Stdout},
		},
	}
)

// The WriterFunc type is an adapter to allow the use of
// ordinary functions as zlog writers. If f is a function
// with the appropriate signature, WriterFunc(f) is a
// [Writer] that calls f.
type WriterFunc func(*Entry) (int, error)

// WriteEntry calls f(e).
func (f WriterFunc) WriteEntry(e *Entry) (int, error) {
	return f(e)
}

// IOWriter wraps an io.Writer to Writer.
type IOWriter struct {
	io.Writer
}

// WriteEntry implements Writer.
func (w IOWriter) WriteEntry(e *Entry) (n int, err error) {
	return w.Writer.Write(e.buf)
}

// IOWriteCloser wraps an io.IOWriteCloser to Writer.
type IOWriteCloser struct {
	io.WriteCloser
}

// WriteEntry implements Writer.
func (w IOWriteCloser) WriteEntry(e *Entry) (n int, err error) {
	return w.WriteCloser.Write(e.buf)
}

// Close implements Writer.
func (w IOWriteCloser) Close() (err error) {
	return w.WriteCloser.Close()
}

// ObjectMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Entry's Object methods.
type ObjectMarshaler interface {
	MarshalObject(e *Entry)
}

// A Logger represents an active logging object that generates lines of JSON output to an io.Writer.
type Logger struct {
	// Level defines zlog levels.
	Level Level

	// Caller determines if adds the file:line of the "caller" key.
	// If Caller is negative, adds the full /path/to/file:line of the "caller" key.
	Caller int

	// TimeField defines the time field name in output.  It uses "time" in if empty.
	TimeField string

	// TimeFormat specifies the time format in output. It uses time.RFC3339 with milliseconds if empty.
	// Strongly recommended to leave TimeFormat empty for optimal built-in zlog formatting performance.
	// If set with `TimeFormatUnix`, `TimeFormatUnixMs`, times are formated as UNIX timestamp.
	TimeFormat string

	// TimeLocation specifics that the location which TimeFormat used. It uses time.Local if empty.
	TimeLocation *time.Location

	// Context specifies an optional context of zlog.
	Context Context

	// Writer specifies the writer of output. It uses a wrapped os.Stderr Writer in if empty.
	Writer Writer
}

// TimeFormatUnix defines a time format that makes time fields to be
// serialized as Unix timestamp integers.
const TimeFormatUnix = "\x01"

// TimeFormatUnixMs defines a time format that makes time fields to be
// serialized as Unix timestamp integers in milliseconds.
const TimeFormatUnixMs = "\x02"

// TimeFormatUnixWithMs defines a time format that makes time fields to be
// serialized as Unix timestamp floats.
const TimeFormatUnixWithMs = "\x03"

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func SetDefaultLogger(l *Logger) {
	defaultLogger = l
}

// Trace starts a new message with trace level.
func Trace() (e *Entry) {
	if defaultLogger.silent(TraceLevel) {
		return nil
	}
	e = defaultLogger.header(TraceLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Debug starts a new message with debug level.
func Debug() (e *Entry) {
	if defaultLogger.silent(DebugLevel) {
		return nil
	}
	e = defaultLogger.header(DebugLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Info starts a new message with info level.
func Info() (e *Entry) {
	if defaultLogger.silent(InfoLevel) {
		return nil
	}
	e = defaultLogger.header(InfoLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Warn starts a new message with warning level.
func Warn() (e *Entry) {
	if defaultLogger.silent(WarnLevel) {
		return nil
	}
	e = defaultLogger.header(WarnLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Error starts a new message with error level.
func Error() (e *Entry) {
	if defaultLogger.silent(ErrorLevel) {
		return nil
	}
	e = defaultLogger.header(ErrorLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Fatal starts a new message with fatal level.
func Fatal() (e *Entry) {
	if defaultLogger.silent(FatalLevel) {
		return nil
	}
	e = defaultLogger.header(FatalLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Panic starts a new message with panic level.
func Panic() (e *Entry) {
	if defaultLogger.silent(PanicLevel) {
		return nil
	}
	e = defaultLogger.header(PanicLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Stat starts a new message with panic level.
func Stat() (e *Entry) {
	if defaultLogger.silent(InfoLevel) {
		return nil
	}
	e = defaultLogger.header(InfoLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Printf sends a zlog entry without extra field. Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...any) {
	e := defaultLogger.header(noLevel)
	if caller, full := defaultLogger.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	e.Msgf(format, v...)
}

// Trace starts a new message with trace level.
func (l *Logger) Trace() (e *Entry) {
	if l.silent(TraceLevel) {
		return nil
	}
	e = l.header(TraceLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Debug starts a new message with debug level.
func (l *Logger) Debug() (e *Entry) {
	if l.silent(DebugLevel) {
		return nil
	}
	e = l.header(DebugLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Info starts a new message with info level.
func (l *Logger) Info() (e *Entry) {
	if l.silent(InfoLevel) {
		return nil
	}
	e = l.header(InfoLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Warn starts a new message with warning level.
func (l *Logger) Warn() (e *Entry) {
	if l.silent(WarnLevel) {
		return nil
	}
	e = l.header(WarnLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Error starts a new message with error level.
func (l *Logger) Error() (e *Entry) {
	if l.silent(ErrorLevel) {
		return nil
	}
	e = l.header(ErrorLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Fatal starts a new message with fatal level.
func (l *Logger) Fatal() (e *Entry) {
	if l.silent(FatalLevel) {
		return nil
	}
	e = l.header(FatalLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Panic starts a new message with panic level.
func (l *Logger) Panic() (e *Entry) {
	if l.silent(PanicLevel) {
		return nil
	}
	e = l.header(PanicLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Log starts a new message with no level.
func (l *Logger) Log() (e *Entry) {
	e = l.header(noLevel)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// WithLevel starts a new message with level.
func (l *Logger) WithLevel(level Level) (e *Entry) {
	if l.silent(level) {
		return nil
	}
	e = l.header(level)
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// Err starts a new message with error level with err as a field if not nil or with info level if err is nil.
func (l *Logger) Err(err error) (e *Entry) {
	var level = InfoLevel
	if err != nil {
		level = ErrorLevel
	}
	if l.silent(level) {
		return nil
	}
	e = l.header(level)
	if e == nil {
		return nil
	}
	if level == ErrorLevel {
		e = e.Err(err)
	}
	if caller, full := l.Caller, false; caller != 0 {
		if caller < 0 {
			caller, full = -caller, true
		}
		var pc uintptr
		e.caller(caller1(caller, &pc, 1, 1), pc, full)
	}
	return
}

// SetLevel changes zlog default level.
func (l *Logger) SetLevel(level Level) {
	atomic.StoreUint32((*uint32)(&l.Level), uint32(level))
}

// Printf sends a zlog entry without extra field. Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...any) {
	e := l.header(noLevel)
	if e != nil {
		if caller, full := l.Caller, false; caller != 0 {
			if caller < 0 {
				caller, full = -caller, true
			}
			var pc uintptr
			e.caller(caller1(caller, &pc, 1, 1), pc, full)
		}
	}
	e.Msgf(format, v...)
}

// WithName returns a new Logger instance with the specified name field added to the context.
func (l *Logger) WithName(name string) *Logger {
	newLogger := *l

	// Create a new context entry with the name field
	e := NewContext(nil)
	e.Str("name", name)
	nameContext := e.Value()

	// Combine existing context with name context
	if l.Context != nil {
		newContext := make([]byte, 0, len(l.Context)+len(nameContext))
		newContext = append(newContext, l.Context...)
		newContext = append(newContext, nameContext...)
		newLogger.Context = newContext
	} else {
		newLogger.Context = nameContext
	}

	return &newLogger
}

// WithValues returns a new Logger instance with the specified key-value pairs added to the context.
func (l *Logger) WithValues(keysAndValues ...any) *Logger {
	newLogger := *l

	// Create a new context entry with the key-value pairs
	e := NewContext(nil)
	e.KeysAndValues(keysAndValues...)
	valuesContext := e.Value()

	// Combine existing context with values context
	if l.Context != nil {
		newContext := make([]byte, 0, len(l.Context)+len(valuesContext))
		newContext = append(newContext, l.Context...)
		newContext = append(newContext, valuesContext...)
		newLogger.Context = newContext
	} else {
		newLogger.Context = valuesContext
	}

	return &newLogger
}

// WithCaller returns a new Logger instance with caller information enabled.
// The depth parameter specifies how many stack frames to skip (0 = current function, 1 = caller, etc.).
func (l *Logger) WithCaller(depth int) *Logger {
	newLogger := *l
	newLogger.Caller = depth + 1 // Add 1 to account for this wrapper function
	return &newLogger
}

var epool = sync.Pool{
	New: func() any {
		return &Entry{
			buf: make([]byte, 0, 1024),
		}
	},
}

const bbcap = 1 << 16

const smallsString = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

var timeNow = time.Now
var timeOffset, timeZone = func() (int64, string) {
	now := timeNow()
	_, n := now.Zone()
	s := now.Format("Z07:00")
	return int64(n), s
}()

func (l *Logger) header(level Level) *Entry {
	e := epool.Get().(*Entry)
	e.buf = e.buf[:0]
	e.Level = level
	if l.Writer != nil {
		e.w = l.Writer
	} else {
		e.w = IOWriter{os.Stderr}
	}
	// time
	if l.TimeField == "" {
		e.buf = append(e.buf, "{\"time\":"...)
	} else {
		e.buf = append(e.buf, '{', '"')
		e.buf = append(e.buf, l.TimeField...)
		e.buf = append(e.buf, '"', ':')
	}
	offset := timeOffset
	if l.TimeLocation != nil {
		if l.TimeLocation == time.UTC {
			offset = 0
		} else if l.TimeLocation == time.Local {
			offset = timeOffset
		} else {
			format := l.TimeFormat
			if format == "" {
				format = "2006-01-02T15:04:05.999Z07:00"
			}
			e.buf = append(e.buf, '"')
			e.buf = timeNow().In(l.TimeLocation).AppendFormat(e.buf, format)
			e.buf = append(e.buf, '"')
			goto headerlevel
		}
	}
	switch l.TimeFormat {
	case "":
		sec, nsec, _ := now()
		var tmp [32]byte
		var buf []byte
		if offset == 0 {
			// "2006-01-02T15:04:05.999Z"
			tmp[25] = '"'
			tmp[24] = 'Z'
			buf = tmp[:26]
		} else {
			// "2006-01-02T15:04:05.999Z07:00"
			tmp[30] = '"'
			tmp[29] = timeZone[5]
			tmp[28] = timeZone[4]
			tmp[27] = timeZone[3]
			tmp[26] = timeZone[2]
			tmp[25] = timeZone[1]
			tmp[24] = timeZone[0]
			buf = tmp[:31]
		}
		// date time
		sec += 9223372028715321600 + offset // unixToInternal + internalToAbsolute + timeOffset
		year, month, day, _ := absDate(uint64(sec), true)
		hour, minute, second := absClock(uint64(sec))
		// year
		a := year / 100 * 2
		b := year % 100 * 2
		tmp[0] = '"'
		tmp[1] = smallsString[a]
		tmp[2] = smallsString[a+1]
		tmp[3] = smallsString[b]
		tmp[4] = smallsString[b+1]
		// month
		month *= 2
		tmp[5] = '-'
		tmp[6] = smallsString[month]
		tmp[7] = smallsString[month+1]
		// day
		day *= 2
		tmp[8] = '-'
		tmp[9] = smallsString[day]
		tmp[10] = smallsString[day+1]
		// hour
		hour *= 2
		tmp[11] = 'T'
		tmp[12] = smallsString[hour]
		tmp[13] = smallsString[hour+1]
		// minute
		minute *= 2
		tmp[14] = ':'
		tmp[15] = smallsString[minute]
		tmp[16] = smallsString[minute+1]
		// second
		second *= 2
		tmp[17] = ':'
		tmp[18] = smallsString[second]
		tmp[19] = smallsString[second+1]
		// milli seconds
		a = int(nsec) / 1000000
		b = a % 100 * 2
		tmp[20] = '.'
		tmp[21] = byte('0' + a/100)
		tmp[22] = smallsString[b]
		tmp[23] = smallsString[b+1]
		// append to e.buf
		e.buf = append(e.buf, buf...)
	case TimeFormatUnix:
		sec, _, _ := now()
		// 1595759807
		var tmp [10]byte
		// seconds
		b := sec % 100 * 2
		sec /= 100
		tmp[9] = smallsString[b+1]
		tmp[8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[7] = smallsString[b+1]
		tmp[6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[5] = smallsString[b+1]
		tmp[4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[3] = smallsString[b+1]
		tmp[2] = smallsString[b]
		b = sec % 100 * 2
		tmp[1] = smallsString[b+1]
		tmp[0] = smallsString[b]
		// append to e.buf
		e.buf = append(e.buf, tmp[:]...)
	case TimeFormatUnixMs:
		sec, nsec, _ := now()
		// 1595759807105
		var tmp [13]byte
		// milli seconds
		a := int64(nsec) / 1000000
		b := a % 100 * 2
		tmp[12] = smallsString[b+1]
		tmp[11] = smallsString[b]
		tmp[10] = byte('0' + a/100)
		// seconds
		b = sec % 100 * 2
		sec /= 100
		tmp[9] = smallsString[b+1]
		tmp[8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[7] = smallsString[b+1]
		tmp[6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[5] = smallsString[b+1]
		tmp[4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[3] = smallsString[b+1]
		tmp[2] = smallsString[b]
		b = sec % 100 * 2
		tmp[1] = smallsString[b+1]
		tmp[0] = smallsString[b]
		// append to e.buf
		e.buf = append(e.buf, tmp[:]...)
	case TimeFormatUnixWithMs:
		sec, nsec, _ := now()
		// 1595759807.105
		var tmp [14]byte
		// milli seconds
		a := int64(nsec) / 1000000
		b := a % 100 * 2
		tmp[13] = smallsString[b+1]
		tmp[12] = smallsString[b]
		tmp[11] = byte('0' + a/100)
		tmp[10] = '.'
		// seconds
		b = sec % 100 * 2
		sec /= 100
		tmp[9] = smallsString[b+1]
		tmp[8] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[7] = smallsString[b+1]
		tmp[6] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[5] = smallsString[b+1]
		tmp[4] = smallsString[b]
		b = sec % 100 * 2
		sec /= 100
		tmp[3] = smallsString[b+1]
		tmp[2] = smallsString[b]
		b = sec % 100 * 2
		tmp[1] = smallsString[b+1]
		tmp[0] = smallsString[b]
		// append to e.buf
		e.buf = append(e.buf, tmp[:]...)
	default:
		e.buf = append(e.buf, '"')
		if l.TimeLocation == time.UTC {
			e.buf = timeNow().UTC().AppendFormat(e.buf, l.TimeFormat)
		} else {
			e.buf = timeNow().AppendFormat(e.buf, l.TimeFormat)
		}
		e.buf = append(e.buf, '"')
	}

headerlevel:
	// level
	switch level {
	case DebugLevel:
		e.buf = append(e.buf, ",\"level\":\"debug\""...)
	case InfoLevel:
		e.buf = append(e.buf, ",\"level\":\"info\""...)
	case WarnLevel:
		e.buf = append(e.buf, ",\"level\":\"warn\""...)
	case ErrorLevel:
		e.buf = append(e.buf, ",\"level\":\"error\""...)
	case TraceLevel:
		e.buf = append(e.buf, ",\"level\":\"trace\""...)
	case FatalLevel:
		e.buf = append(e.buf, ",\"level\":\"fatal\""...)
	case PanicLevel:
		e.buf = append(e.buf, ",\"level\":\"panic\""...)
	}

	// context
	if l.Context != nil {
		e.buf = append(e.buf, l.Context...)
	}
	return e
}
