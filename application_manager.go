// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/go-utils/log"
)

// ApplicationManager manages an application.
type ApplicationManager interface {
	// Get the payload format used in this application. If the payload format is "custom", you can get the custom JS
	// payload functions with the GetCustomPayloadFunctions() function.
	GetPayloadFormat() (string, error)

	// Set the payload format to use in this application. If you want to use custom JS payload functions, use the
	// SetCustomPayloadFunctions() function instead. If you want to disable payload conversion, pass an empty string.
	SetPayloadFormat(format string) error

	// Get the custom JS payload functions.
	GetCustomPayloadFunctions() (jsDecoder, jsConverter, jsValidator, jsEncoder string, err error)

	// Set the custom JS payload functions.
	//
	// Example Decoder:
	//
	// // Decoder (Array<byte>, uint8) returns (Object)
	// function Decoder(bytes, port) {
	//   var decoded = {};
	//   return decoded;
	// }
	//
	// Example Converter:
	//
	// // Converter (Object, uint8) returns (Object)
	// function Converter(decoded, port) {
	//   var converted = decoded;
	//   return converted;
	// }
	//
	// Example Validator:
	// // Validator (Object, uint8) returns (Boolean)
	// function Validator(converted, port) {
	//   return true;
	// }
	//
	// Example Encoder:
	//
	// // Validator (Object, uint8) returns (Array<byte>)
	// function Encoder(object, port) {
	//   var bytes = [];
	//   return bytes;
	// }
	SetCustomPayloadFunctions(jsDecoder, jsConverter, jsValidator, jsEncoder string) error
}

func (c *client) ManageApplication() (ApplicationManager, error) {
	if err := c.connectHandler(); err != nil {
		return nil, err
	}
	return &applicationManager{
		logger:         c.Logger,
		client:         handler.NewApplicationManagerClient(c.handler.conn),
		getContext:     c.getContext,
		requestTimeout: c.RequestTimeout,
		appID:          c.appID,
	}, nil
}

type applicationManager struct {
	logger         log.Interface
	client         handler.ApplicationManagerClient
	getContext     func(context.Context) context.Context
	requestTimeout time.Duration

	appID string
}

func (a *applicationManager) getApplication() (*handler.Application, error) {
	ctx, cancel := context.WithTimeout(a.getContext(context.Background()), a.requestTimeout)
	defer cancel()
	return a.client.GetApplication(ctx, &handler.ApplicationIdentifier{AppID: a.appID})
}

func (a *applicationManager) setApplication(app *handler.Application) error {
	ctx, cancel := context.WithTimeout(a.getContext(context.Background()), a.requestTimeout)
	defer cancel()
	_, err := a.client.SetApplication(ctx, app)
	return err
}

func (a *applicationManager) GetPayloadFormat() (string, error) {
	app, err := a.getApplication()
	if err != nil {
		return "", err
	}
	return app.PayloadFormat, nil
}

func (a *applicationManager) SetPayloadFormat(format string) error {
	app, err := a.getApplication()
	if err != nil {
		return err
	}
	app.PayloadFormat = format
	if app.PayloadFormat != "custom" {
		app.Decoder, app.Converter, app.Validator, app.Encoder = "", "", "", ""
	}
	return a.setApplication(app)
}

func (a *applicationManager) GetCustomPayloadFunctions() (jsDecoder, jsConverter, jsValidator, jsEncoder string, err error) {
	app, err := a.getApplication()
	if err != nil {
		return jsDecoder, jsConverter, jsValidator, jsEncoder, err
	}
	if app.PayloadFormat != "custom" {
		return jsDecoder, jsConverter, jsValidator, jsEncoder, fmt.Errorf("ttn-sdk: application does not have custom payload functions, but uses \"%s\"", app.PayloadFormat)
	}
	return app.Decoder, app.Converter, app.Validator, app.Encoder, nil
}

func (a *applicationManager) SetCustomPayloadFunctions(jsDecoder, jsConverter, jsValidator, jsEncoder string) error {
	app, err := a.getApplication()
	if err != nil {
		return err
	}
	app.PayloadFormat = "custom"
	app.Decoder, app.Converter, app.Validator, app.Encoder = jsDecoder, jsConverter, jsValidator, jsEncoder
	return a.setApplication(app)
}

func (a *applicationManager) TestCustomUplinkPayloadFunctions(jsDecoder, jsConverter, jsValidator string, payload []byte, port uint8) (*handler.DryUplinkResult, error) {
	ctx, cancel := context.WithTimeout(a.getContext(context.Background()), a.requestTimeout)
	defer cancel()
	return a.client.DryUplink(ctx, &handler.DryUplinkMessage{
		Payload: payload,
		App: handler.Application{
			AppID:         a.appID,
			PayloadFormat: "custom",
			Decoder:       jsDecoder,
			Converter:     jsConverter,
			Validator:     jsValidator,
		},
		Port: uint32(port),
	})
}

func (a *applicationManager) TestCustomDownlinkPayloadFunctions(jsEncoder string, fields map[string]interface{}, port uint8) (*handler.DryDownlinkResult, error) {
	ctx, cancel := context.WithTimeout(a.getContext(context.Background()), a.requestTimeout)
	defer cancel()
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	return a.client.DryDownlink(ctx, &handler.DryDownlinkMessage{
		Fields: string(fieldsJSON),
		App: handler.Application{
			AppID:         a.appID,
			PayloadFormat: "custom",
			Encoder:       jsEncoder,
		},
		Port: uint32(port),
	})
}
