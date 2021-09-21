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
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type P2PLinkType uint

const (
	P2PLinkUnknown P2PLinkType = iota
	P2PLinkCrossCPU
	P2PLinkSameCPU
	P2PLinkHostBridge
	P2PLinkMultiSwitch
	P2PLinkSingleSwitch
	P2PLinkSameBoard
	SingleNVLINKLink
	TwoNVLINKLinks
	ThreeNVLINKLinks
	FourNVLINKLinks
	FiveNVLINKLinks
	SixNVLINKLinks
	SevenNVLINKLinks
	EightNVLINKLinks
	NineNVLINKLinks
	TenNVLINKLinks
	ElevenNVLINKLinks
	TwelveNVLINKLinks
)

func GetP2PLink(dev1 nvml.Device, dev2 nvml.Device) (P2PLinkType, error) {
	level, ret := dev1.GetTopologyCommonAncestor(dev2)
	if ret != nvml.SUCCESS {
		return P2PLinkUnknown, fmt.Errorf("failed to get common ancestor: %v", nvml.ErrorString(ret))
	}

	var link P2PLinkType

	switch level {
	case TOPOLOGY_INTERNAL:
		link = P2PLinkSameBoard
	case TOPOLOGY_SINGLE:
		link = P2PLinkSingleSwitch
	case TOPOLOGY_MULTIPLE:
		link = P2PLinkMultiSwitch
	case TOPOLOGY_HOSTBRIDGE:
		link = P2PLinkHostBridge
	case TOPOLOGY_NODE: // NOTE: TOPOLOGY_CPU was renamed TOPOLOGY_NODE
		link = P2PLinkSameCPU
	case TOPOLOGY_SYSTEM:
		link = P2PLinkCrossCPU
	default:
		return P2PLinkUnknown, fmt.Errorf("unsupported topology level: %v", level)
	}

	return link, nil
}

func GetNVLink(dev1 nvml.Device, dev2 nvml.Device) (link P2PLinkType, err error) {
	pciInfo2, ret := dev2.GetPciInfo()
	if ret != nvml.SUCCESS {
		return P2PLinkUnknown, fmt.Errorf("failed to get PciInfo: %v", nvml.ErrorString(ret))
	}

	pciInfos, err := deviceGetAllNvLinkRemotePciInfo(dev1)
	if err != nil {
		return P2PLinkUnknown, err
	}

	nvlink := P2PLinkUnknown
	for _, pciInfo1 := range pciInfos {
		if pciInfo1.BusId == pciInfo2.BusId {
			nvlink = nvlink.add()
		}
	}

	// TODO(klueska): Handle NVSwitch semantics

	return nvlink, nil
}

func (l P2PLinkType) add() P2PLinkType {
	if l == P2PLinkUnknown {
		return SingleNVLINKLink
	}
	if l == TwelveNVLINKLinks {
		return TwelveNVLINKLinks
	}
	return l + 1
}

func deviceGetAllNvLinkRemotePciInfo(dev nvml.Device) ([]nvml.PciInfo, error) {
	var pciInfos []nvml.PciInfo
	for i := 0; i < nvml.NVLINK_MAX_LINKS; i++ {
		state, ret := dev.GetNvLinkState(i)
		if ret == nvml.ERROR_NOT_SUPPORTED || ret == nvml.ERROR_INVALID_ARGUMENT {
			continue
		}
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("failed to query link %d state: %v", i, nvml.ErrorString(ret))
		}
		if state != nvml.FEATURE_ENABLED {
			continue
		}

		info, ret := dev.GetNvLinkRemotePciInfo(i)
		if ret == nvml.ERROR_NOT_SUPPORTED || ret == nvml.ERROR_INVALID_ARGUMENT {
			continue
		}
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("failed to query remote link %d: %v", i, nvml.ErrorString(ret))
		}
		pciInfos = append(pciInfos, info)
	}

	return pciInfos, nil

}
