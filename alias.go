package ttnsdk

import "github.com/TheThingsNetwork/ttn/core/types"

// TODO(NeuralSpaz) replace AppEUI()/DevEUI() with type aliases.

// AppEUI converts [8]byte to ttn/core/types.AppEUI
// which is a unique identifier for applications.
func AppEUI(b [8]byte) types.AppEUI { return (types.AppEUI(b)) }

// DevEUI converts [8]byte to ttn/core/types.DevEUI
// which is a unique is a unique identifier for devices.
func DevEUI(b [8]byte) types.DevEUI { return (types.DevEUI(b)) }

// These alias types in ttn/core/types to avoid venderoed type errors.

// DevAddr is a non-unique address for LoRaWAN devices.
type DevAddr types.DevAddr

// NwkSKey (Network Session Key) is used for LoRaWAN MIC calculation.
type NwkSKey types.NwkSKey

// AppSKey (Application Session Key) is used for LoRaWAN payload encryption.
type AppSKey types.AppSKey

// AppKey (Application Key) is used for LoRaWAN OTAA.
type AppKey types.AppKey

// DownlinkMessage represents an application-layer downlink message.
type DownlinkMessage types.DownlinkMessage

// UplinkMessage represents an application-layer uplink message.
type UplinkMessage types.UplinkMessage

// DeviceEvent represents an application-layer event message for a device event.
type DeviceEvent types.DeviceEvent

// Activation messages are used to notify application of a device activation.
type Activation types.Activation
