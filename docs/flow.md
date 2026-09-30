# Android / AOSP
App
  ↓
Surface
  ↓
BufferQueue
  ↓
SurfaceControl.Transaction
  ↓
SurfaceFlinger
  ↓
HWC
  ↓
Display


# Chromium / Viz
Renderer
  ↓
cc::Layer / SharedImage
  ↓
CompositorFrame
  ↓
Viz Surface
  ↓
SurfaceAggregator
  ↓
SkiaRenderer
  ↓
OutputSurface
  ↓
Platform Display

Wayland
  Surface-centric
  Client produces pixel buffers
  Compositor owns presentation and window policy

Android / AOSP
  Surface + Layer + Transaction-centric
  BufferQueue connects producers and consumers
  SurfaceFlinger owns system composition

Chromium / Viz
  Frame + Scene Description + Resource-centric
  Renderers submit CompositorFrames
  Viz aggregates surfaces and produces final frames