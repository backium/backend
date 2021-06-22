package http

import "github.com/spf13/viper"

type Config struct {
	DBURI         string `mapstructure:"BACKIUM_DB_URI"`
	DBName        string `mapstructure:"BACKIUM_DB_NAME"`
	RedisURI      string `mapstructure:"BACKIUM_REDIS_URI"`
	RedisPassword string `mapstructure:"BACKIUM_REDIS_PASSWORD"`
	Port          string `mapstructure:"BACKIUM_APP_PORT"`
}

func LoadConfig(path string) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	config := Config{}
	if err := viper.ReadInConfig(); err != nil {
		return config, err
	}
	if err := viper.Unmarshal(&config); err != nil {
		return config, err
	}
	return config, nil
}
