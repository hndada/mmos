# Minimal Mobile OS Reference Map

This document maps the parts of a minimal mobile UI / windowing system to useful
reference implementations. The names are **conceptual analogies**, not proposed
API names and not necessarily 1:1 object mappings.

The boundary is deliberate: the OS owns process trust, window lifecycle, input
routing, composition policy, and presentation. An application owns its UI tree,
layout, paint commands, and app-local animation state. Flutter and Qt Quick are
therefore mainly references for the *client runtime*; Wayland, AOSP,
Chromium/Viz, and DRM/KMS describe more of the server-to-display path.

For a deeper Android-first reading of the same boundaries, see
[AOSP-Centred Mobile Windowing Reference](aosp.md).

## Reference scope

| Reference | Best used for | Do not copy as-is |
|---|---|---|
| Wayland / wlroots | Small IPC surface protocol; surface state and callbacks | Desktop shell policy and every optional extension |
| Android / AOSP | Separate app, window-policy, input, composition, and hardware services; atomic transactions | Activity framework, Binder details, and vendor compatibility layers |
| Chromium / Viz | Frame production, aggregation, resource lifetime, begin-frame scheduling | Browser-specific frame-sink topology and web security machinery |
| Flutter | Client-owned retained widget/render tree, layout, paint, semantics, animation | Treating the application tree as OS-visible window state |
| Qt Quick | Retained scene graph and GUI/render-thread handoff | QML/Qt object model and platform abstraction layers |
| DRM/KMS | Connectors, CRTCs, planes, atomic modesets, page flips, direct scanout | Linux kernel-driver API as an application-facing contract |
| Skia | 2D raster, paths, images, and text blobs | A scene graph or window manager |
| Vulkan | Explicit GPU queues, images, command buffers, and synchronization | A UI toolkit, compositor policy, or display API |
| HarfBuzz / FreeType / ICU | Text shaping, glyph rasterization, Unicode and locale data | Windowing, layout policy, or input routing |
| libinput | Linux device normalization and seat input | App-level gestures or focus policy |
| Core Animation | Transactional retained layers and presentation-oriented animation | A portable OS/display backend |

## 1. Process, IPC, and authority

| Concept | Wayland / wlroots | Android / AOSP | Chromium / Viz | Minimal OS takeaway |
|---|---|---|---|---|
| UI client | Wayland client | App process | Renderer process | App process owns its UI and rendering input. |
| Window-policy authority | Compositor / shell | WindowManager in `system_server` | Browser process / Aura | Keep focus, visibility, z-order, launch, and system overlays server-owned. |
| Composition authority | Compositor | SurfaceFlinger | Viz | Compose only server-accepted app content. |
| Transport | Unix socket / Wayland protocol | Binder | Mojo | Start with a narrow request/reply transport; authenticate every server-visible object. |
| Object identity | Per-client Wayland object ID | Binder token / layer handle | `FrameSinkId`, `SurfaceId` | Use opaque, server-issued IDs; never trust app-supplied ownership. |
| Capability | Bound protocol object | Binder permission / token | Interface binding and process isolation | Session capability should scope window creation and presentation. |
| Lifecycle | Client connection closes | Process/activity lifecycle | Renderer lifetime | Server removes process-owned windows and resources on termination. |
| Trusted system UI | Shell-owned surface | SystemUI / privileged windows | Browser UI | Lock, splash, IME, and notice UI need a separate trusted path. |

## 2. Window, surface, and configuration

| Concept | Wayland / wlroots | Android / AOSP | Chromium / Viz | Minimal OS takeaway |
|---|---|---|---|---|
| Top-level window | `xdg_toplevel` | `Window` / `WindowState` | Widget / Aura `Window` | Server record: owner, bounds, role, visibility, focus, display. |
| Drawable endpoint | `wl_surface` | `Surface` | `viz::Surface` | A surface is where a client submits content, not necessarily an OS window. |
| Role / policy | `xdg_surface` role | Window type and token | Widget / view type | Make roles explicit: app, launcher, lock, IME, popup. |
| Child content | `wl_subsurface` | Child `SurfaceControl` | Embedded `SurfaceId` | Add only when child composition is needed; do not make app widgets server surfaces. |
| Configure | `xdg_surface.configure` | Relayout / configuration dispatch | Visual-properties update | Server supplies display bounds, safe area, orientation, scale, and configuration revision. |
| Configure acknowledgement | `ack_configure` | App redraw after configuration | Visual-properties ACK | Associate presented content with the configuration it was drawn for. |
| Show / hide | Buffer attach / null buffer | Show / hide transaction | Surface activation / eviction | Visibility is policy state, distinct from buffer lifetime. |
| Focus | Seat focus | Focused window | Focused RenderWidget | Server alone chooses focus. |
| Resize / rotation | Configure + buffer scale/transform | Configuration change + transaction | Visual properties + local surface ID | Reject or retain stale-size content by explicit revision, not frame-number guesses. |

