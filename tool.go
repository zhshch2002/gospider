package gospider

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"os"
	"runtime"
)

var Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Stack().Logger()

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

func SprintStack() string {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

func MetaCopy(a map[string]interface{}) (b map[string]interface{}) {
	b = map[string]interface{}{}
	for key, value := range a {
		b[key] = value
	}
	return
}
