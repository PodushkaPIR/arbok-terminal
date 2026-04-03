# Arbok Terminal

A terminal emulator written in Go with Fyne GUI framework.


## Requirements

- Go 1.21+
- GCC (for CGO)
- X11 development libraries (Linux)

### Linux dependencies

On Fedora/RHEL:
```bash
sudo dnf install gcc libX11-devel libXrandr-devel libXcursor-devel libXi-devel
```

On Debian/Ubuntu:
```bash
sudo apt install build-essential libx11-dev libxrandr-dev libxcursor-dev libxi-dev
```

On Arch Linux:
```bash
sudo pacman -S base-devel libx11 libxrandr libxcursor libxi
```

### X11 library setup (if needed)

Some systems need a symlink for Fyne:
```bash
mkdir -p ~/lib
ln -sf /usr/lib64/libXxf86vm.so.1 ~/lib/libXxf86vm.so
```

## Building

```bash
./build.sh
```

Or manually:
```bash
export CGO_LDFLAGS="-L$HOME/lib"  # if symlink was created
go build -o arbok ./cmd/arbok
```

## Running

### On X11
```bash
./arbok
```

### On Wayland (KDE, GNOME)
```bash
./arbok-launcher.sh
```

The launcher script handles XWayland auth automatically.


### Components

- **PTY Manager**: Creates pseudo-terminal, spawns shell, handles I/O streams
- **Screen Buffer**: 2D grid of cells (character + colors + attributes)
- **ANSI Parser**: Parses escape sequences, updates buffer state
- **Fyne UI**: Renders buffer to window, captures keyboard input

## In Progress

