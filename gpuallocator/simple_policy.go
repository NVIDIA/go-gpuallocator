// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

type simplePolicy struct{}

// NewSimplePolicy creates a new SimplePolicy.
func NewSimplePolicy() Policy {
	return &simplePolicy{}
}

// Allocate GPUs following a simple policy.
func (p *simplePolicy) Allocate(devices []*Device, size int) []*Device {
	if size <= 0 {
		return []*Device{}
	}

	if len(devices) < size {
		return []*Device{}
	}

	return append([]*Device{}, devices[:size]...)
}
