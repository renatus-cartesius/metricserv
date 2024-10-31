package utils

import (
	"os"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"go.uber.org/zap"
)

func CloseFile(file *os.File) {
	err := file.Close()
	if err != nil {
		logger.Log.Fatal(
			"error on closing file",
			zap.String("filepath", file.Name()),
		)
	}
}
