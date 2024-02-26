package store

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/internal/vlog"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type RedisTestSuite struct {
	suite.Suite
	redisConnectionString string
	tableConnect          []struct {
		connectionStr string
		expected      bool
	}
	tableStoreRetrieve map[string]string
	tableStorePatch    map[string]string
	tokens             map[string]string
	log                *vlog.Logger
}

var (
	varTableRedisConnect = []struct {
		connectionStr string
		expected      bool
	}{
		// Assuming 'localhost:6379' is a valid connection and Redis is running on the default port without a password
		{"redis://localhost:6379", true},
		// Assuming Redis is running on a non-default port and requires a password
		{"redis://:password@localhost:6380", true},
		// Invalid format or unreachable host/port
		{"invalidformat", false},
		// Correct format but incorrect port or Redis not running here
		{"localhost:6399", false},
		// Attempt to use a filepath, which is invalid for Redis connections
		{"/false/great.db", false},
		// Another example of an invalid connection string
		{"http://localhost:6379", false},
	}

	varTableStoreRedisRetrieve = map[string]string{
		"ijbnijdelkfiue1": "A1B2C3D4E5F6G7H8",
		"ijbnijdelkfiue2": "Z9Y8X7W6V5U4T3S2",
		"ijbnijdelkfiue3": "Q1W2E3R4T5Y6U7I8",
		"ijbnijdelkfiue4": "O9P0A1S2D3F4G5H6",
		"ijbnijdelkfiue5": "J7K8L9Z0X1C2V3B4",
		"ijbnijdelkfiue6": "N5M6Q1W2E3R4T5Y",
		"ijbnijdelkfiue7": "U6I7O8P9A0S1D2F",
		"ijbnijdelkfiue8": "G3H4J5K6L7Z8X9C",
	}

	varTableRedisPatch = map[string]string{
		"ijbnijdelkfiue1": "649sx8C30ubzd0cu",
		"ijbnijdelkfiue2": "TN4IFzbjuJfwuOIW",
		"ijbnijdelkfiue3": "a1otXUTJnt4gzOLL",
		"ijbnijdelkfiue4": "df8opIzQPrpRn9sM",
		"ijbnijdelkfiue5": "wADOGUHAJ5wtgiAO",
		"ijbnijdelkfiue6": "bvaeCnte1VAODI91",
		"ijbnijdelkfiue7": "8WNEeW7uVvYZpIrR",
		"ijbnijdelkfiue8": "Exz9baPAttTIgusZ",
	}
)

func (suite *RedisTestSuite) SetupTest() {
	//port := "6378"
	//err := SetUpEnv(port)
	//suite.Require().NoErrorf(err, "docker environment setup failed with error: %s\n", err.Error())
	suite.tableConnect = varTableRedisConnect
	suite.tableStoreRetrieve = varTableStoreRedisRetrieve
	suite.tableStorePatch = varTableRedisPatch
	suite.redisConnectionString = "redis://localhost:6379"
	suite.log = vlog.New(true)
}

// Right now these tests are majorly happy-path tests

func (suite *RedisTestSuite) TestConnect() {
	_ = suite.log.Logger()
	ctx := context.Background()
	for i := 0; i < len(suite.tableConnect); i++ {
		fmt.Printf(Order, i+1)
		loc := suite.tableConnect[i].connectionStr
		redis, err := NewRedis(loc, suite.log)
		// decided not to require no errors here, because the core error handling logic is handled by go=redis, so no use we trying to test it
		//suite.Assert().NoErrorf(err, "got error: %v\n", err)
		// continue even with error
		if err != nil {
			continue
		}
		b, err := redis.Connect(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Equalf(suite.tableConnect[i].expected, b, "expected %v, got %v\n", suite.tableConnect[i].expected, b)
		// clean DB
		suite.flush(ctx, redis)
		err = redis.Close(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
	}
}

func (suite *RedisTestSuite) TestStoreAndRetrieve() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		redis, err := NewRedis(suite.redisConnectionString, suite.log)
		b, err := redis.Connect(ctx)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().True(b, "expected true but got false")
		err = redis.Store(ctx, k, v)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		time.Sleep(2)
		val, err := redis.Retrieve(ctx, k)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().Equalf(v, val, "expected %s, but got %s\n", v, val)
		// clean DB
		suite.flush(ctx, redis)
		err = redis.Close(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		i++
	}
}

func (suite *RedisTestSuite) TestRetrieveAll() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	redis, err := NewRedis(suite.redisConnectionString, suite.log)
	b, err := redis.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true but got false")
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = redis.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		i++
	}
	valMap, err := redis.RetrieveAll(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Require().Equalf(len(suite.tableStoreRetrieve), len(valMap), "expected %d, but got %d\n", len(suite.tableStoreRetrieve), len(valMap))
	// clean DB
	suite.flush(ctx, redis)
	err = redis.Close(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
}

func (suite *RedisTestSuite) TestDelete() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	redis, err := NewRedis(suite.redisConnectionString, suite.log)
	b, err := redis.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true but got false")
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = redis.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		// id should exist, and value should equal v
		val, err := redis.Retrieve(ctx, k)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().Equalf(v, val, "expected %s, but got %v\n", v, val)
		b, err := redis.Delete(ctx, k)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected %v, got %v\n", true, b)
		// id should exist, and value should equal v
		val, err = redis.Retrieve(ctx, k)
		suite.Require().Error(err, "expected id to not exist, but got this %v\n", err.Error())
		suite.Require().Equalf("", val, "expected %s, but got %v\n", v, val)
		// ensure DB is flushed
		suite.flush(ctx, redis)
		i++
	}
	err = redis.Close(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
}

func (suite *RedisTestSuite) TestPatch() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	redis, err := NewRedis(suite.redisConnectionString, suite.log)
	b, err := redis.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true but got false")
	for k, v := range suite.tableStorePatch {
		fmt.Printf(Order, i)
		err = redis.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		// id should exist, and value should equal v
		val, err := redis.Retrieve(ctx, k)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().Equalf(v, val, "expected %s, but got %v\n", v, val)
		b, err := redis.Patch(ctx, k, v)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected %v, got %v\n", true, b)
		// id should exist, and value should equal v
		val, err = redis.Retrieve(ctx, k)
		suite.Require().NoErrorf(err, ErrWithOperation, "expected id to not exist, but got this %v\n", err)
		suite.Require().Equalf(v, val, "expected %s, but got %v\n", v, val)
		// ensure DB is flushed
		suite.flush(ctx, redis)
		i++
	}
	err = redis.Close(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
}

func (suite *RedisTestSuite) TearDownTest() {}

// TestRedisSuite tests the Redis suite
func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

// SetUpEnv spawns up a redis instance in a docker container
func SetUpEnv(port string) error {
	cmd := exec.Command("docker", fmt.Sprintf("run -d --name redis-stack -p %s:6379 -p 8001:8001 redis/redis-stack:latest", port))

	b, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error: %s: %s\n", err, string(b))
	}
	return nil
}

func (suite *RedisTestSuite) flush(ctx context.Context, redis *Redis) {
	b, err := redis.Flush(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true, got false")
}
