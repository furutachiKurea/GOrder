package logging

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	Method   = "method"
	Args     = "args"
	Cost     = "cost_ms"
	Response = "response"
	Error    = "err"
)

type ArgFormatter interface {
	FormatArg() (string, error)
}

func WhenMySQL(ctx context.Context, method string, args ...any) (zerolog.Logger, func(resp any, errP *error)) {
	l := log.With().Str(Method, method).Str(Args, formatArgs(args)).Logger()
	start := time.Now()

	return l, func(resp any, err *error) {
		l = l.With().
			Int(Cost, int(time.Since(start).Milliseconds())).
			Any(Response, resp).Logger()

		if err != nil && *err != nil {
			l.Error().
				Ctx(ctx).
				Str(Error, (*err).Error()).
				Msg("mysql_error")
		}

		l.Info().Ctx(ctx).Msg("mysql_success")
	}
}

func formatArgs(args []any) string {
	var items []string
	for _, arg := range args {
		items = append(items, formatArg(arg))
	}
	return strings.Join(items, "||")
}

func formatArg(arg any) string {
	var (
		str string
		err error
	)
	defer func() {
		if err != nil {
			str = "unsupported type in formatArg, err=" + err.Error()
		}
	}()

	switch v := arg.(type) {
	case ArgFormatter:
		str, err = v.FormatArg()
	default:
		str, err = marshalString(v)
	}
	return str
}

func marshalString(v any) (string, error) {
	bytes, err := json.Marshal(v)
	return string(bytes), err
}
