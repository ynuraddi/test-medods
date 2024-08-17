package integrational

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres() (closer func() error, endpoint string, err error) {
	ctx := context.Background()

	dbName := "medods"
	dbUser := "user"
	dbPassword := "1234"

	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				ExposedPorts: []string{"5432/tcp"},
			},
		}),
	)
	if err != nil {
		return nil, "", err
	}

	endpoint, err = postgresContainer.Endpoint(context.Background(), "")
	if err != nil {
		return nil, "", err
	}

	return func() error {
		if err := postgresContainer.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
		return nil
	}, endpoint, nil
}

// docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
func setupMailHog() (closer func() error, smtpEndpoint, apiEndpoint string, err error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mailhog/mailhog",
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor:   wait.ForListeningPort("1025/tcp").WithStartupTimeout(15 * time.Second),
	}
	smtpC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", "", err
	}

	host, err := smtpC.Host(ctx)
	if err != nil {
		return nil, "", "", err
	}

	smtpPort, err := smtpC.MappedPort(ctx, nat.Port("1025"))
	if err != nil {
		return nil, "", "", err
	}

	apiPort, err := smtpC.MappedPort(ctx, nat.Port("8025"))
	if err != nil {
		return nil, "", "", err
	}

	return func() error {
			if err := smtpC.Terminate(ctx); err != nil {
				return err
			}
			return nil
		}, net.JoinHostPort(host, string(smtpPort)),
		net.JoinHostPort(host, apiPort.Port()), nil
}
