package wifi

import (
	"errors"
	"fmt"
	"github.com/mdlayher/wifi"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

// CheckHouseConnection checks if the system is connected to the specified SSID.
func checkHouseConnection(targetSSID string) (bool, error) {
	// Initialize a new Wi-Fi client.
	client, err := wifi.New()
	if err != nil {
		return false, fmt.Errorf("failed to create Wi-Fi client: %w", err)
	}
	defer client.Close()

	// Retrieve the list of Wi-Fi interfaces.
	interfaces, err := client.Interfaces()
	if err != nil {
		return false, fmt.Errorf("failed to get Wi-Fi interfaces: %w", err)
	}

	// Iterate over the interfaces to find the connected SSID.
	for _, iface := range interfaces {
		// Skip non-Wi-Fi interfaces
		if iface.Type != wifi.InterfaceTypeStation {
			continue
		}

		// Fetch the BSS (Basic Service Set) information for the interface.
		bss, err := client.BSS(iface)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return false, fmt.Errorf("failed to get BSS info for interface %s: %w", iface.Name, err)
		}

		// Compare the current SSID with the target SSID.
		if bss.SSID == targetSSID {
			return true, nil
		}
	}

	// If no interfaces are connected to the target SSID.
	return false, nil
}

// MonitorWiFiConnection ensures the device is connected to the target SSID.
func MonitorWiFiConnection(ssid string) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		connected, err := checkHouseConnection(ssid)
		if err != nil {
			logrus.Fatalf("[RSP-PI] Error checking Wi-Fi connection: %s", err.Error())
		}

		if connected {
			logrus.Printf("[RSP-PI] Successfully connected to '%s'", ssid)
			return
		}

		logrus.Printf("[RSP-PI] Not connected to '%s'. Re-attempting in 5 minutes...", ssid)
		<-ticker.C
	}
}
