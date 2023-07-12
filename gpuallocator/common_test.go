// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"fmt"
	"sort"
	"testing"

	"github.com/NVIDIA/go-gpuallocator/internal/gpulib"
)

const pad = ^int(0)

type TestGPU Device
type TestNode []*TestGPU

type AllocTest struct {
	size   int
	result []int
}

type PolicyAllocTest struct {
	description string
	devices     []*Device
	available   []int
	required    []int
	size        int
	result      []int
}

func GetDevicesFromIndices(allDevices []*Device, indices []int) []*Device {
	var input []*Device
	for _, i := range indices {
		for _, device := range allDevices {
			if i == pad {
				input = append(input, nil)
				break
			}
			if i == device.Index {
				input = append(input, device)
				break
			}
		}
	}
	return input
}

func sortGPUSet(set []*Device) {
	sort.Slice(set, func(i, j int) bool {
		if set[i] == nil {
			return true
		}
		if set[j] == nil {
			return false
		}
		return set[i].Index < set[j].Index
	})
}

func RunAllocTests(t *testing.T, allocator *Allocator, tests []AllocTest) {
	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			result := GetDevicesFromIndices(allocator.GPUs, tc.result)

			allocated := allocator.Allocate(tc.size)
			if len(allocated) != len(tc.result) {
				t.Errorf("got %v, want %v", allocated, tc.result)
				return
			}

			sortGPUSet(allocated)
			sortGPUSet(result)

			for _, device := range allocated {
				if !NewDeviceSet(result...).Contains(device) {
					t.Errorf("got %v, want %v", allocated, tc.result)
					break
				}
			}
		})
	}
}

func RunPolicyAllocTests(t *testing.T, policy Policy, tests []PolicyAllocTest) {
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			available := GetDevicesFromIndices(tc.devices, tc.available)
			required := GetDevicesFromIndices(tc.devices, tc.required)
			result := GetDevicesFromIndices(tc.devices, tc.result)

			allocated := policy.Allocate(available, required, tc.size)
			if len(allocated) != len(tc.result) {
				t.Errorf("got %v, want %v", allocated, tc.result)
				return
			}

			sortGPUSet(allocated)
			sortGPUSet(result)

			for _, device := range allocated {
				if !NewDeviceSet(result...).Contains(device) {
					t.Errorf("got %v, want %v", allocated, tc.result)
					break
				}
			}
		})
	}
}

func NewTestGPU(index int) *TestGPU {
	id := fmt.Sprintf("GPU-%d", index)

	mockedSuccess := new(gpulib.Return)
	mockedDevice := &gpulib.DeviceLiteMock{
		GetUUIDFunc: func() (string, gpulib.Return) {
			return id, *mockedSuccess
		},
	}
	return &TestGPU{
		Index:      index,
		DeviceLite: mockedDevice,
		Links:      make(map[int][]P2PLink),
	}
}

func (from *TestGPU) AddLink(to *TestGPU, linkType gpulib.P2PLinkType) {
	link := P2PLink{(*Device)(to), linkType}
	from.Links[to.Index] = append(from.Links[to.Index], link)
}

func (n TestNode) AddLink(from, to int, linkType gpulib.P2PLinkType) {
	n[from].AddLink(n[to], linkType)
}

func (n TestNode) Devices() []*Device {
	var devices []*Device
	for _, gpu := range n {
		devices = append(devices, (*Device)(gpu))
	}
	return devices
}

func New4xRTX8000Node() TestNode {
	node := TestNode{
		NewTestGPU(0),
		NewTestGPU(1),
		NewTestGPU(2),
		NewTestGPU(3),
	}

	// NVLinks
	node.AddLink(0, 3, gpulib.TwoNVLINKLinks)
	node.AddLink(1, 2, gpulib.TwoNVLINKLinks)
	node.AddLink(2, 1, gpulib.TwoNVLINKLinks)
	node.AddLink(3, 0, gpulib.TwoNVLINKLinks)

	// P2PLinks
	node.AddLink(0, 1, gpulib.P2PLinkSameCPU)
	node.AddLink(0, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 0, gpulib.P2PLinkSameCPU)
	node.AddLink(1, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 3, gpulib.P2PLinkSameCPU)
	node.AddLink(3, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 2, gpulib.P2PLinkSameCPU)

	return node
}

