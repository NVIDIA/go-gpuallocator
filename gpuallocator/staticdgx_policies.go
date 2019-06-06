// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

// GPUType represents the valid set of GPU
// types a Static DGX policy can be created for.
type GPUType int

// Valid GPUTypes
const (
	GPUTypePascal GPUType = iota // Pascal GPUs
	GPUTypeVolta
)

// Policy Definitions
type staticDGX1PascalPolicy struct{}
type staticDGX1VoltaPolicy struct{}
type staticDGX2VoltaPolicy struct{}

// NewStaticDGX1Policy creates a new StaticDGX1Policy for gpuType.
func NewStaticDGX1Policy(gpuType GPUType) Policy {
	if gpuType == GPUTypePascal {
		return &staticDGX1PascalPolicy{}
	}
	if gpuType == GPUTypeVolta {
		return &staticDGX1VoltaPolicy{}
	}
	return nil
}

// NewStaticDGX2Policy creates a new StaticDGX2Policy.
func NewStaticDGX2Policy() Policy {
	return &staticDGX1VoltaPolicy{}
}

// Allocate GPUs following the Static DGX-1 policy for Pascal GPUs.
func (p *staticDGX1PascalPolicy) Allocate(devices []*Device, size int) []*Device {
	if size <= 0 {
		return []*Device{}
	}

	if len(devices) < size {
		return []*Device{}
	}

	validSets := map[int][][]int{
		1: {{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}},
		2: {{0, 2}, {1, 3}, {4, 6}, {5, 7}},
		4: {{0, 1, 2, 3}, {4, 5, 6, 7}},
		8: {{0, 1, 2, 3, 4, 5, 6, 7}},
	}

	return findGPUSet(devices, size, validSets[size])
}

// Allocate GPUs following the Static DGX-1 policy for Volta GPUs.
func (p *staticDGX1VoltaPolicy) Allocate(devices []*Device, size int) []*Device {
	if size <= 0 {
		return []*Device{}
	}

	if len(devices) < size {
		return []*Device{}
	}

	validSets := map[int][][]int{
		1: {{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}},
		2: {{0, 3}, {1, 2}, {4, 7}, {5, 6}},
		4: {{0, 1, 2, 3}, {4, 5, 6, 7}},
		8: {{0, 1, 2, 3, 4, 5, 6, 7}},
	}

	return findGPUSet(devices, size, validSets[size])
}

// Allocate GPUs following the Static DGX-2 policy for Volta GPUs.
func (p *staticDGX2VoltaPolicy) Allocate(devices []*Device, size int) []*Device {
	if size <= 0 {
		return []*Device{}
	}

	if len(devices) < size {
		return []*Device{}
	}

	validSets := map[int][][]int{
		1:  {{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}, {14}, {15}},
		2:  {{0, 1}, {2, 3}, {4, 5}, {6, 7}, {8, 9}, {10, 11}, {12, 13}, {14, 15}},
		4:  {{0, 1, 2, 3}, {4, 5, 6, 7}, {8, 9, 10, 11}, {12, 13, 14, 15}},
		8:  {{0, 1, 2, 3, 4, 5, 6, 7}, {8, 9, 10, 11, 12, 13, 14, 15}},
		16: {{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}},
	}

	return findGPUSet(devices, size, validSets[size])
}

// Find a GPU set of size 'size' in the list of devices that is contained in 'validSets'.
func findGPUSet(devices []*Device, size int, validSets [][]int) []*Device {
	solutionSet := []*Device{}

	for _, validSet := range validSets {
		for _, i := range validSet {
			for _, device := range devices {
				if device.Index == i {
					solutionSet = append(solutionSet, device)
					break
				}
			}
		}

		if len(solutionSet) == size {
			break
		}

		solutionSet = []*Device{}
	}

	return solutionSet
}
