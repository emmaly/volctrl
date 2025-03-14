package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

type AudioDevice struct {
	ID          string
	Name        string
	EndpointID  string
	DeviceState uint32
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	args := os.Args[1:]

	if args[0] == "list" {
		listDevices()
		return
	}

	if len(args) < 2 {
		fmt.Println("Error: Missing command")
		printUsage()
		os.Exit(1)
	}

	deviceName := args[0]
	command := args[1]

	// Initialize COM library
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		fmt.Printf("Error initializing COM: %v\n", err)
		os.Exit(1)
	}
	defer ole.CoUninitialize()

	// Handle commands
	switch command {
	case "up":
		amount := 5
		if len(args) > 2 {
			var err error
			amount, err = strconv.Atoi(args[2])
			if err != nil {
				fmt.Println("Error: Invalid amount value")
				os.Exit(1)
			}
		}
		increaseVolume(deviceName, amount)
	case "down":
		amount := 5
		if len(args) > 2 {
			var err error
			amount, err = strconv.Atoi(args[2])
			if err != nil {
				fmt.Println("Error: Invalid amount value")
				os.Exit(1)
			}
		}
		decreaseVolume(deviceName, amount)
	case "set":
		if len(args) < 3 {
			fmt.Println("Error: Missing volume value")
			os.Exit(1)
		}
		value, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("Error: Invalid volume value")
			os.Exit(1)
		}
		setVolume(deviceName, value)
	case "mute":
		setMute(deviceName, true)
	case "unmute":
		setMute(deviceName, false)
	case "toggle":
		toggleMute(deviceName)
	case "status":
		getStatus(deviceName)
	default:
		fmt.Println("Error: Unknown command:", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  volctrl <device-name> up [amount]    - Increase volume by amount (default 5%)")
	fmt.Println("  volctrl <device-name> down [amount]  - Decrease volume by amount (default 5%)")
	fmt.Println("  volctrl <device-name> set <value>    - Set volume to value (0-100%)")
	fmt.Println("  volctrl <device-name> mute           - Mute volume")
	fmt.Println("  volctrl <device-name> unmute         - Unmute volume")
	fmt.Println("  volctrl <device-name> toggle         - Toggle mute state")
	fmt.Println("  volctrl <device-name> status         - Display current volume and mute state")
	fmt.Println("  volctrl list                         - List all available audio devices")
}

func listDevices() {
	// Initialize COM library
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		fmt.Printf("Error initializing COM: %v\n", err)
		os.Exit(1)
	}
	defer ole.CoUninitialize()

	// Get IMMDeviceEnumerator
	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		fmt.Printf("Error creating device enumerator: %v\n", err)
		os.Exit(1)
	}
	defer mmde.Release()

	// Get devices
	var dc *wca.IMMDeviceCollection
	if err := mmde.EnumAudioEndpoints(wca.EAll, wca.DEVICE_STATE_ACTIVE, &dc); err != nil {
		fmt.Printf("Error enumerating devices: %v\n", err)
		os.Exit(1)
	}
	defer dc.Release()

	var count uint32
	if err := dc.GetCount(&count); err != nil {
		fmt.Printf("Error getting device count: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available audio devices:")
	fmt.Println("--------------------------------------------------")

	for i := uint32(0); i < count; i++ {
		var mmd *wca.IMMDevice
		if err := dc.Item(i, &mmd); err != nil {
			continue
		}
		defer mmd.Release()

		var ps *wca.IPropertyStore
		if err := mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
			continue
		}
		defer ps.Release()

		var pv wca.PROPVARIANT
		if err := ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
			continue
		}

		var id string
		mmd.GetId(&id)

		name := pv.String()
		fmt.Printf("Device: %s\n", name)
		fmt.Printf("ID: %s\n", id)
		fmt.Println("--------------------------------------------------")
	}
}

func getDeviceByName(name string) (*wca.IMMDevice, error) {
	// Get IMMDeviceEnumerator
	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return nil, fmt.Errorf("error creating device enumerator: %v", err)
	}
	defer mmde.Release()

	// Get devices
	var dc *wca.IMMDeviceCollection
	if err := mmde.EnumAudioEndpoints(wca.EAll, wca.DEVICE_STATE_ACTIVE, &dc); err != nil {
		return nil, fmt.Errorf("error enumerating devices: %v", err)
	}
	defer dc.Release()

	var count uint32
	if err := dc.GetCount(&count); err != nil {
		return nil, fmt.Errorf("error getting device count: %v", err)
	}

	// Find device by name
	for i := uint32(0); i < count; i++ {
		var mmd *wca.IMMDevice
		if err := dc.Item(i, &mmd); err != nil {
			continue
		}

		var ps *wca.IPropertyStore
		if err := mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
			mmd.Release()
			continue
		}

		var pv wca.PROPVARIANT
		if err := ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
			ps.Release()
			mmd.Release()
			continue
		}

		deviceName := pv.String()
		ps.Release()

		if strings.Contains(strings.ToLower(deviceName), strings.ToLower(name)) {
			return mmd, nil
		}

		mmd.Release()
	}

	return nil, fmt.Errorf("device not found: %s", name)
}

