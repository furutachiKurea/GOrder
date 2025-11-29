package discovery

import (
	"context"
	"time"

	"github.com/furutachiKurea/gorder/common/discovery/consul"
	"github.com/sirupsen/logrus"

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
				logrus.Panicf("ho heart beat from %s to registry, err=%v", instanceID, err)
			}

			time.Sleep(1 * time.Second)
		}
	}()

	logrus.WithFields(logrus.Fields{
		"serviceName": serviceName,
		"addr":        grpcAddr,
	}).Info("registered to consul")

	return func() error {
		return registry.DeRegister(ctx, instanceID, serviceName)
	}, nil
}
