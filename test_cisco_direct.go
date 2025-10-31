package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Simple test program to validate Cisco DevNet connectivity
func main() {
	fmt.Println("🚀 Testing Cisco DevNet Sandbox Connectivity...")

	// Your DevNet credentials
	host := "devnetsandboxiosxec8k.cisco.com"
	username := "adarshkumarrr07"
	password := "48_ECdl9-gKuhQ"

	// Create HTTP client with SSL verification disabled for lab environments
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr, Timeout: 30 * time.Second}

	// Test 1: Get device hostname
	fmt.Println("\n1️⃣ Testing RESTCONF API - Getting device hostname...")
	url := fmt.Sprintf("https://%s:443/restconf/data/Cisco-IOS-XE-native:native/hostname", host)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	req.SetBasicAuth(username, password)
	req.Header.Set("Accept", "application/yang-data+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response Status: %s\n", resp.Status)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ API request failed: %s\n", string(body))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("❌ Failed to decode response: %v\n", err)
		return
	}

	fmt.Printf("✅ Success! Device response: %+v\n", result)

	// Test 2: Get interfaces
	fmt.Println("\n2️⃣ Testing RESTCONF API - Getting interfaces...")
	interfacesURL := fmt.Sprintf("https://%s:443/restconf/data/ietf-interfaces:interfaces", host)

	req2, err := http.NewRequest("GET", interfacesURL, nil)
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	req2.SetBasicAuth(username, password)
	req2.Header.Set("Accept", "application/yang-data+json")

	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	fmt.Printf("📡 Response Status: %s\n", resp2.Status)

	if resp2.StatusCode == http.StatusOK {
		var interfacesResult map[string]interface{}
		if err := json.NewDecoder(resp2.Body).Decode(&interfacesResult); err == nil {
			fmt.Printf("✅ Interfaces retrieved successfully!\n")

			// Count interfaces
			if interfaces, ok := interfacesResult["ietf-interfaces:interfaces"]; ok {
				if interfaceMap, ok := interfaces.(map[string]interface{}); ok {
					if interfaceList, ok := interfaceMap["interface"].([]interface{}); ok {
						fmt.Printf("📊 Found %d network interfaces\n", len(interfaceList))

						// Show first few interfaces
						for i, iface := range interfaceList {
							if i >= 3 { // Show only first 3
								break
							}
							if ifaceMap, ok := iface.(map[string]interface{}); ok {
								name := "Unknown"
								if n, ok := ifaceMap["name"].(string); ok {
									name = n
								}
								fmt.Printf("   📶 Interface %d: %s\n", i+1, name)
							}
						}
					}
				}
			}
		}
	} else {
		body, _ := io.ReadAll(resp2.Body)
		fmt.Printf("⚠️  Interface request status: %s, body: %s\n", resp2.Status, string(body))
	}

	fmt.Println("\n🎉 Cisco DevNet Sandbox Connectivity Test Complete!")
	fmt.Println("✨ Your credentials work and the device is reachable!")
	fmt.Println("🚀 Ready to integrate into NetOrchestrator platform!")
}
