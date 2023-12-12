// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"sort"
	"testing"
)

func sortGPUSetOfSets(sos [][]*Device) {
	sort.Slice(sos, func(i, j int) bool {
		return calculateGPUSetScore(sos[i]) < calculateGPUSetScore(sos[j])
	})
	for _, set := range sos {
		sortGPUSet(set)
	}
}

func sortGPUSetOfSetOfSets(sosos [][][]*Device) {
	sort.Slice(sosos, func(i, j int) bool {
		return calculateGPUPartitionScore(sosos[i]) < calculateGPUPartitionScore(sosos[j])
	})
	for _, sos := range sosos {
		sortGPUSetOfSets(sos)
	}
}

func TestBestEffortAllocate(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()
	policy := NewBestEffortPolicy()

	tests := []PolicyAllocTest{
		{
			"Single NVLINK, prefer same switch",
			devices,
			[]int{0, 2, 4, 5},
			[]int{},
			2,
			[]int{4, 5},
		},
		{
			"Single NVLINK, prefer same socket",
			devices,
			[]int{0, 4, 5, 6},
			[]int{},
			2,
			[]int{5, 6},
		},
		{
			"Prefer dual NVLINK same socket",
			devices,
			[]int{0, 3, 1},
			[]int{},
			2,
			[]int{0, 3},
		},
		{
			"Prefer dual NVLINK cross socket",
			devices,
			[]int{0, 1, 4},
			[]int{},
			2,
			[]int{0, 4},
		},
		{
			"Must include with allocation size 1",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{4},
			1,
			[]int{4},
		},
		{
			"Must include with allocation size 2",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{4},
			2,
			[]int{4, 7},
		},
		{
			"Must include with allocation size 4",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{4},
			4,
			[]int{4, 5, 6, 7},
		},
		{
			"Must include with unavailable device",
			devices,
			[]int{1, 2, 3, 4, 5, 6, 7},
			[]int{0, 1, 2, 3},
			4,
			[]int{},
		},
		{
			"Must include with full set specified",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{0, 2, 4, 6},
			4,
			[]int{0, 2, 4, 6},
		},
		{
			"Must include with too many devices specified",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{1, 2, 3, 4, 5, 6},
			4,
			[]int{},
		},
		{
			"Required too many devices than available",
			devices,
			[]int{1, 2, 3, 4, 5},
			[]int{1, 2, 3, 4, 5, 6},
			1,
			[]int{},
		},
	}

	RunPolicyAllocTests(t, policy, tests)
}

func TestBestEffortIterateGPUSets(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()

	type IterateGPUSetsTest struct {
		description string
		devices     []*Device
		input       []int
		size        int
		result      [][]int
	}

	tests := []IterateGPUSetsTest{
		{
			"Iterate 0",
			devices,
			[]int{0, 1, 2, 3},
			0,
			[][]int{},
		},
		{
			"Iterate 1",
			devices,
			[]int{0, 1, 2, 3},
			1,
			[][]int{{0}, {1}, {2}, {3}},
		},
		{
			"Iterate 2",
			devices,
			[]int{0, 1, 2, 3},
			2,
			[][]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}},
		},
		{
			"Iterate 3",
			devices,
			[]int{0, 1, 2, 3},
			3,
			[][]int{{0, 1, 2}, {0, 1, 3}, {0, 2, 3}, {1, 2, 3}},
		},
		{
			"Iterate 4",
			devices,
			[]int{0, 1, 2, 3},
			4,
			[][]int{{0, 1, 2, 3}},
		},
		{
			"Iterate Too Many",
			devices,
			[]int{0, 1, 2, 3},
			5,
			[][]int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			input := GetDevicesFromIndices(tc.devices, tc.input)

			result := make([][]*Device, len(tc.result))
			for i, indices := range tc.result {
				result[i] = GetDevicesFromIndices(tc.devices, indices)
			}

			var accumulator [][]*Device
			iterateGPUSets(input, tc.size, func(set []*Device) {
				set = append([]*Device{}, set...) // Requires a copy out
				accumulator = append(accumulator, set)
			})

			if len(accumulator) != len(result) {
				t.Errorf("got %v, want %v", accumulator, result)
				return
			}

			for i := range accumulator {
				if len(accumulator[i]) != len(result[i]) {
					t.Errorf("got %v, want %v", accumulator, result)
					return
				}
			}

			sortGPUSetOfSets(accumulator)
			sortGPUSetOfSets(result)

			for i := range accumulator {
				for _, device := range accumulator[i] {
					if !gpuSetContains(result[i], device) {
						t.Errorf("got %v, want %v", accumulator, result)
						return
					}
				}
			}
		})
	}
}

