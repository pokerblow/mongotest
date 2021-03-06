package mongotest

import (
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

type ContainerCurator struct {
	pool *dockertest.Pool
	resource *dockertest.Resource
}

func StartMongoContainer(version string) *ContainerCurator {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mongo", version, nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	cc := &ContainerCurator{pool: pool, resource: resource}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(cc.GetMongoURI("local")))
		if err != nil {
			log.Fatalf("Couldn't create client: %s", err)
		}

		err = client.Ping(ctx, nil)
		fmt.Println(err)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource.Expire(360)
	return cc
}

// You can't defer this because os.Exit doesn't care for defer
func (cc *ContainerCurator) KillMongoContainer() {
	if err := cc.pool.Purge(cc.resource); err != nil {
		log.Fatalf("ERROR!!! Please kill the mongodb container manually: %s", err)
	}
}

func (cc *ContainerCurator) GetMongoURI(dbName string) string {
	return fmt.Sprintf("mongodb://%s:%d/%s", GetMongoHost(), cc.GetMongoPort(), dbName)
}

func (cc *ContainerCurator) GetMongoPort() int {
	port, err := strconv.Atoi(cc.resource.GetPort("27017/tcp"))
	if err != nil {
		log.Fatalf("Couldn't convert port: %s", err)
	}
	return port
}

func GetMongoHost() string {
	var endpoint string
	if os.Getenv("DOCKER_HOST") != "" {
		endpoint = os.Getenv("DOCKER_HOST")
	} else if os.Getenv("DOCKER_URL") != "" {
		endpoint = os.Getenv("DOCKER_URL")
	} else {
		return "localhost"
	}
	return extractHostname(endpoint)
}

func extractHostname(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal("Couldn't parse docker host", err)
	}
	return u.Hostname()
}
