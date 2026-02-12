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
	RedirectURI  string `mapstructure:"redirect_uri"`
	// Optional custom endpoints (for testing)
	AuthURL     string `mapstructure:"auth_url"`
	TokenURL    string `mapstructure:"token_url"`
	UserInfoURL string `mapstructure:"user_info_url"`
}

func defaultOauthConfig(key string, vConfig *viper.Viper) {
	vConfig.SetDefault(key+".enabled", false)

	// google
	vConfig.SetDefault(key+".google.enabled", false)
	vConfig.SetDefault(key+".google.client_id", "")
	vConfig.SetDefault(key+".google.client_secret", "")
	vConfig.SetDefault(key+".google.redirect_uri", "")
	// Optional custom endpoints (empty by default)
	vConfig.SetDefault(key+".google.auth_url", "")
	vConfig.SetDefault(key+".google.token_url", "")
	vConfig.SetDefault(key+".google.user_info_url", "")
}
