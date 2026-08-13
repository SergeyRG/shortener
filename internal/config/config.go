package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
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

func validateURL(val string) error {
	uParsed, err := url.ParseRequestURI(val)
	if err != nil {
		return fmt.Errorf("incorrect URL")
	}

	if uParsed.Scheme != "http" && uParsed.Scheme != "https" {
		return fmt.Errorf("incorrect URL")
	}

	if uParsed.User != nil {
		return fmt.Errorf("incorrect URL")
	}

	if uParsed.Path != "" {
		return fmt.Errorf("incorrect URL")
	}

	if err := validatePortString(uParsed.Port()); err != nil {
		return err
	}

	return nil
}

// generate:reset
type Config struct {
	ServerAddress       string `json:"server_address"`
	GRPCServerAddress   string `json:"grpc_server_address"`
	EnableGRPCTLS       bool   `json:"-"`
	BaseShortURLAddress string `json:"base_url"`
	FileStoragePath     string `json:"file_storage_path"`
	DBDSN               string `json:"database_dsn"`
	SecretKey           string `json:"-"`
	AuditFilePath       string `json:"-"`
	AuditURL            string `json:"-"`
	EnableHTTPS         bool   `json:"enable_https"`
	TrustedSubnet       string `json:"trusted_subnet"`
}

func NewConfig() (Config, error) {
	binPath, err := os.Executable()
	if err != nil {
		return Config{}, err
	}
	defKey := string("DEFAULT_SECRET_KEY")
	SecretKey := &defKey

	binDir := filepath.Dir(binPath)

	ServerAddress := flag.String("a", ":8080", "address and port to run server")
	GRPCServerAddress := flag.String("g", ":50051", "address and port to run grpc server")
	EnableGRPCTLS := flag.Bool("gtls", false, "enable TLS for grpc server")
	BaseShortURLAddress := flag.String("b", "http://localhost:8080", "base URL for short URLs")
	FileStoragePath := flag.String("f", binDir+"/file_storage.NDJSON", "base URL for short URLs")
	DBDSN := flag.String("d", "", "DSN to connect to the database.")
	AuditFilePath := flag.String("audit-file", "", "path to request audit file")
	AuditURL := flag.String("audit-url", "", "URL for request audit")
	EnableHTTPS := flag.Bool("s", false, "enable TLS")
	configPath := flag.String("c", "", "path to config file")
	flag.StringVar(configPath, "config", "", "path to config file")
	trustedSubnet := flag.String("t", "", "CIDR доверенной сети")

	flag.Parse()

	if val, exist := os.LookupEnv("CONFIG"); exist {
		*configPath = val
	}

	if *configPath != "" {
		cfg := Config{}
		fileCFG, err := os.ReadFile(*configPath)
		if err != nil {
			return Config{}, fmt.Errorf("ошибка чтения файла конфигурации: %v", err)
		}

		err = json.Unmarshal(fileCFG, &cfg)
		if err != nil {
			return Config{}, fmt.Errorf("ошибка разбора файла конфигурации: %v", err)
		}

		userFlags := make(map[string]bool)
		flag.Visit(func(f *flag.Flag) {
			userFlags[f.Name] = true
		})

		if !userFlags["a"] && cfg.ServerAddress != "" {
			*ServerAddress = cfg.ServerAddress
		}
		if !userFlags["b"] && cfg.BaseShortURLAddress != "" {
			*BaseShortURLAddress = cfg.BaseShortURLAddress
		}
		if !userFlags["f"] && cfg.FileStoragePath != "" {
			*FileStoragePath = cfg.FileStoragePath
		}
		if !userFlags["d"] && cfg.DBDSN != "" {
			*DBDSN = cfg.DBDSN
		}
		if !userFlags["s"] {
			*EnableHTTPS = cfg.EnableHTTPS
		}
		if !userFlags["t"] && cfg.TrustedSubnet != "" {
			*trustedSubnet = cfg.TrustedSubnet
		}
		if !userFlags["g"] && cfg.GRPCServerAddress != "" {
			*GRPCServerAddress = cfg.GRPCServerAddress
		}

	}

	if val, exist := os.LookupEnv("SERVER_ADDRESS"); exist {
		*ServerAddress = val
	}
	if val, exist := os.LookupEnv("BASE_URL"); exist {
		*BaseShortURLAddress = val
	}
	if val, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		*FileStoragePath = val
	}
	if val, exist := os.LookupEnv("DATABASE_DSN"); exist {
		*DBDSN = val
	}
	if val, exist := os.LookupEnv("SECRET_KEY"); exist {
		*SecretKey = val
	} else {
		log.Printf("Используется ключ по умолчанию")
	}
	if val, exist := os.LookupEnv("AUDIT_FILE"); exist {
		*AuditFilePath = val
	}
	if val, exist := os.LookupEnv("AUDIT_URL"); exist {
		*AuditURL = val
	}
	if _, exist := os.LookupEnv("ENABLE_HTTPS"); exist {
		*EnableHTTPS = true
	}
	if val, exist := os.LookupEnv("TRUSTED_SUBNET"); exist {
		*trustedSubnet = val
	}
	if val, exist := os.LookupEnv("GRPC_SERVER_ADDRESS"); exist {
		*GRPCServerAddress = val
	}

	if err := validateServerAddress(*ServerAddress); err != nil {
		return Config{}, err
	}

	if err := validateURL(*BaseShortURLAddress); err != nil {
		return Config{}, fmt.Errorf("flag -b value must be in form http[s]://host:port, "+
			"received: %s", *BaseShortURLAddress)
	}

	if *AuditURL != "" {
		if err := validateURL(*AuditURL); err != nil {
			return Config{}, fmt.Errorf("flag --audit-url value must be in form http[s]://host:port, "+
				"received: %s", *AuditURL)
		}
	}

	return Config{
			ServerAddress:       *ServerAddress,
			GRPCServerAddress:   *GRPCServerAddress,
			EnableGRPCTLS:       *EnableGRPCTLS,
			BaseShortURLAddress: *BaseShortURLAddress,
			FileStoragePath:     *FileStoragePath,
			DBDSN:               *DBDSN,
			SecretKey:           *SecretKey,
			AuditFilePath:       *AuditFilePath,
			AuditURL:            *AuditURL,
			EnableHTTPS:         *EnableHTTPS,
			TrustedSubnet:       *trustedSubnet,
		},
		nil
}
