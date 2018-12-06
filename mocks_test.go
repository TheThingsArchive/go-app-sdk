// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.
//
// +build go1.9

package ttnsdk

import (
	"github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/api/protocol/lorawan"
	ptypes "github.com/gogo/protobuf/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type mockApplicationManagerClient struct {
	application            *handler.Application
	applicationIdentifier  *handler.ApplicationIdentifier
	ctx                    context.Context
	device                 *handler.Device
	deviceIdentifier       *handler.DeviceIdentifier
	deviceList             *handler.DeviceList
	dryDownlinkMessage     *handler.DryDownlinkMessage
	dryDownlinkResult      *handler.DryDownlinkResult
	dryUplinkMessage       *handler.DryUplinkMessage
	dryUplinkResult        *handler.DryUplinkResult
	empty                  *ptypes.Empty
	err                    error
	SimulatedUplinkMessage *handler.SimulatedUplinkMessage
}

func (m *mockApplicationManagerClient) reset() {
	m.application = nil
	m.applicationIdentifier = nil
	m.ctx = nil
	m.device = nil
	m.deviceIdentifier = nil
	m.deviceList = nil
	m.dryDownlinkMessage = nil
	m.dryDownlinkResult = nil
	m.dryUplinkMessage = nil
	m.dryUplinkResult = nil
	m.empty = nil
	m.err = nil
	m.SimulatedUplinkMessage = nil
}

func (m *mockApplicationManagerClient) RegisterApplication(ctx context.Context, in *handler.ApplicationIdentifier, opts ...grpc.CallOption) (*ptypes.Empty, error) {
	m.ctx = ctx
	m.applicationIdentifier = in
	return m.empty, m.err
}
func (m *mockApplicationManagerClient) GetApplication(ctx context.Context, in *handler.ApplicationIdentifier, opts ...grpc.CallOption) (*handler.Application, error) {
	m.ctx = ctx
	m.applicationIdentifier = in
	return m.application, m.err
}
func (m *mockApplicationManagerClient) SetApplication(ctx context.Context, in *handler.Application, opts ...grpc.CallOption) (*ptypes.Empty, error) {
	m.ctx = ctx
	m.application = in
	return m.empty, m.err
}
func (m *mockApplicationManagerClient) DeleteApplication(ctx context.Context, in *handler.ApplicationIdentifier, opts ...grpc.CallOption) (*ptypes.Empty, error) {
	m.ctx = ctx
	m.applicationIdentifier = in
	return m.empty, m.err
}
func (m *mockApplicationManagerClient) GetDevice(ctx context.Context, in *handler.DeviceIdentifier, opts ...grpc.CallOption) (*handler.Device, error) {
	m.ctx = ctx
	m.deviceIdentifier = in
	return m.device, m.err
}
func (m *mockApplicationManagerClient) SetDevice(ctx context.Context, in *handler.Device, opts ...grpc.CallOption) (*ptypes.Empty, error) {
	m.ctx = ctx
	m.device = in
	return m.empty, m.err
}
func (m *mockApplicationManagerClient) DeleteDevice(ctx context.Context, in *handler.DeviceIdentifier, opts ...grpc.CallOption) (*ptypes.Empty, error) {
	m.ctx = ctx
	m.deviceIdentifier = in
	return m.empty, m.err
}
func (m *mockApplicationManagerClient) GetDevicesForApplication(ctx context.Context, in *handler.ApplicationIdentifier, opts ...grpc.CallOption) (*handler.DeviceList, error) {
	m.ctx = ctx
	m.applicationIdentifier = in
	return m.deviceList, m.err
}
func (m *mockApplicationManagerClient) DryDownlink(ctx context.Context, in *handler.DryDownlinkMessage, opts ...grpc.CallOption) (*handler.DryDownlinkResult, error) {
	m.ctx = ctx
	m.dryDownlinkMessage = in
	return m.dryDownlinkResult, m.err
}
func (m *mockApplicationManagerClient) DryUplink(ctx context.Context, in *handler.DryUplinkMessage, opts ...grpc.CallOption) (*handler.DryUplinkResult, error) {
	m.ctx = ctx
	m.dryUplinkMessage = in
	return m.dryUplinkResult, m.err
}
func (m *mockApplicationManagerClient) SimulateUplink(ctx context.Context, in *handler.SimulatedUplinkMessage, opts ...grpc.CallOption) (*ptypes.Empty, error) {
	m.ctx = ctx
	m.SimulatedUplinkMessage = in
	return m.empty, m.err
}

type mockDevAddrManagerClient struct {
	ctx              context.Context
	devAddrRequest   *lorawan.DevAddrRequest
	devAddrResponse  *lorawan.DevAddrResponse
	err              error
	prefixesRequest  *lorawan.PrefixesRequest
	prefixesResponse *lorawan.PrefixesResponse
}

func (m *mockDevAddrManagerClient) reset() {
	m.ctx = nil
	m.devAddrRequest = nil
	m.devAddrResponse = nil
	m.err = nil
	m.prefixesRequest = nil
	m.prefixesResponse = nil
}

func (m *mockDevAddrManagerClient) GetPrefixes(ctx context.Context, in *lorawan.PrefixesRequest, opts ...grpc.CallOption) (*lorawan.PrefixesResponse, error) {
	m.ctx = ctx
	m.prefixesRequest = in
	return m.prefixesResponse, m.err
}
func (m *mockDevAddrManagerClient) GetDevAddr(ctx context.Context, in *lorawan.DevAddrRequest, opts ...grpc.CallOption) (*lorawan.DevAddrResponse, error) {
	m.ctx = ctx
	m.devAddrRequest = in
	return m.devAddrResponse, m.err
}
