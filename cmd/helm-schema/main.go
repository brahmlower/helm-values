package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	err := GenerateCommand(logger).Execute()
	if err != nil {
		os.Exit(1)
	}
}