func TestBestEffortIterateGPUPartitions(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()

	type IterateGPUPartitionsTest struct {
		description string
		devices     []*Device
		input       []int
		size        int
		result      [][][]int
	}

	tests := []IterateGPUPartitionsTest{
		{
			"Iterate 0",
			devices,
			[]int{0, 1, 2, 3},
			0,
			[][][]int{},
		},
		{
			"Iterate 1",
			devices,
			[]int{0, 1, 2, 3},
			1,
			[][][]int{{{0}}, {{1}}, {{2}}, {{3}}},
		},
		{
			"Iterate 2",
			devices,
			[]int{0, 1, 2, 3},
			2,
			[][][]int{
				{{0, 1}, {2, 3}},
				{{0, 2}, {1, 3}},
				{{0, 3}, {1, 2}},
			},
		},
		{
			"Iterate 3",
			devices,
			[]int{0, 1, 2, 3},
			3,
			[][][]int{
				{{0, 1, 2}, {3, pad, pad}},
				{{0, 1, 3}, {2, pad, pad}},
				{{0, 2, 3}, {1, pad, pad}},
				{{1, 2, 3}, {0, pad, pad}},
			},
		},
		{
			"Iterate 4",
			devices,
			[]int{0, 1, 2, 3},
			4,
			[][][]int{{{0, 1, 2, 3}}},
		},
		{
			"Iterate Too Many",
			devices,
			[]int{0, 1, 2, 3},
			5,
			[][][]int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			input := GetDevicesFromIndices(tc.devices, tc.input)

			result := make([][][]*Device, len(tc.result))
			for i, partition := range tc.result {
				result[i] = make([][]*Device, len(partition))
				for j, indices := range partition {
					result[i][j] = GetDevicesFromIndices(tc.devices, indices)
				}
			}

			var accumulator [][][]*Device
			iterateGPUPartitions(input, tc.size, func(partition [][]*Device) {
				partition = append([][]*Device{}, partition...) // Requires a copy out
				accumulator = append(accumulator, partition)
			})

			if len(accumulator) != len(result) {
				t.Errorf("got %v, want %v", accumulator, result)
				return
			}

			for i := range accumulator {
				if len(accumulator[i]) != len(result[i]) {
					t.Errorf("got %v, want %v", accumulator, result)
					return
				}
			}

			for i := range accumulator {
				for j := range accumulator[i] {
					if len(accumulator[i][j]) != len(result[i][j]) {
						t.Errorf("got %v, want %v", accumulator, result)
						return
					}
				}
			}

			sortGPUSetOfSetOfSets(accumulator)
			sortGPUSetOfSetOfSets(result)

			for i := range accumulator {
				for j, set := range accumulator[i] {
					for _, device := range set {
						if !gpuSetContains(result[i][j], device) {
							t.Errorf("got %v, want %v", accumulator, result)
							return
						}
					}
				}
			}
		})
	}
}

func TestBestEffort4xRTX8000GPUAllocOne(t *testing.T) {
	devices := New4xRTX8000Node().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{1, []int{0}},
		{1, []int{1}},
		{1, []int{2}},
		{1, []int{3}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffort4xRTX8000GPUAllocTwo(t *testing.T) {
	devices := New4xRTX8000Node().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{2, []int{0, 3}},
		{2, []int{1, 2}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffort4xRTX8000GPUAllocFour(t *testing.T) {
	devices := New4xRTX8000Node().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{4, []int{0, 1, 2, 3}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1PascalGPUAllocOne(t *testing.T) {
	devices := NewDGX1PascalNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{1, []int{0}},
		{1, []int{1}},
		{1, []int{2}},
		{1, []int{3}},
		{1, []int{4}},
		{1, []int{5}},
		{1, []int{6}},
		{1, []int{7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1PascalGPUAllocTwo(t *testing.T) {
	devices := NewDGX1PascalNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{2, []int{0, 2}},
		{2, []int{1, 3}},
		{2, []int{4, 6}},
		{2, []int{5, 7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1PascalGPUAllocFour(t *testing.T) {
	devices := NewDGX1PascalNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{4, []int{0, 1, 2, 3}},
		{4, []int{4, 5, 6, 7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1PascalGPUAllocEight(t *testing.T) {
	devices := NewDGX1PascalNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{8, []int{0, 1, 2, 3, 4, 5, 6, 7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1VoltaGPUAllocOne(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{1, []int{0}},
		{1, []int{1}},
		{1, []int{2}},
		{1, []int{3}},
		{1, []int{4}},
		{1, []int{5}},
		{1, []int{6}},
		{1, []int{7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1VoltaGPUAllocTwo(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{2, []int{0, 3}},
		{2, []int{1, 2}},
		{2, []int{4, 7}},
		{2, []int{5, 6}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1VoltaGPUAllocFour(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{4, []int{0, 1, 2, 3}},
		{4, []int{4, 5, 6, 7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}

func TestBestEffortDGX1VoltaGPUAllocEight(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()
	policy := NewBestEffortPolicy()
	allocator := newAllocatorFrom(devices, policy)

	tests := []AllocTest{
		{8, []int{0, 1, 2, 3, 4, 5, 6, 7}},
		{1, []int{}},
	}

	RunAllocTests(t, allocator, tests)
}
