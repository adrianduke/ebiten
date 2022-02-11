// Copyright 2022 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package directx

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func boolToUintptr(v bool) uintptr {
	if v {
		return 1
	}
	return 0
}

// Reference:
// * https://github.com/microsoft/DirectX-Headers
// * https://github.com/microsoft/win32metadata
// * https://raw.githubusercontent.com/microsoft/win32metadata/master/generation/WinSDK/RecompiledIdlHeaders/um/d3d12.h

type _D3D_FEATURE_LEVEL int32

const (
	_D3D_FEATURE_LEVEL_11_0 _D3D_FEATURE_LEVEL = 0xb000
)

type _D3D12_COMMAND_LIST_TYPE int32

const (
	_D3D12_COMMAND_LIST_TYPE_DIRECT _D3D12_COMMAND_LIST_TYPE = 0
)

type _D3D12_COMMAND_QUEUE_FLAGS int32

const (
	_D3D12_COMMAND_QUEUE_FLAG_NONE _D3D12_COMMAND_QUEUE_FLAGS = 0
)

type _D3D12_CPU_PAGE_PROPERTY int32

const (
	_D3D12_CPU_PAGE_PROPERTY_UNKNOWN _D3D12_CPU_PAGE_PROPERTY = 0
)

type _D3D12_DESCRIPTOR_HEAP_TYPE int32

const (
	_D3D12_DESCRIPTOR_HEAP_TYPE_RTV _D3D12_DESCRIPTOR_HEAP_TYPE = 2
)

type _D3D12_DESCRIPTOR_HEAP_FLAGS int32

const (
	_D3D12_DESCRIPTOR_HEAP_FLAG_NONE _D3D12_DESCRIPTOR_HEAP_FLAGS = 0
)

type _D3D12_FENCE_FLAGS int32

const (
	_D3D12_FENCE_FLAG_NONE _D3D12_FENCE_FLAGS = 0
)

type _D3D12_HEAP_FLAGS int32

const (
	_D3D12_HEAP_FLAG_NONE _D3D12_HEAP_FLAGS = 0
)

type _D3D12_HEAP_TYPE int32

const (
	_D3D12_HEAP_TYPE_UPLOAD _D3D12_HEAP_TYPE = 2
)

type _D3D12_MEMORY_POOL int32

const (
	_D3D12_MEMORY_POOL_UNKNOWN _D3D12_MEMORY_POOL = 0
)

const (
	_D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES = 0xffffffff
)

type _D3D12_RESOURCE_BARRIER_FLAGS int32

const (
	_D3D12_RESOURCE_BARRIER_FLAG_NONE _D3D12_RESOURCE_BARRIER_FLAGS = 0
)

type _D3D12_RESOURCE_BARRIER_TYPE int32

const (
	_D3D12_RESOURCE_BARRIER_TYPE_TRANSITION _D3D12_RESOURCE_BARRIER_TYPE = 0
)

type _D3D12_RESOURCE_DIMENSION int32

const (
	_D3D12_RESOURCE_DIMENSION_BUFFER _D3D12_RESOURCE_DIMENSION = 1
)

type _D3D12_RESOURCE_FLAGS int32

const (
	_D3D12_RESOURCE_FLAG_NONE _D3D12_RESOURCE_FLAGS = 0
)

type _D3D12_RESOURCE_STATES int32

const (
	_D3D12_RESOURCE_STATE_RENDER_TARGET _D3D12_RESOURCE_STATES = 0x4
	_D3D12_RESOURCE_STATE_GENERIC_READ  _D3D12_RESOURCE_STATES = 0x1 | 0x2 | 0x40 | 0x80 | 0x200 | 0x800
	_D3D12_RESOURCE_STATE_PRESENT       _D3D12_RESOURCE_STATES = 0
)

type _D3D12_RTV_DIMENSION int32

type _D3D12_TEXTURE_LAYOUT int32

const (
	_D3D12_TEXTURE_LAYOUT_ROW_MAJOR _D3D12_TEXTURE_LAYOUT = 1
)

type _DXGI_ALPHA_MODE int32

type _DXGI_FORMAT int32

const (
	_DXGI_FORMAT_UNKNOWN        _DXGI_FORMAT = 0
	_DXGI_FORMAT_R8G8B8A8_UNORM _DXGI_FORMAT = 28
	_DXGI_FORMAT_R16_UINT       _DXGI_FORMAT = 57
)

type _DXGI_MODE_SCANLINE_ORDER int32

type _DXGI_MODE_SCALING int32

type _DXGI_SCALING int32

type _DXGI_SWAP_EFFECT int32

const (
	_DXGI_SWAP_EFFECT_FLIP_DISCARD _DXGI_SWAP_EFFECT = 4
)

type _DXGI_USAGE uint32

const (
	_DXGI_USAGE_RENDER_TARGET_OUTPUT _DXGI_USAGE = 1 << (1 + 4)
)

const (
	_DXGI_ADAPTER_FLAG_SOFTWARE = 2

	_DXGI_CREATE_FACTORY_DEBUG = 0x01

	_DXGI_ERROR_NOT_FOUND = windows.Errno(0x887A0002)
)

