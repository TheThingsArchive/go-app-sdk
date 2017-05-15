// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/log/apex"
	"github.com/TheThingsNetwork/go-utils/random"
	"github.com/TheThingsNetwork/ttn/core/types"
)

const (
	sdkClientName = "my-amazing-app"
)

func Example() {
	log := apex.Stdout() // We use a cli logger at Stdout
	ttnlog.Set(log)      // Set the logger as default for TTN

	// We get the application ID and application access key from the environment
	appID := os.Getenv("TTN_APP_ID")
	appAccessKey := os.Getenv("TTN_APP_ACCESS_KEY")

	// Create a new SDK configuration for the public community network
	config := ttnsdk.NewCommunityConfig(sdkClientName)
	config.ClientVersion = "2.0.5" // The version of the application

	// If you connect to a private network that does not use trusted certificates (from Let's Encrypt for example), you
	// have to manually trust the certificates. If you use the public community network, you can just delete the next
	// code block.
	if caCert := os.Getenv("TTN_CA_CERT"); caCert != "" {
		config.TLSConfig = new(tls.Config)
		certBytes, err := ioutil.ReadFile(caCert)
		if err != nil {
			log.WithError(err).Fatal("my-amazing-app: could not read CA certificate file")
		}
		config.TLSConfig.RootCAs = x509.NewCertPool()
		if ok := config.TLSConfig.RootCAs.AppendCertsFromPEM(certBytes); !ok {
			log.Fatal("my-amazing-app: could not read CA certificates")
		}
	}

	// Create a new SDK client for the application
	client := config.NewClient(appID, appAccessKey)

	// Manage devices for the application.
	devices, err := client.ManageDevices()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not get device manager")
	}

	// List the first 10 devices
	deviceList, err := devices.List(10, 0)
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not get devices")
	}
	log.Info("my-amazing-app: found devices")
	for _, device := range deviceList {
		fmt.Printf("- %s", device.DevID)
	}

	// Create a new device
	dev := new(ttnsdk.Device)
	dev.AppID = appID
	dev.DevID = "my-new-device"
	dev.Description = "A new device in my amazing app"
	dev.AppEUI = types.AppEUI{0x70, 0xB3, 0xD5, 0x7E, 0xF0, 0x00, 0x00, 0x24} // Use the real AppEUI here
	dev.DevEUI = types.DevEUI{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08} // Use the real DevEUI here
	random.FillBytes(dev.AppKey[:])                                           // Generate a random AppKey

	err = devices.Set(dev)
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not create device")
	}

	// Get the device
	dev, err = devices.Get("my-new-device")
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not get device")
	}

	// Personalize the device with random session keys
	err = dev.PersonalizeRandom()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not personalize device")
	}
	log.WithFields(ttnlog.Fields{
		"devAddr": dev.DevAddr,
		"nwkSKey": dev.NwkSKey,
		"appSKey": dev.AppSKey,
	}).Info("my-amazing-app: personalized device")

	// Start Publish/Subscribe client (MQTT)
	pubsub, err := client.PubSub()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not get application pub/sub")
	}

	allDevicesPubSub := pubsub.Device("my-new-device")

	// Subscribe to activations
	activations, err := allDevicesPubSub.SubscribeActivations()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not subscribe to activations")
	}
	for activation := range activations {
		log.WithFields(ttnlog.Fields{
			"appEUI":  activation.AppEUI.String(),
			"devEUI":  activation.DevEUI.String(),
			"devAddr": activation.DevAddr.String(),
		}).Info("my-amazing-app: received activation")
	}

	// Subscribe to events
	events, err := allDevicesPubSub.SubscribeEvents()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not subscribe to events")
	}
	for event := range events {
		log.WithFields(ttnlog.Fields{
			"devID":     event.DevID,
			"eventType": event.Event,
		}).Info("my-amazing-app: received event")
		if event.Data != nil {
			eventJSON, _ := json.Marshal(event.Data)
			fmt.Println("Event data:" + string(eventJSON))
		}
	}

	myNewDevicePubSub := pubsub.Device("my-new-device")

	// Subscribe to uplink messages
	uplink, err := myNewDevicePubSub.SubscribeUplink()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not subscribe to uplink messages")
	}
	for message := range uplink {
		hexPayload := hex.EncodeToString(message.PayloadRaw)
		log.WithField("data", hexPayload).Info("my-amazing-app: received uplink")
	}

	// Publish downlink message
	err = myNewDevicePubSub.Publish(&types.DownlinkMessage{
		AppID:      appID,           // can be left out, the SDK will fill this
		DevID:      "my-new-device", // can be left out, the SDK will fill this
		PayloadRaw: []byte{0xaa, 0xbc},
		FPort:      10,
		Schedule:   types.ScheduleLast, // allowed values: "replace" (default), "first", "last"
		Confirmed:  false,              // can be left out, default is false
	})
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not schedule downlink message")
	}

}
