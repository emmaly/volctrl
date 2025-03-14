# volctrl

## Description

`volctrl` is a simple command-line utility for controlling the volume of your system. It allows you to increase, decrease, set, and mute the volume using CLI arguments. The device name is a required argument.

## Commands

- `volctrl <device-name> up [amount]`: Increase the volume by the specified amount (default is 5%).
- `volctrl <device-name> down [amount]`: Decrease the volume by the specified amount (default is 5%).
- `volctrl <device-name> set <value>`: Set the volume to a specific value (0-100%).
- `volctrl <device-name> mute`: Mute the volume.
- `volctrl <device-name> unmute`: Unmute the volume.
- `volctrl <device-name> toggle`: Toggle the mute state (mute if unmuted, unmute if muted).
- `volctrl <device-name> status`: Display the current volume and mute state.
- `volctrl list`: List all available audio devices.

## Implementation Details

- Written in Go 1.24 or later.
- Support for Windows.