var (
	_IID_ID3D12CommandAllocator    = windows.GUID{0x6102dee4, 0xaf59, 0x4b09, [...]byte{0xb9, 0x99, 0xb4, 0x4d, 0x73, 0xf0, 0x9b, 0x24}}
	_IID_ID3D12CommandQueue        = windows.GUID{0x0ec870a6, 0x5d7e, 0x4c22, [...]byte{0x8c, 0xfc, 0x5b, 0xaa, 0xe0, 0x76, 0x16, 0xed}}
	_IID_ID3D12Debug               = windows.GUID{0x344488b7, 0x6846, 0x474b, [...]byte{0xb9, 0x89, 0xf0, 0x27, 0x44, 0x82, 0x45, 0xe0}}
	_IID_ID3D12DescriptorHeap      = windows.GUID{0x8efb471d, 0x616c, 0x4f49, [...]byte{0x90, 0xf7, 0x12, 0x7b, 0xb7, 0x63, 0xfa, 0x51}}
	_IID_ID3D12Device              = windows.GUID{0x189819f1, 0x1db6, 0x4b57, [...]byte{0xbe, 0x54, 0x18, 0x21, 0x33, 0x9b, 0x85, 0xf7}}
	_IID_ID3D12Fence               = windows.GUID{0x0a753dcf, 0xc4d8, 0x4b91, [...]byte{0xad, 0xf6, 0xbe, 0x5a, 0x60, 0xd9, 0x5a, 0x76}}
	_IID_ID3D12GraphicsCommandList = windows.GUID{0x5b160d0f, 0xac1b, 0x4185, [...]byte{0x8b, 0xa8, 0xb3, 0xae, 0x42, 0xa5, 0xa4, 0x55}}
	_IID_ID3D12Resource1           = windows.GUID{0x9D5E227A, 0x4430, 0x4161, [...]byte{0x88, 0xB3, 0x3E, 0xCA, 0x6B, 0xB1, 0x6E, 0x19}}

	_IID_IDXGIAdapter1 = windows.GUID{0x29038f61, 0x3839, 0x4626, [...]byte{0x91, 0xfd, 0x08, 0x68, 0x79, 0x01, 0x1a, 0x05}}
	_IID_IDXGIFactory4 = windows.GUID{0x1bc6ea02, 0xef36, 0x464f, [...]byte{0xbf, 0x0c, 0x21, 0xca, 0x39, 0xe5, 0x16, 0x8a}}
)

type _D3D12_CLEAR_VALUE struct {
	Format _DXGI_FORMAT
	Color  [4]float32 // Union
}

type _D3D12_CPU_DESCRIPTOR_HANDLE struct {
	ptr uintptr
}

type _D3D12_GPU_VIRTUAL_ADDRESS uint64

type _D3D12_HEAP_PROPERTIES struct {
	Type                 _D3D12_HEAP_TYPE
	CPUPageProperty      _D3D12_CPU_PAGE_PROPERTY
	MemoryPoolPreference _D3D12_MEMORY_POOL
	CreationNodeMask     uint32
	VisibleNodeMask      uint32
}

type _D3D12_INDEX_BUFFER_VIEW struct {
	BufferLocation _D3D12_GPU_VIRTUAL_ADDRESS
	SizeInBytes    uint32
	Format         _DXGI_FORMAT
}

type _D3D12_RANGE struct {
	Begin uintptr
	End   uintptr
}

type _D3D12_RECT struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

func (h *_D3D12_CPU_DESCRIPTOR_HANDLE) Offset(offsetInDescriptors int32, descriptorIncrementSize uint32) {
	h.ptr += uintptr(offsetInDescriptors) * uintptr(descriptorIncrementSize)
}

type _D3D12_RESOURCE_BARRIER_Transition struct {
	Type       _D3D12_RESOURCE_BARRIER_TYPE
	Flags      _D3D12_RESOURCE_BARRIER_FLAGS
	Transition _D3D12_RESOURCE_TRANSITION_BARRIER
}

type _D3D12_RESOURCE_DESC struct {
	Dimension        _D3D12_RESOURCE_DIMENSION
	Alignment        uint64
	Width            uint64
	Height           uint32
	DepthOrArraySize uint16
	MipLevels        uint16
	Format           _DXGI_FORMAT
	SampleDesc       _DXGI_SAMPLE_DESC
	Layout           _D3D12_TEXTURE_LAYOUT
	Flags            _D3D12_RESOURCE_FLAGS
}

type _D3D12_RESOURCE_TRANSITION_BARRIER struct {
	pResource   *iD3D12Resource1
	Subresource uint32
	StateBefore _D3D12_RESOURCE_STATES
	StateAfter  _D3D12_RESOURCE_STATES
}

type _D3D12_VERTEX_BUFFER_VIEW struct {
	BufferLocation _D3D12_GPU_VIRTUAL_ADDRESS
	SizeInBytes    uint32
	StrideInBytes  uint32
}

var (
	d3d12 = windows.NewLazySystemDLL("d3d12.dll")
	dxgi  = windows.NewLazySystemDLL("dxgi.dll")

	procD3D12CreateDevice      = d3d12.NewProc("D3D12CreateDevice")
	procD3D12GetDebugInterface = d3d12.NewProc("D3D12GetDebugInterface")

	procCreateDXGIFactory2 = dxgi.NewProc("CreateDXGIFactory2")
)

