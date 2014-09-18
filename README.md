keyfwd
======

Forward (media-)keys to another computer

The tool intercepts key-presses on one or more Windows PCs and forwards them (encrypted) to another (target-) Windows computer, where it simulates a key-press accordingly.

It is intended to be used to forward media-keys (play, pause, next-track, etc.) to my laptop-computer, where I plugged in the headphones, while I work on my workstation, where I connected the keyboard. The "Synergy" tool can do something similar, but as of a bug in a Windows 8.1 Update, I was forced to switch to a RDP based "KVM" solution.

Usage
-----

### On the target machine
```
keyfwd.exe configure server
```
Enter the UDP port number and the encryption secret.
The encryption secret will be stored inside the Windows credential store.
The port number is stored inside the Windows registry.

Start the server using the following command:
```
keyfwd.exe server
```

### On the client machine
```
keyfwd.exe configure client
```
Enter the hostname of the target machine, the UDP port number (same as on the target machine) and the encryption secret.
Again, the encryption secret will be stored inside the Windows credential store and the hostname and port number are stored inside the Windows registry.

Start the client using the following command:
```
keyfwd.exe client
```


TODO
----

* Add a tray-icon and hide the console window
* maybe add Linux support (KDE?)
