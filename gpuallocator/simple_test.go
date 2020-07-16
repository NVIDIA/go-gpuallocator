// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package gpuallocator

import (
	"testing"
)

func TestSimpleAllocate(t *testing.T) {
	devices := NewDGX1VoltaNode().Devices()
	policy := NewSimplePolicy()

	tests := []PolicyAllocTest{
		{
			"Allocate 1",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			1,
			[]int{0},
		},
		{
			"Allocate 2",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			2,
			[]int{0, 1},
		},
		{
			"Allocate 3",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			3,
			[]int{0, 1, 2},
		},
		{
			"Allocate 4",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			4,
			[]int{0, 1, 2, 3},
		},
		{
			"Allocate 5",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			5,
			[]int{0, 1, 2, 3, 4},
		},
		{
			"Allocate 6",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			6,
			[]int{0, 1, 2, 3, 4, 5},
		},
		{
			"Allocate 7",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			7,
			[]int{0, 1, 2, 3, 4, 5, 6},
		},
		{
			"Allocate 8",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{},
			8,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
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
			[]int{4, 0},
		},
		{
			"Must include with allocation size 4",
			devices,
			[]int{0, 1, 2, 3, 4, 5, 6, 7},
			[]int{4},
			4,
			[]int{4, 0, 1, 2},
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
	}

	RunPolicyAllocTests(t, policy, tests)
}
