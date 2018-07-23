# The Things Network Go SDK

[![Build Status](https://travis-ci.org/TheThingsNetwork/go-app-sdk.svg?branch=master)](https://travis-ci.org/TheThingsNetwork/go-app-sdk) [![Coverage Status](https://coveralls.io/repos/github/TheThingsNetwork/go-app-sdk/badge.svg?branch=master)](https://coveralls.io/github/TheThingsNetwork/go-app-sdk?branch=master) [![GoDoc](https://godoc.org/github.com/TheThingsNetwork/go-app-sdk?status.svg)](https://godoc.org/github.com/TheThingsNetwork/go-app-sdk)

![The Things Network](https://thethings.blob.core.windows.net/ttn/logo.svg)

## Usage

To avoid issues with incompatible dependencies, we recommend using [vgo](https://github.com/golang/go/wiki/vgo).

Assuming you're working on a project `github.com/your-username/your-project`:

```
go get -u golang.org/x/vgo
cd $GOPATH/src/github.com/your-username/your-project
vgo mod -init -module github.com/your-username/your-project
vgo get github.com/TheThingsNetwork/go-app-sdk
vgo mod -vendor
```

See the examples [on GoDoc](https://godoc.org/github.com/TheThingsNetwork/go-app-sdk#example-package).

## License

Source code for The Things Network is released under the MIT License, which can be found in the [LICENSE](LICENSE) file. A list of authors can be found in the [AUTHORS](AUTHORS) file.