## 3. Client UI, layout, and retained scene

| Concept | Flutter | Qt Quick | Chromium renderer | Minimal OS takeaway |
|---|---|---|---|---|
| Declarative UI tree | Widget tree | QML `Item` tree | Blink DOM / layout tree | App-private tree; it is not part of the server protocol. |
| Layout | Constraints → size → position | Anchors, positioners, layouts | Blink layout | App turns its UI state plus configuration into geometry. |
| Retained render tree | RenderObject / layer tree | `QSGNode` tree | `cc::Layer` tree | Retain client-side structure to avoid repainting unchanged content. |
| Painting | `Canvas` / `Picture` | Scene graph geometry/material | Paint artifacts / display lists | Produce pixels or compositable primitives inside the app runtime. |
| Animation | Ticker / animation controller | Animation / render loop | Compositor animation | Advance app-local animation on a display-aligned tick. |
| Semantics / accessibility | Semantics tree | Accessibility tree | AX tree | Keep a parallel semantic tree; pixels alone are insufficient. |
| Render threading | UI and raster threads | GUI/render thread handoff | Main, compositor, raster threads | Model thread ownership before adding concurrency. |

## 4. Buffers, resources, and ownership

| Concept | Wayland / wlroots | Android / AOSP | Chromium / Viz | Skia / Vulkan | Minimal OS takeaway |
|---|---|---|---|---|---|
| Pixel buffer | `wl_buffer` | `GraphicBuffer` / `AHardwareBuffer` | `SharedImage` | `SkSurface` / `VkImage` | Presentable content with width, height, format, and revision. |
| CPU buffer | `wl_shm` | Shared or gralloc buffer | SharedBitmap / software image | Raster surface / mapped memory | Useful first implementation; make ownership and copying explicit. |
| GPU handle | dma-buf FD | Native handle | Mailbox | Image + memory binding | Put foreign-handle import behind a later interop boundary. |
| Allocation | Client allocator / GBM | gralloc | GMB / SharedImage factory | allocator / `vkAllocateMemory` | Allocation policy is separate from window policy. |
| Producer | Client | App, media, GL producer | Renderer / compositor | Raster or GPU queue | Producer must finish writing before submission. |
| Consumer | Compositor | SurfaceFlinger | Viz | Compositor renderer | Consumer releases only after reading finishes. |
| Queue | Client protocol cadence | `BufferQueue` | Frame/resource pipeline | Swapchain-like reuse | Start with one accepted latest buffer per window; add queue depth only when needed. |
| Resource return | `wl_buffer.release` | Release fence | `ReturnedResource` | Fence / timeline semaphore | Returned means reusable, not necessarily presented. |
| Protected content | Protocol / compositor policy | Protected buffers | Protected-video path | Protected memory support | Mark policy at the server layer and exclude it from capture. |

## 5. State, transactions, and scene composition

| Concept | Wayland / wlroots | Android / AOSP | Chromium / Viz | Core Animation | Minimal OS takeaway |
|---|---|---|---|---|---|
| Pending state | `wl_surface` pending state | `SurfaceControl.Transaction` | Pending compositor state | Layer transaction | Keep mutations pending until one apply point. |
| Commit | `wl_surface.commit()` | `Transaction.apply()` | `SubmitCompositorFrame()` | `CATransaction` commit | Buffer and metadata become visible atomically. |
| Multi-object update | Synchronized subsurfaces | One transaction across layers | Aggregated frame | Transaction across layers | Update multiple windows coherently when required. |
| Scene node | Surface / compositor node | SurfaceFlinger `Layer` | Surface, render pass, draw quad | `CALayer` | Server scene contains windows/layers, not app widgets. |
| Geometry | Surface state / shell policy | Transaction properties | Shared quad state | Layer properties | Position, transform, crop, alpha, z, visibility belong in server state. |
| Damage | `damage_buffer()` | Layer/buffer damage | Damage rect | Invalidated layer region | Damage is a work-reduction hint, never a correctness boundary. |
| Occlusion | Compositor decision | WM / SurfaceFlinger | cc / Viz | Compositor decision | Skip hidden work only after correctness is established. |
| Animation | Toolkit/compositor-specific | Layer transaction animation | Compositor animation | Implicit/explicit animation | System transitions animate server-owned scene properties, not app pixels. |

