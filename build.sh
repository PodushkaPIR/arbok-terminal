#!/bin/bash
# Build script for arbok terminal

# Set up library paths for X11 (required for Fyne)
# Create symlink if needed:
#   mkdir -p ~/lib && ln -sf /usr/lib64/libXxf86vm.so.1 ~/lib/libXxf86vm.so

LIB_DIR="$HOME/lib"
if [ -d "$LIB_DIR" ] && [ -f "$LIB_DIR/libXxf86vm.so" ]; then
    export CGO_LDFLAGS="-L$LIB_DIR"
fi

# Build
echo "Building arbok..."
go build -o arbok ./cmd/arbok

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo ""
    echo "To run:"
    echo "  Wayland/KDE/GNOME: ./arbok-launcher.sh"
    echo "  X11:               ./arbok"
else
    echo "Build failed!"
    exit 1
fi