func d3D12CreateDevice(pAdapter unsafe.Pointer, minimumFeatureLevel _D3D_FEATURE_LEVEL, riid *windows.GUID, ppDevice *unsafe.Pointer) error {
	r, _, _ := procD3D12CreateDevice.Call(uintptr(pAdapter), uintptr(minimumFeatureLevel), uintptr(unsafe.Pointer(riid)), uintptr(unsafe.Pointer(ppDevice)))
	if ppDevice == nil && windows.Handle(r) != windows.S_FALSE {
		return fmt.Errorf("directx: D3D12CreateDevice failed: %w", windows.Errno(r))
	}
	if ppDevice != nil && windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: D3D12CreateDevice failed: %w", windows.Errno(r))
	}
	return nil
}

func d3D12GetDebugInterface() (*iD3D12Debug, error) {
	var debug *iD3D12Debug
	r, _, _ := procD3D12GetDebugInterface.Call(uintptr(unsafe.Pointer(&_IID_ID3D12Debug)), uintptr(unsafe.Pointer(&debug)))
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: D3D12GetDebugInterface failed: %w", windows.Errno(r))
	}
	return debug, nil
}

func createDXGIFactory2(flags uint32) (*iDXGIFactory4, error) {
	var factory *iDXGIFactory4
	r, _, _ := procCreateDXGIFactory2.Call(uintptr(flags), uintptr(unsafe.Pointer(&_IID_IDXGIFactory4)), uintptr(unsafe.Pointer(&factory)))
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: CreateDXGIFactory2 failed: %w", windows.Errno(r))
	}
	return factory, nil
}

type _D3D12_COMMAND_QUEUE_DESC struct {
	Type     _D3D12_COMMAND_LIST_TYPE
	Priority int32
	Flags    _D3D12_COMMAND_QUEUE_FLAGS
	NodeMask uint32
}

type _D3D12_DESCRIPTOR_HEAP_DESC struct {
	Type           _D3D12_DESCRIPTOR_HEAP_TYPE
	NumDescriptors uint32
	Flags          _D3D12_DESCRIPTOR_HEAP_FLAGS
	NodeMask       uint32
}

type _D3D12_RENDER_TARGET_VIEW_DESC struct {
	Format        _DXGI_FORMAT
	ViewDimension _D3D12_RTV_DIMENSION
	_             [3]uint32 // Union: D3D12_BUFFER_RTV seems the biggest
}

type _DXGI_ADAPTER_DESC1 struct {
	Description           [128]uint16
	VendorId              uint32
	DeviceId              uint32
	SubSysId              uint32
	Revision              uint32
	DedicatedVideoMemory  uint
	DedicatedSystemMemory uint
	SharedSystemMemory    uint
	AdapterLuid           _LUID
	Flags                 uint32
}

type _DXGI_SWAP_CHAIN_FULLSCREEN_DESC struct {
	RefreshRate      _DXGI_RATIONAL
	ScanlineOrdering _DXGI_MODE_SCANLINE_ORDER
	Scaling          _DXGI_MODE_SCALING
	Windowed         int32
}

type _DXGI_RATIONAL struct {
	Numerator   uint32
	Denominator uint32
}

type _DXGI_SAMPLE_DESC struct {
	Count   uint32
	Quality uint32
}

type _DXGI_SWAP_CHAIN_DESC1 struct {
	Width       uint32
	Height      uint32
	Format      _DXGI_FORMAT
	Stereo      int32
	SampleDesc  _DXGI_SAMPLE_DESC
	BufferUsage _DXGI_USAGE
	BufferCount uint32
	Scaling     _DXGI_SCALING
	SwapEffect  _DXGI_SWAP_EFFECT
	AlphaMode   _DXGI_ALPHA_MODE
	Flags       uint32
}

type _LUID struct {
	LowPart  uint32
	HighPart int32
}

type iD3D12CommandAllocator struct {
	vtbl *iD3D12CommandAllocator_Vtbl
}

type iD3D12CommandAllocator_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData          uintptr
	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	SetName                 uintptr
	GetDevice               uintptr
	Reset                   uintptr
}

func (i *iD3D12CommandAllocator) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

func (i *iD3D12CommandAllocator) Reset() error {
	r, _, _ := syscall.Syscall(i.vtbl.Reset, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12CommandAllocator::Reset failed: %w", windows.Errno(r))
	}
	return nil
}

type iD3D12CommandQueue struct {
	vtbl *iD3D12CommandQueue_Vtbl
}

type iD3D12CommandQueue_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData          uintptr
	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	SetName                 uintptr
	GetDevice               uintptr
	UpdateTileMappings      uintptr
	CopyTileMappings        uintptr
	ExecuteCommandLists     uintptr
	SetMarker               uintptr
	BeginEvent              uintptr
	EndEvent                uintptr
	Signal                  uintptr
	Wait                    uintptr
	GetTimestampFrequency   uintptr
	GetClockCalibration     uintptr
	GetDesc                 uintptr
}

func (i *iD3D12CommandQueue) ExecuteCommandLists(numCommandLists uint32, ppCommandLists **iD3D12GraphicsCommandList) {
	syscall.Syscall(i.vtbl.ExecuteCommandLists, 3, uintptr(unsafe.Pointer(i)),
		uintptr(numCommandLists), uintptr(unsafe.Pointer(ppCommandLists)))
	runtime.KeepAlive(ppCommandLists)
}

func (i *iD3D12CommandQueue) Signal(signal *iD3D12Fence, value uint64) error {
	r, _, _ := syscall.Syscall(i.vtbl.Signal, 3, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(signal)), uintptr(value))
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12CommandQueue::Signal failed: %w", windows.Errno(r))
	}
	return nil
}

