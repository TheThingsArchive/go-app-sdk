// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/go-account-lib/account"
	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"github.com/TheThingsNetwork/go-utils/grpc/rpclog"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/mwitkow/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ClientVersion to use
var ClientVersion = "2.x.x"

// ClientConfig contains the configuration for the API client. Use the NewConfig() or NewCommunityConfig() functions to
// initialize your configuration, otherwise NewClient will panic.
type ClientConfig struct {
	initialized bool

	Logger log.Interface

	// The name of this client
	ClientName string

	// The version of this client (in the default config, this is the value of ttnsdk.ClientVersion)
	ClientVersion string

	// TLS Configuration only has to be set if connecting with servers that do not have trusted certificates.
	TLSConfig *tls.Config

	// Address of the Account Server (in the default config, this is https://account.thethingsnetwork.org)
	AccountServerAddress string

	// Client ID for the account server (if you registered your client)
	AccountServerClientID string

	// Client Secret for the account server (if you registered your client)
	AccountServerClientSecret string

	// Address of the Discovery Server (in the default config, this is discovery.thethings.network:1900)
	DiscoveryServerAddress string

	// Set this to true if the Discovery Server is insecure (not recommended)
	DiscoveryServerInsecure bool

	// Address of the Handler (optional)
	HandlerAddress string

	// Timeout for requests (in the default config, this is 10 seconds)
	RequestTimeout time.Duration

	appID        string
	appAccessKey string
}

// NewCommunityConfig creates a new configuration for the API client that is pre-configured for the Public Community Network.
func NewCommunityConfig(clientName string) ClientConfig {
	return NewConfig(clientName, "https://account.thethingsnetwork.org", "discovery.thethings.network:1900")
}

// NewConfig creates a new configuration for the API client.
func NewConfig(clientName, accountServerAddress, discoveryServerAddress string) ClientConfig {
	return ClientConfig{
		initialized:            true,
		Logger:                 log.Get(),
		ClientName:             clientName,
		ClientVersion:          ClientVersion,
		AccountServerAddress:   accountServerAddress,
		AccountServerClientID:  clientName,
		DiscoveryServerAddress: discoveryServerAddress,
		RequestTimeout:         10 * time.Second,
	}
}

// NewClient creates a new API client from the configuration, using the given Application ID and Application access key.
func (c ClientConfig) NewClient(appID, appAccessKey string) Client {
	c.appID = appID
	c.appAccessKey = appAccessKey
	return newClient(c)
}

// DialOptions to use when connecting to components
var DialOptions = []grpc.DialOption{
	grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		rpclog.UnaryClientInterceptor(nil),
	)),
	grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
		restartstream.Interceptor(restartstream.DefaultSettings),
		rpclog.StreamClientInterceptor(nil),
	)),
	grpc.WithBlock(),
}

func newClient(config ClientConfig) Client {
	if !config.initialized {
		panic("ttn-sdk: ClientConfig not initialized. Use ttnsdk.NewConfig or ttnsdk.NewCommunityConfig to generate your configuration")
	}
	client := &client{
		ClientConfig:         config,
		transportCredentials: credentials.NewTLS(config.TLSConfig),
	}
	if config.AccountServerAddress != "" {
		client.account = account.New(config.AccountServerAddress)
	}
	return client
}

// Client interface for The Things Network's API.
type Client interface {
	// Close the client and clean up all connections
	Close() error

	// Subscribe to uplink and events, publish downlink
	PubSub() (ApplicationPubSub, error)

	// Manage the application
	ManageApplication() (ApplicationManager, error)

	// Manage devices in the application
	ManageDevices() (DeviceManager, error)

	// Simulate uplink messages for a device (for testing)
	Simulate(devID string) (Simulator, error)
}

type client struct {
	ClientConfig
	transportCredentials credentials.TransportCredentials
	account              *account.Account
	discovery            struct {
		sync.RWMutex
		conn *grpc.ClientConn
	}
	handler struct {
		sync.RWMutex
		announcement *discovery.Announcement
		conn         *grpc.ClientConn
	}
	mqtt struct {
		sync.RWMutex
		client mqtt.Client
		ctx    context.Context
		cancel context.CancelFunc
	}
}

func (c *client) getContext(ctx context.Context) context.Context {
	ctx = ttnctx.OutgoingContextWithServiceInfo(ctx, c.ClientConfig.ClientName, c.ClientConfig.ClientVersion, "")
	if c.appAccessKey != "" {
		ctx = ttnctx.OutgoingContextWithKey(ctx, c.appAccessKey)
	}
	return ctx
}

func (c *client) Close() (closeErr error) {
	if err := c.closeHandler(); err != nil {
		closeErr = err
	}
	if err := c.closeDiscovery(); err != nil && closeErr == nil {
		closeErr = err
	}
	if err := c.closeMQTT(); err != nil && closeErr == nil {
		closeErr = err
	}
	return
}
