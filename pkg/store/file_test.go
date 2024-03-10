package store

import (
	"context"
	"fmt"
	"github.com/dark-enstein/vault/pkg/vlog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	suite.Suite
	locs         []string
	tableConnect []struct {
		loc      string
		expected bool
	}
	tableStoreRetrieve map[string]string
	tokens             map[string]string
	log                *vlog.Logger
}

var (
	varTableConnect = []struct {
		loc      string
		expected bool
	}{
		{"test.db", true},
		{filepath.Join("false", "great.db"), true},
		{filepath.Join("false", "true.db"), true},
		{"test.yaml", true},
	}

	varTableStoreRetrieve = map[string]string{
		"ijbnijdelkfiue1": "A1B2C3D4E5F6G7H8",
		"ijbnijdelkfiue2": "Z9Y8X7W6V5U4T3S2",
		"ijbnijdelkfiue3": "Q1W2E3R4T5Y6U7I8",
		"ijbnijdelkfiue4": "O9P0A1S2D3F4G5H6",
		"ijbnijdelkfiue5": "J7K8L9Z0X1C2V3B4",
		"ijbnijdelkfiue6": "N5M6Q1W2E3R4T5Y",
		"ijbnijdelkfiue7": "U6I7O8P9A0S1D2F",
		"ijbnijdelkfiue8": "G3H4J5K6L7Z8X9C",
	}

	varTablePatch = map[string]string{
		"ijbnijdelkfiue1": "649sx8C30ubzd0cu",
		"ijbnijdelkfiue2": "TN4IFzbjuJfwuOIW",
		"ijbnijdelkfiue3": "a1otXUTJnt4gzOLL",
		"ijbnijdelkfiue4": "df8opIzQPrpRn9sM",
		"ijbnijdelkfiue5": "wADOGUHAJ5wtgiAO",
		"ijbnijdelkfiue6": "bvaeCnte1VAODI91",
		"ijbnijdelkfiue7": "8WNEeW7uVvYZpIrR",
		"ijbnijdelkfiue8": "Exz9baPAttTIgusZ",
	}

	Order = `
-------------------------------
Test Order: %d
-------------------------------
`
	TESTDIR = "./.store"
)

func (suite *FileTestSuite) SetupTest() {
	suite.tableConnect = varTableConnect
	suite.tableStoreRetrieve = varTableStoreRetrieve
	suite.log = vlog.New(true)
}

// Right now these tests are majorly happy-path tests

