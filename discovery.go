// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk

import (
	"context"

	"github.com/TheThingsNetwork/ttn/api/discovery"
	"google.golang.org/grpc"
)

func (c *client) discover() (err error) {
	var discoveryConn *grpc.ClientConn
	logger := c.Logger.WithField("Address", c.DiscoveryServerAddress)
	logger.Debug("ttn-sdk: Connecting to discovery...")
	if c.DiscoveryServerInsecure {
		discoveryConn, err = c.connPool.DialInsecure(c.DiscoveryServerAddress)
	} else {
		discoveryConn, err = c.connPool.DialSecure(c.DiscoveryServerAddress, c.transportCredentials)
	}
	if err != nil {
		logger.WithError(err).Debug("ttn-sdk: Could not connect to discovery")
		return err
	}
	logger.Debug("ttn-sdk: Connected to discovery")
	discoveryClient := discovery.NewDiscoveryClient(discoveryConn)
	ctx, cancel := context.WithTimeout(c.getContext(context.Background()), c.RequestTimeout)
	defer cancel()
	c.Logger.Debug("ttn-sdk: Finding handler...")
	handler, err := discoveryClient.GetByAppID(ctx, &discovery.GetByAppIDRequest{AppId: c.appID})
	if err != nil {
		c.Logger.WithError(err).Debug("ttn-sdk: Could not find handler for application")
		return err
	}
	c.handler.announcement = handler
	return nil
}