func (i *iD3D12CommandQueue) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

type iD3D12Debug struct {
	vtbl *iD3D12Debug_Vtbl
}

type iD3D12Debug_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	EnableDebugLayer uintptr
}

func (i *iD3D12Debug) EnableDebugLayer() {
	syscall.Syscall(i.vtbl.EnableDebugLayer, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

func (i *iD3D12Debug) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

type iD3D12DescriptorHeap struct {
	vtbl *iD3D12DescriptrHeap_Vtbl
}

type iD3D12DescriptrHeap_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData                     uintptr
	SetPrivateData                     uintptr
	SetPrivateDataInterface            uintptr
	SetName                            uintptr
	GetDevice                          uintptr
	GetDesc                            uintptr
	GetCPUDescriptorHandleForHeapStart uintptr
	GetGPUDescriptorHandleForHeapStart uintptr
}

func (i *iD3D12DescriptorHeap) GetCPUDescriptorHandleForHeapStart() _D3D12_CPU_DESCRIPTOR_HANDLE {
	// There is a bug in the header file:
	// https://stackoverflow.com/questions/34118929/getcpudescriptorhandleforheapstart-stack-corruption
	var handle _D3D12_CPU_DESCRIPTOR_HANDLE
	syscall.Syscall(i.vtbl.GetCPUDescriptorHandleForHeapStart, 2, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&handle)), 0)
	return handle
}

func (i *iD3D12DescriptorHeap) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

type iD3D12Device struct {
	vtbl *iD3D12Device_Vtbl
}

type iD3D12Device_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData                   uintptr
	SetPrivateData                   uintptr
	SetPrivateDataInterface          uintptr
	SetName                          uintptr
	GetNodeCount                     uintptr
	CreateCommandQueue               uintptr
	CreateCommandAllocator           uintptr
	CreateGraphicsPipelineState      uintptr
	CreateComputePipelineState       uintptr
	CreateCommandList                uintptr
	CheckFeatureSupport              uintptr
	CreateDescriptorHeap             uintptr
	GetDescriptorHandleIncrementSize uintptr
	CreateRootSignature              uintptr
	CreateConstantBufferView         uintptr
	CreateShaderResourceView         uintptr
	CreateUnorderedAccessView        uintptr
	CreateRenderTargetView           uintptr
	CreateDepthStencilView           uintptr
	CreateSampler                    uintptr
	CopyDescriptors                  uintptr
	CopyDescriptorsSimple            uintptr
	GetResourceAllocationInfo        uintptr
	GetCustomHeapProperties          uintptr
	CreateCommittedResource          uintptr
	CreateHeap                       uintptr
	CreatePlacedResource             uintptr
	CreateReservedResource           uintptr
	CreateSharedHandle               uintptr
	OpenSharedHandle                 uintptr
	OpenSharedHandleByName           uintptr
	MakeResident                     uintptr
	Evict                            uintptr
	CreateFence                      uintptr
	GetDeviceRemovedReason           uintptr
	GetCopyableFootprints            uintptr
	CreateQueryHeap                  uintptr
	SetStablePowerState              uintptr
	CreateCommandSignature           uintptr
	GetResourceTiling                uintptr
	GetAdapterLuid                   uintptr
}

func (i *iD3D12Device) CreateCommandAllocator(typ _D3D12_COMMAND_LIST_TYPE) (*iD3D12CommandAllocator, error) {
	var commandAllocator *iD3D12CommandAllocator
	r, _, _ := syscall.Syscall6(i.vtbl.CreateCommandAllocator, 4, uintptr(unsafe.Pointer(i)),
		uintptr(typ), uintptr(unsafe.Pointer(&_IID_ID3D12CommandAllocator)), uintptr(unsafe.Pointer(&commandAllocator)),
		0, 0)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Device::CreateCommandAllocator failed: %w", windows.Errno(r))
	}
	return commandAllocator, nil
}

func (i *iD3D12Device) CreateCommandList(nodeMask uint32, typ _D3D12_COMMAND_LIST_TYPE, pCommandAllocator *iD3D12CommandAllocator, pInitialState *iD3D12PipelineState) (*iD3D12GraphicsCommandList, error) {
	var commandList *iD3D12GraphicsCommandList
	r, _, _ := syscall.Syscall9(i.vtbl.CreateCommandList, 7,
		uintptr(unsafe.Pointer(i)), uintptr(nodeMask), uintptr(typ),
		uintptr(unsafe.Pointer(pCommandAllocator)), uintptr(unsafe.Pointer(pInitialState)), uintptr(unsafe.Pointer(&_IID_ID3D12GraphicsCommandList)),
		uintptr(unsafe.Pointer(&commandList)), 0, 0)
	runtime.KeepAlive(pCommandAllocator)
	runtime.KeepAlive(pInitialState)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Device::CreateCommandList failed: %w", windows.Errno(r))
	}
	return commandList, nil
}

