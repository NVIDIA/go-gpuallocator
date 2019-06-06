// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

// Device represents a GPU device as reported by NVML, including all of its
// Point-to-Point link information.
type Device struct {
	*nvml.Device
	Index int
	Links map[int][]P2PLink
}

// P2PLink represents a Point-to-Point link between two GPU devices. The link
// is between the Device struct this struct is embedded in and the GPU Device
// contained in the P2PLink struct itself.
type P2PLink struct {
	GPU  *Device
	Type nvml.P2PLinkType
}

// DeviceSet is used to hold and manipulate a set of unique GPU devices.
type DeviceSet map[string]*Device

// Create a list of all Devices with P2PLink information prepopulated.
func newDevices() ([]*Device, error) {
	count, err := nvml.GetDeviceCount()
	if err != nil {
		return nil, fmt.Errorf("error calling nvml.GetDeviceCount: %v", err)
	}

	devices := []*Device{}
	for i := 0; i < int(count); i++ {
		device, err := nvml.NewDevice(uint(i))
		if err != nil {
			return nil, fmt.Errorf("error creating nvml.Device %v: %v", i, err)
		}

		devices = append(devices, &Device{device, i, make(map[int][]P2PLink)})
	}

	for i, d1 := range devices {
		for j, d2 := range devices {
			if d1 != d2 {
				p2plink, err := nvml.GetP2PLink(d1.Device, d2.Device)
				if err != nil {
					return nil, fmt.Errorf("error getting P2PLink for devices (%v, %v): %v", i, j, err)
				}
				if p2plink != nvml.P2PLinkUnknown {
					d1.Links[d2.Index] = append(d1.Links[d2.Index], P2PLink{d2, p2plink})
				}

				nvlink, err := nvml.GetNVLink(d1.Device, d2.Device)
				if err != nil {
					return nil, fmt.Errorf("error getting NVLink for devices (%v, %v): %v", i, j, err)
				}
				if nvlink != nvml.P2PLinkUnknown {
					d1.Links[d2.Index] = append(d1.Links[d2.Index], P2PLink{d2, nvlink})
				}
			}
		}
	}

	return devices, nil
}

// String returns a compact representation of a Device as string of its index.
func (d *Device) String() string {
	return fmt.Sprintf("%v", d.Index)
}

// Details returns all details of a Device as a multi-line string.
func (d *Device) Details() string {
	s := ""
	s += fmt.Sprintf("Device %v:\n", d.Index)
	s += fmt.Sprintf("  UUID: %v\n", d.UUID)
	s += fmt.Sprintf("  PCI BusID: %v\n", d.PCI.BusID)
	s += fmt.Sprintf("  SocketAffinity: %v\n", *d.CPUAffinity)
	s += fmt.Sprintf("  Topology: \n")
	for gpu, links := range d.Links {
		s += fmt.Sprintf("    GPU %v Links:\n", gpu)
		for _, link := range links {
			s += fmt.Sprintf("      %v\n", link.Type)
		}
	}

	return strings.TrimSuffix(s, "\n")
}

// NewDeviceSet creates a new DeviceSet.
func NewDeviceSet() DeviceSet {
	return make(DeviceSet)
}

// Insert inserts a list of devices into a DeviceSet.
func (ds DeviceSet) Insert(devices ...*Device) {
	for _, device := range devices {
		ds[device.UUID] = device
	}
}

// Delete deletes a list of devices from a DeviceSet.
func (ds DeviceSet) Delete(devices ...*Device) {
	for _, device := range devices {
		delete(ds, device.UUID)
	}
}

// Contains checks if a device is present in a DeviceSet.
func (ds DeviceSet) Contains(device *Device) bool {
	if device == nil {
		return false
	}

	_, ok := ds[device.UUID]
	return ok
}

// SortedSlice etunrs returns a slice of devices,
// sorted by device index from a DeviceSet.
func (ds DeviceSet) SortedSlice() []*Device {
	devices := make([]*Device, 0, len(ds))

	for _, device := range ds {
		devices = append(devices, device)
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Index < devices[j].Index
	})

	return devices
}
