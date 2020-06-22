package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/jhunt/go-ansi"
)

func main() {
	portStr := os.Getenv("PORT")
	if portStr == "" {
		bailWith("Must provide PORT envvar")
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		bailWith(err.Error())
	}

	rabbitURI := os.Getenv("RABBIT_URI")
	if rabbitURI == "" {
		bailWith("Must provide RABBIT_URI envvar")
	}

	rmqSkipVerify := false
	rmqSkipVerifyStr := strings.ToLower(os.Getenv("RMQ_SKIP_VERIFY"))
	if rmqSkipVerifyStr != "" && rmqSkipVerifyStr != "no" && rmqSkipVerifyStr != "false" {
		rmqSkipVerify = true
	}

	err = StartServer(&ServerConfig{
		Port:                     uint16(port),
		RabbitMQConnectionString: rabbitURI,
		RMQSkipVerify:            rmqSkipVerify,
	})
	if err != nil {
		bailWith(err.Error())
	}
}

func bailWith(f string, args ...interface{}) {
	ansi.Fprintf(os.Stderr, "@R{"+f+"}\n", args...)
	os.Exit(1)
}
