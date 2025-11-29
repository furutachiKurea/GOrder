package consul

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

var (
	once         sync.Once
	consulClient *api.Client
	initErr      error
)

type Registry struct {
	client *api.Client
}

func New(consulAddr string) (*Registry, error) {
	once.Do(func() {
		config := api.DefaultConfig()
		config.Address = consulAddr
		client, err := api.NewClient(config)
		if err != nil {
			initErr = err
			return
		}

		consulClient = client
	})

	if initErr != nil {
		return nil, initErr
	}

	return &Registry{client: consulClient}, nil
}

func (r *Registry) Register(ctx context.Context, instanceID, serviceName, hostPort string) error {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return errors.New("invalid host:port format")
	}

	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	return r.client.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      instanceID,
		Address: host,
		Port:    port,
		Name:    serviceName,
		Check: &api.AgentServiceCheck{
			CheckID:                        instanceID,
			TLSSkipVerify:                  true,
			TTL:                            "5s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "10s",
		},
	})
}

func (r *Registry) DeRegister(ctx context.Context, instanceID, serviceName string) error {
	logrus.WithFields(logrus.Fields{
		"instance_id":  instanceID,
		"service_name": serviceName,
	}).Info("deregister from consul")
	defer func() {

	}()
	return r.client.Agent().CheckDeregister(instanceID)
}

func (r *Registry) Discover(ctx context.Context, serviceName string) (ips []string, err error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		ips = append(ips, fmt.Sprintf("%s:%d", e.Service.Address, e.Service.Port))
	}

	return ips, nil
}

func (r *Registry) HealthCheck(instanceID, _ string) error {
	return r.client.Agent().UpdateTTL(instanceID, "online", api.HealthPassing)
}
