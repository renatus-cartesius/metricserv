package utils

import (
	"io"
	"reflect"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"go.uber.org/zap"
)

func SafeClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logger.Log.Error(
			"error on closing",
			zap.String("closer", reflect.ValueOf(closer).String()),
		)
	}
}
