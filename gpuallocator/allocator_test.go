// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/NVIDIA/go-nvml/pkg/nvml/mock"
	"github.com/stretchr/testify/require"
)

func TestAllocatorNVMLLifecycle(t *testing.T) {
	constructors := map[string]func() (*Allocator, error){
		"custom":      func() (*Allocator, error) { return NewAllocator(NewSimplePolicy()) },
		"simple":      NewSimpleAllocator,
		"best effort": NewBestEffortAllocator,
	}
	testCases := map[string]struct {
		failedInitCall int32
		countError     nvml.Return
		expectedError  string
	}{
		"initialization failure": {
			failedInitCall: 1,
			expectedError:  "error initializing NVML",
		},
		"device list initialization failure": {
			failedInitCall: 2,
			expectedError:  "error enumerating GPU devices",
		},
		"device enumeration failure": {
			countError:    nvml.ERROR_UNKNOWN,
			expectedError: "error enumerating GPU devices",
		},
		"successful construction": {},
	}
	for constructorName, construct := range constructors {
		for name, tc := range testCases {
			t.Run(constructorName+"/"+name, func(t *testing.T) {
				var initCalls, successfulInits, shutdownCalls atomic.Int32
				nvmllib := &mock.Interface{
					InitFunc: func() nvml.Return {
						if initCalls.Add(1) == tc.failedInitCall {
							return nvml.ERROR_UNKNOWN
						}
						successfulInits.Add(1)
						return nvml.SUCCESS
					},
					ShutdownFunc: func() nvml.Return {
						shutdownCalls.Add(1)
						return nvml.SUCCESS
					},
					DeviceGetCountFunc: func() (int, nvml.Return) {
						return 0, tc.countError
					},
				}
				t.Cleanup(setNVMLNewDuringTest(nvmllib))
				allocator, err := construct()
				if tc.expectedError != "" {
					require.ErrorContains(t, err, tc.expectedError)
					require.Nil(t, allocator)
					require.Equal(t, successfulInits.Load(), shutdownCalls.Load())
					return
				}
				require.NoError(t, err)
				require.NotNil(t, allocator)
				// The live allocator keeps one NVML reference for its devices.
				require.EqualValues(t, 1, successfulInits.Load()-shutdownCalls.Load())
				runtime.KeepAlive(allocator)
			})
		}
	}
}

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
