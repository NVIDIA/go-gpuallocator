/**
# Copyright (c) 2023, NVIDIA CORPORATION.  All rights reserved.
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

package gpulib

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

func Init() Return {
	return nvmlReturn(nvml.Init())
}

func Shutdown() Return {
	return nvmlReturn(nvml.Shutdown())
}

func DeviceGetCount() (int, Return) {
	count, ret := nvml.DeviceGetCount()
	return count, nvmlReturn(ret)
}

func DeviceGetHandleByUUID(uuid string) (Device, Return) {
	d1, ret := nvml.DeviceGetHandleByUUID(uuid)
	return nvmlDevice(d1), nvmlReturn(ret)
}