func NewDGX1PascalNode() TestNode {
	node := TestNode{
		NewTestGPU(0),
		NewTestGPU(1),
		NewTestGPU(2),
		NewTestGPU(3),
		NewTestGPU(4),
		NewTestGPU(5),
		NewTestGPU(6),
		NewTestGPU(7),
	}

	// NVLinks
	node.AddLink(0, 1, gpulib.SingleNVLINKLink)
	node.AddLink(0, 2, gpulib.SingleNVLINKLink)
	node.AddLink(0, 3, gpulib.SingleNVLINKLink)
	node.AddLink(0, 4, gpulib.SingleNVLINKLink)

	node.AddLink(1, 0, gpulib.SingleNVLINKLink)
	node.AddLink(1, 2, gpulib.SingleNVLINKLink)
	node.AddLink(1, 3, gpulib.SingleNVLINKLink)
	node.AddLink(1, 5, gpulib.SingleNVLINKLink)

	node.AddLink(2, 0, gpulib.SingleNVLINKLink)
	node.AddLink(2, 1, gpulib.SingleNVLINKLink)
	node.AddLink(2, 3, gpulib.SingleNVLINKLink)
	node.AddLink(2, 6, gpulib.SingleNVLINKLink)

	node.AddLink(3, 0, gpulib.SingleNVLINKLink)
	node.AddLink(3, 1, gpulib.SingleNVLINKLink)
	node.AddLink(3, 2, gpulib.SingleNVLINKLink)
	node.AddLink(3, 7, gpulib.SingleNVLINKLink)

	node.AddLink(4, 0, gpulib.SingleNVLINKLink)
	node.AddLink(4, 5, gpulib.SingleNVLINKLink)
	node.AddLink(4, 6, gpulib.SingleNVLINKLink)
	node.AddLink(4, 7, gpulib.SingleNVLINKLink)

	node.AddLink(5, 1, gpulib.SingleNVLINKLink)
	node.AddLink(5, 4, gpulib.SingleNVLINKLink)
	node.AddLink(5, 6, gpulib.SingleNVLINKLink)
	node.AddLink(5, 7, gpulib.SingleNVLINKLink)

	node.AddLink(6, 2, gpulib.SingleNVLINKLink)
	node.AddLink(6, 4, gpulib.SingleNVLINKLink)
	node.AddLink(6, 5, gpulib.SingleNVLINKLink)
	node.AddLink(6, 7, gpulib.SingleNVLINKLink)

	node.AddLink(7, 3, gpulib.SingleNVLINKLink)
	node.AddLink(7, 4, gpulib.SingleNVLINKLink)
	node.AddLink(7, 5, gpulib.SingleNVLINKLink)
	node.AddLink(7, 6, gpulib.SingleNVLINKLink)

	// P2PLinks
	node.AddLink(0, 1, gpulib.P2PLinkHostBridge)
	node.AddLink(0, 2, gpulib.P2PLinkSingleSwitch)
	node.AddLink(0, 3, gpulib.P2PLinkHostBridge)
	node.AddLink(0, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(0, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(0, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(0, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(1, 0, gpulib.P2PLinkHostBridge)
	node.AddLink(1, 2, gpulib.P2PLinkHostBridge)
	node.AddLink(1, 3, gpulib.P2PLinkSingleSwitch)
	node.AddLink(1, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(2, 0, gpulib.P2PLinkSingleSwitch)
	node.AddLink(2, 1, gpulib.P2PLinkHostBridge)
	node.AddLink(2, 3, gpulib.P2PLinkHostBridge)
	node.AddLink(2, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(3, 0, gpulib.P2PLinkHostBridge)
	node.AddLink(3, 1, gpulib.P2PLinkSingleSwitch)
	node.AddLink(3, 2, gpulib.P2PLinkHostBridge)
	node.AddLink(3, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(4, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 5, gpulib.P2PLinkHostBridge)
	node.AddLink(4, 6, gpulib.P2PLinkSingleSwitch)
	node.AddLink(4, 7, gpulib.P2PLinkHostBridge)

	node.AddLink(5, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 4, gpulib.P2PLinkHostBridge)
	node.AddLink(5, 6, gpulib.P2PLinkHostBridge)
	node.AddLink(5, 7, gpulib.P2PLinkSingleSwitch)

	node.AddLink(6, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 4, gpulib.P2PLinkSingleSwitch)
	node.AddLink(6, 5, gpulib.P2PLinkHostBridge)
	node.AddLink(6, 7, gpulib.P2PLinkHostBridge)

	node.AddLink(7, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 4, gpulib.P2PLinkHostBridge)
	node.AddLink(7, 5, gpulib.P2PLinkSingleSwitch)
	node.AddLink(7, 6, gpulib.P2PLinkHostBridge)

	return node
}

func NewDGX1VoltaNode() TestNode {
	node := TestNode{
		NewTestGPU(0),
		NewTestGPU(1),
		NewTestGPU(2),
		NewTestGPU(3),
		NewTestGPU(4),
		NewTestGPU(5),
		NewTestGPU(6),
		NewTestGPU(7),
	}

	// NVLinks
	node.AddLink(0, 1, gpulib.SingleNVLINKLink)
	node.AddLink(0, 2, gpulib.SingleNVLINKLink)
	node.AddLink(0, 3, gpulib.TwoNVLINKLinks)
	node.AddLink(0, 4, gpulib.TwoNVLINKLinks)

	node.AddLink(1, 0, gpulib.SingleNVLINKLink)
	node.AddLink(1, 2, gpulib.TwoNVLINKLinks)
	node.AddLink(1, 3, gpulib.SingleNVLINKLink)
	node.AddLink(1, 5, gpulib.TwoNVLINKLinks)

	node.AddLink(2, 0, gpulib.SingleNVLINKLink)
	node.AddLink(2, 1, gpulib.TwoNVLINKLinks)
	node.AddLink(2, 3, gpulib.TwoNVLINKLinks)
	node.AddLink(2, 6, gpulib.SingleNVLINKLink)

	node.AddLink(3, 0, gpulib.TwoNVLINKLinks)
	node.AddLink(3, 1, gpulib.SingleNVLINKLink)
	node.AddLink(3, 2, gpulib.TwoNVLINKLinks)
	node.AddLink(3, 7, gpulib.SingleNVLINKLink)

	node.AddLink(4, 0, gpulib.TwoNVLINKLinks)
	node.AddLink(4, 5, gpulib.SingleNVLINKLink)
	node.AddLink(4, 6, gpulib.SingleNVLINKLink)
	node.AddLink(4, 7, gpulib.TwoNVLINKLinks)

	node.AddLink(5, 1, gpulib.TwoNVLINKLinks)
	node.AddLink(5, 4, gpulib.SingleNVLINKLink)
	node.AddLink(5, 6, gpulib.TwoNVLINKLinks)
	node.AddLink(5, 7, gpulib.SingleNVLINKLink)

	node.AddLink(6, 2, gpulib.SingleNVLINKLink)
	node.AddLink(6, 4, gpulib.SingleNVLINKLink)
	node.AddLink(6, 5, gpulib.TwoNVLINKLinks)
	node.AddLink(6, 7, gpulib.TwoNVLINKLinks)

	node.AddLink(7, 3, gpulib.SingleNVLINKLink)
	node.AddLink(7, 4, gpulib.TwoNVLINKLinks)
	node.AddLink(7, 5, gpulib.SingleNVLINKLink)
	node.AddLink(7, 6, gpulib.TwoNVLINKLinks)

	// P2PLinks
	node.AddLink(0, 1, gpulib.P2PLinkSingleSwitch)
	node.AddLink(0, 2, gpulib.P2PLinkHostBridge)
	node.AddLink(0, 3, gpulib.P2PLinkHostBridge)
	node.AddLink(0, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(0, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(0, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(0, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(1, 0, gpulib.P2PLinkSingleSwitch)
	node.AddLink(1, 2, gpulib.P2PLinkHostBridge)
	node.AddLink(1, 3, gpulib.P2PLinkHostBridge)
	node.AddLink(1, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(1, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(2, 0, gpulib.P2PLinkHostBridge)
	node.AddLink(2, 1, gpulib.P2PLinkHostBridge)
	node.AddLink(2, 3, gpulib.P2PLinkSingleSwitch)
	node.AddLink(2, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(2, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(3, 0, gpulib.P2PLinkHostBridge)
	node.AddLink(3, 1, gpulib.P2PLinkHostBridge)
	node.AddLink(3, 2, gpulib.P2PLinkSingleSwitch)
	node.AddLink(3, 4, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 5, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 6, gpulib.P2PLinkCrossCPU)
	node.AddLink(3, 7, gpulib.P2PLinkCrossCPU)

	node.AddLink(4, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(4, 5, gpulib.P2PLinkSingleSwitch)
	node.AddLink(4, 6, gpulib.P2PLinkHostBridge)
	node.AddLink(4, 7, gpulib.P2PLinkHostBridge)

	node.AddLink(5, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(5, 4, gpulib.P2PLinkSingleSwitch)
	node.AddLink(5, 6, gpulib.P2PLinkHostBridge)
	node.AddLink(5, 7, gpulib.P2PLinkHostBridge)

	node.AddLink(6, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(6, 4, gpulib.P2PLinkHostBridge)
	node.AddLink(6, 5, gpulib.P2PLinkHostBridge)
	node.AddLink(6, 7, gpulib.P2PLinkSingleSwitch)

	node.AddLink(7, 0, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 1, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 2, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 3, gpulib.P2PLinkCrossCPU)
	node.AddLink(7, 4, gpulib.P2PLinkHostBridge)
	node.AddLink(7, 5, gpulib.P2PLinkHostBridge)
	node.AddLink(7, 6, gpulib.P2PLinkSingleSwitch)

	return node
}

func NewDGX2VoltaNode() TestNode {
	return nil
}
