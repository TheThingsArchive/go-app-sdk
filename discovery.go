// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk

import (
	"context"

	"github.com/TheThingsNetwork/api/discovery"
	"google.golang.org/grpc"
)

func (c *client) discover() (err error) {
	logger := c.Logger.WithField("Address", c.DiscoveryServerAddress)
	logger.Debug("ttn-sdk: Connecting to discovery...")
	if c.DiscoveryServerInsecure {
		c.discovery.conn, err = grpc.Dial(c.DiscoveryServerAddress, append(DialOptions, grpc.WithInsecure())...)
	} else {
		c.discovery.conn, err = grpc.Dial(c.DiscoveryServerAddress, append(DialOptions, grpc.WithTransportCredentials(c.transportCredentials))...)
	}
	if err != nil {
		logger.WithError(err).Debug("ttn-sdk: Could not connect to discovery")
		return err
	}
	logger.Debug("ttn-sdk: Connected to discovery")
	discoveryClient := discovery.NewDiscoveryClient(c.discovery.conn)
	ctx, cancel := context.WithTimeout(c.getContext(context.Background()), c.RequestTimeout)
	defer cancel()
	c.Logger.Debug("ttn-sdk: Finding handler...")
	handler, err := discoveryClient.GetByAppID(ctx, &discovery.GetByAppIDRequest{AppID: c.appID})
	if err != nil {
		c.Logger.WithError(err).Debug("ttn-sdk: Could not find handler for application")
		return err
	}
	c.handler.announcement = handler
	return nil
}

func (c *client) closeDiscovery() error {
	c.discovery.Lock()
	defer c.discovery.Unlock()
	if c.discovery.conn != nil {
		c.discovery.conn.Close()
	}
	c.discovery.conn = nil
	return nil
}