func (i *iD3D12Device) CreateCommittedResource(pHeapProperties *_D3D12_HEAP_PROPERTIES, heapFlags _D3D12_HEAP_FLAGS, pDesc *_D3D12_RESOURCE_DESC, initialResourceState _D3D12_RESOURCE_STATES, pOptimizedClearValue *_D3D12_CLEAR_VALUE) (*iD3D12Resource1, error) {
	var resource *iD3D12Resource1
	r, _, _ := syscall.Syscall9(i.vtbl.CreateCommittedResource, 8,
		uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(pHeapProperties)), uintptr(heapFlags),
		uintptr(unsafe.Pointer(pDesc)), uintptr(initialResourceState), uintptr(unsafe.Pointer(pOptimizedClearValue)),
		uintptr(unsafe.Pointer(&_IID_ID3D12Resource1)), uintptr(unsafe.Pointer(&resource)), 0)
	runtime.KeepAlive(pHeapProperties)
	runtime.KeepAlive(pDesc)
	runtime.KeepAlive(pOptimizedClearValue)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Device::CreateCommittedResource failed: %w", windows.Errno(r))
	}
	return resource, nil
}

func (i *iD3D12Device) CreateCommandQueue(desc *_D3D12_COMMAND_QUEUE_DESC) (*iD3D12CommandQueue, error) {
	var commandQueue *iD3D12CommandQueue
	r, _, _ := syscall.Syscall6(i.vtbl.CreateCommandQueue, 4, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(desc)), uintptr(unsafe.Pointer(&_IID_ID3D12CommandQueue)), uintptr(unsafe.Pointer(&commandQueue)),
		0, 0)
	runtime.KeepAlive(desc)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Device::CreateCommandQueue failed: %w", windows.Errno(r))
	}
	return commandQueue, nil
}

func (i *iD3D12Device) GetDescriptorHandleIncrementSize(descriptorHeapType _D3D12_DESCRIPTOR_HEAP_TYPE) uint32 {
	r, _, _ := syscall.Syscall(i.vtbl.GetDescriptorHandleIncrementSize, 2, uintptr(unsafe.Pointer(i)),
		uintptr(descriptorHeapType), 0)
	return uint32(r)
}

func (i *iD3D12Device) CreateDescriptorHeap(desc *_D3D12_DESCRIPTOR_HEAP_DESC) (*iD3D12DescriptorHeap, error) {
	var descriptorHeap *iD3D12DescriptorHeap
	r, _, _ := syscall.Syscall6(i.vtbl.CreateDescriptorHeap, 4, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(desc)), uintptr(unsafe.Pointer(&_IID_ID3D12DescriptorHeap)), uintptr(unsafe.Pointer(&descriptorHeap)),
		0, 0)
	runtime.KeepAlive(desc)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Device::CreateDescriptorHeap failed: %w", windows.Errno(r))
	}
	return descriptorHeap, nil
}

func (i *iD3D12Device) CreateFence(initialValue uint64, flags _D3D12_FENCE_FLAGS) (*iD3D12Fence, error) {
	// TODO: Does this work on a 32bit machine?
	var fence *iD3D12Fence
	r, _, _ := syscall.Syscall6(i.vtbl.CreateFence, 5, uintptr(unsafe.Pointer(i)),
		uintptr(initialValue), uintptr(flags), uintptr(unsafe.Pointer(&_IID_ID3D12Fence)), uintptr(unsafe.Pointer(&fence)),
		0)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Device::CreateFence failed: %w", windows.Errno(r))
	}
	return fence, nil
}

func (i *iD3D12Device) CreateRenderTargetView(pResource *iD3D12Resource1, pDesc *_D3D12_RENDER_TARGET_VIEW_DESC, destDescriptor _D3D12_CPU_DESCRIPTOR_HANDLE) error {
	r, _, _ := syscall.Syscall6(i.vtbl.CreateRenderTargetView, 4, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pResource)), uintptr(unsafe.Pointer(pDesc)), destDescriptor.ptr,
		0, 0)
	runtime.KeepAlive(pResource)
	runtime.KeepAlive(pDesc)
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12Device::CreateRenderTargetView failed: %w", windows.Errno(r))
	}
	return nil
}

type iD3D12Fence struct {
	vtbl *iD3D12Fence_Vtbl
}

type iD3D12Fence_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData          uintptr
	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	SetName                 uintptr
	GetDevice               uintptr
	GetCompletedValue       uintptr
	SetEventOnCompletion    uintptr
	Signal                  uintptr
}

func (i *iD3D12Fence) GetCompletedValue() uint64 {
	// TODO: Does this work on a 32bit machine?
	r, _, _ := syscall.Syscall(i.vtbl.GetCompletedValue, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	return uint64(r)
}

func (i *iD3D12Fence) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

func (i *iD3D12Fence) SetEventOnCompletion(value uint64, hEvent windows.Handle) error {
	// TODO: Does this work on a 32bit machine?
	r, _, _ := syscall.Syscall(i.vtbl.SetEventOnCompletion, 3, uintptr(unsafe.Pointer(i)),
		uintptr(value), uintptr(hEvent))
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12Fence::SetEventOnCompletion failed: %w", windows.Errno(r))
	}
	return nil
}

type iD3D12GraphicsCommandList struct {
	vtbl *iD3D12GraphicsCommandList_Vtbl
}

