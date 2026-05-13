package config

import "strconv"

type JWTConfig struct {
	Secret                    string
	AccessTokenExpiresMinutes int
}

func LoadJWTConfig() JWTConfig {
	_ = LoadAppConfig()

	accessTokenExpiresMinutes, err := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_EXPIRES_MINUTES", "60"))
	if err != nil {
		accessTokenExpiresMinutes = 60
	}

	return JWTConfig{
		Secret:                    getEnv("JWT_SECRET", "pos_saas_local_secret_change_me"),
		AccessTokenExpiresMinutes: accessTokenExpiresMinutes,
	}
}
