# UI / Windowing System Concept Mapping

> This table compares roughly equivalent concepts across Wayland, Android/AOSP, and Chromium/Viz.
> The mappings are conceptual, not always 1:1.

## 1. Object / Window Model

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| UI Client | Wayland client | App process | Renderer process |
| System / Window Authority | Wayland compositor | WindowManager / `system_server` | Browser process |
| Composition Service | Wayland compositor | SurfaceFlinger | Viz |
| GPU Service | Compositor / driver | GPU driver / vendor HAL | GPU process |
| IPC | Wayland protocol / socket | Binder | Mojo |
| IPC Object Identity | Wayland object ID | Binder object / token | Mojo remote + IDs |
| Top-level Window | `xdg_toplevel` | `Window` / `WindowState` | Widget / Aura `Window` |
| Popup | `xdg_popup` | Sub-window / popup window | Popup widget |
| Drawable Surface | `wl_surface` | `Surface` | `viz::Surface` |
| Surface Identity | Wayland object ID | Surface / layer handle | `SurfaceId` |
| Surface Namespace | Per-client object IDs | Binder / layer handles | `FrameSinkId` |
| Surface Generation | Object lifetime | Surface generation | `LocalSurfaceId` |
| Window Role | `xdg_surface` role | Window type / token | Widget / view type |
| Child Surface | `wl_subsurface` | Child `SurfaceControl` | Embedded `SurfaceId` |
| Layer Control | Surface state | `SurfaceControl` | `cc::Layer` / frame state |
| Window Hierarchy | Surface / subsurface tree | `WindowContainer` / layer tree | FrameSink / Surface tree |
| Configure Window | `xdg_surface.configure` | Relayout / configuration | Visual-properties update |
| Configure ACK | `ack_configure` | Relayout / draw ACK | Visual-properties ACK |
| Map / Show | Commit buffer | Show transaction + buffer | Surface activation |
| Unmap / Hide | Null buffer / unmap | Hide / remove transaction | Surface eviction / deactivation |

## 2. Buffer / Frame Model

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Pixel Buffer | `wl_buffer` | `GraphicBuffer` / `AHardwareBuffer` | `SharedImage` |
| GPU Buffer Handle | dma-buf FD | Native handle / `AHardwareBuffer` | `gpu::Mailbox` |
| CPU Shared Buffer | `wl_shm` | Shared / gralloc buffer | SharedBitmap / software SharedImage |
| GPU Buffer Import | `linux-dmabuf` | gralloc / EGL / Vulkan import | SharedImage backing |
| Buffer Allocator | Client allocator / GBM | gralloc | GMB / SharedImage factory |
| Buffer Queue | Client-managed | `BufferQueue` | Frame / resource queues |
| Producer | Client | App / media / GL producer | Renderer / compositor |
| Consumer | Compositor | SurfaceFlinger | Viz |
| Attach Buffer | `wl_surface.attach()` | `queueBuffer()` | Resource referenced by `CompositorFrame` |
| Submit State | `wl_surface.commit()` | `SurfaceControl.Transaction.apply()` | `SubmitCompositorFrame()` |
| Frame Payload | Surface state + `wl_buffer` | Buffer + transaction | `CompositorFrame` |
| Resource List | Attached buffers | BufferQueue slots | `TransferableResource` |
| Resource Handle | `wl_buffer` | Native buffer handle | Mailbox |
| Resource Return | `wl_buffer.release` | Release fence | `ReturnedResource` |
| Buffer Reuse | After release | Dequeue / release | Returned resources |

## 3. Transaction / Atomic Update

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Pending State | `wl_surface` pending state | `SurfaceControl.Transaction` | Pending frame / compositor state |
| Atomic Commit | `wl_surface.commit()` | `Transaction.apply()` | `SubmitCompositorFrame()` |
| Multi-object Transaction | Synchronized subsurfaces | `SurfaceControl.Transaction` | Frame aggregation |
| Geometry Update | Surface / shell state | Transaction properties | Quad / shared state |
| Buffer + Metadata Atomicity | Commit | Transaction | `CompositorFrame` |
| Transaction Callback | Protocol events | Transaction listeners | Frame ACK |
| State Latching | Compositor | SurfaceFlinger | Viz Surface activation |
| Pending vs Current State | Explicit | Explicit | Pending / active frame |

## 4. Damage / Invalidation

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Damage | `wl_surface.damage*()` | Buffer / layer damage | Damage rect |
| Buffer-space Damage | `damage_buffer()` | Buffer damage | Resource / pass damage |
| Surface-space Damage | `damage()` | Layer damage | Compositor damage |
| Dirty-region Tracking | Until commit | Renderer / SF | `DamageTracker` |
| Partial Redraw | Client responsibility | Renderer + SF | cc / Viz |
| No-change Frame | No commit | Reuse previous buffer | `DidNotProduceFrame()` |
| Occlusion | Compositor | SurfaceFlinger / WM | cc / Viz |

