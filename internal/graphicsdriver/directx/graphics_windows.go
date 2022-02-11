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
	"errors"
	"fmt"
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/hajimehoshi/ebiten/v2/internal/graphics"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver"
	"github.com/hajimehoshi/ebiten/v2/internal/shaderir"
)

const frameCount = 2

var isDirectXAvailable = theGraphics.initializeDevice() == nil

var theGraphics Graphics

func Get() *Graphics {
	if !isDirectXAvailable {
		return nil
	}
	return &theGraphics
}

type Graphics struct {
	debug             *iD3D12Debug
	device            *iD3D12Device
	commandQueue      *iD3D12CommandQueue
	rtvDescriptorHeap *iD3D12DescriptorHeap
	rtvDescriptorSize uint32
	renderTargets     [frameCount]*iD3D12Resource1
	commandAllocators [frameCount]*iD3D12CommandAllocator
	fences            [frameCount]*iD3D12Fence
	fenceValues       [frameCount]uint64
	fenceWaitEvent    windows.Handle
	commandList       *iD3D12GraphicsCommandList
	vertices          *iD3D12Resource1
	indices           *iD3D12Resource1
	verticesView      _D3D12_VERTEX_BUFFER_VIEW
	indicesView       _D3D12_INDEX_BUFFER_VIEW

	factory   *iDXGIFactory4
	adapter   *iDXGIAdapter1
	swapChain *iDXGISwapChain4

	window windows.HWND

	frameIndex int
}

func (g *Graphics) initializeDevice() (ferr error) {
	if err := d3d12.Load(); err != nil {
		return err
	}

	// As g's lifetime is the same as the process's lifetime, debug and other objects are never released
	// if this initialization succeeds.

	d, err := d3D12GetDebugInterface()
	if err != nil {
		return err
	}
	g.debug = d
	defer func() {
		if ferr != nil {
			g.debug.Release()
			g.debug = nil
		}
	}()
	g.debug.EnableDebugLayer()

	f, err := createDXGIFactory2(_DXGI_CREATE_FACTORY_DEBUG)
	if err != nil {
		return err
	}
	g.factory = f
	defer func() {
		if ferr != nil {
			g.factory.Release()
			g.factory = nil
		}
	}()

	if useWARP {
		a, err := g.factory.EnumWarpAdapter()
		if err != nil {
			return err
		}

		g.adapter = a
		defer func() {
			if ferr != nil {
				g.adapter.Release()
				g.adapter = nil
			}
		}()
	} else {
		for i := 0; ; i++ {
			a, err := g.factory.EnumAdapters1(uint32(i))
			if errors.Is(err, _DXGI_ERROR_NOT_FOUND) {
				break
			}
			if err != nil {
				return err
			}

			desc, err := a.GetDesc1()
			if err != nil {
				return err
			}
			if desc.Flags&_DXGI_ADAPTER_FLAG_SOFTWARE != 0 {
				a.Release()
				continue
			}
			if err := d3D12CreateDevice(unsafe.Pointer(a), _D3D_FEATURE_LEVEL_11_0, &_IID_ID3D12Device, nil); err != nil {
				a.Release()
				continue
			}
			g.adapter = a
			defer func() {
				if ferr != nil {
					g.adapter.Release()
					g.adapter = nil
				}
			}()
			break
		}
	}

	if g.adapter == nil {
		return errors.New("directx: DirectX 12 is not supported")
	}

	if err := d3D12CreateDevice(unsafe.Pointer(g.adapter), _D3D_FEATURE_LEVEL_11_0, &_IID_ID3D12Device, (*unsafe.Pointer)(unsafe.Pointer(&g.device))); err != nil {
		return err
	}

	return nil
}

