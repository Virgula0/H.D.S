The Raspberry Pi component is designed to **send network captures performed by `bettercap` to the server**. Its functionality is straightforward: you run the `bettercap` daemon on the Raspberry Pi, and it automatically transmits captured handshakes to the server whenever your local Wi-Fi SSID is detected nearby.

> [!NOTE]  
> This feature is **disabled** if the `TEST` environment variable is set to `False`. Ensure that your Raspberry Pi has at least **two network interfaces**, with one interface always connected to a stable network.

> **Tested on Raspberry Pi 5 Model B Rev. 2**

> [!TIP]  
> Update the `wlan1` interface to match your network interface. Make sure your Wi-Fi card supports **monitor mode** and **packet injection**.

---

## **What Does the Daemon Do?**

The daemon performs the following tasks:

1. Acts as a **TCP/IP client** to establish raw network connections.
2. Scans all `.PCAP` files located in the `~/handshakes` directory (typically where `bettercap` saves handshakes).
3. Utilizes the **`gopacket`** library to read `.PCAP` file layers, extracting **BSSID** and **SSID** information, and verifying if a **valid 4-way handshake** exists.
4. If a valid handshake is detected, the daemon **encodes the file in Base64** and sends it to the server.
5. Waits for a predefined **delay period** before repeating the process.

---

## **Run Bettercap**

This configuration sets up `bettercap` to capture handshakes and save them in the correct directory.

```bash
sudo bettercap -iface wlan1 -eval 'set wifi.handshakes.aggregate false; set wifi.handshakes.file ~/handshakes; wifi.recon on; set wifi.show.sort clients desc; set ticker.commands "wifi.deauth *; clear; wifi.show"; set ticker.period 60; ticker on';
```

> [!WARNING]  
> The daemon uses the `HOME` directory as its base. Since `bettercap` requires `sudo`, you must run the daemon as `root`.

---

## **Other Useful Bettercap Commands**

These commands can help you fine-tune your `bettercap` setup:

```bash
sudo bettercap -iface wlan1
wifi.recon BBSID
wifi.recon on
wifi.recon.channel N; # N is the channel to recon
```

---

## **Compile and Run the Daemon**

Make sure the following requirements are met before building and running the daemon:

> [!IMPORTANT]  
> The daemon requires `libpcap0.8-dev` to be installed on your system.

> [!IMPORTANT]  
> The file `/etc/machine-id` must exist on your machine.

1. Comile daemon with:

```bash
cd raspberry-pi
make build
```

2. Run with

```bash
./build/daemon --help
```

but remeber to export these env var first, change them according to your needs

```bash
export SERVER_HOST=localhost
export SERVER_PORT=4747
export TCP_ADDRESS=localhost
export TCP_PORT=4749
export TEST=False
export HOME_WIFI=Vodafone-A60818803 # Change with your SSID of your home Wireless Network
export BETTERCAP=True
```