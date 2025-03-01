package network

import (
	"github.com/openziti/fabric/controller/command"
	"github.com/openziti/fabric/controller/db"
	"github.com/openziti/fabric/controller/models"
	"github.com/openziti/fabric/controller/xt"
	"github.com/openziti/fabric/event"
	"github.com/openziti/fabric/logcontext"
	"github.com/openziti/foundation/v2/versions"
	"github.com/openziti/identity"
	"github.com/openziti/metrics"
	"github.com/openziti/storage/boltz"
	"github.com/openziti/transport/v2/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
	"time"
)

type testConfig struct {
	ctx             *db.TestContext
	options         *Options
	metricsRegistry metrics.Registry
	versionProvider versions.VersionProvider
	closeNotify     chan struct{}
}

func newTestConfig(ctx *db.TestContext) *testConfig {
	options := DefaultOptions()
	options.MinRouterCost = 0

	closeNotify := make(chan struct{})
	return &testConfig{
		ctx:             ctx,
		options:         options,
		metricsRegistry: metrics.NewRegistry("test", nil),
		versionProvider: NewVersionProviderTest(),
		closeNotify:     closeNotify,
	}
}

func (self *testConfig) GetEventDispatcher() event.Dispatcher {
	return event.DispatcherMock{}
}

func (self *testConfig) GetId() *identity.TokenId {
	return &identity.TokenId{Token: "test"}
}

func (self *testConfig) GetMetricsRegistry() metrics.Registry {
	return self.metricsRegistry
}

func (self *testConfig) GetOptions() *Options {
	return self.options
}

func (self *testConfig) GetCommandDispatcher() command.Dispatcher {
	return nil
}

func (self *testConfig) GetDb() boltz.Db {
	return self.ctx.GetDb()
}

func (self *testConfig) GetVersionProvider() versions.VersionProvider {
	return self.versionProvider
}

func (self *testConfig) GetCloseNotify() <-chan struct{} {
	return self.closeNotify
}

func TestNetwork_parseServiceAndIdentity(t *testing.T) {
	req := require.New(t)
	instanceId, serviceId := parseInstanceIdAndService("hello")
	req.Equal("", instanceId)
	req.Equal("hello", serviceId)

	instanceId, serviceId = parseInstanceIdAndService("@hello")
	req.Equal("", instanceId)
	req.Equal("hello", serviceId)

	instanceId, serviceId = parseInstanceIdAndService("a@hello")
	req.Equal("a", instanceId)
	req.Equal("hello", serviceId)

	instanceId, serviceId = parseInstanceIdAndService("bar@hello")
	req.Equal("bar", instanceId)
	req.Equal("hello", serviceId)

	instanceId, serviceId = parseInstanceIdAndService("@@hello")
	req.Equal("", instanceId)
	req.Equal("@hello", serviceId)

	instanceId, serviceId = parseInstanceIdAndService("a@@hello")
	req.Equal("a", instanceId)
	req.Equal("@hello", serviceId)

	instanceId, serviceId = parseInstanceIdAndService("a@foo@hello")
	req.Equal("a", instanceId)
	req.Equal("foo@hello", serviceId)
}

func TestCreateCircuit(t *testing.T) {
	ctx := db.NewTestContext(t)
	defer ctx.Cleanup()

	config := newTestConfig(ctx)
	defer close(config.closeNotify)

	network, err := NewNetwork(config)
	assert.Nil(t, err)

	addr := "tcp:0.0.0.0:0"
	transportAddr, err := tcp.AddressParser{}.Parse(addr)
	assert.Nil(t, err)

	r0 := newRouterForTest("r0", "", transportAddr, nil, 0, false)

	svc := &Service{
		BaseEntity:         models.BaseEntity{Id: "svc"},
		Name:               "svc",
		TerminatorStrategy: "smartrouting",
	}

	/*
		Terminators: []*Terminator{
		{
		},
	*/
	lc := logcontext.NewContext()
	_, _, _, cerr := network.selectPath(r0, svc, "", lc)
	assert.Error(t, cerr)
	assert.Equal(t, CircuitFailureNoTerminators, cerr.Cause())

	svc.Terminators = []*Terminator{
		{
			BaseEntity: models.BaseEntity{Id: "t0"},
			Service:    "svc",
			Router:     "r0",
			Binding:    "transport",
			Address:    "tcp:localhost:1001",
			InstanceId: "",
			Precedence: xt.Precedences.Default,
		},
	}

	_, _, _, cerr = network.selectPath(r0, svc, "", lc)
	assert.Error(t, cerr)
	assert.Equal(t, CircuitFailureNoOnlineTerminators, cerr.Cause())

	network.Routers.markConnected(r0)
	_, _, _, cerr = network.selectPath(r0, svc, "", lc)
	assert.NoError(t, cerr)

	_, _, _, cerr = network.selectPath(r0, svc, "test", lc)
	assert.Error(t, cerr)
	assert.Equal(t, CircuitFailureNoTerminators, cerr.Cause())
}

type VersionProviderTest struct {
}

func (v VersionProviderTest) Branch() string {
	return "local"
}

func (v VersionProviderTest) EncoderDecoder() versions.VersionEncDec {
	return &versions.StdVersionEncDec
}

func (v VersionProviderTest) Version() string {
	return "v0.0.0"
}

func (v VersionProviderTest) BuildDate() string {
	return time.Now().String()
}

func (v VersionProviderTest) Revision() string {
	return ""
}

func (v VersionProviderTest) AsVersionInfo() *versions.VersionInfo {
	return &versions.VersionInfo{
		Version:   v.Version(),
		Revision:  v.Revision(),
		BuildDate: v.BuildDate(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func NewVersionProviderTest() versions.VersionProvider {
	return &VersionProviderTest{}
}
