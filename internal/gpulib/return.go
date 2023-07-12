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

type Return interface {
	Value() nvml.Return
	String() string
	Error() string
}

type nvmlReturn nvml.Return

var _ Return = (*nvmlReturn)(nil)

func (r nvmlReturn) Value() nvml.Return {
	return nvml.Return(r)
}

func (r nvmlReturn) String() string {
	return r.Error()
}

func (r nvmlReturn) Error() string {
	return nvml.ErrorString(nvml.Return(r))
}
