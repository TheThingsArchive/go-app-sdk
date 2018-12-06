// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheThingsNetwork/api/handler"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	testlog "github.com/TheThingsNetwork/go-utils/log/test"
	. "github.com/smartystreets/assertions"
)

func TestApplicationManager(t *testing.T) {
	a := New(t)

	log := testlog.NewLogger()
	ttnlog.Set(log)
	defer log.Print(t)

	mock := new(mockApplicationManagerClient)

	manager := &applicationManager{
		logger:         log,
		client:         mock,
		getContext:     func(ctx context.Context) context.Context { return ctx },
		requestTimeout: time.Second,
		appID:          "test",
	}

	someErr := errors.New("some error")

	{
		mock.reset()
		mock.err = someErr
		_, err := manager.GetPayloadFormat()
		a.So(err, ShouldNotBeNil)

		mock.reset()
		mock.application = &handler.Application{PayloadFormat: "custom"}
		pf, err := manager.GetPayloadFormat()
		a.So(err, ShouldBeNil)
		a.So(pf, ShouldEqual, "custom")

		err = manager.SetPayloadFormat("other")
		a.So(err, ShouldBeNil)
		a.So(mock.application.PayloadFormat, ShouldEqual, "other")

		mock.err = someErr
		err = manager.SetPayloadFormat("")
		a.So(err, ShouldNotBeNil)
	}

	{
		mock.reset()
		mock.err = someErr
		_, _, _, _, err := manager.GetCustomPayloadFunctions()
		a.So(err, ShouldNotBeNil)

		mock.reset()
		mock.application = &handler.Application{PayloadFormat: "other"}
		_, _, _, _, err = manager.GetCustomPayloadFunctions()
		a.So(err, ShouldNotBeNil)

		mock.application = &handler.Application{PayloadFormat: "custom", Decoder: "decoder", Converter: "converter", Validator: "validator", Encoder: "encoder"}
		jsDecoder, jsConverter, jsValidator, jsEncoder, err := manager.GetCustomPayloadFunctions()
		a.So(err, ShouldBeNil)
		a.So(jsDecoder, ShouldEqual, "decoder")
		a.So(jsConverter, ShouldEqual, "converter")
		a.So(jsValidator, ShouldEqual, "validator")
		a.So(jsEncoder, ShouldEqual, "encoder")

		err = manager.SetCustomPayloadFunctions("newdecoder", "newconverter", "newvalidator", "newencoder")
		a.So(err, ShouldBeNil)
		a.So(mock.application.PayloadFormat, ShouldEqual, "custom")
		a.So(mock.application.Decoder, ShouldEqual, "newdecoder")
		a.So(mock.application.Converter, ShouldEqual, "newconverter")
		a.So(mock.application.Validator, ShouldEqual, "newvalidator")
		a.So(mock.application.Encoder, ShouldEqual, "newencoder")

		mock.err = someErr
		err = manager.SetCustomPayloadFunctions("", "", "", "")
		a.So(err, ShouldNotBeNil)
	}

	// TODO: TestCustomUplinkPayloadFunctions(jsDecoder, jsConverter, jsValidator string, payload []byte, port uint8) (*handler.DryUplinkResult, error) {
	// TODO: TestCustomDownlinkPayloadFunctions(jsEncoder string, fields map[string]interface{}, port uint8) (*handler.DryDownlinkResult, error) {
}
