// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk

import (
	"context"
	"os"
	"testing"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	testlog "github.com/TheThingsNetwork/go-utils/log/test"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/smartystreets/assertions"
)

var mqttAddress string

func init() {
	mqttAddress = os.Getenv("MQTT_ADDRESS")
	if mqttAddress == "" {
		mqttAddress = "localhost:1883"
	}
}

func TestMQTT(t *testing.T) {
	a := New(t)

	log := testlog.NewLogger()
	ttnlog.Set(log)
	defer log.Print(t)

	c := new(client)
	c.Logger = log
	defer c.Close()

	c.appID = "test"
	c.mqtt.ctx, c.mqtt.cancel = context.WithCancel(context.Background())
	c.mqtt.client = mqtt.NewClient(log, "test-mqtt", "", "", "tcp://"+mqttAddress)
	err := c.mqtt.client.Connect()
	a.So(err, ShouldBeNil)

	pubsub, err := c.PubSub()
	a.So(err, ShouldBeNil)

	defer pubsub.Close()

	testDevice := pubsub.Device("test")
	defer testDevice.Close()
	allDevices := pubsub.AllDevices()
	defer allDevices.Close()

	testUplink, err := testDevice.SubscribeUplink()
	a.So(err, ShouldBeNil)

	allUplink, err := allDevices.SubscribeUplink()
	a.So(err, ShouldBeNil)

	c.mqtt.client.PublishUplink(types.UplinkMessage{
		AppID:      "test",
		DevID:      "other",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()

	c.mqtt.client.PublishUplink(types.UplinkMessage{
		AppID:      "test",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()

	select {
	case msg := <-testUplink:
		a.So(msg.DevID, ShouldEqual, "test")
		a.So(msg.PayloadRaw, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
	case <-time.After(time.Second):
		t.Fatal("Did not receive from testUplink within a second")
	}

	for i := 0; i < 2; i++ {
		select {
		case msg := <-allUplink:
			a.So(msg.AppID, ShouldEqual, "test")
			a.So(msg.PayloadRaw, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		case <-time.After(time.Second):
			t.Fatalf("Did not receive %d from allUplink within a second", i)
		}
	}

	a.So(testDevice.UnsubscribeUplink(), ShouldBeNil)
	a.So(allDevices.UnsubscribeUplink(), ShouldBeNil)

	testEvent, err := testDevice.SubscribeEvents()
	a.So(err, ShouldBeNil)

	allEvent, err := allDevices.SubscribeEvents()
	a.So(err, ShouldBeNil)

	c.mqtt.client.PublishDeviceEvent("test", "other", types.UplinkErrorEvent, types.ErrorEventData{
		Error: "some error",
	}).Wait()

	c.mqtt.client.PublishDeviceEvent("test", "test", types.UplinkErrorEvent, types.ErrorEventData{
		Error: "some error",
	}).Wait()

	select {
	case msg := <-testEvent:
		a.So(msg.DevID, ShouldEqual, "test")
		a.So(msg.Data, ShouldNotBeNil)
		a.So(msg.Data, ShouldHaveSameTypeAs, new(types.ErrorEventData))
	case <-time.After(time.Second):
		t.Fatal("Did not receive from testEvent within a second")
	}

	for i := 0; i < 2; i++ {
		select {
		case msg := <-allEvent:
			a.So(msg.AppID, ShouldEqual, "test")
			a.So(msg.Data, ShouldNotBeNil)
			a.So(msg.Data, ShouldHaveSameTypeAs, new(types.ErrorEventData))
		case <-time.After(time.Second):
			t.Fatalf("Did not receive %d from allEvent within a second", i)
		}
	}

	a.So(testDevice.UnsubscribeEvents(), ShouldBeNil)
	a.So(allDevices.UnsubscribeEvents(), ShouldBeNil)

	testActivation, err := testDevice.SubscribeActivations()
	a.So(err, ShouldBeNil)

	allActivation, err := allDevices.SubscribeActivations()
	a.So(err, ShouldBeNil)

	c.mqtt.client.PublishActivation(types.Activation{
		AppID:  "test",
		DevID:  "other",
		AppEUI: types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
	}).Wait()

	c.mqtt.client.PublishActivation(types.Activation{
		AppID:  "test",
		DevID:  "test",
		AppEUI: types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
	}).Wait()

	select {
	case msg := <-testActivation:
		a.So(msg.DevID, ShouldEqual, "test")
	case <-time.After(time.Second):
		t.Fatal("Did not receive from testActivation within a second")
	}

	for i := 0; i < 2; i++ {
		select {
		case msg := <-allActivation:
			a.So(msg.AppID, ShouldEqual, "test")

		case <-time.After(time.Second):
			t.Fatalf("Did not receive %d from allActivation within a second", i)
		}
	}

	a.So(testDevice.UnsubscribeActivations(), ShouldBeNil)
	a.So(allDevices.UnsubscribeActivations(), ShouldBeNil)

	downlink := make(chan *types.DownlinkMessage)
	c.mqtt.client.SubscribeDownlink(func(_ mqtt.Client, appID string, devID string, msg types.DownlinkMessage) {
		downlink <- &msg
	})

	err = pubsub.Publish("test", &DownlinkMessage{
		AppID:      "test",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	a.So(err, ShouldBeNil)

	select {
	case msg := <-downlink:
		a.So(msg.AppID, ShouldEqual, "test")
		a.So(msg.DevID, ShouldEqual, "test")
		a.So(msg.PayloadRaw, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
	case <-time.After(time.Second):
		t.Fatal("Did not receive on downlink within a second")
	}
}
