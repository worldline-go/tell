package tell

import "github.com/rs/zerolog"

type Logger interface {
	Error(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
}

// adapterKV fit for msg, keyvalue interface, Ex: retryablehttp.
//
//	myLogFormat := log.With().Str("log_source", "mycomponent").Logger()
//	kvLogger := logz.adapterKV{Log: myLogFormat}
//	kvLogger.Error("this is message", "error", "failed x")
type adapterKV struct {
	Log zerolog.Logger
}

var _ Logger = adapterKV{}

func (l adapterKV) frameUp() zerolog.Logger {
	return l.Log.With().CallerWithSkipFrameCount(3).Logger()
}

func (l adapterKV) Error(msg string, keysAndValues ...interface{}) {
	f := l.frameUp()
	f.Error().Fields(keysAndValues).Msg(msg)
}

func (l adapterKV) Info(msg string, keysAndValues ...interface{}) {
	f := l.frameUp()
	f.Info().Fields(keysAndValues).Msg(msg)
}

func (l adapterKV) Debug(msg string, keysAndValues ...interface{}) {
	f := l.frameUp()
	f.Debug().Fields(keysAndValues).Msg(msg)
}

func (l adapterKV) Warn(msg string, keysAndValues ...interface{}) {
	f := l.frameUp()
	f.Warn().Fields(keysAndValues).Msg(msg)
}