### Shared visual properties

| Property | Wayland / wlroots | Android / AOSP | Chromium / Viz | Minimal OS contract |
|---|---|---|---|---|
| Position / size | Shell + subsurface position | Transaction position / crop | Transform / rect | Logical display coordinates and bounds. |
| Transform / rotation | Buffer transform / viewport | Transaction matrix | `gfx::Transform` | State coordinate conversion for both content and input. |
| Scale | Buffer scale / viewport | Display density / matrix | Device scale factor | Separate logical units from buffer pixels. |
| Clip / crop | Viewporter / compositor | Crop / window crop | Clip / visible rect | Clip before final composition. |
| Alpha | Compositor extension/policy | Layer alpha | Opacity | Server-controlled per layer. |
| Z order | Surface hierarchy / shell | Layer Z | Quad/pass order | Policy determines ordering; clients do not choose arbitrary global Z. |
| Opaque region | `set_opaque_region()` | Opaque-layer hint | Opaque metadata | Optional optimization hint. |
| Input region | `set_input_region()` | Input window region | Hit-test regions | Server uses it only within an authorized window. |
| Rounded corners / shadow | Toolkit / compositor | Layer metadata / SystemUI | Compositor UI | Visual policy for system chrome, not a protocol prerequisite. |

## 6. Frame scheduling, synchronization, and presentation

| Concept | Wayland / wlroots | Android / AOSP | Chromium / Viz | DRM/KMS / Vulkan | Minimal OS takeaway |
|---|---|---|---|---|---|
| Refresh source | Compositor/backend | HWC VSYNC | Platform VSync | Vblank / present mode | Per-display pacing begins at the display backend. |
| Client tick request | `wl_surface.frame` | `Choreographer` | `SetNeedsBeginFrame()` | — | Client asks for one future production opportunity. |
| Tick delivery | `wl_callback.done` | `doFrame()` | `OnBeginFrame()` | Vblank timestamp | Include monotonic time and display/configuration identity. |
| Submission | Commit | Buffer queue + transaction | Compositor-frame submit | Queue submit | Submission transfers ownership/dependencies, not presentation. |
| Acquire sync | Explicit-sync acquire point | Acquire fence | `SyncToken` | Semaphore / fence | Consumer waits before sampling a buffer. |
| Release sync | Release point | Release fence | Returned resource + token | Semaphore / fence | Producer waits before reusing a buffer. |
| Latching | Compositor chooses state | SurfaceFlinger latches | Viz activates surface | Page-flip preparation | Define which submitted state joins next composition. |
| Present | Presentation-time feedback | Present fence / frame timeline | `PresentationFeedback` | Atomic commit / page flip | Report actual presentation separately from acceptance. |
| Backpressure | Frame callback + release | `BufferQueue` depth | Frame ACK | Swapchain image availability | Bound outstanding buffers and drop superseded work deliberately. |
| No-update frame | No commit | Reuse previous buffer | `DidNotProduceFrame()` | No new flip needed | Reuse accepted content without inventing a revision. |

## 7. Input, text, and accessibility

