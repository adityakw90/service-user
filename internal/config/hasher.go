package config

import "github.com/spf13/viper"

type HasherConfig struct {
	Type        string `mapstructure:"type"`        // used by bcrypt, argon2, sha256
	Cost        int    `mapstructure:"cost"`        // used by bcrypt
	Salt        string `mapstructure:"salt"`        // used by sha256
	Memory      int    `mapstructure:"memory"`      // used by argon2
	Iterations  int    `mapstructure:"iterations"`  // used by argon2
	Parallelism int    `mapstructure:"parallelism"` // used by argon2
	SaltLength  int    `mapstructure:"salt_length"` // used by argon2
	KeyLength   int    `mapstructure:"key_length"`  // used by argon2
}

func defaultHasherConfig(prefix string, v *viper.Viper) {
	v.SetDefault(prefix+".type", "argon2")
	v.SetDefault(prefix+".cost", 10)
	v.SetDefault(prefix+".salt", "")
	v.SetDefault(prefix+".memory", 65536)
	v.SetDefault(prefix+".iterations", 2)
	v.SetDefault(prefix+".parallelism", 1)
	v.SetDefault(prefix+".salt_length", 16)
	v.SetDefault(prefix+".key_length", 32)
}
