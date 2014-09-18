// +build windows
package main

import (
	"encoding/binary"
	"encoding/json"
	"github.com/AllenDang/w32"
	"github.com/danieljoos/wincred"
	"syscall"
	"unsafe"
)

const (
	CLIENT_CONFIGURATION_KEY            = "Software\\danieljoos\\keyfwd\\client"
	CLIENT_CONFIGURATION_HOSTNAME       = "Hostname"
	CLIENT_CONFIGURATION_PORT           = "Port"
	CLIENT_CONFIGURATION_FORWARDED_KEYS = "ForwardedKeys"
	CLIENT_CONFIGURATION_SECRET         = "danieljoos/keyfwd/client"
	SERVER_CONFIGURATION_KEY            = "Software\\danieljoos\\keyfwd\\server"
	SERVER_CONFIGURATION_PORT           = "Port"
	SERVER_CONFIGURATION_SECRET         = "danieljoos/keyfwd/server"
)

var (
	modadvapi32       = syscall.NewLazyDLL("advapi32.dll")
	procRegSetValueEx = modadvapi32.NewProc("RegSetValueExW")
)

// Store the given string value into the Windows registry.
// Calls the 'RegSetValueEx' function with type REG_SZ:
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms724923(v=vs.85).aspx
func regSetString(hKey w32.HKEY, subKey string, value string) (errno int) {
	var lptr, vptr unsafe.Pointer
	if len(subKey) > 0 {
		ptr, _ := syscall.UTF16PtrFromString(subKey)
		lptr = unsafe.Pointer(ptr)
	}
	if len(value) > 0 {
		ptr, _ := syscall.UTF16PtrFromString(value)
		vptr = unsafe.Pointer(ptr)
	}
	ret, _, _ := procRegSetValueEx.Call(
		uintptr(hKey),
		uintptr(lptr),
		uintptr(0),
		uintptr(w32.REG_SZ),
		uintptr(vptr),
		uintptr(len(value)*2))
	return int(ret)
}

// Store the given QWORD value (64 bit integer) into the Windows registry.
// Calls the 'RegSetValueEx' function with type REG_QWORD:
// http://msdn.microsoft.com/en-us/library/windows/desktop/ms724923(v=vs.85).aspx
func regSetQWORD(hKey w32.HKEY, subKey string, value uint64) (errno int) {
	var lptr, vptr unsafe.Pointer
	if len(subKey) > 0 {
		ptr, _ := syscall.UTF16PtrFromString(subKey)
		lptr = unsafe.Pointer(ptr)
	}
	vptr = unsafe.Pointer(&value)
	ret, _, _ := procRegSetValueEx.Call(
		uintptr(hKey),
		uintptr(lptr),
		uintptr(0),
		uintptr(w32.REG_QWORD),
		uintptr(vptr),
		8)
	return int(ret)
}

// Returns a new ClientConfiguration object, filled with the configuration data, loaded from the
// Windows registry (Hostname, Port, ForwardedKeys) and Windows credential store (encryption secret).
func LoadClientConfiguration() *ClientConfiguration {
	ret := new(ClientConfiguration)
	ret.Hostname = w32.RegGetString(w32.HKEY_CURRENT_USER, CLIENT_CONFIGURATION_KEY, CLIENT_CONFIGURATION_HOSTNAME)
	ret.Port = binary.LittleEndian.Uint64(w32.RegGetRaw(w32.HKEY_CURRENT_USER, CLIENT_CONFIGURATION_KEY, CLIENT_CONFIGURATION_PORT))
	json.Unmarshal([]byte(w32.RegGetString(w32.HKEY_CURRENT_USER, CLIENT_CONFIGURATION_KEY, CLIENT_CONFIGURATION_FORWARDED_KEYS)), &ret.ForwardedKeys)
	cred, err := wincred.GetGenericCredential(CLIENT_CONFIGURATION_SECRET)
	if err == nil {
		ret.Secret = cred.CredentialBlob
	}
	return ret
}

// Saves the given ClientConfiguration object to the Windows registry and Windows credential store.
func StoreClientConfiguration(configuration *ClientConfiguration) {
	jsonForwardedKeys, _ := json.Marshal(configuration.ForwardedKeys)

	regKey := w32.RegCreateKey(w32.HKEY_CURRENT_USER, CLIENT_CONFIGURATION_KEY)
	regSetString(regKey, CLIENT_CONFIGURATION_HOSTNAME, configuration.Hostname)
	regSetQWORD(regKey, CLIENT_CONFIGURATION_PORT, configuration.Port)
	regSetString(regKey, CLIENT_CONFIGURATION_FORWARDED_KEYS, string(jsonForwardedKeys))

	cred := wincred.NewGenericCredential(CLIENT_CONFIGURATION_SECRET)
	cred.CredentialBlob = configuration.Secret
	cred.Write()
}

// Load the server related configuration data from the Windows registry and Windows credential store.
// Returns a ServerConfiguration object.
func LoadServerConfiguration() *ServerConfiguration {
	ret := new(ServerConfiguration)
	ret.Port = binary.LittleEndian.Uint64(w32.RegGetRaw(w32.HKEY_CURRENT_USER, SERVER_CONFIGURATION_KEY, SERVER_CONFIGURATION_PORT))
	cred, err := wincred.GetGenericCredential(SERVER_CONFIGURATION_SECRET)
	if err == nil {
		ret.Secret = cred.CredentialBlob
	}
	return ret
}

// Saves the given ServerConfiguration object to the Windows registry and Windows credential store.
func StoreServerConfiguration(configuration *ServerConfiguration) {
	regKey := w32.RegCreateKey(w32.HKEY_CURRENT_USER, SERVER_CONFIGURATION_KEY)
	regSetQWORD(regKey, SERVER_CONFIGURATION_PORT, configuration.Port)

	cred := wincred.NewGenericCredential(SERVER_CONFIGURATION_SECRET)
	cred.CredentialBlob = configuration.Secret
	cred.Write()
}

// Returns an array of keys to forward to the remote host.
// As this little tool was intended to forward media keys, the default
// set consists of those keys.
func GetDefaultForwardedKeys() []int {
	return []int{
		w32.VK_VOLUME_MUTE,
		w32.VK_VOLUME_DOWN,
		w32.VK_VOLUME_UP,
		w32.VK_MEDIA_NEXT_TRACK,
		w32.VK_MEDIA_PREV_TRACK,
		w32.VK_MEDIA_STOP,
		w32.VK_MEDIA_PLAY_PAUSE,
		w32.VK_PLAY,
		w32.VK_PAUSE,
	}
}
