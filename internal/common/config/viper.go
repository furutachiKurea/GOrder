package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/spf13/viper"
)

var once sync.Once

func init() {
	if err := NewViperConfig(); err != nil {
		panic(err)
	}
}

func NewViperConfig() (err error) {
	once.Do(func() {
		err = newViperConfig()
	})
	return
}

func newViperConfig() error {
	relPath, err := getRelativePathFromCaller()
	if err != nil {
		return err
	}

	viper.SetConfigName("global")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(relPath)

	viper.AutomaticEnv()

	_ = viper.BindEnv("stripe-key", "STRIPE_KEY")
	_ = viper.BindEnv("endpoint-stripe-secret", "ENDPOINT_STRIPE_SECRET")

	return viper.ReadInConfig()
}

func getRelativePathFromCaller() (relPath string, err error) {
	callerPwd, err := os.Getwd()
	if err != nil {
		return
	}

	_, here, _, _ := runtime.Caller(0)
	relPath, err = filepath.Rel(callerPwd, filepath.Dir(here))
	fmt.Printf("caller from: %s, here: %s, relpath: %s\n", callerPwd, here, relPath)
	return
}
