package store

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/internal/vlog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type MapTestSuite struct {
	suite.Suite
	tableStoreRetrieve map[string]string
	tableStorePatch    map[string]string
	tokens             map[string]string
	log                *vlog.Logger
}

var (
	varTableStoreMapRetrieve = map[string]string{
		"ijbnijdelkfiue1": "A1B2C3D4E5F6G7H8",
		"ijbnijdelkfiue2": "Z9Y8X7W6V5U4T3S2",
		"ijbnijdelkfiue3": "Q1W2E3R4T5Y6U7I8",
		"ijbnijdelkfiue4": "O9P0A1S2D3F4G5H6",
		"ijbnijdelkfiue5": "J7K8L9Z0X1C2V3B4",
		"ijbnijdelkfiue6": "N5M6Q1W2E3R4T5Y",
		"ijbnijdelkfiue7": "U6I7O8P9A0S1D2F",
		"ijbnijdelkfiue8": "G3H4J5K6L7Z8X9C",
	}

	varTableMapPatch = map[string]string{
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

func (suite *MapTestSuite) SetupTest() {
	suite.tableStoreRetrieve = varTableStoreMapRetrieve
	suite.tableStorePatch = varTableMapPatch
	suite.log = vlog.New(true)
}

func (suite *MapTestSuite) TestStoreAndRetrieve() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		syncMap := NewSyncMap(ctx, suite.log)
		b, err := syncMap.Connect(ctx)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().True(b, "expected true but got false")
		err = syncMap.Store(ctx, k, v)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		time.Sleep(2)
		val, err := syncMap.Retrieve(ctx, k)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().Equalf(v, val, "expected %s, but got %s\n", v, val)
		// clean DB
		suite.flush(ctx, syncMap)
		err = syncMap.Close(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		i++
	}
}

func (suite *MapTestSuite) TestRetrieveAll() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	syncMap := NewSyncMap(ctx, suite.log)
	b, err := syncMap.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true but got false")
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = syncMap.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		i++
	}
	valMap, err := syncMap.RetrieveAll(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Require().Equalf(len(suite.tableStoreRetrieve), len(valMap), "expected %d, but got %d\n", len(suite.tableStoreRetrieve), len(valMap))
	// clean DB
	suite.flush(ctx, syncMap)
	err = syncMap.Close(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
}

func (suite *MapTestSuite) TestDelete() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	syncMap := NewSyncMap(ctx, suite.log)
	b, err := syncMap.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true but got false")
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = syncMap.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		// id should exist, and value should equal v
		val, err := syncMap.Retrieve(ctx, k)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().Equalf(v, val, "expected %s, but got %v\n", v, val)
		b, err := syncMap.Delete(ctx, k)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected %v, got %v\n", true, b)
		// id should exist, and value should equal v
		val, err = syncMap.Retrieve(ctx, k)
		suite.Require().Error(err, "expected id to not exist, but got this %v\n", err.Error())
		suite.Require().Equalf("", val, "expected %s, but got %v\n", v, val)
		// ensure DB is flushed
		suite.flush(ctx, syncMap)
		i++
	}
	err = syncMap.Close(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
}

func (suite *MapTestSuite) TestPatch() {
	_ = suite.log.Logger()
	ctx := context.Background()
	i := 1
	syncMap := NewSyncMap(ctx, suite.log)
	b, err := syncMap.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true but got false")
	for k, v := range suite.tableStorePatch {
		fmt.Printf(Order, i)
		err = syncMap.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		// id should exist, and value should equal v
		val, err := syncMap.Retrieve(ctx, k)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().Equalf(v, val, "expected %s, but got %v\n", v, val)
		b, err := syncMap.Patch(ctx, k, v)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected %v, got %v\n", true, b)
		// id should exist, and value should equal v
		val, err = syncMap.Retrieve(ctx, k)
		suite.Require().NoErrorf(err, ErrWithOperation, "expected id to not exist, but got this %v\n", err)
		suite.Require().Equalf(v, val, "expected %s, but got %v\n", v, val)
		// ensure DB is flushed
		suite.flush(ctx, syncMap)
		i++
	}
	err = syncMap.Close(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
}

func (suite *MapTestSuite) TearDownTest() {}

// TestMapSuite tests the Map suite
func TestMapSuite(t *testing.T) {
	suite.Run(t, new(MapTestSuite))
}

func (suite *MapTestSuite) flush(ctx context.Context, syncMap *Map) {
	b, err := syncMap.Flush(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Assert().True(b, "expected true, got false")
}
