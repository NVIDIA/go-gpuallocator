/**
# Copyright 2026 NVIDIA CORPORATION
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

package links

import (
	"testing"

	"github.com/NVIDIA/go-nvlib/pkg/nvlib/device"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// testDevice is a stub device.Device that only implements the two NVLink
// methods exercised by getAllNvLinkRemotePciInfo.
type testDevice struct {
	device.Device
	getNvLinkState         func(int) (nvml.EnableState, nvml.Return)
	getNvLinkRemotePciInfo func(int) (nvml.PciInfo, nvml.Return)
}

func (t *testDevice) GetNvLinkState(i int) (nvml.EnableState, nvml.Return) {
	return t.getNvLinkState(i)
}

func (t *testDevice) GetNvLinkRemotePciInfo(i int) (nvml.PciInfo, nvml.Return) {
	return t.getNvLinkRemotePciInfo(i)
}

// pciInfoWithBusID builds an nvml.PciInfo whose BusId field encodes the
// provided ASCII string. BusId is a fixed-size []int8 in the nvml C struct.
func pciInfoWithBusID(busID string) nvml.PciInfo {
	var info nvml.PciInfo
	for i := 0; i < len(busID) && i < len(info.BusId); i++ {
		info.BusId[i] = busID[i]
	}
	return info
}

func TestGetAllNvLinkRemotePciInfo(t *testing.T) {
	t.Run("GetNvLinkState ERROR_GPU_IS_LOST is skipped", func(t *testing.T) {
		dev := &testDevice{
			getNvLinkState: func(int) (nvml.EnableState, nvml.Return) {
				return nvml.FEATURE_DISABLED, nvml.ERROR_GPU_IS_LOST
			},
			getNvLinkRemotePciInfo: func(int) (nvml.PciInfo, nvml.Return) {
				t.Fatalf("GetNvLinkRemotePciInfo should not be called when state lookup is skipped")
				return nvml.PciInfo{}, nvml.SUCCESS
			},
		}

		pciInfos, err := getAllNvLinkRemotePciInfo(dev)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(pciInfos) != 0 {
			t.Fatalf("expected no pci infos, got %d", len(pciInfos))
		}
	})

	t.Run("GetNvLinkRemotePciInfo ERROR_GPU_IS_LOST skips that link only", func(t *testing.T) {
		const lostLink = 3
		dev := &testDevice{
			getNvLinkState: func(int) (nvml.EnableState, nvml.Return) {
				return nvml.FEATURE_ENABLED, nvml.SUCCESS
			},
			getNvLinkRemotePciInfo: func(i int) (nvml.PciInfo, nvml.Return) {
				if i == lostLink {
					return nvml.PciInfo{}, nvml.ERROR_GPU_IS_LOST
				}
				return pciInfoWithBusID("0000:01:00.0"), nvml.SUCCESS
			},
		}

		pciInfos, err := getAllNvLinkRemotePciInfo(dev)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		want := nvml.NVLINK_MAX_LINKS - 1
		if len(pciInfos) != want {
			t.Fatalf("expected %d pci infos, got %d", want, len(pciInfos))
		}
		for _, info := range pciInfos {
			if info.BusID() != ":01:00.0" {
				t.Errorf("unexpected bus id %q", info.BusID())
			}
		}
	})

	t.Run("unrelated error from GetNvLinkState still surfaces", func(t *testing.T) {
		dev := &testDevice{
			getNvLinkState: func(int) (nvml.EnableState, nvml.Return) {
				return nvml.FEATURE_DISABLED, nvml.ERROR_UNKNOWN
			},
			getNvLinkRemotePciInfo: func(int) (nvml.PciInfo, nvml.Return) {
				t.Fatalf("GetNvLinkRemotePciInfo should not be called when state lookup fails")
				return nvml.PciInfo{}, nvml.SUCCESS
			},
		}

		pciInfos, err := getAllNvLinkRemotePciInfo(dev)
		if err == nil {
			t.Fatalf("expected an error, got nil (pciInfos=%v)", pciInfos)
		}
	})

	t.Run("unrelated error from GetNvLinkRemotePciInfo still surfaces", func(t *testing.T) {
		dev := &testDevice{
			getNvLinkState: func(int) (nvml.EnableState, nvml.Return) {
				return nvml.FEATURE_ENABLED, nvml.SUCCESS
			},
			getNvLinkRemotePciInfo: func(int) (nvml.PciInfo, nvml.Return) {
				return nvml.PciInfo{}, nvml.ERROR_UNKNOWN
			},
		}

		pciInfos, err := getAllNvLinkRemotePciInfo(dev)
		if err == nil {
			t.Fatalf("expected an error, got nil (pciInfos=%v)", pciInfos)
		}
	})
}
