# Arbok Terminal

A terminal emulator written in Go with Fyne GUI framework. Aims for xterm-256color compatibility — vim, tmux, htop, less should work.

## Features

- **ANSI escape sequences**: Full CSI support — cursor movement, erase display/line, insert/delete lines/chars, scroll up/down, scroll regions
- **256-color palette**: Standard xterm-256color + 24-bit true color (RGB)
- **SGR attributes**: Bold, dim, italic, underline, blink, reverse, strikethrough
- **Alt screen buffer**: `\e[?1049h/l` for vim, tmux, htop
- **Scrollback**: 1000-line ring buffer, populated on scroll
- **Cell-level diff rendering**: Only changed cells are redrawn
- **Dynamic resize**: Window resize updates terminal size via polling
- **Graceful shutdown**: Context-based goroutine lifecycle, SIGHUP → wait → SIGKILL
- **UTF-8**: Full Unicode support including multi-byte characters

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

## Testing

```bash
go test ./internal/terminal/ -v
```

90+ tests covering screen operations, ANSI parser, and VirtualTerminal lifecycle.

## Project Structure

```
cmd/arbok/main.go              ← entry point, goroutine wiring, graceful shutdown
internal/terminal/
  ├── screen.go                ← Screen: Grid + cursor + scroll region
  ├── virtual.go               ← VirtualTerminal: main/alt screen + scrollback + DEC modes
  ├── parser.go                ← ANSI state machine (7 states)
  └── color.go                 ← Color type + 256-palette
internal/pty/
  └── manager.go               ← PTY spawn, readLoop, graceful shutdown
internal/input/
  └── handler.go               ← key → ANSI escape sequences
internal/ui/
  ├── widget.go                ← Fyne widget
  ├── renderer.go              ← Cell-level diff renderer
  └── colors.go                ← Color → RGBA
```

See `docs/architecture.md` for detailed architecture documentation.
