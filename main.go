package main

import (
	"github.com/sirupsen/logrus"
	"github.com/supreethrao/automated-rota-manager/cmd"
)

func main() {
	err := cmd.Exec()
	if err != nil {
		logrus.Fatalf("error starting automated rota manager: %v", err)
	}
}
