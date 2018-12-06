// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import (
	"os"
	"strings"
	"testing"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	testlog "github.com/TheThingsNetwork/go-utils/log/test"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestClient(t *testing.T) {
	a := New(t)

	for _, env := range strings.Split("APP_ID APP_ACCESS_KEY", " ") {
		if os.Getenv(env) == "" {
			t.Skipf("Skipping client test: %s not configured", env)
		}
	}

	log := testlog.NewLogger()
	ttnlog.Set(log)
	defer log.Print(t)

	appID := os.Getenv("APP_ID")

	config := NewCommunityConfig("client-test")
	client := config.NewClient(appID, os.Getenv("APP_ACCESS_KEY")).(*client)

	client.connectHandler()
	client.connectMQTT()

	time.Sleep(10 * time.Millisecond)

	client.Close()

	time.Sleep(time.Second)

	client.connectHandler()
	client.connectMQTT()

	time.Sleep(10 * time.Millisecond)

	devs, err := client.ManageDevices()
	a.So(err, ShouldBeNil)

	var appEUI types.AppEUI
	var devEUI types.DevEUI
	var appKey types.AppKey
	random.FillBytes(appEUI[:])
	random.FillBytes(devEUI[:])
	random.FillBytes(appKey[:])

	dev := &Device{
		SparseDevice: SparseDevice{
			AppID:       appID,
			DevID:       "sdk-test",
			AppEUI:      AppEUI(appEUI),
			DevEUI:      DevEUI(devEUI),
			Description: "SDK Test Device",
			AppKey:      (*AppKey)(&appKey),
		},
		Uses32BitFCnt: true,
	}

	err = devs.Set(dev)
	a.So(err, ShouldBeNil)
	defer devs.Delete("sdk-test")

	deviceList, err := devs.List(0, 0)
	a.So(err, ShouldBeNil)
	a.So(deviceList, ShouldNotBeEmpty)

	dev, err = devs.Get("sdk-test")
	a.So(err, ShouldBeNil)

	err = dev.PersonalizeRandom()
	a.So(err, ShouldBeNil)
	a.So(dev.DevAddr, ShouldNotBeNil)

	pubsub, err := client.PubSub()
	a.So(err, ShouldBeNil)

	uplink, err := pubsub.AllDevices().SubscribeUplink()
	a.So(err, ShouldBeNil)

	sim, err := client.Simulate("sdk-test")
	a.So(err, ShouldBeNil)

	err = sim.Uplink(1, []byte{0xaa, 0xbc})
	a.So(err, ShouldBeNil)

	select {
	case msg := <-uplink:
		a.So(msg.AppID, ShouldEqual, appID)
		a.So(msg.DevID, ShouldEqual, "sdk-test")
	case <-time.After(time.Second):
		t.Fatal("Did not receive uplink within a second")
	}

	err = dev.Delete()
	a.So(err, ShouldBeNil)
}

func TestCleanMQTTAddress(t *testing.T) {
	a := New(t)

	addr, err := cleanMQTTAddress("localhost:1883")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, "tcp://localhost:1883")

	// if `host:port` then `mqtt://host:port`
	addr, err = cleanMQTTAddress("localhost:1234")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, "tcp://localhost:1234")

	// if `host:8883` then `mqtts://host:8883`
	addr, err = cleanMQTTAddress("localhost:8883")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, "ssl://localhost:8883")

	// if `host` then `mqtt://host:1883` and `mqtts://host:8883` (we choose mqtts)
	addr, err = cleanMQTTAddress("localhost")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, "ssl://localhost:8883")

	// if `mqtt://host` then `mqtt://host:1883`
	addr, err = cleanMQTTAddress("mqtt://localhost")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, "tcp://localhost:1883")

	// if `mqtts://host` then `mqtt://host:1883` and `mqtts://host:8883` (we choose mqtts)
	addr, err = cleanMQTTAddress("mqtts://localhost")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, "ssl://localhost:8883")

}
