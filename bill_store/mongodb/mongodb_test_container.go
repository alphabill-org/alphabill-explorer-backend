package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type MongoTestContainer struct {
	Container testcontainers.Container
	URI       string
}

func StartMongoContainer(ctx context.Context) (*MongoTestContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:7.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start MongoDB container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}
	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	uri := fmt.Sprintf("mongodb://%s:%s", host, mappedPort.Port())

	return &MongoTestContainer{
		Container: container,
		URI:       uri,
	}, nil
}

// Stop stops and removes the MongoDB test container
func (m *MongoTestContainer) Stop(ctx context.Context) error {
	if err := m.Container.Terminate(ctx); err != nil {
		return fmt.Errorf("failed to stop MongoDB container: %w", err)
	}
	return nil
}
