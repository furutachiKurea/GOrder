package logging

import (
	"os"
	"strconv"
	"time"

	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO 持久化日志, 日志流转

func Init() {
	zerolog.ErrorFieldName = "err"

	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_ENV")); isLocal {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.TimeOnly,
		})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	registerHook(traceHook{})
}

func registerHook(hooks ...zerolog.Hook) {
	for _, hook := range hooks {
		log.Logger = log.Logger.Hook(hook)
	}
}

type traceHook struct {
}

func (t traceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if e.GetCtx() != nil && tracing.TraceID(e.GetCtx()) != "" {
		e.Str("trace", tracing.TraceID(e.GetCtx()))
	}
}
