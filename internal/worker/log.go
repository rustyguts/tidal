package worker

import "github.com/rs/zerolog/log"

// zerologAdapter implements asynq.Logger over zerolog.
type zerologAdapter struct{}

func (zerologAdapter) Debug(args ...any) { log.Debug().Msgf("%v", concat(args)) }
func (zerologAdapter) Info(args ...any)  { log.Info().Msgf("%v", concat(args)) }
func (zerologAdapter) Warn(args ...any)  { log.Warn().Msgf("%v", concat(args)) }
func (zerologAdapter) Error(args ...any) { log.Error().Msgf("%v", concat(args)) }
func (zerologAdapter) Fatal(args ...any) { log.Fatal().Msgf("%v", concat(args)) }

func concat(args []any) string {
	if len(args) == 0 {
		return ""
	}
	s := ""
	for i, a := range args {
		if i > 0 {
			s += " "
		}
		s += valueOf(a)
	}
	return s
}

func valueOf(a any) string {
	if s, ok := a.(string); ok {
		return s
	}
	return ""
}
