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

func WhenMySQL(_ context.Context, method string, args ...any) (zerolog.Logger, func(resp any, errP *error)) {
	fields := make(map[string]any)
	fields[Method] = method
	fields[Args] = formatMySQLArgs(args)
	start := time.Now()

	return log.Logger.With().Fields(fields).Logger(), func(resp any, err *error) {
		ffields := make(map[string]any)
		for k, v := range fields {
			ffields[k] = v
		}
		ffields[Cost] = time.Since(start).Milliseconds()
		ffields[Response] = resp

		event := log.Info()
		msg := "mysql_success"

		if err != nil && *err != nil {
			msg = "mysql_error"
			ffields[Error] = (*err).Error()
			event = log.Error()
		}

		event.Fields(ffields).Msg(msg)
	}
}

func formatMySQLArgs(args []any) string {
	var items []string
	for _, arg := range args {
		items = append(items, formatMySQLArg(arg))
	}
	return strings.Join(items, "||")
}

func formatMySQLArg(arg any) string {
	switch v := arg.(type) {
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return "unsupported type in formatMySQLArg, err=" + err.Error()
		}

		return string(bytes)
	}
}
