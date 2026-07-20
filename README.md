# Arbok Terminal

Terminal emulator in Go + Fyne. xterm-256color compatible — vim, tmux, htop, less work.

## Features

- CSI commands: cursor movement, erase, insert/delete lines/chars, scroll regions
- 256-color palette + 24-bit true color
- SGR attributes: bold, dim, italic, underline, blink, reverse, strikethrough
- Alt screen buffer (mode 1049) for vim, tmux, htop
- Scrollback with mouse wheel and PageUp/PageDown
- Cell-level diff rendering
- Dynamic resize via polling
- Graceful shutdown (context-based)
- UTF-8

## Requirements

- Go 1.26+
- GCC (for CGO)
- X11 development libraries (Linux)

### Linux dependencies

Fedora/RHEL:
```bash
sudo dnf install gcc libX11-devel libXrandr-devel libXcursor-devel libXi-devel
```

Debian/Ubuntu:
```bash
sudo apt install build-essential libx11-dev libxrandr-dev libxcursor-dev libxi-dev
```

Arch Linux:
```bash
sudo pacman -S base-devel libx11 libxrandr libxcursor libxi
```

## Build & Run

```bash
./run.sh build
./arbok                  # X11
./arbok-launcher.sh      # Wayland
```

## Test

```bash
./run.sh test
./run.sh all             # vet + fmt + test
```
