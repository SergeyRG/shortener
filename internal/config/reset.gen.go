package config

func (s *Config) Reset(){
	s.ServerAddress = ""
	s.BaseShortURLAddress = ""
	s.FileStoragePath = ""
	s.DBDSN = ""
	s.SecretKey = ""
	s.AuditFilePath = ""
	s.AuditURL = ""
}
