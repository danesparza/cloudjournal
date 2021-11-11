package main

import (
	"github.com/danesparza/cloudjournal/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	cmd.Execute()
}