type iD3D12GraphicsCommandList_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData                     uintptr
	SetPrivateData                     uintptr
	SetPrivateDataInterface            uintptr
	SetName                            uintptr
	GetDevice                          uintptr
	GetType                            uintptr
	Close                              uintptr
	Reset                              uintptr
	ClearState                         uintptr
	DrawInstanced                      uintptr
	DrawIndexedInstanced               uintptr
	Dispatch                           uintptr
	CopyBufferRegion                   uintptr
	CopyTextureRegion                  uintptr
	CopyResource                       uintptr
	CopyTiles                          uintptr
	ResolveSubresource                 uintptr
	IASetPrimitiveTopology             uintptr
	RSSetViewports                     uintptr
	RSSetScissorRects                  uintptr
	OMSetBlendFactor                   uintptr
	OMSetStencilRef                    uintptr
	SetPipelineState                   uintptr
	ResourceBarrier                    uintptr
	ExecuteBundle                      uintptr
	SetDescriptorHeaps                 uintptr
	SetComputeRootSignature            uintptr
	SetGraphicsRootSignature           uintptr
	SetComputeRootDescriptorTable      uintptr
	SetGraphicsRootDescriptorTable     uintptr
	SetComputeRoot32BitConstant        uintptr
	SetGraphicsRoot32BitConstant       uintptr
	SetComputeRoot32BitConstants       uintptr
	SetGraphicsRoot32BitConstants      uintptr
	SetComputeRootConstantBufferView   uintptr
	SetGraphicsRootConstantBufferView  uintptr
	SetComputeRootShaderResourceView   uintptr
	SetGraphicsRootShaderResourceView  uintptr
	SetComputeRootUnorderedAccessView  uintptr
	SetGraphicsRootUnorderedAccessView uintptr
	IASetIndexBuffer                   uintptr
	IASetVertexBuffers                 uintptr
	SOSetTargets                       uintptr
	OMSetRenderTargets                 uintptr
	ClearDepthStencilView              uintptr
	ClearRenderTargetView              uintptr
	ClearUnorderedAccessViewUint       uintptr
	ClearUnorderedAccessViewFloat      uintptr
	DiscardResource                    uintptr
	BeginQuery                         uintptr
	EndQuery                           uintptr
	ResolveQueryData                   uintptr
	SetPredication                     uintptr
	SetMarker                          uintptr
	BeginEvent                         uintptr
	EndEvent                           uintptr
	ExecuteIndirect                    uintptr
}

func (i *iD3D12GraphicsCommandList) ClearRenderTargetView(pRenderTargetView _D3D12_CPU_DESCRIPTOR_HANDLE, colorRGBA [4]float32, numRects uint32, pRects *_D3D12_RECT) {
	syscall.Syscall6(i.vtbl.ClearRenderTargetView, 5, uintptr(unsafe.Pointer(i)),
		pRenderTargetView.ptr, uintptr(unsafe.Pointer(&colorRGBA[0])), uintptr(numRects), uintptr(unsafe.Pointer(pRects)),
		0)
	runtime.KeepAlive(pRenderTargetView)
}

func (i *iD3D12GraphicsCommandList) Close() error {
	r, _, _ := syscall.Syscall(i.vtbl.Close, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12GraphicsCommandList::Close failed: %w", windows.Errno(r))
	}
	return nil
}

func (i *iD3D12GraphicsCommandList) OMSetRenderTargets(numRenderTargetDescriptors uint32, pRenderTargetDescriptors *_D3D12_CPU_DESCRIPTOR_HANDLE, rtsSingleHandleToDescriptorRange bool, pDepthStencilDescriptor *_D3D12_CPU_DESCRIPTOR_HANDLE) {
	syscall.Syscall6(i.vtbl.OMSetRenderTargets, 5, uintptr(unsafe.Pointer(i)),
		uintptr(numRenderTargetDescriptors), uintptr(unsafe.Pointer(pRenderTargetDescriptors)), boolToUintptr(rtsSingleHandleToDescriptorRange), uintptr(unsafe.Pointer(pDepthStencilDescriptor)),
		0)
}

func (i *iD3D12GraphicsCommandList) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

func (i *iD3D12GraphicsCommandList) Reset(pAllocator *iD3D12CommandAllocator, pInitialState *iD3D12PipelineState) error {
	r, _, _ := syscall.Syscall(i.vtbl.Reset, 3, uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(pAllocator)), uintptr(unsafe.Pointer(pInitialState)))
	runtime.KeepAlive(pAllocator)
	runtime.KeepAlive(pInitialState)
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12GraphicsCommandList::Reset failed: %w", windows.Errno(r))
	}
	return nil
}

func (i *iD3D12GraphicsCommandList) ResourceBarrier(numBarriers uint32, pBarriers *_D3D12_RESOURCE_BARRIER_Transition) {
	syscall.Syscall(i.vtbl.ResourceBarrier, 3, uintptr(unsafe.Pointer(i)),
		uintptr(numBarriers), uintptr(unsafe.Pointer(pBarriers)))
}

type iD3D12PipelineState struct {
	vtbl *iD3D12PipelineState_Vtbl
}

type iD3D12PipelineState_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData          uintptr
	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	SetName                 uintptr
	GetDevice               uintptr
	GetCachedBlob           uintptr
}

type iD3D12Resource1 struct {
	vtbl *iD3D12Resource1_Vtbl
}

type iD3D12Resource1_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetPrivateData              uintptr
	SetPrivateData              uintptr
	SetPrivateDataInterface     uintptr
	SetName                     uintptr
	GetDevice                   uintptr
	Map                         uintptr
	Unmap                       uintptr
	GetDesc                     uintptr
	GetGPUVirtualAddress        uintptr
	WriteToSubresource          uintptr
	ReadFromSubresource         uintptr
	GetHeapProperties           uintptr
	GetProtectedResourceSession uintptr
}

