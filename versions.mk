# Copyright (c) NVIDIA CORPORATION.  All rights reserved.
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

GIT_TAG ?= $(patsubst v%,%,$(shell git describe --tags 2>/dev/null))
GIT_COMMIT ?= $(shell git describe --match="" --dirty --long --always --abbrev=40 2> /dev/null || echo "")

MODULE := github.com/NVIDIA/go-gpuallocator

REGISTRY ?= nvcr.io/nvidia/cloud-native

VERSION  ?= $(GIT_TAG)

GOLANG_VERSION ?= 1.24.4

BUILDIMAGE_TAG ?= $(GOLANG_VERSION)-bookworm
BUILDIMAGE ?=  golang:$(BUILDIMAGE_TAG)
