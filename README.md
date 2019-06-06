The `gpuallocator` package
-----------------------
The `gpuallocator` package provides a generic abstraction for performing GPU
allocations independent of the larger system the `gpuallocator` is integrated
with.

The abstractions provided by this package are not meant to do actual
"allocation" of GPUs to any specific entity, but rather run the algorithm
responsible for deciding which GPUs should be chosen for allocation based on
the set of GPUs available in the system and the number of GPUs being requested.

Different policies can be hooked in to run different allocation algorithms
depending on the specific needs of the system.

For example, a system like Kubernetes would use this package to help decide
which subset of GPUs to hand out to a container once it has decided how many
GPUs it should be granted. The policy it would choose would be based on the
topological ordering of GPUs to ensure optimal affinity for groups of GPUs
allocated to the same container.

The primary object provided by this package is the `Allocator` object, and the
primary interface used to decide how allocation should actually occur is called
`Policy`. 

More details on each of these can be found below.

The `Allocator` Object
----------------------
The primary object provided by the `gpuallocator` package is that of an
`Allocator`.

A new `Allocator` can be instantiated as follows:

```
func NewAllocator(policy Policy) (*Allocator, error)
```

Once instantiated, an `Allocator` relies on NVML to do GPU discovery and
maintains an internal list of all GPUs available on a node.

Using this list it then allocates GPUs to callers of the `Allocate()` or
`AllocateSpecific()` function and frees GPUs back to this list from callers of
the `Free()` function, as seen below:

```
func (a *Allocator) Allocate(num int) []*Device
func (a *Allocator) AllocateSpecific(devices... *Device) []*Device
func (a *Allocator) Free(devices... *Device)
```

The `Policy` Interface
----------------------
```
type Policy interface {
	Allocate(devices []*Device, num int) []*Device
}
```

The `Policy` interface contains a single function `Allocate()`, which takes a
slice of devices and size `num` as arguments and returns a subset of that
slice of length `num`.

Implementers of this interface take on the heavy lifting
of implementing the actual allocation logic used by the `Allocator`.

The following default policies are implemented as part of this package:
```
func NewSimplePolicy() Policy
func NewBestEffortPolicy() Policy
func NewStaticDGX1Policy(gpuType GPUType) Policy
func NewStaticDGX2Policy() Policy
```

With the following convenience wrappers for simple and best effort allocators:
```
func NewSimpleAllocator() (*Allocator, error)
func NewBestEffortAllocator() (*Allocator, error)
```

`Simple` takes a slice of GPU devices and simply allocates `num` GPUs from the
front of it.

`BestEffort` attempts to allocate GPUs in topological order, considering both
NVLINKs between GPUs and their placement in the PCIe hierarchy.
The choice of GPUs to allocate is optimized to assume that all future
allocations will be of size 'num' as well.

Sample Usage
------------
```
package main

import (
	"fmt"
	"os"

	"github.com/NVIDIA/go-gpuallocator/gpuallocator"
)

func main() {
	allocator, err := gpuallocator.NewSimpleAllocator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	for _, gpu := range allocator.GPUs {
		fmt.Printf("%v\n", gpu.Details())
	}

	fmt.Printf("\n")

	for _, i := range []int{1, 2, 4, 8} {
		gpus := allocator.Allocate(i)
		fmt.Printf("Simple allocation of %v GPUs: %v\n", i, gpus)
		allocator.Free(gpus...)
	}
}
```
