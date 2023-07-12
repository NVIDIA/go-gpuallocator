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
	"gitlab.com/nvidia/cloud-native/go-nvlib/pkg/nvml"
)

type EventData nvml.EventData
type EventSet nvml.EventSet

func EventSetCreate() (EventSet, Return) {
	e1, ret := nvml.EventSetCreate()
	return EventSet(e1), nvmlReturn(ret)
}

func (e EventSet) Free() Return {
	ret := nvml.EventSet(e).Free()
	return nvmlReturn(ret)
}

func (e EventSet) Wait(timeoutms uint32) (EventData, Return) {
	d1, ret := nvml.EventSet(e).Wait(timeoutms)
	return EventData(d1), nvmlReturn(ret)
}
