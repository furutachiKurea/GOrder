package discovery

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/furutachiKurea/gorder/common/discovery/consul"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func RegisterToConsul(ctx context.Context, serviceName string) (func() error, error) {
	registry, err := consul.New(viper.GetString("consul.addr"))
	if err != nil {
		return func() error { return nil }, err
	}

	instanceID := GenerateInstanceID(serviceName)
	grpcAddr := viper.Sub(serviceName).GetString("grpc-addr")
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		return func() error { return nil }, err
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Panic().Msgf("ho heart beat from %s to registry, err=%v", instanceID, err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	log.Info().
		Str("serviceName", serviceName).
		Str("addr", grpcAddr).
		Msg("registered to consul")

	return func() error {
		return registry.DeRegister(ctx, instanceID, serviceName)
	}, nil
}

func GetServiceAddr(ctx context.Context, serviceName string) (string, error) {
	registry, err := consul.New(viper.GetString("consul.addr"))
	if err != nil {
		return "", err
	}

	addrs, err := registry.Discover(ctx, serviceName)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("got empty %s addrs from registry", serviceName)
	}

	i := rand.Intn(len(addrs))
	log.Info().Msgf("Discoverd %d instance of %s, addrs=%v", len(addrs), serviceName, addrs)
	return addrs[i], nil
}
