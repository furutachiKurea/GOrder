package config

import (
	"github.com/spf13/viper"
)

func NewViperConfig() error {
	viper.SetConfigName("global")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../common/config")

	viper.AutomaticEnv()

	_ = viper.BindEnv("stripe-key", "STRIPE_KEY")
	_ = viper.BindEnv("endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET")

	return viper.ReadInConfig()
}
