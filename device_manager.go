// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk

import (
	"context"
	"strings"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// DeviceManager manages devices within an application
type DeviceManager interface {
	// List devices in an application. Use the limit and offset for pagination. Requests that fetch many devices will be
	// very slow, which is often not necessary. If you use this function too often, the response will be cached by the
	// server, and you might receive outdated data.
	List(limit, offset uint64) ([]*SparseDevice, error)

	// Get details for a device
	Get(devID string) (*Device, error)

	// Create or Update a device.
	Set(*Device) error

	// Delete a device
	Delete(devID string) error
}

func (c *client) ManageDevices() (DeviceManager, error) {
	if err := c.connectHandler(); err != nil {
		return nil, err
	}
	return &deviceManager{
		logger:         c.Logger,
		client:         handler.NewApplicationManagerClient(c.handler.conn),
		devAddrClient:  lorawan.NewDevAddrManagerClient(c.handler.conn),
		getContext:     c.getContext,
		requestTimeout: c.RequestTimeout,
		appID:          c.appID,
	}, nil
}

type deviceManager struct {
	logger         log.Interface
	client         handler.ApplicationManagerClient
	devAddrClient  lorawan.DevAddrManagerClient
	getContext     func(context.Context) context.Context
	requestTimeout time.Duration

	appID string
}

func (d *deviceManager) List(limit, offset uint64) (devices []*SparseDevice, err error) {
	ctx, cancel := context.WithTimeout(d.getContext(context.Background()), d.requestTimeout)
	defer cancel()
	ctx = ttnctx.OutgoingContextWithLimitAndOffset(ctx, limit, offset)
	res, err := d.client.GetDevicesForApplication(ctx, &handler.ApplicationIdentifier{AppId: d.appID})
	if err != nil {
		return nil, err
	}
	for _, res := range res.Devices {
		dev := new(SparseDevice)
		dev.fromProto(res)
		devices = append(devices, dev)
	}
	return devices, nil
}

func (d *deviceManager) Get(devID string) (*Device, error) {
	ctx, cancel := context.WithTimeout(d.getContext(context.Background()), d.requestTimeout)
	defer cancel()
	res, err := d.client.GetDevice(ctx, &handler.DeviceIdentifier{AppId: d.appID, DevId: devID})
	if err != nil {
		return nil, err
	}
	dev := &Device{deviceManager: d}
	dev.fromProto(res)
	return dev, nil
}

func (d *deviceManager) Set(dev *Device) error {
	if dev.AppID != d.appID {
		dev.AppID = d.appID
	}
	req := new(handler.Device)
	dev.toProto(req)
	ctx, cancel := context.WithTimeout(d.getContext(context.Background()), d.requestTimeout)
	defer cancel()
	_, err := d.client.SetDevice(ctx, req) // TODO: fill dev from response and set deviceManager when the server actually returns the device
	return err
}

func (d *deviceManager) Delete(devID string) error {
	ctx, cancel := context.WithTimeout(d.getContext(context.Background()), d.requestTimeout)
	defer cancel()
	_, err := d.client.DeleteDevice(ctx, &handler.DeviceIdentifier{AppId: d.appID, DevId: devID})
	return err
}

// SparseDevice contains most, but not all fields of the device. It's returned by List operations to save server resources
type SparseDevice struct {
	AppID       string         `json:"app_id"`
	DevID       string         `json:"dev_id"`
	AppEUI      types.AppEUI   `json:"app_eui"`
	DevEUI      types.DevEUI   `json:"dev_eui"`
	Description string         `json:"description,omitempty"`
	DevAddr     *types.DevAddr `json:"dev_addr,omitempty"`
	NwkSKey     *types.NwkSKey `json:"nwk_s_key,omitempty"`
	AppSKey     *types.AppSKey `json:"app_s_key,omitempty"`
	AppKey      *types.AppKey  `json:"app_key,omitempty"`
	Latitude    float32        `json:"latitude,omitempty"`
	Longitude   float32        `json:"longitude,omitempty"`
	Altitude    int32          `json:"altitude,omitempty"`
}

func (d *SparseDevice) fromProto(dev *handler.Device) {
	d.AppID = dev.AppId
	d.DevID = dev.DevId
	d.Description = dev.Description
	if lorawanDevice := dev.GetLorawanDevice(); lorawanDevice != nil {
		if lorawanDevice.AppEui != nil {
			d.AppEUI = *lorawanDevice.AppEui
		}
		if lorawanDevice.DevEui != nil {
			d.DevEUI = *lorawanDevice.DevEui
		}
		d.DevAddr = lorawanDevice.DevAddr
		d.NwkSKey = lorawanDevice.NwkSKey
		d.AppSKey = lorawanDevice.AppSKey
		d.AppKey = lorawanDevice.AppKey
	}
	d.Latitude = dev.Latitude
	d.Longitude = dev.Longitude
	d.Altitude = dev.Altitude
}

func (d *SparseDevice) toProto(dev *handler.Device) {
	dev.AppId = d.AppID
	dev.DevId = d.DevID
	dev.Description = d.Description
	dev.Latitude = d.Latitude
	dev.Longitude = d.Longitude
	dev.Altitude = d.Altitude
	if dev.Device == nil {
		dev.Device = &handler.Device_LorawanDevice{LorawanDevice: &lorawan.Device{}}
	}
	lorawanDevice := dev.GetLorawanDevice()
	lorawanDevice.AppId = d.AppID
	lorawanDevice.DevId = d.DevID
	lorawanDevice.AppEui = &d.AppEUI
	lorawanDevice.DevEui = &d.DevEUI
	lorawanDevice.DevAddr = d.DevAddr
	lorawanDevice.NwkSKey = d.NwkSKey
	lorawanDevice.AppSKey = d.AppSKey
	lorawanDevice.AppKey = d.AppKey
}

// Device in an application
type Device struct {
	deviceManager *deviceManager

	SparseDevice
	FCntUp                uint32    `json:"f_cnt_up"`
	FCntDown              uint32    `json:"f_cnt_down"`
	DisableFCntCheck      bool      `json:"disable_f_cnt_check"`
	Uses32BitFCnt         bool      `json:"uses32_bit_f_cnt"`
	ActivationConstraints string    `json:"activation_constraints"`
	LastSeen              time.Time `json:"last_seen"`
}

func (d *Device) addActivationConstraint(c string) {
	constraints := strings.Split(d.ActivationConstraints, ",")
	for _, constraint := range constraints {
		if constraint == c {
			return
		}
	}
	constraints = append(constraints, c)
	d.ActivationConstraints = strings.Join(constraints, ",")
}

// Update the device. This function panics if this is a new device.
func (d *Device) Update() error {
	if d.deviceManager == nil {
		panic("ttn-sdk: you can not update new devices")
	}
	return d.deviceManager.Set(d)
}

// Delete the device. This function panics if this is a new device.
func (d *Device) Delete() error {
	if d.deviceManager == nil {
		panic("ttn-sdk: you can not update new devices")
	}
	return d.deviceManager.Delete(d.DevID)
}

// PersonalizeRandom calls Personalize with a random NwkSKey and random AppSKey. This function panics if this is a new device.
func (d *Device) PersonalizeRandom() error {
	var nwkSKey types.NwkSKey
	var appSKey types.AppSKey
	random.FillBytes(nwkSKey[:])
	random.FillBytes(appSKey[:])
	return d.Personalize(nwkSKey, appSKey)
}

// Personalize the device. This function panics if this is a new device.
func (d *Device) Personalize(nwkSKey types.NwkSKey, appSKey types.AppSKey) error {
	if d.deviceManager == nil {
		panic("ttn-sdk: you can not update new devices. Use the Get() function to retrieve the device from the server first.")
	}
	d.addActivationConstraint("abp")
	ctx, cancel := context.WithTimeout(d.deviceManager.getContext(context.Background()), d.deviceManager.requestTimeout)
	defer cancel()
	res, err := d.deviceManager.devAddrClient.GetDevAddr(ctx, &lorawan.DevAddrRequest{Usage: strings.Split(d.ActivationConstraints, ",")})
	if err != nil {
		return err
	}
	d.DevAddr = res.DevAddr
	d.NwkSKey = &nwkSKey
	d.AppSKey = &appSKey
	return d.Update()
}

func (d *Device) fromProto(dev *handler.Device) {
	d.SparseDevice.fromProto(dev)
	if lorawanDevice := dev.GetLorawanDevice(); lorawanDevice != nil {
		d.FCntUp = lorawanDevice.FCntUp
		d.FCntDown = lorawanDevice.FCntDown
		d.DisableFCntCheck = lorawanDevice.DisableFCntCheck
		d.Uses32BitFCnt = lorawanDevice.Uses32BitFCnt
		d.ActivationConstraints = lorawanDevice.ActivationConstraints
		d.LastSeen = time.Unix(0, lorawanDevice.LastSeen)
	}
}

func (d *Device) toProto(dev *handler.Device) {
	d.SparseDevice.toProto(dev)
	lorawanDevice := dev.GetLorawanDevice()
	lorawanDevice.FCntUp = d.FCntUp
	lorawanDevice.FCntDown = d.FCntDown
	lorawanDevice.DisableFCntCheck = d.DisableFCntCheck
	lorawanDevice.Uses32BitFCnt = d.Uses32BitFCnt
	lorawanDevice.ActivationConstraints = d.ActivationConstraints
}