func (g *Graphics) Initialize() (ferr error) {
	e, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("directx: CreateEvent failed: %w", err)
	}
	g.fenceWaitEvent = e

	// Create a command queue.
	desc := _D3D12_COMMAND_QUEUE_DESC{
		Type:  _D3D12_COMMAND_LIST_TYPE_DIRECT,
		Flags: _D3D12_COMMAND_QUEUE_FLAG_NONE,
	}
	c, err := g.device.CreateCommandQueue(&desc)
	if err != nil {
		return err
	}
	g.commandQueue = c
	defer func() {
		if ferr != nil {
			g.commandQueue.Release()
			g.commandQueue = nil
		}
	}()

	// Create command allocators.
	for i := 0; i < frameCount; i++ {
		ca, err := g.device.CreateCommandAllocator(_D3D12_COMMAND_LIST_TYPE_DIRECT)
		if err != nil {
			return err
		}
		g.commandAllocators[i] = ca
		defer func(i int) {
			if ferr != nil {
				g.commandAllocators[i].Release()
				g.commandAllocators[i] = nil
			}
		}(i)
	}

	// Create frame fences.
	for i := 0; i < frameCount; i++ {
		f, err := g.device.CreateFence(0, _D3D12_FENCE_FLAG_NONE)
		if err != nil {
			return err
		}
		g.fences[i] = f
		defer func(i int) {
			if ferr != nil {
				g.fences[i].Release()
				g.fences[i] = nil
			}
		}(i)
	}

	// Create a command list.
	cl, err := g.device.CreateCommandList(0, _D3D12_COMMAND_LIST_TYPE_DIRECT, g.commandAllocators[0], nil)
	if err != nil {
		return err
	}
	g.commandList = cl
	defer func() {
		if ferr != nil {
			g.commandList.Release()
			g.commandList = nil
		}
	}()
	if err := g.commandList.Close(); err != nil {
		return err
	}

	// Create buffers.
	// TODO: Use the default heap for efficienty. See the official example HelloTriangle.
	vs, err := g.createBuffer(graphics.IndicesNum * graphics.VertexFloatNum * uint64(unsafe.Sizeof(float32(0))))
	if err != nil {
		return err
	}
	g.vertices = vs
	defer func() {
		if ferr != nil {
			g.vertices.Release()
			g.vertices = nil
		}
	}()

	is, err := g.createBuffer(graphics.IndicesNum * uint64(unsafe.Sizeof(uint16(0))))
	if err != nil {
		return err
	}
	g.indices = is
	defer func() {
		if ferr != nil {
			g.indices.Release()
			g.indices = nil
		}
	}()

	return nil
}

