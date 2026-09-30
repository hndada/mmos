# AOSP-Centred Mobile Windowing Reference

This document uses Android/AOSP as the primary reference for evolving this
minimal mobile OS. It maps responsibilities; it does not propose reproducing
Android's framework, Binder ABI, or vendor interfaces.

The core lesson is separation of authority. The app draws and submits content,
system_server decides policy, SurfaceFlinger latches and composes accepted
layers, and HWC/display hardware presents the result.

In particular, an app may request lifecycle or visual changes, but only the
server decides which process owns a window, which windows are visible, where
they appear, which one receives input, and whether submitted content enters the
composed frame.

## 1. AOSP responsibility map

| AOSP component | Primary responsibility | Boundary to preserve in this OS |
|---|---|---|
| App process | App state, UI tree, layout, raster/GPU rendering, buffer production | Client owns its pixels and widgets, never global window policy. |
| Framework UI runtime | View hierarchy, in-app input dispatch, IME integration, accessibility | Client-runtime concern; not a server scene graph. |
| system_server | System services and privileged policy | Keep process/session authority and policy outside every app process. |
| WindowManagerService (WMS) | Window lifecycle, placement, visibility, focus, transitions, display configuration | Corresponds to server-owned Window state and lifecycle policy. |
| InputManager / InputDispatcher | Device events, focus and touch-target selection, dispatch | Corresponds to server input routing and pointer capture. |
| SurfaceControl | Layer-state handle controlled through transactions | Useful split: content endpoint versus server-visible layer state. |
| Surface / ANativeWindow | Producer-facing endpoint for buffers | Conceptual future per-window presentation endpoint. |
| BufferQueue | Bounded producer/consumer buffer handoff | Later optimization; POC can retain only the latest accepted buffer. |
| SurfaceFlinger | Layer latching, scene composition, display timing, present feedback | Corresponds to compositor plus display scheduler. |
| Hardware Composer (HWC) | Hardware-plane assignment and display presentation | Keep behind a display backend, never client-facing. |
| Gralloc | Cross-process graphics-buffer allocation | Defer until real foreign/GPU buffer interop is needed. |

## 2. The key AOSP split: window, layer, surface, buffer

These nouns are deliberately different in AOSP. Keeping their responsibilities
separate avoids treating an app's pixel buffer as the window itself.

| Concept | AOSP meaning | POC equivalent | Ownership |
|---|---|---|---|
| Window | WMS policy record with attributes, token, bounds, visibility, focus, and input participation | Window record | Server |
| Layer | SurfaceFlinger compositing node with visual properties | Composed Layer / visible window state | Server/compositor |
| SurfaceControl | Control handle for layer hierarchy and properties | Future server transaction target | Server or specifically authorized client |
| Surface | Client producer endpoint backed by a BufferQueue | Future per-window presentation endpoint | App receives it; server authorizes it |
| Buffer | Pixel allocation queued to a surface | Buffer | Produced by app; accepted/released by server |
| BufferQueue | Producer/consumer queue and fence exchange | Future queue behind Present | Shared protocol with explicit ownership |
| Transaction | Atomic change set for layer properties and buffers | Future server transaction or atomic Present | Server validates before apply |
| Frame | Composed result for one display opportunity | Frame | Compositor |

One server-owned Window can initially carry its accepted latest Buffer. That is
smaller than Android's Surface + BufferQueue + SurfaceControl split, but it must
not collapse their authority: the app cannot choose global Z order, focus,
system-overlay status, or another app's visibility.

## 3. Control path versus buffer path

Android has two paths that meet at SurfaceFlinger but have different purposes.

| Concern | Control path | Buffer path | POC rule |
|---|---|---|---|
| Changes | Geometry and scene policy | Pixel content | Keep both atomic at Present until a transaction API is needed. |
| Authority | WMS/system policy; limited client controls | Owning app writes only its buffer | Validate that the session owns the target window. |
| Cadence | Infrequent or transition-driven | Usually one per animation/display tick | A new buffer is optional; reuse prior accepted content. |
| Completion signal | Transaction/latch/present callback | Release fence | Distinguish accepted, composed, presented, and reusable. |
| Security | Parentage, layer role, Z constraints | Protected content, buffer ownership | Enforce in the server, not through client convention. |

Conceptually, the paths are:

    app / WMS -> SurfaceControl transaction -> layer properties -> SurfaceFlinger
    app renderer -> draw -> queue buffer + acquire fence -> BufferQueue -> SurfaceFlinger

## 4. Lifecycle and launch

| AOSP flow | Purpose | POC mapping |
|---|---|---|
| Package manager resolves launch target | Chooses a trusted installed app identity | AppPackage lookup and launch request |
| Activity/task lifecycle starts or resumes app | Produces a process with lifecycle state | Process registry and app activation |
| WMS creates/attaches window state | Creates policy record before content is trusted | Server creates owned Window |
| WMS shows starting window / splash | Covers launch until real content is ready | Trusted splash stays visible while app prepares |
| App obtains a surface and draws | Produces first client buffer | App submits first Buffer |
| SurfaceFlinger latches and composes | Makes accepted content eligible for display | Present then Compose |
| First real frame appears | Removes temporary launch UI at visual boundary | Hide splash only after app revision is in composed frame |
| Process death / task removal | Removes owned state and releases graphics resources | Terminate process, remove owned windows, clear capture/input state |

The POC already preserves the vital policy: the server owns the splash and
dismisses it only after an accepted app frame has replaced it. It need not copy
AOSP's asynchronous Binder or BufferQueue mechanics to preserve that rule.

## 5. Window policy and system UI

