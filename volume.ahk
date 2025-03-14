#Requires AutoHotkey v2.0

VolumeChange(dir) {
  ToolTip "Volume " . dir
  Run "volctrl.exe WH-1000XM5 " . dir,, "Hide"
}

; Volume control hotkeys (RALT+RSHIFT+F24 for volume up, RCTRL+RSHIFT+F24 for volume down)
>!>+F24::VolumeChange("up")
>^>+F24::VolumeChange("down")
