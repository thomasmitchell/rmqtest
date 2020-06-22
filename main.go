package main

import (
	"encoding/json"
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
		rabbitURI = parseVcapServices()
		if rabbitURI == "" {
			bailWith("Must provide RABBIT_URI or properly formatted VCAP_SERVICES envvar")
		}
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

type services struct {
	PRabbitMQ []struct {
		Credentials struct {
			Protocols struct {
				AMQPS struct {
					URI string `json:"uri"`
				} `json:"amqp+ssl"`
			} `json:"protocols"`
		} `json:"credentials"`
	} `json:"p-rabbitmq"`
}

func (s *services) URI() string {
	if len(s.PRabbitMQ) == 0 {
		return ""
	}

	return s.PRabbitMQ[0].Credentials.Protocols.AMQPS.URI
}

func parseVcapServices() string {
	servicesJSON := os.Getenv("VCAP_SERVICES")
	services := services{}
	err := json.Unmarshal([]byte(servicesJSON), &services)
	if err != nil {
		return ""
	}

	return services.URI()
}
