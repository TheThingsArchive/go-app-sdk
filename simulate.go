// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import (
	"context"
	"time"

	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/go-utils/log"
)

// Simulator simulates messages for devices
type Simulator interface {
	Uplink(port uint8, payload []byte) error
}

type simulator struct {
	logger         log.Interface
	client         handler.ApplicationManagerClient
	getContext     func(context.Context) context.Context
	requestTimeout time.Duration

	appID string
	devID string
}

func (c *client) Simulate(devID string) (Simulator, error) {
	if err := c.connectHandler(); err != nil {
		return nil, err
	}
	return &simulator{
		logger:         c.Logger,
		client:         handler.NewApplicationManagerClient(c.handler.conn),
		getContext:     c.getContext,
		requestTimeout: c.RequestTimeout,
		appID:          c.appID,
		devID:          devID,
	}, nil
}

func (s *simulator) Uplink(port uint8, payload []byte) error {
	ctx, cancel := context.WithTimeout(s.getContext(context.Background()), s.requestTimeout)
	defer cancel()
	_, err := s.client.SimulateUplink(ctx, &handler.SimulatedUplinkMessage{
		AppID:   s.appID,
		DevID:   s.devID,
		Payload: payload,
		Port:    uint32(port),
	})
	return err
}
