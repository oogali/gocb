package gocb

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/couchbaselabs/gojcbmock"
	"github.com/stretchr/testify/suite"
)

const (
	defaultServerVersion = "5.1.0"
)

var globalBucket *Bucket
var globalCollection *Collection
var globalCluster *testCluster

type IntegrationTestSuite struct {
	suite.Suite
}

func (suite *IntegrationTestSuite) SetupSuite() {
	var err error
	var connStr string
	var mock *gojcbmock.Mock
	var auth PasswordAuthenticator
	if globalConfig.Server == "" {
		if globalConfig.Version != "" {
			panic("version cannot be specified with mock")
		}

		mpath, err := gojcbmock.GetMockPath()
		if err != nil {
			panic(err.Error())
		}

		mock, err = gojcbmock.NewMock(mpath, 4, 1, 64, []gojcbmock.BucketSpec{
			{Name: "default", Type: gojcbmock.BCouchbase},
		}...)

		mock.Control(gojcbmock.NewCommand(gojcbmock.CSetCCCP,
			map[string]interface{}{"enabled": "true"}))
		mock.Control(gojcbmock.NewCommand(gojcbmock.CSetSASLMechanisms,
			map[string]interface{}{"mechs": []string{"SCRAM-SHA512"}}))

		if err != nil {
			panic(err.Error())
		}

		globalConfig.Version = mock.Version()

		var addrs []string
		for _, mcport := range mock.MemcachedPorts() {
			addrs = append(addrs, fmt.Sprintf("127.0.0.1:%d", mcport))
		}
		connStr = fmt.Sprintf("couchbase://%s", strings.Join(addrs, ","))
		auth = PasswordAuthenticator{
			Username: "default",
			Password: "",
		}
	} else {
		connStr = globalConfig.Server

		auth = PasswordAuthenticator{
			Username: globalConfig.User,
			Password: globalConfig.Password,
		}

		if globalConfig.Version == "" {
			globalConfig.Version = defaultServerVersion
		}
	}

	cluster, err := Connect(connStr, ClusterOptions{Authenticator: auth})

	time.Sleep(1000)

	if err != nil {
		panic(err.Error())
	}

	nodeVersion, err := newNodeVersion(globalConfig.Version, mock != nil)
	if err != nil {
		panic(err.Error())
	}

	globalCluster = &testCluster{Cluster: cluster, Mock: mock, Version: nodeVersion}

	globalBucket = globalCluster.Bucket(globalConfig.Bucket)

	if globalConfig.Collection != "" {
		globalCollection = globalBucket.Collection(globalConfig.Collection)
	} else {
		globalCollection = globalBucket.DefaultCollection()
	}
}

type UnitTestSuite struct {
	suite.Suite
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		return
	}

	suite.Run(t, new(IntegrationTestSuite))
}

func TestUnit(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}