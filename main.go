package main

import (
	"github.com/sirupsen/logrus"
	"github.com/supreethrao/automated-rota-manager/cmd"
	"github.com/supreethrao/automated-rota-manager/pkg/localdb"
)

func main() {
	defer localdb.Close()
	err := cmd.Exec()
	if err != nil {
		logrus.Fatalf("error starting automated rota manager: %v", err)
	}
}
