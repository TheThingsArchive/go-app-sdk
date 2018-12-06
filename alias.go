// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import "github.com/TheThingsNetwork/ttn/core/types"

// AppEUI is a unique identifier for applications.
type AppEUI = types.AppEUI

// DevEUI is a unique identifier for devices.
type DevEUI = types.DevEUI

// DevAddr is a non-unique address for LoRaWAN devices.
type DevAddr = types.DevAddr

// NwkSKey (Network Session Key) is used for LoRaWAN MIC calculation.
type NwkSKey = types.NwkSKey

// AppSKey (Application Session Key) is used for LoRaWAN payload encryption.
type AppSKey = types.AppSKey

// AppKey (Application Key) is used for LoRaWAN OTAA.
type AppKey = types.AppKey

// DownlinkMessage represents an application-layer downlink message.
type DownlinkMessage = types.DownlinkMessage

// UplinkMessage represents an application-layer uplink message.
type UplinkMessage = types.UplinkMessage

// DeviceEvent represents an application-layer event message for a device event.
type DeviceEvent = types.DeviceEvent

// Activation messages are used to notify application of a device activation.
type Activation = types.Activation