| Policy | AOSP reference | Required server decision |
|---|---|---|
| Focus | WMS + input focus | Select one keyboard target according to display policy. |
| Touch target | InputDispatcher hit test | Choose the initial window; no app redirects a new stream to another app. |
| Pointer capture | Input policy / target | Scope capture to target window; cancel on up, hide, destroy, or interruption. |
| Z order | Window types, tokens, layer hierarchy | Limit app windows to an app band; system overlays have privileged bands. |
| Back / Home / Recents | System navigation and task policy | Keep fallback navigation and task restoration server-owned. |
| IME | System IME window and insets | IME is trusted system UI; app receives text/configuration events, not overlay authority. |
| Lock screen | SystemUI and keyguard | Lock content and unlock input cannot be an untrusted app window. |
| Notifications / shade | SystemUI | Publication is an app request; history, display, and dismissal are server policy. |
| Multi-window | WMS task/window configuration | Server computes bounds and asks app to relayout. |
| Multi-display | WMS display-area policy | Display ownership and window migration remain server decisions. |

## 6. Frame lifecycle and timing

| Moment | AOSP terminology | Meaning for the POC |
|---|---|---|
| Submitted | Producer queues a buffer | Server received a candidate buffer. |
| Accepted | Valid surface/layer transaction | Ownership, visibility, revision, and configuration checks passed. |
| Latched | SurfaceFlinger adopts buffer for composition | Candidate becomes compositor state for a display opportunity. |
| Composed | SurfaceFlinger produced display scene | Frame enumerates visible ordered layers. |
| Presented | HWC/display completed a present | Future feedback; Compose alone cannot claim this. |
| Released | Consumer no longer reads buffer | Future reuse signal after multi-buffer/foreign-buffer support. |

This prevents two incorrect shortcuts:

1. A successful Present is not proof that hardware displayed content.
2. Buffer release and presentation are separate; neither necessarily implies the
   timing of the other.

## 7. Display composition and hardware acceleration

| AOSP layer | What it decides | Minimal implementation now | Later implementation |
|---|---|---|---|
| SurfaceFlinger scene | Visible layers and properties | Ordered Frame.Layers | Retained scene plus damage/occlusion |
| Client composition | GPU renders composited target | CPU/simple buffer composition | Skia or Vulkan renderer |
| HWC validation | Which layers can use hardware planes | None | Plane eligibility and device/client split |
| HWC present | Sends target/layers to display hardware | VSync-driven simulated frame | DRM/KMS or platform backend |
| Direct scanout | Bypass GPU for eligible fullscreen layer | None | Only after format, transform, protection, occlusion checks |
| Color / HDR | Dataspace, color mode, display capabilities | One baseline color space | Explicit color metadata and conversion |

Do not expose an HWC abstraction in the app API. It is a backend optimization
selected after server policy has formed the display scene.

## 8. Input and text

| AOSP component | AOSP role | POC boundary |
|---|---|---|
| EventHub | Reads raw Linux input devices | Backend concern. |
| InputReader | Converts device data to normalized events | Produce stable pointer, key, scroll, text, system events. |
| InputDispatcher | Applies policy, targets a window, dispatches | Server owns hit testing, focus, coordinate conversion, capture. |
| App input dispatch | Routes inside app hierarchy | App owns widget hit testing and gestures after delivery. |
| IME / InputConnection | Produces and edits text with composition state | Text input is distinct from physical key presses. |
| Accessibility services | Consume semantics and perform actions | Add semantic tree/action contract with client runtime. |

Input delivery alone does not yield correct international text. A future client
text stack needs ICU for Unicode/bidi/locale, HarfBuzz for shaping, and FreeType
or an equivalent rasterizer for glyphs.

## 9. Mapping to the current repository

| Current concept | AOSP analogue | Important difference |
|---|---|---|
| Runtime | Small combination of WMS policy and SurfaceFlinger orchestration | Intentionally in-process and synchronous. |
| Session | Permission-bearing client connection / token | POC capability, not Binder identity. |
| Window | WMS window record plus simplified layer relationship | No separate SurfaceControl yet. |
| Buffer | Producer buffer | No gralloc handle, queue slots, or fences. |
| Present | Queue buffer plus allowed layer update | One validated server operation rather than distributed AOSP APIs. |
| Compose | SurfaceFlinger composition | Logical Frame, not GPU target or hardware present. |
| VSync | Scheduler/display timing | Simulated composition opportunity, not panel completion. |
| System windows | SystemUI / keyguard / IME windows | Same Go process, but retained as server-trusted roles. |

## 10. Sequenced growth path

| Stage | Add | Preserve from the start |
|---|---|---|
| Current | Lifecycle, visibility, input routing, composed logical frames | Sessions and ownership checks on every server mutation. |
| IPC boundary | Per-process connection and server commands | Server remains authoritative after transport is introduced. |
| Async first frame | Client readiness/frame acknowledgement | Splash hides after correct accepted/composed revision. |
| Buffer reuse | Bounded queue, acquire/release lifecycle | No reuse until consumer declares safety. |
| GPU interop | Foreign pixel buffers, fences, renderer | Validate provenance, format, protection. |
| Real display | Backend present feedback, DRM/KMS or platform path | Composition result and actual presentation remain separate. |
| Hardware planes | Overlay/direct-scanout selection | Compositor policy, never app-requested global privilege. |

## What not to import from AOSP yet

- Binder, Java framework classes, ActivityTaskManager, and compatibility APIs
- A full SurfaceControl hierarchy exposed to every app
- BufferQueue slots, gralloc handles, sync fences, and frame timelines
- HWC validation/present protocol, DRM atomic details, or vendor HAL contracts
- Transition frameworks that hide the basic ownership and lifecycle model

These mechanisms solve Android-scale compatibility, multi-process, and hardware
integration problems. The POC should first make equivalent authority, state
transitions, and observable frame results explicit.