func (suite *FileTestSuite) TestConnect() {
	_ = suite.log.Logger()
	ctx := context.Background()
	for i := 0; i < len(suite.tableConnect); i++ {
		fmt.Printf(Order, i+1)
		loc := suite.tableConnect[i].loc
		file := NewFile(loc, suite.log)
		b, err := file.Connect(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Equalf(suite.tableConnect[i].expected, b, "expected %v, got %s\n", suite.tableConnect[i].expected, b)
		// just for appropriateness. the same principle applies to actual databases too. clean after use.
		b, err = file.Flush(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected true, but received false")
		_ = file.Close(ctx)
	}
}

func (suite *FileTestSuite) TestStoreAndRetrieve() {
	_ = suite.log.Logger()
	loc := "test_file.db"
	suite.locs = append(suite.locs, loc)
	ctx := context.Background()
	i := 1
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		file := NewFile(loc, suite.log)
		_, err := file.Connect(ctx)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		err = file.Store(ctx, k, v)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		val, err := file.Retrieve(ctx, k)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().Equalf(v, val, "expected %s, but got %s\n", v, val)
		// just for appropriateness. the same principle applies to actual databases too. clean after use.
		b, err := file.Flush(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected true, but received false")
		// just for appropriateness. the same principle applies to actual databases too. clean after use.
		b, err = file.Flush(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected true, but received false")
		_ = file.Close(ctx)
		i++
	}
}

func (suite *FileTestSuite) TestRetrieveAll() {
	_ = suite.log.Logger()
	loc := "test_file.db"
	suite.locs = append(suite.locs, loc)
	ctx := context.Background()
	i := 1
	file := NewFile(loc, suite.log)
	_, err := file.Connect(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = file.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		i++
	}
	valMap, err := file.RetrieveAll(ctx)
	suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Require().Equalf(len(suite.tableStoreRetrieve), len(valMap), "expected %d, but got %d\n", len(suite.tableStoreRetrieve), len(valMap))
	// just for appropriateness. the same principle applies to actual databases too. clean after use.
	b, err := file.Flush(ctx)
	suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
	suite.Require().True(b, "expected true, but received false")
	_ = file.Close(ctx)
}

func (suite *FileTestSuite) TestDelete() {
	_ = suite.log.Logger()
	loc := "test_file.db"
	suite.locs = append(suite.locs, loc)
	ctx := context.Background()
	i := 1
	file := NewFile(loc, suite.log)
	_, err := file.Connect(ctx)
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = file.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		// id should exist, and value should equal v
		val, err := file.Retrieve(ctx, k)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().Equalf(v, val, "expected %s, but got %v\n", v, val)
		b, err := file.Delete(ctx, k)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected %v, got %v\n", true, b)
		// id should exist, and value should equal v
		val, err = file.Retrieve(ctx, k)
		suite.Require().Contains(err.Error(), "doesn't exist", "expected id to not exist, but got this %v\n", err.Error())
		suite.Require().Equalf("", val, "expected %s, but got %v\n", v, val)
		// just for appropriateness. the same principle applies to actual databases too. clean after use.
		b, err = file.Flush(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected true, but received false")
		i++
	}
	_ = file.Close(ctx)
}

func (suite *FileTestSuite) TestPatch() {
	_ = suite.log.Logger()
	loc := "test_file.db"
	suite.locs = append(suite.locs, loc)
	ctx := context.Background()
	i := 1
	file := NewFile(loc, suite.log)
	_, err := file.Connect(ctx)
	for k, v := range suite.tableStoreRetrieve {
		fmt.Printf(Order, i)
		err = file.Store(ctx, k, v)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		// id should exist, and value should equal v
		val, err := file.Retrieve(ctx, k)
		suite.Assert().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Assert().Equalf(v, val, "expected %s, but got %v\n", v, val)
		b, err := file.Patch(ctx, k, varTablePatch[k])
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected %v, got %v\n", true, b)
		// retrieve and check that the value at ID matches the patch map, and not the original
		val, err = file.Retrieve(ctx, k)
		suite.Require().Equalf(varTablePatch[k], val, "expected %s, got %s %v\n", varTablePatch[k], val)
		suite.Require().NotEqualf(v, val, "%s remains the dame and wasn't patched\n", val)
		// just for appropriateness. the same principle applies to actual databases too. clean after use.
		b, err = file.Flush(ctx)
		suite.Require().NoErrorf(err, "expected no errors, but got this %v\n", err)
		suite.Require().True(b, "expected true, but received false")
		i++
	}
	_ = file.Close(ctx)
}

func (suite *FileTestSuite) TearDownTest() {
	_ = context.Background()
	log := suite.log.Logger()
	for i := 0; i < len(suite.locs); i++ {
		if err := os.RemoveAll(suite.locs[i]); err != nil {
			log.Info().Msgf("encountered error while removing test artifacts %s: %s\n", suite.locs[i], err.Error())
		}
	}
	for i := 0; i < len(suite.tableConnect); i++ {
		if err := os.RemoveAll(suite.tableConnect[i].loc); err != nil {
			log.Info().Msgf("encountered error while removing test artifacts %s: %s\n", suite.tableConnect[i].loc, err.Error())
		}
	}
}

// TestLoggerSuite tests that the Options values are correctly passed into
// the Logger implementation,
// and that the logging depends on the GlobalLogLevel set
func TestFileSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}
