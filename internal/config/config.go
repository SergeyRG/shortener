package config

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

func validatePortString(port string) error {

	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("incorrect port")
	}
	return nil

}

func validateServerAddress(val string) error {
	_, port, err := net.SplitHostPort(val)
	if err != nil {
		return fmt.Errorf("flag -a value must be in form host:port, "+
			"received: %s", val)
	}

	if err := validatePortString(port); err != nil {
		return errors.New("port value is specified outside the acceptable range")
	}
	return nil
}

func validateBaseURL(val string) error {
	uParsed, err := url.ParseRequestURI(val)
	if err != nil {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", val)
	}

	if uParsed.Scheme != "http" && uParsed.Scheme != "https" {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", val)
	}

	if uParsed.User != nil {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", val)
	}

	if uParsed.Path != "" {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", val)
	}

	if err := validatePortString(uParsed.Port()); err != nil {
		return err
	}

	return nil
}

type Config struct {
	ServerAddress       string
	BaseShortURLAddress string
}

func NewConfig() (Config, error) {
	ServerAddress := flag.String(
		"a", ":8080", "address and port to run server")
	BaseShortURLAddress := flag.String(
		"b", "http://localhost:8080", "base URL for short URLs")

	flag.Parse()

	if val, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		*ServerAddress = val
	}
	if val, exist := os.LookupEnv("BASE_URL"); exist {
		*BaseShortURLAddress = val
	}

	if err := validateServerAddress(*ServerAddress); err != nil {
		return Config{}, err
	}

	if err := validateBaseURL(*BaseShortURLAddress); err != nil {
		return Config{}, err
	}

	return Config{
			ServerAddress:       *ServerAddress,
			BaseShortURLAddress: *BaseShortURLAddress,
		},
		nil
}