## 5. Scene / Composition Model

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Scene Node | `wl_surface` | SurfaceFlinger `Layer` | `viz::Surface` / quad |
| Client-side Scene Node | Toolkit-specific | View / RenderNode | `cc::Layer` |
| Scene Hierarchy | Surface tree | Layer tree | Surface / RenderPass tree |
| Render Layer | Compositor-specific | SurfaceFlinger Layer | `cc::Layer` / DrawQuad |
| Composition Unit | Surface | Layer | `DrawQuad` |
| Render Pass | Compositor-specific | Layer subtree | `CompositorRenderPass` |
| Shared Visual State | Surface state | `SurfaceControl` state | `SharedQuadState` |
| Transform | Surface transform | Transaction matrix | `gfx::Transform` |
| Position | Subsurface position | `setPosition()` | Transform / rect |
| Scale | `buffer_scale` / viewport | Matrix / transform | Device scale / transform |
| Rotation | Buffer transform | Matrix | Display / quad transform |
| Crop | Viewporter | Crop / window crop | Clip / visible rect |
| Clip | Toolkit / compositor | Crop | Clip rect |
| Alpha | Extension / compositor | Layer alpha | Opacity |
| Z-order | Surface hierarchy | Layer Z | Quad / pass order |
| Opaque Region | `set_opaque_region()` | Opaque layer state | Opaque metadata |
| Input Region | `set_input_region()` | Input window region | HitTestRegion |
| Rounded Corner | Extension-specific | Layer metadata | Rounded-corner metadata |
| Shadow | Toolkit / compositor | Shell / render layer | Views / Blink / compositor |

## 6. Frame Scheduling

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Display Refresh Source | Compositor / backend | HWC VSYNC | Platform VSync |
| Client Frame Tick | `wl_surface.frame` | `Choreographer` | `BeginFrame` |
| Tick Request | `frame()` | Frame callback | `SetNeedsBeginFrame()` |
| Tick Event | `wl_callback.done` | `doFrame()` | `OnBeginFrame()` |
| Frame Scheduler | Compositor | SurfaceFlinger scheduler | `DisplayScheduler` |
| Frame ID | Callback / serial | Frame timeline token | BeginFrame ID / frame token |
| No Frame Produced | No commit | Missed / skipped frame | `DidNotProduceFrame()` |
| Backpressure | Buffer release / callback | BufferQueue | Frame ACK |
| Submission ACK | Release semantics | Buffer / transaction callback | `DidReceiveCompositorFrameAck()` |
| Presentation Feedback | Presentation-time protocol | Present fence / stats | `PresentationFeedback` |
| Refresh-rate Handling | Output mode | SurfaceFlinger / HWC | BeginFrame / VSync |

## 7. Synchronization

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Implicit GPU Sync | dma-buf implicit sync | Supported | Platform-dependent |
| Explicit Sync | DRM syncobj protocol | Sync fence | `gpu::SyncToken` |
| Producer Finished Writing | Acquire point / fence | Acquire fence | SyncToken |
| Consumer Finished Reading | Release point / fence | Release fence | Resource return + SyncToken |
| Frame Presented | Presentation feedback | Present fence | PresentationFeedback |
| GPU Dependency | DRM syncobj | Fence FD | SyncToken |
| Cross-process Sync | FD / timeline | Fence FD | GPU service SyncToken |

## 8. Input Routing

| Concept | Wayland | Android / AOSP | Chromium |
|---|---|---|---|
| Input Subsystem | Compositor / libinput | Native input stack | Browser / Viz input |
| Device Collection | `wl_seat` | `InputDevice` | Platform event source |
| Raw Device Read | libinput / evdev | `EventHub` | OS backend |
| Input Normalization | libinput | `InputReader` | `ui/events` |
| Event Router | Compositor | `InputDispatcher` | InputRouter |
| Pointer | `wl_pointer` | MotionEvent | WebMouseEvent / PointerEvent |
| Keyboard | `wl_keyboard` | KeyEvent | WebKeyboardEvent |
| Touch | `wl_touch` | MotionEvent | WebTouchEvent |
| Gesture Recognition | Client / toolkit | App / framework | Blink / compositor |
| Hit Testing | Surface selection | WM / InputDispatcher | Viz HitTestRegion |
| Input Region | `set_input_region()` | Input window region | `HitTestRegionList` |
| Pointer Focus | `wl_pointer.enter/leave` | Focused / touched window | Target RenderWidget |
| Keyboard Focus | `wl_keyboard.enter/leave` | Focused Window | Focused RenderWidget |
| Event ACK | Protocol semantics | InputDispatcher ACK | InputEvent ACK |
| Pointer Capture | Protocol extensions | Pointer capture | Pointer lock / capture |
| Cursor | Cursor surface | PointerIcon / Sprite | Platform cursor |

