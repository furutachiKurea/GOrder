package logging

import (
	"os"
	"strconv"
	"time"

	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

	log.Logger = log.Hook(traceHook{})
}

type traceHook struct {
}

func (t traceHook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	if e.GetCtx() != nil && tracing.TraceID(e.GetCtx()) != "" {
		e.Str("trace", tracing.TraceID(e.GetCtx()))
	}
}
