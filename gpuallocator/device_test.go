/**
# Copyright 2024 NVIDIA CORPORATION
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

package gpuallocator

import (
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/NVIDIA/go-nvml/pkg/nvml/mock"
	"github.com/stretchr/testify/require"
)

func TestDeviceListFilter(t *testing.T) {
	singleDeviceNVML := &mock.Interface{
		InitFunc: func() nvml.Return {
			return nvml.SUCCESS
		},
		ShutdownFunc: func() nvml.Return {
			return nvml.SUCCESS
		},
		DeviceGetCountFunc: func() (int, nvml.Return) {
			return 1, nvml.SUCCESS
		},
		DeviceGetHandleByIndexFunc: func(Index int) (nvml.Device, nvml.Return) {
			device := &mock.Device{
				GetNameFunc: func() (string, nvml.Return) {
					return "Device0", nvml.SUCCESS
				},
				GetUUIDFunc: func() (string, nvml.Return) {
					return "GPU-0", nvml.SUCCESS
				},
				GetPciInfoFunc: func() (nvml.PciInfo, nvml.Return) {
					return nvml.PciInfo{}, nvml.SUCCESS
				},
			}
			return device, nvml.SUCCESS
		},
	}

	testCases := []struct {
		description        string
		uuids              []string
		nvmllib            nvml.Interface
		expectedDeviceList DeviceList
		expectedError      error
	}{
		{
			description:        "nil uuids returns empty list",
			nvmllib:            singleDeviceNVML,
			expectedDeviceList: DeviceList{},
		},
		{
			description:        "empty uuids returns empty list",
			uuids:              []string{},
			nvmllib:            singleDeviceNVML,
			expectedDeviceList: DeviceList{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			defer setNVMLNewDuringTest(tc.nvmllib)()
			deviceList, err := NewDevicesFrom(tc.uuids)
			require.ErrorIs(t, tc.expectedError, err)
			require.EqualValues(t, tc.expectedDeviceList, deviceList)
		})
	}
}

func setNVMLNewDuringTest(to nvml.Interface) func() {
	original := nvmlNew
	nvmlNew = func() nvml.Interface {
		return to
	}

	return func() {
		nvmlNew = original
	}
}
