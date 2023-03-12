package log

import (
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func (l Logger) Fields(fields map[string]any) Logger {
	return Logger{
		logger: l.logger.With().Fields(fields).Logger(),
	}
}

func (l Logger) Service(name string) Logger {
	return l.Field("service_name", fmt.Sprintf("%-8s", name))
}

func (l Logger) Field(key string, val any) Logger {
	zerologContext := l.logger.With()

	switch v := val.(type) {
	case string:
		zerologContext = zerologContext.Str(key, v)
	case []string:
		zerologContext = zerologContext.Strs(key, v)
	case []byte:
		zerologContext = zerologContext.Bytes(key, v)
	case error:
		zerologContext = zerologContext.AnErr(key, v)
	case []error:
		zerologContext = zerologContext.Errs(key, v)
	case bool:
		zerologContext = zerologContext.Bool(key, v)
	case []bool:
		zerologContext = zerologContext.Bools(key, v)
	case int8:
		zerologContext = zerologContext.Int8(key, v)
	case []int8:
		zerologContext = zerologContext.Ints8(key, v)
	case int16:
		zerologContext = zerologContext.Int16(key, v)
	case []int16:
		zerologContext = zerologContext.Ints16(key, v)
	case int32:
		zerologContext = zerologContext.Int32(key, v)
	case []int32:
		zerologContext = zerologContext.Ints32(key, v)
	case int64:
		zerologContext = zerologContext.Int64(key, v)
	case []int64:
		zerologContext = zerologContext.Ints64(key, v)
	case int:
		zerologContext = zerologContext.Int(key, v)
	case []int:
		zerologContext = zerologContext.Ints(key, v)
	case uint8:
		zerologContext = zerologContext.Uint8(key, v)
	case uint16:
		zerologContext = zerologContext.Uint16(key, v)
	case []uint16:
		zerologContext = zerologContext.Uints16(key, v)
	case uint32:
		zerologContext = zerologContext.Uint32(key, v)
	case []uint32:
		zerologContext = zerologContext.Uints32(key, v)
	case uint64:
		zerologContext = zerologContext.Uint64(key, v)
	case []uint64:
		zerologContext = zerologContext.Uints64(key, v)
	case uint:
		zerologContext = zerologContext.Uint(key, v)
	case []uint:
		zerologContext = zerologContext.Uints(key, v)
	case float32:
		zerologContext = zerologContext.Float32(key, v)
	case []float32:
		zerologContext = zerologContext.Floats32(key, v)
	case float64:
		zerologContext = zerologContext.Float64(key, v)
	case []float64:
		zerologContext = zerologContext.Floats64(key, v)
	case time.Time:
		zerologContext = zerologContext.Time(key, v)
	case []time.Time:
		zerologContext = zerologContext.Times(key, v)
	case time.Duration:
		zerologContext = zerologContext.Dur(key, v)
	case []time.Duration:
		zerologContext = zerologContext.Durs(key, v)
	case net.IP:
		zerologContext = zerologContext.IPAddr(key, v)
	case net.IPNet:
		zerologContext = zerologContext.IPPrefix(key, v)
	case net.HardwareAddr:
		zerologContext = zerologContext.MACAddr(key, v)
	case fmt.Stringer:
		zerologContext = zerologContext.Stringer(key, v)
	default:
		zerologContext = zerologContext.Interface(key, v)
	}

	return Logger{
		logger: zerologContext.Logger(),
	}
}

func (l Logger) Trace(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.TraceLevel {
		l.logger.Trace().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Debug(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.DebugLevel {
		l.logger.Debug().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Print(msgs ...any) {
	l.Debug(msgs...)
}

func (l Logger) Info(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.InfoLevel {
		l.logger.Info().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Warn(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.WarnLevel {
		l.logger.Warn().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Error(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.ErrorLevel {
		l.logger.Error().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Fatal(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.FatalLevel {
		l.logger.Fatal().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Panic(msgs ...any) {
	if l.logger.GetLevel() <= zerolog.PanicLevel {
		l.logger.Panic().Msg(fmt.Sprint(msgs...))
	}
}

func (l Logger) Tracef(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.TraceLevel {
		l.logger.Trace().Msgf(msg, v...)
	}
}

func (l Logger) Debugf(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.DebugLevel {
		l.logger.Debug().Msgf(msg, v...)
	}
}

func (l Logger) Printf(msg string, v ...any) {
	l.Debugf(msg, v...)
}

func (l Logger) Infof(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.InfoLevel {
		l.logger.Info().Msgf(msg, v...)
	}
}

func (l Logger) Warnf(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.WarnLevel {
		l.logger.Warn().Msgf(msg, v...)
	}
}

func (l Logger) Errorf(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.ErrorLevel {
		l.logger.Error().Msgf(msg, v...)
	}
}

func (l Logger) Fatalf(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.FatalLevel {
		l.logger.Fatal().Msgf(msg, v...)
	}
}

func (l Logger) Panicf(msg string, v ...any) {
	if l.logger.GetLevel() <= zerolog.PanicLevel {
		l.logger.Panic().Msgf(msg, v...)
	}
}
