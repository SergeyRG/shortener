package config

import (
	"flag"
	"fmt"
	"net"
	"net/url"
	"strconv"
)

func validatePortString(port string) error {

	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("incorrect port")
	}
	return nil

}

type netAddress struct {
	host string
	port string
}

func (na *netAddress) String() string {
	result := na.host
	if na.port != "" {
		result = result + ":" + na.port
	}
	return result
}

func (na *netAddress) Set(flagValue string) error {
	host, port, err := net.SplitHostPort(flagValue)
	if err != nil {
		return fmt.Errorf("flag -a value must be in form host:port, "+
			"received: %s", flagValue)
	}

	if err := validatePortString(port); err != nil {
		return err
	}

	na.host = host
	na.port = port
	return nil
}

type urlAddress struct {
	scheme string
	host   string
}

func (ua *urlAddress) String() string {
	return ua.scheme + `://` + ua.host
}

func (ua *urlAddress) Set(flagValue string) error {
	uParsed, err := url.ParseRequestURI(flagValue)
	if err != nil {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", flagValue)
	}

	if uParsed.Scheme != "http" && uParsed.Scheme != "https" {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", flagValue)
	}

	if uParsed.User != nil {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", flagValue)
	}

	if uParsed.Path != "" {
		return fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", flagValue)
	}

	if err := validatePortString(uParsed.Port()); err != nil {
		return err
	}

	ua.scheme = uParsed.Scheme
	ua.host = uParsed.Host
	return nil
}

type Config struct {
	ServerAddress       string
	BaseShortURLAddress string
}

func NewConfig() Config {
	ServerAddress := netAddress{
		host: "",
		port: "8080",
	}
	BaseShortURLAddress := urlAddress{
		scheme: "http",
		host:   "localhost:8080",
	}

	parseFlags(&ServerAddress, &BaseShortURLAddress)

	return Config{
		ServerAddress:       ServerAddress.String(),
		BaseShortURLAddress: BaseShortURLAddress.String(),
	}
}

func parseFlags(sa *netAddress, ua *urlAddress) {
	flag.Var(sa, "a", "address and port to run server")
	flag.Var(ua, "b", "base URL for short URLs")

	flag.Parse()
}
