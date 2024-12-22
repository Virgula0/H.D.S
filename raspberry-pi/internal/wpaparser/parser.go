package wpaparser

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
	"strings"
)

const UnknownType = "Unknown"

type HandshakeInfo struct {
	FilePath string
	BSSID    string
	SSID     string
}

// -----------------------------
// BSSID and SSID Extraction
// -----------------------------

// findBSSIDSSID processes packets to extract BSSID and SSID.
func findBSSIDSSID(packets []gopacket.Packet, seenBSSIDs map[string]string) {
	for _, packet := range packets {
		if dot11 := extractDot11Layer(packet); dot11 != nil {
			bssid := dot11.Address3.String()
			if bssid == "ff:ff:ff:ff:ff:ff" || bssid == "" {
				continue // Skip invalid BSSIDs
			}

			// Check if it's a management frame
			if dot11.Type != layers.Dot11TypeMgmt {
				continue
			}

			if ssid := extractSSID(packet); ssid != "" {
				if _, seen := seenBSSIDs[bssid]; !seen {
					seenBSSIDs[bssid] = strings.Trim(ssid, "\"")
					log.Printf("BSSID: %v SSID: %q\n", bssid, ssid)
				}
			}
		}
	}
}

// extractDot11Layer gets the Dot11 layer from a packet.
func extractDot11Layer(packet gopacket.Packet) *layers.Dot11 {
	if layer := packet.Layer(layers.LayerTypeDot11); layer != nil {
		if dot11, ok := layer.(*layers.Dot11); ok {
			return dot11
		}
	}
	return nil
}

// extractSSID retrieves the SSID from a packet.
func extractSSID(packet gopacket.Packet) string {
	if layer := packet.Layer(layers.LayerTypeDot11InformationElement); layer != nil {
		if infoElem, ok := layer.(*layers.Dot11InformationElement); ok && infoElem.ID == layers.Dot11InformationElementIDSSID {
			return string(infoElem.Info)
		}
	}
	return ""
}

// -----------------------------
// WPA Handshake Processing
// -----------------------------

// processWPAHandshake validates a 4-way WPA handshake.
func processWPAHandshake(packets []gopacket.Packet) bool {
	handshakePackets := make([]*layers.EAPOL, 0, 4) // Preallocate capacity for 4 packets
	var keyVersionDesc string

	for _, packet := range packets {
		eapol := extractEAPOLLayer(packet)
		if eapol == nil {
			continue
		}

		switch len(handshakePackets) {
		case 0:
			// Process the first EAPOL packet to get the key descriptor
			keyVersionDesc = processFirstEAPOLPacket(eapol)
			fallthrough
		case 1, 2, 3:
			// Append the next EAPOL packets until we reach 4
			handshakePackets = append(handshakePackets, eapol)
		case 4:
			// Validate handshake completeness
			if keyVersionDesc != UnknownType {
				printHandshakeDetails(handshakePackets)
				return true
			}
		}
	}

	return false
}

// extractEAPOLLayer retrieves the EAPOL layer from a packet.
func extractEAPOLLayer(packet gopacket.Packet) *layers.EAPOL {
	if layer := packet.Layer(layers.LayerTypeEAPOL); layer != nil {
		if eapol, ok := layer.(*layers.EAPOL); ok {
			return eapol
		}
	}
	return nil
}

// processFirstEAPOLPacket processes the first EAPOL packet to extract key descriptor details.
func processFirstEAPOLPacket(eapol *layers.EAPOL) string {
	log.Println("EAPOL M1 detected")
	log.Println("EAPOL Payload:", eapol.Payload)

	if len(eapol.Payload) >= 3 {
		keyInfo := uint16(eapol.Payload[1])<<8 | uint16(eapol.Payload[2])
		keyVersion := keyInfo & 0x07 // Extract the last 3 bits

		keyVersionDesc := map[uint16]string{
			0: UnknownType,
			1: "HMAC-MD5",
			2: "HMAC-SHA1-128",
		}[keyVersion]

		if keyVersionDesc == "" {
			keyVersionDesc = "Invalid"
		}

		log.Printf("Key Information: 0x%04x\n", keyInfo)
		log.Printf(".... .... .... .XXX = Key Descriptor Version: %s (%d)\n", keyVersionDesc, keyVersion)
		return keyVersionDesc
	}
	return UnknownType
}

// printHandshakeDetails logs the details of a successful WPA handshake.
func printHandshakeDetails(handshakePackets []*layers.EAPOL) {
	log.Println("WPA2 Handshake Detected!")
	messages := []string{
		"Message 1: AP -> Client",
		"Message 2: Client -> AP",
		"Message 3: AP -> Client",
		"Message 4: Client -> AP",
	}

	for i, packet := range handshakePackets {
		if packet != nil {
			log.Printf("%s (EAPOL Payload: %v)\n", messages[i], packet.Payload)
		}
	}
}
