// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk

// MoveDevice moves a device to another application
func MoveDevice(devID string, from, to DeviceManager) (err error) {
	dev, err := from.Get(devID)
	if err != nil {
		return err
	}
	err = from.Delete(devID)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			from.Set(dev)
		}
	}()
	err = to.Set(dev)
	if err != nil {
		return err
	}
	return nil
}