func (i *iD3D12Resource1) GetGPUVirtualAddress() _D3D12_GPU_VIRTUAL_ADDRESS {
	r, _, _ := syscall.Syscall(i.vtbl.GetGPUVirtualAddress, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	return _D3D12_GPU_VIRTUAL_ADDRESS(r)
}

func (i *iD3D12Resource1) Map(subresource uint32, pReadRange *_D3D12_RANGE) (unsafe.Pointer, error) {
	var data unsafe.Pointer
	r, _, _ := syscall.Syscall6(i.vtbl.Map, 4, uintptr(unsafe.Pointer(i)),
		uintptr(subresource), uintptr(unsafe.Pointer(pReadRange)), uintptr(unsafe.Pointer(&data)),
		0, 0)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: ID3D12Resource1::Map failed: %w", windows.Errno(r))
	}
	return data, nil
}

func (i *iD3D12Resource1) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

func (i *iD3D12Resource1) Unmap(subresource uint32, pWrittenRange *_D3D12_RANGE) error {
	r, _, _ := syscall.Syscall(i.vtbl.Unmap, 3, uintptr(unsafe.Pointer(i)),
		uintptr(subresource), uintptr(unsafe.Pointer(pWrittenRange)))
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: ID3D12Resource1::Unmap failed: %w", windows.Errno(r))
	}
	return nil
}

type iDXGIAdapter1 struct {
	vtbl *iDXGIAdapter1_Vtbl
}

type iDXGIAdapter1_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	GetPrivateData          uintptr
	GetParent               uintptr
	EnumOutputs             uintptr
	GetDesc                 uintptr
	CheckInterfaceSupport   uintptr
	GetDesc1                uintptr
}

func (i *iDXGIAdapter1) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

func (i *iDXGIAdapter1) GetDesc1() (*_DXGI_ADAPTER_DESC1, error) {
	var desc _DXGI_ADAPTER_DESC1
	r, _, _ := syscall.Syscall(i.vtbl.GetDesc1, 2, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&desc)), 0)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: IDXGIAdapter1::GetDesc1 failed: %w", windows.Errno(r))
	}
	return &desc, nil
}

type iDXGIFactory4 struct {
	vtbl *iDXGIFactory4_Vtbl
}

type iDXGIFactory4_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	SetPrivateData                uintptr
	SetPrivateDataInterface       uintptr
	GetPrivateData                uintptr
	GetParent                     uintptr
	EnumAdapters                  uintptr
	MakeWindowAssociation         uintptr
	GetWindowAssociation          uintptr
	CreateSwapChain               uintptr
	CreateSoftwareAdapter         uintptr
	EnumAdapters1                 uintptr
	IsCurrent                     uintptr
	IsWindowedStereoEnabled       uintptr
	CreateSwapChainForHwnd        uintptr
	CreateSwapChainForCoreWindow  uintptr
	GetSharedResourceAdapterLuid  uintptr
	RegisterStereoStatusWindow    uintptr
	RegisterStereoStatusEvent     uintptr
	UnregisterStereoStatus        uintptr
	RegisterOcclusionStatusWindow uintptr
	RegisterOcclusionStatusEvent  uintptr
	UnregisterOcclusionStatus     uintptr
	CreateSwapChainForComposition uintptr
	GetCreationFlags              uintptr
	EnumAdapterByLuid             uintptr
	EnumWarpAdapter               uintptr
}

func (i *iDXGIFactory4) CreateSwapChainForHwnd(pDevice unsafe.Pointer, hWnd windows.HWND, pDesc *_DXGI_SWAP_CHAIN_DESC1, pFullscreenDesc *_DXGI_SWAP_CHAIN_FULLSCREEN_DESC, pRestrictToOutput *iDXGIOutput) (*iDXGISwapChain1, error) {
	var swapChain *iDXGISwapChain1
	r, _, _ := syscall.Syscall9(i.vtbl.CreateSwapChainForHwnd, 7,
		uintptr(unsafe.Pointer(i)), uintptr(pDevice), uintptr(hWnd),
		uintptr(unsafe.Pointer(pDesc)), uintptr(unsafe.Pointer(pFullscreenDesc)), uintptr(unsafe.Pointer(pRestrictToOutput)),
		uintptr(unsafe.Pointer(&swapChain)), 0, 0)
	runtime.KeepAlive(pDesc)
	runtime.KeepAlive(pFullscreenDesc)
	runtime.KeepAlive(pRestrictToOutput)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: IDXGIFactory4::CreateSwapChainForHwnd failed: %w", windows.Errno(r))
	}
	return swapChain, nil
}

func (i *iDXGIFactory4) EnumAdapters1(adapter uint32) (*iDXGIAdapter1, error) {
	var ptr *iDXGIAdapter1
	r, _, _ := syscall.Syscall(i.vtbl.EnumAdapters1, 3, uintptr(unsafe.Pointer(i)), uintptr(adapter), uintptr(unsafe.Pointer(&ptr)))
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: IDXGIFactory4::EnumAdapters1 failed: %w", windows.Errno(r))
	}
	return ptr, nil
}

func (i *iDXGIFactory4) EnumWarpAdapter() (*iDXGIAdapter1, error) {
	var ptr *iDXGIAdapter1
	r, _, _ := syscall.Syscall(i.vtbl.EnumWarpAdapter, 3, uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&_IID_IDXGIAdapter1)), uintptr(unsafe.Pointer(&ptr)))
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: IDXGIFactory4::EnumWarpAdapter failed: %w", windows.Errno(r))
	}
	return ptr, nil
}

