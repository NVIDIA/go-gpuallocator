// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/NVIDIA/go-gpuallocator/internal/gpulib"
)

// Device represents a GPU device as reported by gpulib, including all of its
// Point-to-Point link information.
type Device struct {
	gpulib.DeviceLite
	Index int
	Links map[int][]P2PLink
}

// P2PLink represents a Point-to-Point link between two GPU devices. The link
// is between the Device struct this struct is embedded in and the GPU Device
// contained in the P2PLink struct itself.
type P2PLink struct {
	GPU  *Device
	Type gpulib.P2PLinkType
}

// DeviceSet is used to hold and manipulate a set of unique GPU devices.
type DeviceSet map[string]*Device

// NewDevices creates a list of Devices from all available gpulib.Devices.
func NewDevices() ([]*Device, error) {
	count, ret := gpulib.DeviceGetCount()
	if ret.Value() != gpulib.SUCCESS {
		return nil, fmt.Errorf("error calling gpulib.GetDeviceCount: %v", ret.Error())
	}

	devices := []*Device{}
	for i := 0; i < count; i++ {
		gpulibDevice, ret := gpulib.NewDeviceLite(i)
		if ret.Value() != gpulib.SUCCESS {
			return nil, fmt.Errorf("error creating gpulib.Device %v: %v", i, ret.Error())
		}

		device := Device{
			DeviceLite: gpulibDevice,
			Index:      i,
			Links:      make(map[int][]P2PLink),
		}
		devices = append(devices, &device)
	}

	for i, d1 := range devices {
		for j, d2 := range devices {
			if d1 != d2 {
				p2plink, err := gpulib.GetP2PLink(d1, d2)
				if err != nil {
					return nil, fmt.Errorf("error getting P2PLink for devices (%v, %v): %v", i, j, err)
				}
				if p2plink != gpulib.P2PLinkUnknown {
					d1.Links[d2.Index] = append(d1.Links[d2.Index], P2PLink{d2, p2plink})
				}

				nvlink, err := gpulib.GetNVLink(d1, d2)
				if err != nil {
					return nil, fmt.Errorf("error getting NVLink for devices (%v, %v): %v", i, j, err)
				}
				if nvlink != gpulib.P2PLinkUnknown {
					d1.Links[d2.Index] = append(d1.Links[d2.Index], P2PLink{d2, nvlink})
				}
			}
		}
	}

	return devices, nil
}

// NewDevicesFrom creates a list of Devices from the specific set of GPU uuids passed in.
func NewDevicesFrom(uuids []string) ([]*Device, error) {
	devices, err := NewDevices()
	if err != nil {
		return nil, err
	}

	filtered := []*Device{}

	found := make(map[string]bool)
	for _, uuid := range uuids {
		for _, device := range devices {
			id, ret := device.GetUUID()
			if ret.Value() != gpulib.SUCCESS {
				return nil, fmt.Errorf("failed to get device UUID: %v", ret.Error())
			}
			if id == uuid {
				filtered = append(filtered, device)
				found[id] = true
				break
			}
		}
		if len(filtered) == 0 || !found[uuid] {
			return nil, fmt.Errorf("no device with uuid: %v", uuid)
		}
	}

	return filtered, nil
}

// String returns a compact representation of a Device as string of its index.
func (d *Device) String() string {
	return fmt.Sprintf("%v", d.Index)
}

// Details returns all details of a Device as a multi-line string.
func (d *Device) Details() string {
	uuid, ret := d.GetUUID()
	if ret.Value() != gpulib.SUCCESS {
		uuid = "UNKNOWN"
	}

	var pciBusID string
	pciInfo, ret := d.GetPciInfo()
	if ret.Value() != gpulib.SUCCESS {
		pciBusID = "UNKNOWN"
	} else {
		pciBusID = gpulib.NewPCIBusID(pciInfo).String()
	}

	s := ""
	s += fmt.Sprintf("Device %v:\n", d.Index)
	s += fmt.Sprintf("  UUID: %v\n", uuid)
	s += fmt.Sprintf("  PCI BusID: %v\n", pciBusID)
	s += fmt.Sprintf("  SocketAffinity: %v\n", d.CPUAffinity())
	s += fmt.Sprintf("  Topology: \n")
	for gpu, links := range d.Links {
		s += fmt.Sprintf("    GPU %v Links:\n", gpu)
		for _, link := range links {
			s += fmt.Sprintf("      %v\n", link.Type)
		}
	}

	return strings.TrimSuffix(s, "\n")
}

func (d Device) uuidOrPanic() string {
	uuid, ret := d.GetUUID()
	if ret.Value() != gpulib.SUCCESS {
		log.Panicf("could not get UUID for device: %v\n", ret.Error())
	}
	return uuid
}

// NewDeviceSet creates a new DeviceSet.
func NewDeviceSet(devices ...*Device) DeviceSet {
	set := make(DeviceSet)
	set.Insert(devices...)
	return set
}

// Insert inserts a list of devices into a DeviceSet.
func (ds DeviceSet) Insert(devices ...*Device) {
	for _, device := range devices {
		uuid := device.uuidOrPanic()
		ds[uuid] = device
	}
}

// Delete deletes a list of devices from a DeviceSet.
func (ds DeviceSet) Delete(devices ...*Device) {
	for _, device := range devices {
		uuid := device.uuidOrPanic()
		delete(ds, uuid)
	}
}

// Contains checks if a device is present in a DeviceSet.
func (ds DeviceSet) Contains(device *Device) bool {
	if device == nil {
		return false
	}

	uuid := device.uuidOrPanic()
	_, ok := ds[uuid]
	return ok
}

// ContainsAll checks if a list of devices is present in a DeviceSet.
func (ds DeviceSet) ContainsAll(devices []*Device) bool {
	if len(devices) > len(ds) {
		return false
	}

	for _, d := range devices {
		if !ds.Contains(d) {
			return false
		}
	}

	return true
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
