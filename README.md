# MESNEVI External

A high-performance external player tracking and visualization tool developed in Go, utilizing native Windows GDI and Win32 APIs via direct memory reading (RPM).

---

## 🎨 Visual (ESP) Features

- **2D Bounding Box ESP:** Draws dynamic bounding boxes around players based on their position and scale.
- **Skeleton ESP:** Renders real-time player bone hierarchies (Head, Neck, Spine, Pelvis, Arms, Legs).
  - **Customization:** Cycle instantly through Blue, Green, and Red skeleton color profiles.
- **Head Circle:** Displays a circular indicator targeting head coordinates.
- **Dynamic Health Bar:** Color-shifting HP bar (Red ↔ Green) dynamically calculated from current health.
- **Numeric Health Text:** Displays exact numeric HP values.
- **Player Name Display:** Renders sanitized player nicknames above the entity box.
- **Distance Display:** Real-time distance calculation relative to the local player, displayed in meters.

---

## 🛠️ Technical & System Architecture

- **Stream Proof / Capture Bypass:** Utilizes `SetWindowDisplayAffinity` (`WDA_EXCLUDEFROMCAPTURE`) to hide the overlay from streaming and screen capture tools (OBS, Discord, etc.).
- **Embedded Offsets:** Integrates JSON memory offsets directly into the compiled binary via Go's `go:embed`.
- **Transparent Overlay:** Built using Win32 API (`WS_EX_TRANSPARENT`, `WS_EX_LAYERED`, `WS_EX_TOPMOST`) for a click-through, flicker-free double-buffered (`BitBlt`) layer.
- **WorldToScreen Engine:** Mathematical transformation matrix pipeline converting 3D world coordinates to 2D screen positions via View Matrix.
- **Team Check:** Configurable filter to ignore team members and render enemy targets only.

---

## 💻 Console UI & Hotkeys

- **Animated RGB ASCII Banner:** Dynamic color-wave header rendered directly in the terminal using HSV-to-RGB color space conversions.
- **Live Status Dashboard:** Real-time console UI reflecting active/inactive module states without screen flickering.
- **Audible Feedback:** System audio cues (beeps) confirming hotkey state toggles.

### Keybindings

| Key | Function |
| :--- | :--- |
| **F1** | Toggle Box ESP |
| **F2** | Toggle Skeleton ESP |
| **F3** | Toggle Head Circle |
| **F4** | Toggle Team Check |
| **F5** | Toggle Health Bar |
| **F6** | Toggle Player Name |
| **F7** | Toggle Distance Display |
| **F8** | Cycle Skeleton Color (Blue / Green / Red) |
| **END** | Emergency Exit / Panic Key |