func (i *iDXGIFactory4) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}

type iDXGIOutput struct {
	vtbl *iDXGIOutput_Vtbl
}

type iDXGIOutput_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	SetPrivateData              uintptr
	SetPrivateDataInterface     uintptr
	GetPrivateData              uintptr
	GetParent                   uintptr
	GetDesc                     uintptr
	GetDisplayModeList          uintptr
	FindClosestMatchingMode     uintptr
	WaitForVBlank               uintptr
	TakeOwnership               uintptr
	ReleaseOwnership            uintptr
	GetGammaControlCapabilities uintptr
	SetGammaControl             uintptr
	GetGammaControl             uintptr
	SetDisplaySurface           uintptr
	GetDisplaySurfaceData       uintptr
	GetFrameStatistics          uintptr
}

type iDXGISwapChain1 struct {
	vtbl *iDXGISwapChain1_Vtbl
}

type iDXGISwapChain1_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	SetPrivateData           uintptr
	SetPrivateDataInterface  uintptr
	GetPrivateData           uintptr
	GetParent                uintptr
	GetDevice                uintptr
	Present                  uintptr
	GetBuffer                uintptr
	SetFullscreenState       uintptr
	GetFullscreenState       uintptr
	GetDesc                  uintptr
	ResizeBuffers            uintptr
	ResizeTarget             uintptr
	GetContainingOutput      uintptr
	GetFrameStatistics       uintptr
	GetLastPresentCount      uintptr
	GetDesc1                 uintptr
	GetFullscreenDesc        uintptr
	GetHwnd                  uintptr
	GetCoreWindow            uintptr
	Present1                 uintptr
	IsTemporaryMonoSupported uintptr
	GetRestrictToOutput      uintptr
	SetBackgroundColor       uintptr
	GetBackgroundColor       uintptr
	SetRotation              uintptr
	GetRotation              uintptr
}

func (i *iDXGISwapChain1) As(swapChain **iDXGISwapChain4) {
	*swapChain = (*iDXGISwapChain4)(unsafe.Pointer(i))
}

type iDXGISwapChain4 struct {
	vtbl *iDXGISwapChain4_Vtbl
}

type iDXGISwapChain4_Vtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	SetPrivateData           uintptr
	SetPrivateDataInterface  uintptr
	GetPrivateData           uintptr
	GetParent                uintptr
	GetDevice                uintptr
	Present                  uintptr
	GetBuffer                uintptr
	SetFullscreenState       uintptr
	GetFullscreenState       uintptr
	GetDesc                  uintptr
	ResizeBuffers            uintptr
	ResizeTarget             uintptr
	GetContainingOutput      uintptr
	GetFrameStatistics       uintptr
	GetLastPresentCount      uintptr
	GetDesc1                 uintptr
	GetFullscreenDesc        uintptr
	GetHwnd                  uintptr
	GetCoreWindow            uintptr
	Present1                 uintptr
	IsTemporaryMonoSupported uintptr
	GetRestrictToOutput      uintptr
	SetBackgroundColor       uintptr
	GetBackgroundColor       uintptr
	SetRotation              uintptr
	GetRotation              uintptr

	SetSourceSize                 uintptr
	GetSourceSize                 uintptr
	SetMaximumFrameLatency        uintptr
	GetMaximumFrameLatency        uintptr
	GetFrameLatencyWaitableObject uintptr
	SetMatrixTransform            uintptr
	GetMatrixTransform            uintptr
	GetCurrentBackBufferIndex     uintptr
	CheckColorSpaceSupport        uintptr
	SetColorSpace1                uintptr
	ResizeBuffers1                uintptr
	SetHDRMetaData                uintptr
}

func (i *iDXGISwapChain4) GetBuffer(buffer uint32) (*iD3D12Resource1, error) {
	var resource *iD3D12Resource1
	r, _, _ := syscall.Syscall6(i.vtbl.GetBuffer, 4, uintptr(unsafe.Pointer(i)),
		uintptr(buffer), uintptr(unsafe.Pointer(&_IID_ID3D12Resource1)), uintptr(unsafe.Pointer(&resource)),
		0, 0)
	if windows.Handle(r) != windows.S_OK {
		return nil, fmt.Errorf("directx: IDXGISwapChain4::GetBuffer failed: %w", windows.Errno(r))
	}
	return resource, nil
}

func (i *iDXGISwapChain4) GetCurrentBackBufferIndex() uint32 {
	r, _, _ := syscall.Syscall(i.vtbl.GetCurrentBackBufferIndex, 1, uintptr(unsafe.Pointer(i)), 0, 0)
	return uint32(r)
}

func (i *iDXGISwapChain4) Present(syncInterval uint32, flags uint32) error {
	r, _, _ := syscall.Syscall(i.vtbl.Present, 3, uintptr(unsafe.Pointer(i)), uintptr(syncInterval), uintptr(flags))
	if windows.Handle(r) != windows.S_OK {
		return fmt.Errorf("directx: IDXGISwapChain4::Present failed: %w", windows.Errno(r))
	}
	return nil
}

func (i *iDXGISwapChain4) Release() {
	syscall.Syscall(i.vtbl.Release, 1, uintptr(unsafe.Pointer(i)), 0, 0)
}
