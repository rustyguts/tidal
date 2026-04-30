package logging

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Options struct {
	Level   string
	Pretty  bool
	Service string
}

func Setup(opts Options) zerolog.Logger {
	level, err := zerolog.ParseLevel(strings.ToLower(opts.Level))
	if err != nil || opts.Level == "" {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var l zerolog.Logger
	if opts.Pretty {
		l = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Timestamp().Logger()
	} else {
		l = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}
	if opts.Service != "" {
		l = l.With().Str("svc", opts.Service).Logger()
	}
	log.Logger = l
	return l
}
