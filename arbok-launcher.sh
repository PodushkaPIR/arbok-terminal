#!/bin/bash
# Launcher for arbok terminal emulator
# Works with KDE Plasma / GNOME on Wayland via XWayland

# Disable Wayland for this process - arbok will use XWayland
export WAYLAND_DISPLAY=""
export XDG_SESSION_TYPE=x11

# Use current X authority if available, otherwise try system locations
if [ -z "$XAUTHORITY" ] || [ ! -f "$XAUTHORITY" ]; then
    if [ -f "$HOME/.Xauthority" ]; then
        export XAUTHORITY="$HOME/.Xauthority"
    elif [ -f "/run/user/$(id -u)/xauth_"* ]; then
        export XAUTHORITY=$(ls -1 /run/user/$(id -u)/xauth_* 2>/dev/null | head -1)
    fi
fi

# Launch arbok
exec ./arbok "$@"
