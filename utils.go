// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package ttnsdk

import (
	"fmt"
	"net/url"
	"strings"
)

func cleanMQTTAddress(in string) (address string, err error) {
	if !strings.Contains(in, "://") {
		in = "detect://" + in
	}
	url, err := url.Parse(in)
	if err != nil {
		return address, err
	}
	switch url.Scheme {
	case "detect", "mqtt", "mqtts":
	default:
		return address, fmt.Errorf("ttn-sdk: unknown mqtt scheme: %s", url.Scheme)
	}
	scheme, host, port := url.Scheme, url.Hostname(), url.Port()
	if scheme == "detect" {
		switch port {
		case "8883", "":
			scheme = "ssl"
		default:
			scheme = "tcp"
		}
	}
	if port == "" {
		switch scheme {
		case "ssl", "mqtts":
			scheme = "ssl"
			port = "8883"
		case "tcp", "mqtt":
			scheme = "tcp"
			port = "1883"
		}
	}
	return fmt.Sprintf("%s://%s:%s", scheme, host, port), nil
}