func (g *Graphics) createBuffer(bufferSize uint64) (*iD3D12Resource1, error) {
	heapProps := _D3D12_HEAP_PROPERTIES{
		Type:                 _D3D12_HEAP_TYPE_UPLOAD,
		CPUPageProperty:      _D3D12_CPU_PAGE_PROPERTY_UNKNOWN,
		MemoryPoolPreference: _D3D12_MEMORY_POOL_UNKNOWN,
		CreationNodeMask:     1,
		VisibleNodeMask:      1,
	}
	resDesc := _D3D12_RESOURCE_DESC{
		Dimension:        _D3D12_RESOURCE_DIMENSION_BUFFER,
		Alignment:        0,
		Width:            bufferSize,
		Height:           1,
		DepthOrArraySize: 1,
		MipLevels:        1,
		Format:           _DXGI_FORMAT_UNKNOWN,
		SampleDesc: _DXGI_SAMPLE_DESC{
			Count:   1,
			Quality: 0,
		},
		Layout: _D3D12_TEXTURE_LAYOUT_ROW_MAJOR,
		Flags:  _D3D12_RESOURCE_FLAG_NONE,
	}

	r, err := g.device.CreateCommittedResource(&heapProps, _D3D12_HEAP_FLAG_NONE, &resDesc, _D3D12_RESOURCE_STATE_GENERIC_READ, nil)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (g *Graphics) updateSwapChain(width, height int) error {
	if g.window == 0 {
		return errors.New("directx: the window handle is not initialized yet")
	}

	if g.swapChain == nil {
		if err := g.initSwapChain(width, height); err != nil {
			return err
		}
	} else {
		// TODO: Resize the chain buffer size if exists?
	}

	return nil
}

func (g *Graphics) initSwapChain(width, height int) (ferr error) {
	// Create a swap chain.
	desc := _DXGI_SWAP_CHAIN_DESC1{
		Width:       uint32(width),
		Height:      uint32(height),
		Format:      _DXGI_FORMAT_R8G8B8A8_UNORM,
		BufferUsage: _DXGI_USAGE_RENDER_TARGET_OUTPUT,
		BufferCount: frameCount,
		SwapEffect:  _DXGI_SWAP_EFFECT_FLIP_DISCARD,
		SampleDesc: _DXGI_SAMPLE_DESC{
			Count:   1,
			Quality: 0,
		},
	}
	s, err := g.factory.CreateSwapChainForHwnd(unsafe.Pointer(g.commandQueue), g.window, &desc, nil, nil)
	if err != nil {
		return err
	}
	s.As(&g.swapChain)
	defer func() {
		if ferr != nil {
			g.swapChain.Release()
			g.swapChain = nil
		}
	}()

	// TODO: Call factory.MakeWindowAssociation not to support fullscreen transitions?
	// TODO: Get the current buffer index?

	// Create descriptor heaps for RTV.
	rtvHeapDesc := _D3D12_DESCRIPTOR_HEAP_DESC{
		Type:           _D3D12_DESCRIPTOR_HEAP_TYPE_RTV,
		NumDescriptors: frameCount,
		Flags:          _D3D12_DESCRIPTOR_HEAP_FLAG_NONE,
	}
	h, err := g.device.CreateDescriptorHeap(&rtvHeapDesc)
	if err != nil {
		return err
	}
	g.rtvDescriptorHeap = h
	defer func() {
		if ferr != nil {
			g.rtvDescriptorHeap.Release()
			g.rtvDescriptorHeap = nil
		}
	}()

	g.rtvDescriptorSize = g.device.GetDescriptorHandleIncrementSize(_D3D12_DESCRIPTOR_HEAP_TYPE_RTV)

	// Create frame resources.
	rtvHandle := g.rtvDescriptorHeap.GetCPUDescriptorHandleForHeapStart()
	for i := 0; i < frameCount; i++ {
		r, err := g.swapChain.GetBuffer(uint32(i))
		if err != nil {
			return err
		}
		g.renderTargets[i] = r
		defer func(i int) {
			if ferr != nil {
				g.renderTargets[i].Release()
				g.renderTargets[i] = nil
			}
		}(i)

		if err := g.device.CreateRenderTargetView(r, nil, rtvHandle); err != nil {
			return err
		}
		rtvHandle.Offset(1, g.rtvDescriptorSize)
	}

	return nil
}

func (g *Graphics) SetWindow(window uintptr) {
	g.window = windows.HWND(window)
	// TODO: need to update the swap chain?
}

func (g *Graphics) Begin() error {
	g.frameIndex = -1
	if g.swapChain != nil {
		g.frameIndex = int(g.swapChain.GetCurrentBackBufferIndex())
	}

	idx := g.frameIndex
	if idx < 0 {
		idx = 0
	}
	if err := g.commandAllocators[idx].Reset(); err != nil {
		return err
	}
	if err := g.commandList.Reset(g.commandAllocators[idx], nil); err != nil {
		return err
	}

	if g.frameIndex >= 0 {
		barrierToRT := _D3D12_RESOURCE_BARRIER_Transition{
			Type:  _D3D12_RESOURCE_BARRIER_TYPE_TRANSITION,
			Flags: _D3D12_RESOURCE_BARRIER_FLAG_NONE,
			Transition: _D3D12_RESOURCE_TRANSITION_BARRIER{
				pResource:   g.renderTargets[idx],
				Subresource: _D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES,
				StateBefore: _D3D12_RESOURCE_STATE_PRESENT,
				StateAfter:  _D3D12_RESOURCE_STATE_RENDER_TARGET,
			},
		}
		g.commandList.ResourceBarrier(1, &barrierToRT)

		rtv := g.rtvDescriptorHeap.GetCPUDescriptorHandleForHeapStart()
		rtv.Offset(int32(idx), g.rtvDescriptorSize)

		clearColor := [...]float32{0.1, 0.25, 0.5, 1}
		g.commandList.ClearRenderTargetView(rtv, clearColor, 0, nil)

		g.commandList.OMSetRenderTargets(1, &rtv, false, nil)
	}

	return nil
}

func (g *Graphics) End() error {
	if g.frameIndex >= 0 {
		barrierToPresent := _D3D12_RESOURCE_BARRIER_Transition{
			Type:  _D3D12_RESOURCE_BARRIER_TYPE_TRANSITION,
			Flags: _D3D12_RESOURCE_BARRIER_FLAG_NONE,
			Transition: _D3D12_RESOURCE_TRANSITION_BARRIER{
				pResource:   g.renderTargets[g.frameIndex],
				Subresource: _D3D12_RESOURCE_BARRIER_ALL_SUBRESOURCES,
				StateBefore: _D3D12_RESOURCE_STATE_RENDER_TARGET,
				StateAfter:  _D3D12_RESOURCE_STATE_PRESENT,
			},
		}
		g.commandList.ResourceBarrier(1, &barrierToPresent)
	}

	if err := g.commandList.Close(); err != nil {
		return err
	}
	g.commandQueue.ExecuteCommandLists(1, &g.commandList)

	if g.frameIndex >= 0 {
		if err := g.swapChain.Present(1, 0); err != nil {
			return err
		}

		// Wait for the previous frame.
		fence := g.fences[g.frameIndex]
		g.fenceValues[g.frameIndex]++
		if err := g.commandQueue.Signal(fence, g.fenceValues[g.frameIndex]); err != nil {
			return err
		}

		nextIndex := (g.frameIndex + 1) % frameCount
		expected := g.fenceValues[nextIndex]
		actual := g.fences[nextIndex].GetCompletedValue()
		if actual < expected {
			if err := g.fences[nextIndex].SetEventOnCompletion(expected, g.fenceWaitEvent); err != nil {
				return err
			}
			const gpuWaitTimeout = 10 * 1000 // 10[s]
			if _, err := windows.WaitForSingleObject(g.fenceWaitEvent, gpuWaitTimeout); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Graphics) SetTransparent(transparent bool) {
}

func (g *Graphics) SetVertices(vertices []float32, indices []uint16) error {
	r := _D3D12_RANGE{0, 0}

	m, err := g.vertices.Map(0, &r)
	if err != nil {
		return err
	}
	copyFloat32s(m, vertices)
	if err := g.vertices.Unmap(0, nil); err != nil {
		return err
	}

	m, err = g.indices.Map(0, &r)
	if err != nil {
		return err
	}
	copyUint16s(m, indices)
	if err := g.indices.Unmap(0, nil); err != nil {
		return err
	}

	g.verticesView = _D3D12_VERTEX_BUFFER_VIEW{
		BufferLocation: g.vertices.GetGPUVirtualAddress(),
		SizeInBytes:    uint32(len(vertices)) * uint32(unsafe.Sizeof(float32(0))),
		StrideInBytes:  graphics.VertexFloatNum * uint32(unsafe.Sizeof(float32(0))),
	}
	g.indicesView = _D3D12_INDEX_BUFFER_VIEW{
		BufferLocation: g.indices.GetGPUVirtualAddress(),
		SizeInBytes:    uint32(len(indices)) * uint32(unsafe.Sizeof(uint16(0))),
		Format:         _DXGI_FORMAT_R16_UINT,
	}

	return nil
}

func (g *Graphics) NewImage(width, height int) (graphicsdriver.Image, error) {
	// TODO: Implement this
	return nullImage{}, nil
}

func (g *Graphics) NewScreenFramebufferImage(width, height int) (graphicsdriver.Image, error) {
	if err := g.updateSwapChain(width, height); err != nil {
		return nil, err
	}

	// TODO: Implement this
	return nullImage{}, nil
}

func (g *Graphics) SetVsyncEnabled(enabled bool) {
}

func (g *Graphics) SetFullscreen(fullscreen bool) {
}

func (g *Graphics) FramebufferYDirection() graphicsdriver.YDirection {
	return graphicsdriver.Downward
}

func (g *Graphics) NeedsRestoring() bool {
	return false
}

func (g *Graphics) NeedsClearingScreen() bool {
	// TODO: Confirm this is really true.
	return true
}

func (g *Graphics) IsGL() bool {
	return false
}

func (g *Graphics) HasHighPrecisionFloat() bool {
	return true
}

func (g *Graphics) MaxImageSize() int {
	// TODO: Return a correct value.
	return 4096
}

func (g *Graphics) NewShader(program *shaderir.Program) (graphicsdriver.Shader, error) {
	// TODO: Implement this.
	return nil, nil
}

func (g *Graphics) DrawTriangles(dst graphicsdriver.ImageID, srcs [graphics.ShaderImageNum]graphicsdriver.ImageID, offsets [graphics.ShaderImageNum - 1][2]float32, shader graphicsdriver.ShaderID, indexLen int, indexOffset int, mode graphicsdriver.CompositeMode, colorM graphicsdriver.ColorM, filter graphicsdriver.Filter, address graphicsdriver.Address, dstRegion, srcRegion graphicsdriver.Region, uniforms []graphicsdriver.Uniform, evenOdd bool) error {
	return nil
}

// nullImage is a temporary image which does nothing.
type nullImage struct{}

func (nullImage) ID() graphicsdriver.ImageID {
	return 0
}

func (nullImage) Dispose() {
}

func (nullImage) IsInvalidated() bool {
	return false
}

func (nullImage) Pixels() ([]byte, error) {
	return nil, nil
}

func (nullImage) ReplacePixels(args []*graphicsdriver.ReplacePixelsArgs) {
}

func copyFloat32s(dst unsafe.Pointer, src []float32) {
	var dsts []float32
	h := (*reflect.SliceHeader)(unsafe.Pointer(&dsts))
	h.Data = uintptr(dst)
	h.Len = len(src)
	h.Cap = len(src)
	copy(dsts, src)
}

func copyUint16s(dst unsafe.Pointer, src []uint16) {
	var dsts []uint16
	h := (*reflect.SliceHeader)(unsafe.Pointer(&dsts))
	h.Data = uintptr(dst)
	h.Len = len(src)
	h.Cap = len(src)
	copy(dsts, src)
}