## 9. Display / Final Composition

| Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Logical Display | `wl_output` | `Display` | `viz::Display` |
| Physical Display Backend | DRM / KMS | HWC / display HAL | Platform backend |
| Display Compositor | Wayland compositor | SurfaceFlinger | Viz `Display` |
| Scene Aggregation | Compositor | SurfaceFlinger | `SurfaceAggregator` |
| Renderer | Compositor renderer | SurfaceFlinger composition | SkiaRenderer |
| Composition Output | DRM framebuffer | HWC client target | `OutputSurface` |
| Hardware Overlay | KMS plane | HWC device composition | Overlay processor |
| Direct Scanout | KMS direct scanout | HWC device composition | Platform overlay path |
| Display Plane | DRM plane | HWC layer / plane | Overlay candidate |
| Present / Page Flip | DRM atomic commit | `presentDisplay()` | Swap / present |
| Multiple Displays | Multiple `wl_output` | DisplayManager / SF | Platform displays |
| Virtual Display | Virtual output | VirtualDisplay | Offscreen / capture path |
| Display Transform | Output transform | Display transform | Output transform |
| Refresh Rate | Output mode | HWC display config | VSync / BeginFrame |
| Color Space | Color-management protocol | ColorSpace / SF / HWC | `gfx::ColorSpace` |

# Condensed Mapping

| Abstract Concept | Wayland | Android / AOSP | Chromium / Viz |
|---|---|---|---|
| Client | Wayland client | App | Renderer |
| Window | `xdg_toplevel` | Window / WindowState | Widget / Aura Window |
| Surface | `wl_surface` | `Surface` | `viz::Surface` |
| Surface ID | Object ID | Surface / layer handle | `SurfaceId` |
| Layer Control | Surface state | `SurfaceControl` | `cc::Layer` / frame state |
| Child Surface | `wl_subsurface` | Child SurfaceControl | Embedded Surface |
| Buffer | `wl_buffer` | GraphicBuffer | SharedImage |
| Buffer Handle | dma-buf | AHardwareBuffer | Mailbox |
| Allocator | GBM / client allocator | gralloc | GMB / SharedImage |
| Buffer Queue | Client-managed | BufferQueue | Frame / resource pipeline |
| Frame | Commit state | Buffer + Transaction | CompositorFrame |
| Commit | `wl_surface.commit()` | `Transaction.apply()` | `SubmitCompositorFrame()` |
| Damage | `damage_buffer()` | Damage region | Damage rect |
| Transaction | Surface commit | SurfaceControl.Transaction | CompositorFrame |
| Scene Node | Surface | Layer | Surface / Quad |
| Scene Tree | Surface tree | Layer tree | Surface tree |
| Transform | Surface state | Transaction matrix | `gfx::Transform` |
| Z-order | Surface hierarchy | Layer Z | Draw order |
| Hit Test | Compositor | WM / InputDispatcher | Viz HitTest |
| Frame Tick | Frame callback | Choreographer | BeginFrame |
| Backpressure | Buffer release | BufferQueue / fence | Frame ACK |
| GPU Sync | DRM syncobj | Sync fence | SyncToken |
| Buffer Release | `wl_buffer.release` | Release fence | ReturnedResource |
| Presentation Feedback | Presentation protocol | Present fence / stats | PresentationFeedback |
| Compositor | Wayland compositor | SurfaceFlinger | Viz |
| Scene Aggregation | Compositor | SurfaceFlinger | SurfaceAggregator |
| Output Renderer | Compositor renderer | SF client composition | SkiaRenderer |
| Output Target | DRM framebuffer | HWC client target | OutputSurface |
| Hardware Composition | DRM / KMS | HWC | Platform overlays |
| Display | `wl_output` | Display | Platform display |
| Input Devices | `wl_seat` | EventHub / InputReader | Platform event source |
| Input Routing | Compositor | InputDispatcher | InputRouter / Viz |
| IPC | Wayland protocol | Binder | Mojo |

# Conceptual Architecture

## Wayland

```text
Client
  ↓
wl_surface
  ↓
wl_buffer
  ↓
commit()
  ↓
Wayland Compositor
  ↓
DRM / KMS
  ↓
Display