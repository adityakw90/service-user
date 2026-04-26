package config

import "github.com/spf13/viper"

type OauthConfig struct {
	Enabled bool              `mapstructure:"enabled"`
	Google  OauthGoogleConfig `mapstructure:"google"`
}

type OauthGoogleConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}

func defaultOauthConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".enabled", false)

	// google
	vConfig.SetDefault(key+".google.enabled", false)
	vConfig.SetDefault(key+".google.client_id", "")
	vConfig.SetDefault(key+".google.client_secret", "")
	// // Optional custom endpoints (empty by default)
	// vConfig.SetDefault(key+".google.redirect_uri", "")
	// vConfig.SetDefault(key+".google.auth_url", "")
	// vConfig.SetDefault(key+".google.token_url", "")
	// vConfig.SetDefault(key+".google.user_info_url", "")
}
