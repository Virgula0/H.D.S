package wpaparser

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	log "github.com/sirupsen/logrus"
)

const UnknownType = "Unknown"

// FindBSSIDSSID Function to process the packet and return BSSID, SSID, and a found flag
func FindBSSIDSSID(packets []gopacket.Packet, seenBSSIDs map[string]string) {
	for _, packet := range packets {
		dot11 := getDot11Layer(packet)
		if dot11 == nil {
			continue
		}

		bssid := getBSSID(dot11)
		if bssid == "" || isInvalidBSSID(bssid) {
			continue
		}

		ssid := getSSID(packet)
		if ssid == "" {
			continue
		}

		if _, seen := seenBSSIDs[bssid]; !seen {
			seenBSSIDs[bssid] = ssid
			log.Printf("BSSID: %v SSID: %q\n", bssid, ssid)
		}
	}
}

// Extracts the Dot11 layer from the packet
func getDot11Layer(packet gopacket.Packet) *layers.Dot11 {
	dot11Layer := packet.Layer(layers.LayerTypeDot11)
	if dot11Layer == nil {
		return nil
	}
	dot11, _ := dot11Layer.(*layers.Dot11)
	return dot11
}

// Extracts the BSSID from a Dot11 layer
func getBSSID(dot11 *layers.Dot11) string {
	return dot11.Address3.String()
}

// Checks if a BSSID is invalid (e.g., broadcast address)
func isInvalidBSSID(bssid string) bool {
	return bssid == "ff:ff:ff:ff:ff:ff"
}

// Extracts the SSID from the packet, if present
func getSSID(packet gopacket.Packet) string {
	dot11InfoLayer := packet.Layer(layers.LayerTypeDot11InformationElement)
	if dot11InfoLayer == nil {
		return ""
	}

	dot11Info, _ := dot11InfoLayer.(*layers.Dot11InformationElement)
	if dot11Info.ID == layers.Dot11InformationElementIDSSID {
		return string(dot11Info.Info)
	}

	return ""
}

// ProcessWPAHandshake Function to process WPA handshake by checking EAPOL messages and extracting key info
func ProcessWPAHandshake(packets []gopacket.Packet) bool {
	handshakePackets := make([]*layers.EAPOL, 4) // Store up to 4 EAPOL packets
	eapolCount := 0
	keyVersionDesc := ""

	for _, packet := range packets {
		eapolPacket := getEAPOLLayer(packet)
		if eapolPacket == nil {
			continue
		}

		if eapolCount == 0 {
			keyVersionDesc = processFirstEAPOLPacket(eapolPacket)
		}

		if eapolCount < 4 {
			handshakePackets[eapolCount] = eapolPacket
			eapolCount++
		}

		if isHandshakeComplete(eapolCount, keyVersionDesc) {
			printHandshakeDetails(handshakePackets)
			return true
		}
	}

	return false // Return false if not enough EAPOL packets were found
}

// Extracts the EAPOL layer from the packet
func getEAPOLLayer(packet gopacket.Packet) *layers.EAPOL {
	eapolLayer := packet.Layer(layers.LayerTypeEAPOL)
	if eapolLayer == nil {
		return nil
	}
	eapolPacket, _ := eapolLayer.(*layers.EAPOL)
	return eapolPacket
}

// Processes the first EAPOL packet to extract and print key descriptor details
func processFirstEAPOLPacket(eapolPacket *layers.EAPOL) string {
	log.Println("EAPOL M1 detected")
	log.Println("EAPOL Payload:", eapolPacket.Payload)

	if len(eapolPacket.Payload) >= 3 { // Ensure we have enough bytes to extract Key Information field
		keyInfo := uint16(eapolPacket.Payload[1])<<8 | uint16(eapolPacket.Payload[2])
		keyVersionDesc := getKeyDescriptorVersion(keyInfo)
		log.Printf("Key Information: 0x%04x\n", keyInfo)
		log.Printf("    .... .... .... .XXX = Key Descriptor Version: %s (%d)\n", keyVersionDesc, keyInfo&0x07)
		return keyVersionDesc
	}
	return UnknownType
}

// Checks if the handshake is complete based on EAPOL packet count and key version descriptor
func isHandshakeComplete(eapolCount int, keyVersionDesc string) bool {
	return eapolCount == 4 && keyVersionDesc != UnknownType
}

// Prints handshake details for all 4 EAPOL packets
func printHandshakeDetails(handshakePackets []*layers.EAPOL) {
	log.Println("WPA2 Handshake Detected!")
	log.Println("Message 1: AP -> Client (EAPOL)")
	log.Println("Message 2: Client -> AP (EAPOL)")
	log.Println("Message 3: AP -> Client (EAPOL)")
	log.Println("Message 4: Client -> AP (EAPOL)")

	for i, eapolPacket := range handshakePackets {
		if eapolPacket != nil {
			log.Printf("Message %d: %v\n", i+1, eapolPacket.Payload)
		}
	}
}

// Helper function to extract the Key Descriptor Version (3 bits)
func getKeyDescriptorVersion(keyInfo uint16) string {
	// Mask the last 3 bits to determine the Key Descriptor Version
	descriptorVersion := keyInfo & 0x07 // Mask the last 3 bits (0x07 = 00000111)

	// Interpret based on the last 3 bits
	switch descriptorVersion {
	case 0:
		return UnknownType
	case 1:
		return "HMAC-MD5"
	case 2:
		return "HMAC-SHA1-128"
	default:
		return "Invalid"
	}
}
