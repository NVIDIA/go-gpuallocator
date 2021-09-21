/**
# Copyright (c) 2021, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package nvml

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

const (
	CPUAffinityNotSupported = -1
)

type nvmlDevice nvml.Device

var _ Device = (*nvmlDevice)(nil)

func DeviceGetHandleByIndex(index int) (Device, Return) {
	d, ret := nvml.DeviceGetHandleByIndex(index)
	return nvmlDevice(d), nvmlReturn(ret)
}

func (d nvmlDevice) GetAttributes() (DeviceAttributes, Return) {
	a1, ret := nvml.Device(d).GetAttributes()
	return DeviceAttributes(a1), nvmlReturn(ret)
}

func (d nvmlDevice) GetComputeInstanceId() (int, Return) {
	i1, ret := nvml.Device(d).GetComputeInstanceId()
	return i1, nvmlReturn(ret)
}

func (d nvmlDevice) GetDeviceHandleFromMigDeviceHandle() (Device, Return) {
	d1, ret := nvml.Device(d).GetDeviceHandleFromMigDeviceHandle()
	return nvmlDevice(d1), nvmlReturn(ret)
}

func (d nvmlDevice) GetGpuInstanceId() (int, Return) {
	i1, ret := nvml.Device(d).GetGpuInstanceId()
	return i1, nvmlReturn(ret)
}

func (d nvmlDevice) GetMaxMigDeviceCount() (int, Return) {
	s1, ret := nvml.Device(d).GetMaxMigDeviceCount()
	return s1, nvmlReturn(ret)
}

func (d nvmlDevice) GetMigDeviceHandleByIndex(index int) (Device, Return) {
	h, ret := nvml.Device(d).GetMigDeviceHandleByIndex(index)
	return nvmlDevice(h), nvmlReturn(ret)
}

func (d nvmlDevice) GetMigMode() (int, int, Return) {
	s1, s2, ret := nvml.Device(d).GetMigMode()
	return s1, s2, nvmlReturn(ret)
}

func (d nvmlDevice) GetMinorNumber() (int, Return) {
	i1, ret := nvml.Device(d).GetMinorNumber()
	return i1, nvmlReturn(ret)
}

func (d nvmlDevice) GetPciInfo() (PciInfo, Return) {
	p1, ret := nvml.Device(d).GetPciInfo()
	return PciInfo(p1), nvmlReturn(ret)
}

func (d nvmlDevice) GetUUID() (string, Return) {
	s1, ret := nvml.Device(d).GetUUID()
	return s1, nvmlReturn(ret)
}

func (d nvmlDevice) RegisterEvents(EventTypes uint64, Set EventSet) Return {
	ret := nvml.Device(d).RegisterEvents(EventTypes, nvml.EventSet(Set))
	return nvmlReturn(ret)
}
