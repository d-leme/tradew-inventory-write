package cmd

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/d-leme/tradew-inventory-write/pkg/core"
	"github.com/d-leme/tradew-inventory-write/pkg/inventory"
	"github.com/d-leme/tradew-inventory-write/pkg/inventory/postgres"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

// Container contains all depencies from our api
type Container struct {
	Settings *core.Settings

	DBConnPool *pgxpool.Pool

	Authenticate *core.Authenticate

	Producer *core.MessageBrokerProducer
	SNS      *session.Session
	SQS      *session.Session

	InventoryRepository inventory.Repository
	InventoryService    inventory.Service
	InventoryController inventory.Controller
}

// NewContainer creates new instace of Container
func NewContainer(settings *core.Settings) *Container {

	container := new(Container)

	container.Settings = settings

	container.DBConnPool = connectPostgres(settings.Postgres)

	container.SQS = core.NewSession(
		settings.SQS.Region,
		settings.SQS.Endpoint,
		settings.SQS.Path,
		settings.SQS.Profile,
		settings.SQS.Fake,
	)

	container.SNS = core.NewSession(
		settings.SNS.Region,
		settings.SNS.Endpoint,
		settings.SNS.Path,
		settings.SNS.Profile,
		settings.SNS.Fake,
	)

	container.Producer = core.NewMessageBrokerProducer(container.SNS)

	container.Authenticate = core.NewAuthenticate(settings.JWT.Secret)

	container.InventoryRepository = postgres.NewRepository(container.DBConnPool)
	container.InventoryService = inventory.NewService(container.InventoryRepository)
	container.InventoryController = inventory.NewController(settings, container.Authenticate, container.InventoryService)

	return container
}

// Controllers maps all routes and exposes them
func (c *Container) Controllers() []core.Controller {
	return []core.Controller{
		&c.InventoryController,
	}
}

// Close terminates every opened resource
func (c *Container) Close() {
	c.DBConnPool.Close()
}

func connectPostgres(conf *core.PostgresConfig) *pgxpool.Pool {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", conf.User, conf.Password, conf.Host, conf.Port, conf.Database)

	pool, err := pgxpool.Connect(context.Background(), connString)

	if err != nil {
		logrus.
			WithError(err).
			Fatalf("unable to connect to database")
	}

	if err = pool.Ping(context.Background()); err != nil {
		logrus.
			WithError(err).
			Fatalf("unable to ping database")
	}

	logrus.Info("connected to postgres")

	return pool
}
