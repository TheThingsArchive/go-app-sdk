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
	"github.com/TheThingsNetwork/api/protocol/lorawan"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	testlog "github.com/TheThingsNetwork/go-utils/log/test"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestDeviceManager(t *testing.T) {
	a := New(t)

	log := testlog.NewLogger()
	ttnlog.Set(log)
	defer log.Print(t)

	mock := new(mockApplicationManagerClient)
	devMock := new(mockDevAddrManagerClient)

	manager := &deviceManager{
		logger:         log,
		client:         mock,
		devAddrClient:  devMock,
		getContext:     func(ctx context.Context) context.Context { return ctx },
		requestTimeout: time.Second,
		appID:          "test",
	}

	someErr := errors.New("some error")

	{
		mock.reset()
		mock.err = someErr
		_, err := manager.List(0, 0)
		a.So(err, ShouldNotBeNil)

		mock.reset()
		mock.deviceList = &handler.DeviceList{Devices: []*handler.Device{
			&handler.Device{
				DevID: "dev-id",
				Device: &handler.Device_LoRaWANDevice{LoRaWANDevice: &lorawan.Device{
					DevEUI: types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
					AppEUI: types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
				}},
			},
		}}
		sparseDevices, err := manager.List(0, 0)
		a.So(err, ShouldBeNil)
		a.So(mock.applicationIdentifier, ShouldNotBeNil)
		a.So(mock.applicationIdentifier.AppID, ShouldEqual, "test")
		devices := sparseDevices.AsDevices()
		a.So(devices, ShouldHaveLength, 1)
		a.So(devices[0].DevID, ShouldEqual, "dev-id")
		a.So(devices[0].AppEUI, ShouldEqual, types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8})
		a.So(devices[0].DevEUI, ShouldEqual, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8})
	}

	{
		mock.reset()
		mock.err = someErr
		_, err := manager.Get("dev-id")
		a.So(err, ShouldNotBeNil)

		mock.reset()
		mock.device = &handler.Device{
			DevID: "dev-id",
			Device: &handler.Device_LoRaWANDevice{LoRaWANDevice: &lorawan.Device{
				AppEUI:   types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
				FCntDown: 42,
			}},
		}
		device, err := manager.Get("dev-id")
		a.So(err, ShouldBeNil)
		a.So(mock.deviceIdentifier, ShouldNotBeNil)
		a.So(mock.deviceIdentifier.AppID, ShouldEqual, "test")
		a.So(mock.deviceIdentifier.DevID, ShouldEqual, "dev-id")
		a.So(device.DevID, ShouldEqual, "dev-id")
		a.So(device.AppEUI, ShouldEqual, types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8})
		a.So(device.DevEUI, ShouldEqual, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8})
		a.So(device.FCntDown, ShouldEqual, 42)
	}

	{
		mock.reset()
		mock.err = someErr
		err := manager.Set(&Device{})
		a.So(err, ShouldNotBeNil)

		mock.reset()
		err = manager.Set(&Device{
			SparseDevice: SparseDevice{
				DevID:  "dev-id",
				AppEUI: AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI: DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
			},
			FCntDown: 42,
		})
		a.So(err, ShouldBeNil)
		a.So(mock.device, ShouldNotBeNil)
		a.So(mock.device.DevID, ShouldEqual, "dev-id")
		a.So(mock.device.GetLoRaWANDevice().DevID, ShouldEqual, "dev-id")
		a.So(mock.device.GetLoRaWANDevice().AppEUI, ShouldResemble, types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8})
		a.So(mock.device.GetLoRaWANDevice().DevEUI, ShouldResemble, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8})
		a.So(mock.device.GetLoRaWANDevice().FCntDown, ShouldEqual, 42)
	}

	{
		mock.reset()
		mock.err = someErr
		err := manager.Delete("dev-id")
		a.So(err, ShouldNotBeNil)

		mock.reset()
		err = manager.Delete("dev-id")
		a.So(err, ShouldBeNil)
		a.So(mock.deviceIdentifier, ShouldNotBeNil)
		a.So(mock.deviceIdentifier.AppID, ShouldEqual, "test")
		a.So(mock.deviceIdentifier.DevID, ShouldEqual, "dev-id")
	}

	{
		dev := new(Device)
		a.So(dev.IsNew(), ShouldBeTrue)
		// Can't call these funcs on a new device
		a.So(func() { dev.Update() }, ShouldPanic)
		a.So(func() { dev.Personalize(NwkSKey{}, AppSKey{}) }, ShouldPanic)
		a.So(func() { dev.Delete() }, ShouldPanic)
	}

	{
		dev := new(Device)
		a.So(dev.IsNew(), ShouldBeTrue)
		dev.SetManager(manager)
		a.So(dev.IsNew(), ShouldBeFalse)
		// You can set the same manager
		a.So(func() { dev.SetManager(manager) }, ShouldNotPanic)
		// But you can't change the manager
		otherManager := &deviceManager{}
		a.So(func() { dev.SetManager(otherManager) }, ShouldPanic)
	}

	{
		mock.reset()
		mock.device = &handler.Device{
			DevID: "dev-id",
			Device: &handler.Device_LoRaWANDevice{LoRaWANDevice: &lorawan.Device{
				AppEUI:   types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:   types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
				FCntDown: 42,
			}},
		}
		device, err := manager.Get("dev-id")
		a.So(err, ShouldBeNil)

		device.FCntDown = 0
		err = device.Update()
		a.So(err, ShouldBeNil)
		a.So(mock.device.GetLoRaWANDevice().FCntDown, ShouldEqual, 0)

		mock.reset()
		devMock.reset()
		devMock.err = someErr
		err = device.Personalize(NwkSKey{}, AppSKey{})
		a.So(err, ShouldNotBeNil)

		mock.reset()
		devMock.reset()
		devMock.devAddrResponse = &lorawan.DevAddrResponse{DevAddr: types.DevAddr{1, 2, 3, 4}}
		err = device.Personalize(NwkSKey{}, AppSKey{})
		a.So(err, ShouldBeNil)
		a.So(mock.device.GetLoRaWANDevice().DevAddr, ShouldResemble, &types.DevAddr{1, 2, 3, 4})

		mock.reset()
		err = device.Delete()
		a.So(err, ShouldBeNil)
		a.So(mock.deviceIdentifier.DevID, ShouldEqual, "dev-id")
	}

}
