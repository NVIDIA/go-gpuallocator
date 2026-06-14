// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllocatorFreeReturnsAllocatedDevice(t *testing.T) {
	allocator := newAllocatorFrom(New4xRTX8000Node().Devices(), NewSimplePolicy())

	allocated := allocator.Allocate(1)
	require.Len(t, allocated, 1)

	allocator.Free(allocated[0])

	require.False(t, allocator.allocated.Contains(allocated[0]))
	require.True(t, allocator.remaining.Contains(allocated[0]))
	require.Same(t, allocated[0], allocator.remaining[allocated[0].UUID])
}

func TestAllocatorFreeIgnoresUnknownDevice(t *testing.T) {
	allocator := newAllocatorFrom(New4xRTX8000Node().Devices(), NewSimplePolicy())

	allocated := allocator.Allocate(1)
	require.Len(t, allocated, 1)

	unknown := (*Device)(NewTestGPU(99))
	allocator.Free(unknown)

	require.False(t, allocator.remaining.Contains(unknown))
	require.True(t, allocator.allocated.Contains(allocated[0]))
	require.False(t, allocator.remaining.Contains(allocated[0]))
}

func TestAllocatorFreeIgnoresFabricatedAllocatedDevice(t *testing.T) {
	allocator := newAllocatorFrom(New4xRTX8000Node().Devices(), NewSimplePolicy())

	allocated := allocator.Allocate(1)
	require.Len(t, allocated, 1)

	fabricated := (*Device)(NewTestGPU(allocated[0].Index))
	allocator.Free(fabricated)

	require.True(t, allocator.allocated.Contains(allocated[0]))
	require.Same(t, allocated[0], allocator.allocated[allocated[0].UUID])
	require.False(t, allocator.remaining.Contains(allocated[0]))
}