| Concept | Wayland / wlroots | Android / AOSP | Chromium | Supporting reference | Minimal OS takeaway |
|---|---|---|---|---|---|
| Raw device read | libinput / evdev | `EventHub` | Platform event source | libinput | Keep hardware discovery below the server input API. |
| Normalization | libinput | `InputReader` | `ui/events` | — | Normalize into pointer, key, text, scroll, and system events. |
| Routing | Compositor seat focus | `InputDispatcher` | `InputRouter` | — | Hit test server windows, then deliver only to the selected owner. |
| Pointer / touch | `wl_pointer`, `wl_touch` | `MotionEvent` | Pointer/touch events | — | Preserve stream ID, phases, coordinates, and timestamps. |
| Pointer capture | Protocol extensions | Pointer capture | Pointer capture / lock | — | Server grants capture; end it on up, cancel, hide, or process loss. |
| Keyboard focus | `wl_keyboard` focus | Focused window | Focused RenderWidget | — | One focused target per seat/display policy. |
| Text input / IME | Text-input protocol | IME / `InputConnection` | IME integration | ICU, HarfBuzz | Text editing is not equivalent to raw key events. |
| Text shaping | Client toolkit | Framework | Blink text stack | HarfBuzz + ICU | Shape Unicode runs; apply bidi, script, locale, and fallback before glyph raster. |
| Glyph rasterization | Client toolkit | Framework | Text renderer | FreeType / Skia | Cache glyphs as an app-renderer concern. |
| Gesture recognition | Toolkit/client | Framework | Blink/compositor | — | Keep app gestures out of the OS except reserved system gestures. |
| Accessibility | Client/toolkit | Accessibility services | AX tree | Platform a11y APIs | Expose semantic nodes/actions alongside pixels. |

## 8. Display backend and final composition

| Concept | DRM/KMS | Android / AOSP | Wayland / wlroots | Chromium / Viz | Minimal OS takeaway |
|---|---|---|---|---|---|
| Logical display | Connector + CRTC mode | `Display` | `wl_output` | `viz::Display` | Publish logical bounds, scale, orientation, safe area, and configuration revision. |
| Physical backend | Connector, encoder, CRTC | HWC / display HAL | DRM backend | Platform backend | Hide backend-specific handles behind a display interface. |
| Composition engine | GPU / planes | SurfaceFlinger | Compositor renderer | `SkiaRenderer` | Compose the server scene into an output target. |
| Output target | Framebuffer | HWC client target | DRM framebuffer | `OutputSurface` | Renderer owns final-target lifetime. |
| Hardware plane | KMS plane | HWC layer | Direct scanout / overlay | Overlay candidate | Server selects plane use as an optimization. |
| Direct scanout | Plane to CRTC | Device composition | Direct scanout | Overlay path | Require suitable policy, format, transform, protection, and occlusion. |
| Mode change | Atomic modeset | Display-config change | Output-mode change | Display reconfiguration | Treat as new display configuration and relayout clients. |
| Page flip / present | Atomic commit | `presentDisplay()` | Backend present | Swap/present | Present output on vblank-aligned scheduling. |
| Color management | Properties / blobs | Color mode / dataspace | Color-management protocol | `gfx::ColorSpace` | Defer wide color/HDR but keep color space explicit in a future buffer contract. |
| Capture | Writeback connector | Virtual display / capture | Screencopy | Copy output | Capture composed frames server-side and exclude protected layers. |

## End-to-end frame flow

```text
input device
  -> input normalization and server hit test
  -> app event handler
  -> app UI tree: state -> layout -> paint / retained scene
  -> pixel buffer or GPU image (Skia / Vulkan)
  -> authenticated Present(window, buffer, revision)
  -> server validates owner, window state, revision, and configuration
  -> server latches accepted buffers and composes the window scene
  -> renderer / hardware planes produce output
  -> DRM/KMS, HWC, or platform backend presents at vblank
  -> presentation feedback and buffer release reach the app
```

## POC-sized contract

The current codebase can remain intentionally smaller than any reference above.
The minimum useful boundary is:

| OS contract | Owns | Explicitly does not own |
|---|---|---|
| `Session` | App identity and server capability | Client UI objects |
| `Window` | Owner, role, display, bounds, visibility, focus, Z order | App widget hierarchy |
| `Buffer` | Dimensions, pixel format, revision, protected flag | How the app lays out or paints it |
| `Present` | Ownership/configuration/revision validation and accepted content | GPU allocation details initially |
| `Frame` | Ordered visible layers for one display revision | Proof that hardware displayed it |
| `VSync` / display backend | Composition opportunity and later presentation feedback | Application animation policy |
| Input route | Target selection, coordinate conversion, capture, focus | App-local gestures and text layout |

### Deliberate deferrals

- Real IPC transport and asynchronous server-to-process commands
- Multi-buffer queues, dma-buf / `AHardwareBuffer` import, and fence plumbing
- GPU renderer and Vulkan command submission
- Plane assignment, direct scanout, HDR, and color management
- Full IME protocol, shaping/raster cache, and accessibility bridge

These are growth paths, not missing abstractions to add before the current
server-owned lifecycle and frame contract needs them.
