# Index of Mobile GUI API & Operating System

Checklist legend: `[x]` has a code implementation; `[ ]` is not implemented.

## Volume 2 — Mobile Operating System: Core & Advanced

### Core

#### I. Client and Server Boundary

- [x] 1. Architecture
  - [x] Shared model contract
  - [x] Data-only launch request/reply contract with an in-process transport endpoint
  - [x] Client-owned UI tree and drawing
  - [x] Server-owned lifecycle and system policy
- [x] 2. Client Runtime
  - [x] App process, windows, and UI nodes
  - [x] Command hit testing and focus order
  - [x] Buffer drawing and resize handling
- [x] 3. Server Runtime
  - [x] Installed packages and process registry
  - [x] Opaque sessions and window ownership checks
  - [x] Frame submission, composition, and VSync

#### II. Single-window Display

- [x] 4. Window Model
  - [x] Server-visible window record and bounds
  - [x] Buffer revisions and stale-buffer rejection
  - [x] Visibility, focus, z-order, and Back history
- [x] 5. Input Routing
  - [x] Pointer, key, text, scroll, and system events
  - [x] Display-to-window coordinate conversion
  - [x] Pointer capture and release
- [x] 6. Display and Composition
  - [x] Logical display bounds and safe areas
  - [x] Layered frame composition
  - [x] Screen power and orientation state
- [x] 7. Configuration
  - [x] Theme and configuration revisions
  - [x] Client relayout after configuration changes

#### III. Scenario Flows

- [x] 8. Boot, Splash, and Launch
  - [x] Package launch, process start, and cached resume
  - [x] System-owned splash while the first app frame is prepared
  - [x] Splash dismissal after the submitted revision is presented
- [x] 9. Home, Back, and App Navigation
  - [x] Home launcher input and app activation
  - [x] App-local navigation history
  - [x] Server-owned Back fallback and foreground restoration
- [x] 10. Termination, Cache, and Eviction
  - [x] Process termination and owned-window cleanup
  - [x] Caching after the last visible window closes
  - [x] Memory-pressure eviction and opaque saved state
  - [x] Resource quotas and watchdog checks
- [x] 11. Lock and Screen Power
  - [x] Trusted lock surface and unlock input
  - [x] Screen power state
- [x] 12. Settings UI
  - [x] Server-owned configuration API
  - [x] Settings application and controls
- [x] 13. Notices
  - [x] App notice publication
  - [x] Trusted notice history and dismissal
  - [x] Notification shade gesture and overlay

### Advanced Features

#### IV. Tasks and Displays

- [x] 14. Task and History
  - [x] Recents task cards
  - [x] Task selection and foreground restoration
- [x] 15. Multi-window
  - [x] Multiple visible windows in one scene
  - [x] Split, divider resize, and swap
  - [x] Multiple instances of supported packages
- [x] 16. Multi-display
  - [x] Independent display scenes
  - [x] Window movement between displays
  - [x] Display connection and removal rules
- [ ] 17. Foreign PixelBuffer Interop

#### V. System UI and Motion

- [x] 18. System Chrome
  - [x] Trusted system-window registration and presentation
  - [x] Lock, splash, recents, and notice surfaces
  - [x] Input Method Editor (IME) and On-Screen Keyboard
- [x] 19. Privileged Buffers
  - [x] System-only buffer submission
  - [x] App-owned content-protection requests
- [x] 20. System Transition
  - [x] Per-display fade between composed buffers

#### VI. Trust and Privacy

- [x] 21. Package Trust and Permission
  - [x] Package digest, signature verification, and installation
  - [x] Permission request, resolution, query, and revocation
  - [ ] Runtime Permission Dialog
- [x] 22. Capture and Privacy
  - [x] Protected layers excluded from captures
  - [x] Trusted recording lifecycle

## Volume 3 — Mobile Operating System Performance & Optimization

- [x] 23. Resource Limits and Watchdog
  - [x] Per-package CPU, GPU, and memory limits
  - [x] Usage accounting and reset
  - [x] Unresponsive-process detection

## TODO — IPC-ready Execution

- [ ] Server-to-process commands for process start and activity resume
- [ ] Per-process authenticated connection with server-issued capability binding
- [ ] IPC contracts for window attach, activation, and buffer submission
- [ ] Asynchronous first-frame handshake before dismissing the launch splash
- [ ] Activity, task, and server-owned back-stack records
- [ ] Process lifecycle commands: create, start, resume, pause, stop, and destroy
- [ ] Transport conformance tests for ordering, failures, timeouts, and duplicate delivery
- [ ] Request-to-frame trace IDs for launch, lifecycle, buffer, and composition events
