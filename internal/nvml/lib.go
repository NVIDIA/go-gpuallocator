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
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

type DeviceLite interface {
	Device

	IsMigEnabled() (bool, Return)
	GetMigDevices() ([]DeviceLite, Return)
	Path() string
	CPUAffinity() int64
}

// TODO: These composite functions should return error instead of Return
type nvmlDeviceLite struct {
	Device

	mutex sync.Mutex

	uuid     string
	minor    int
	pciInfo  PciInfo
	path     string
	numaNode *int64
}

var _ DeviceLite = (*nvmlDeviceLite)(nil)

func NewDeviceLite(index int) (DeviceLite, Return) {
	device, ret := DeviceGetHandleByIndex(index)
	if ret.Value() != SUCCESS {
		return nil, ret
	}

	lite, ret := newDeviceLite(device)
	if ret.Value() != SUCCESS {
		return nil, ret
	}

	return lite, nvmlReturn(SUCCESS)
}

func (d *nvmlDeviceLite) GetMinorNumber() (int, Return) {
	return d.minor, nvmlReturn(SUCCESS)
}

func (d *nvmlDeviceLite) GetPciInfo() (PciInfo, Return) {
	return d.pciInfo, nvmlReturn(SUCCESS)
}

func (d *nvmlDeviceLite) GetUUID() (string, Return) {
	return d.uuid, nvmlReturn(SUCCESS)
}

// TODO: Should these composite functions return errors instead?
func newDeviceLite(device Device) (DeviceLite, Return) {
	uuid, ret := device.GetUUID()
	if ret.Value() != SUCCESS {
		return nil, ret
	}
	minor, ret := device.GetMinorNumber()
	if ret.Value() != SUCCESS {
		return nil, ret
	}
	pciInfo, ret := device.GetPciInfo()
	if ret.Value() != SUCCESS {
		return nil, ret
	}

	lite := nvmlDeviceLite{
		Device:  device,
		uuid:    uuid,
		minor:   minor,
		pciInfo: pciInfo,
	}

	return &lite, nvmlReturn(SUCCESS)
}

func (d *nvmlDeviceLite) Path() string {
	if d.path == "" {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.path = fmt.Sprintf("/dev/nvidia%d", d.minor)
	}

	return d.path
}

func (d *nvmlDeviceLite) CPUAffinity() int64 {
	if d.numaNode == nil {
		busID := NewPCIBusID(d.pciInfo)
		node, _ := busID.NumaNode()

		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.numaNode = &node
	}

	return *d.numaNode
}

func (d *nvmlDeviceLite) IsMigEnabled() (bool, Return) {
	cm, pm, ret := d.GetMigMode()
	if ret.Value() == ERROR_NOT_SUPPORTED {
		return false, nvmlReturn(SUCCESS)
	}
	if ret.Value() != SUCCESS {
		return false, ret
	}

	return (cm == nvml.DEVICE_MIG_ENABLE) && (cm == pm), nvmlReturn(SUCCESS)
}

func (d *nvmlDeviceLite) GetMigDevices() ([]DeviceLite, Return) {
	c, ret := d.GetMaxMigDeviceCount()
	if ret.Value() != SUCCESS {
		return nil, ret
	}

	var migHandles []DeviceLite
	for i := 0; i < int(c); i++ {
		mig, ret := d.GetMigDeviceHandleByIndex(i)
		if ret.Value() == ERROR_NOT_FOUND {
			continue
		}
		if ret.Value() != SUCCESS {
			return nil, ret
		}

		migLite, ret := newDeviceLite(mig)
		if ret.Value() != SUCCESS {
			return nil, ret
		}
		migHandles = append(migHandles, migLite)
	}

	return migHandles, nvmlReturn(SUCCESS)
}

// PCIBusID is the ID on the PCI bus of a device
type PCIBusID string

// NewPCIBusID provides a utility function that returns the string representation
// of the bus ID.
func NewPCIBusID(p PciInfo) PCIBusID {
	var bytes []byte
	for _, b := range p.BusId {
		if byte(b) == '\x00' {
			break
		}
		bytes = append(bytes, byte(b))
	}
	return PCIBusID(string(bytes))
}

func (p PCIBusID) String() string {
	return string(p)
}

func (p PCIBusID) NumaNode() (int64, error) {
	b, err := ioutil.ReadFile(p.numaNodePath())
	if err != nil {
		return CPUAffinityNotSupported, nil
	}

	node, err := strconv.ParseInt(string(bytes.TrimSpace(b)), 10, 8)
	if err != nil {
		return CPUAffinityNotSupported, fmt.Errorf("failed to parse numa_node contents: %v", err)
	}

	if node < 0 {
		return CPUAffinityNotSupported, nil
	}

	return node, nil
}

// numaNodePath returns the path for the numa_node file associated with the
// PCIBusID
func (p PCIBusID) numaNodePath() string {
	id := strings.ToLower(p.String())

	if strings.HasPrefix(id, "0000") {
		id = id[4:]
	}
	return filepath.Join("/sys/bus/pci/devices", id, "numa_node")
}

func ParseMigDeviceUUID(uuid string) (string, uint32, uint32, error) {
	migDevice, ret := DeviceGetHandleByUUID(uuid)
	if ret.Value() == SUCCESS {
		parent, ret := migDevice.GetDeviceHandleFromMigDeviceHandle()
		if ret.Value() != SUCCESS {
			return "", 0, 0, fmt.Errorf("failed to get parent device handle: %v", ret.Error())
		}
		parentUUID, ret := parent.GetUUID()
		if ret.Value() != SUCCESS {
			return "", 0, 0, fmt.Errorf("failed to get parent device UUID: %v", ret.Error())
		}

		gi, ret := migDevice.GetGpuInstanceId()
		if ret.Value() != SUCCESS {
			return "", 0, 0, fmt.Errorf("failed to get GPU instance ID: %v", ret.Error())
		}
		ci, ret := migDevice.GetComputeInstanceId()
		if ret.Value() != SUCCESS {
			return "", 0, 0, fmt.Errorf("failed to get compute instance ID: %v", ret.Error())
		}

		return parentUUID, uint32(gi), uint32(ci), nil
	}

	return parseMigDeviceUUID(uuid)
}

func parseMigDeviceUUID(mig string) (string, uint32, uint32, error) {
	tokens := strings.SplitN(mig, "-", 2)
	if len(tokens) != 2 || tokens[0] != "MIG" {
		return "", 0, 0, fmt.Errorf("failed to parse UUID as MIG device")
	}

	tokens = strings.SplitN(tokens[1], "/", 3)
	if len(tokens) != 3 || !strings.HasPrefix(tokens[0], "GPU-") {
		return "", 0, 0, fmt.Errorf("failed to parse UUID as MIG device")
	}

	gi, err := strconv.Atoi(tokens[1])
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse UUID as MIG device")
	}

	ci, err := strconv.Atoi(tokens[2])
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to parse UUID as MIG device")
	}

	return tokens[0], uint32(gi), uint32(ci), nil
}

func (e EventData) GetUUID() (string, error) {
	device := nvmlDevice(e.Device)
	if device.Handle == nil {
		return "", nil
	}
	uuid, ret := device.GetUUID()
	if ret.Value() != SUCCESS {
		return "", fmt.Errorf("failed to get UUID for device: %v", ret.Error())
	}

	return uuid, nil
}