func getAudioEndpointVolume(device *wca.IMMDevice) (*wca.IAudioEndpointVolume, error) {
	var audioEndpointVolume *wca.IAudioEndpointVolume
	if err := device.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &audioEndpointVolume); err != nil {
		return nil, fmt.Errorf("error activating endpoint volume: %v", err)
	}
	return audioEndpointVolume, nil
}

func increaseVolume(deviceName string, amount int) {
	device, err := getDeviceByName(deviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer device.Release()

	aev, err := getAudioEndpointVolume(device)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer aev.Release()

	var currentVolume float32
	if err := aev.GetMasterVolumeLevelScalar(&currentVolume); err != nil {
		fmt.Printf("Error getting current volume: %v\n", err)
		os.Exit(1)
	}

	newVolume := currentVolume + (float32(amount) / 100.0)
	if newVolume > 1.0 {
		newVolume = 1.0
	}

	if err := aev.SetMasterVolumeLevelScalar(newVolume, nil); err != nil {
		fmt.Printf("Error setting volume: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Increased volume for %s by %d%% (now %.0f%%)\n", deviceName, amount, newVolume*100)
}

func decreaseVolume(deviceName string, amount int) {
	device, err := getDeviceByName(deviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer device.Release()

	aev, err := getAudioEndpointVolume(device)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer aev.Release()

	var currentVolume float32
	if err := aev.GetMasterVolumeLevelScalar(&currentVolume); err != nil {
		fmt.Printf("Error getting current volume: %v\n", err)
		os.Exit(1)
	}

	newVolume := currentVolume - (float32(amount) / 100.0)
	if newVolume < 0.0 {
		newVolume = 0.0
	}

	if err := aev.SetMasterVolumeLevelScalar(newVolume, nil); err != nil {
		fmt.Printf("Error setting volume: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Decreased volume for %s by %d%% (now %.0f%%)\n", deviceName, amount, newVolume*100)
}

func setVolume(deviceName string, value int) {
	if value < 0 || value > 100 {
		fmt.Println("Error: Volume value must be between 0 and 100")
		os.Exit(1)
	}

	device, err := getDeviceByName(deviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer device.Release()

	aev, err := getAudioEndpointVolume(device)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer aev.Release()

	newVolume := float32(value) / 100.0
	if err := aev.SetMasterVolumeLevelScalar(newVolume, nil); err != nil {
		fmt.Printf("Error setting volume: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Set volume for %s to %d%%\n", deviceName, value)
}

func setMute(deviceName string, mute bool) {
	device, err := getDeviceByName(deviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer device.Release()

	aev, err := getAudioEndpointVolume(device)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer aev.Release()

	if err := aev.SetMute(mute, nil); err != nil {
		fmt.Printf("Error setting mute state: %v\n", err)
		os.Exit(1)
	}

	action := "Muted"
	if !mute {
		action = "Unmuted"
	}
	fmt.Printf("%s %s\n", action, deviceName)
}

func toggleMute(deviceName string) {
	device, err := getDeviceByName(deviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer device.Release()

	aev, err := getAudioEndpointVolume(device)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer aev.Release()

	var muteState bool
	if err := aev.GetMute(&muteState); err != nil {
		fmt.Printf("Error getting mute state: %v\n", err)
		os.Exit(1)
	}

	if err := aev.SetMute(!muteState, nil); err != nil {
		fmt.Printf("Error setting mute state: %v\n", err)
		os.Exit(1)
	}

	action := "Muted"
	if muteState {
		action = "Unmuted"
	}
	fmt.Printf("%s %s\n", action, deviceName)
}

func getStatus(deviceName string) {
	device, err := getDeviceByName(deviceName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer device.Release()

	aev, err := getAudioEndpointVolume(device)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer aev.Release()

	var currentVolume float32
	if err := aev.GetMasterVolumeLevelScalar(&currentVolume); err != nil {
		fmt.Printf("Error getting current volume: %v\n", err)
		os.Exit(1)
	}

	var muteState bool
	if err := aev.GetMute(&muteState); err != nil {
		fmt.Printf("Error getting mute state: %v\n", err)
		os.Exit(1)
	}

	// Get the real device name
	var ps *wca.IPropertyStore
	if err := device.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		fmt.Printf("Error opening property store: %v\n", err)
		os.Exit(1)
	}
	defer ps.Release()

	var pv wca.PROPVARIANT
	if err := ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		fmt.Printf("Error getting device name: %v\n", err)
		os.Exit(1)
	}

	realDeviceName := pv.String()

	fmt.Printf("Status for %s\n", realDeviceName)
	fmt.Printf("Volume: %.0f%%\n", currentVolume*100)
	fmt.Printf("Mute: %v\n", muteState)
}