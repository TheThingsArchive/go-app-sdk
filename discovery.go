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
	logger := c.Logger.WithField("address", c.DiscoveryServerAddress)
	logger.Debug("ttn-sdk: connecting to discovery")
	if c.DiscoveryServerInsecure {
		discoveryConn, err = c.connPool.DialInsecure(c.DiscoveryServerAddress)
	} else {
		discoveryConn, err = c.connPool.DialSecure(c.DiscoveryServerAddress, c.transportCredentials)
	}
	if err != nil {
		logger.WithError(err).Debug("ttn-sdk: could not connect to discovery")
		return err
	}
	logger.Debug("ttn-sdk: connected to discovery")
	discoveryClient := discovery.NewDiscoveryClient(discoveryConn)
	ctx, cancel := context.WithTimeout(c.getContext(context.Background()), c.RequestTimeout)
	defer cancel()
	c.Logger.Debug("ttn-sdk: fetching handlers")
	handler, err := discoveryClient.GetByAppID(ctx, &discovery.GetByAppIDRequest{AppId: c.appID})
	if err != nil {
		c.Logger.WithError(err).Debug("ttn-sdk: could not find handler for application")
		return err
	}
	c.handler.announcement = handler
	return nil
}
