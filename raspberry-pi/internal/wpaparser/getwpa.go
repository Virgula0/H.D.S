package wpaparser

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	log "github.com/sirupsen/logrus"
)

func GetWPA(handles map[string]*pcap.Handle) []*HandshakeInfo {
	var validFiles = make([]*HandshakeInfo, 0)

	for filePath, handle := range handles {
		// Set up a packet source for reading packets from the capture file
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

		// Store packets in a slice. Since it is an iterator it will give problems when iterated twice
		packets := make([]gopacket.Packet, 0, 1000) // Start with a capacity of 1000
		for packet := range packetSource.Packets() {
			packets = append(packets, packet)
		}

		// Map to track which BSSIDs we've already seen
		seenBSSIDs := make(map[string]string)

		// Call the function to find BSSID and SSID
		findBSSIDSSID(packets, seenBSSIDs)

		// Call the function to process the WPA handshake
		handshakeFound := processWPAHandshake(packets)
		if handshakeFound {
			log.Println("WPA2 Handshake successfully detected!")
		} else {
			log.Println("Not enough EAPOL packets to form a WPA2 handshake.")
		}

		for bssid, ssid := range seenBSSIDs {
			validFiles = append(validFiles, &HandshakeInfo{
				FilePath: filePath,
				BSSID:    bssid,
				SSID:     ssid,
			})
		}

		handle.Close()
	}

	return validFiles
}
