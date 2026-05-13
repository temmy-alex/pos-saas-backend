package config

type DatabaseConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	SSLMode  string
}

func LoadDatabaseConfig() DatabaseConfig {
	_ = LoadAppConfig()

	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		Database: getEnv("DB_DATABASE", "pos_saas_db"),
		Username: getEnv("DB_USERNAME", "pos_user"),
		Password: getEnv("DB_PASSWORD", "pos_password"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}
