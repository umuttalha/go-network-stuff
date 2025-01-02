package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// Structs to parse AWS and Google IP ranges
type AWSIPRanges struct {
	SyncToken  string `json:"syncToken"`
	CreateDate string `json:"createDate"`
	Prefixes   []struct {
		IPPrefix string `json:"ip_prefix"`
	} `json:"prefixes"`
}

type GoogleIPRanges struct {
	SyncToken    string `json:"syncToken"`
	CreationTime string `json:"creationTime"`
	Prefixes     []struct {
		IPv4Prefix string `json:"ipv4Prefix"`
		IPv6Prefix string `json:"ipv6Prefix"`
	} `json:"prefixes"`
}

type AzureIPRanges struct {
	Name            string   `json:"name"`
	ID              string   `json:"id"`
	AddressPrefixes []string `json:"addressPrefixes"`
}

type OracleIPRanges struct {
	LastUpdatedTimestamp string `json:"last_updated_timestamp"`
	Regions              []struct {
		Region string `json:"region"`
		CIDRs  []struct {
			CIDR string   `json:"cidr"`
			Tags []string `json:"tags"`
		} `json:"cidrs"`
	} `json:"regions"`
}

// Fetch IP ranges from a URL
func fetchIPRanges(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func readIPRangesFromFile(filename string, target interface{}) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, target)
}

// Check if an IP is within a list of CIDR ranges
func isIPInRanges(ip net.IP, cidrs []string) bool {
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			fmt.Printf("Error parsing CIDR: %v\n", err)
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

func main() {
	// Check if a domain name is provided as an argument
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <domain>")
		return
	}

	// Get the domain name from the command line argument
	domain := os.Args[1]

	// Perform DNS lookup
	ips, err := net.LookupIP(domain)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Fetch AWS IP ranges
	var awsRanges AWSIPRanges
	err = fetchIPRanges("https://ip-ranges.amazonaws.com/ip-ranges.json", &awsRanges)
	if err != nil {
		fmt.Printf("Error fetching AWS IP ranges: %v\n", err)
	} else {
		// Extract CIDR ranges
		var awsCIDRs []string
		for _, prefix := range awsRanges.Prefixes {
			awsCIDRs = append(awsCIDRs, prefix.IPPrefix)
		}

		// Check if any IP belongs to AWS
		for _, ip := range ips {
			if isIPInRanges(ip, awsCIDRs) {
				fmt.Printf("The domain '%s' is using AWS.\n", domain)
				return
			}
		}
	}

	// Fetch Google IP ranges
	var googleRanges GoogleIPRanges
	err = fetchIPRanges("https://www.gstatic.com/ipranges/goog.json", &googleRanges)
	if err != nil {
		fmt.Printf("Error fetching Google IP ranges: %v\n", err)
	} else {
		// Extract CIDR ranges
		var googleCIDRs []string
		for _, prefix := range googleRanges.Prefixes {
			if prefix.IPv4Prefix != "" {
				googleCIDRs = append(googleCIDRs, prefix.IPv4Prefix)
			}
			if prefix.IPv6Prefix != "" {
				googleCIDRs = append(googleCIDRs, prefix.IPv6Prefix)
			}
		}

		// Check if any IP belongs to Google
		for _, ip := range ips {
			if isIPInRanges(ip, googleCIDRs) {
				fmt.Printf("The domain '%s' is using Google.\n", domain)
				return
			}
		}
	}

	// Manually added Cloudflare IP ranges
	cloudflareCIDRs := []string{
		"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22",
		"141.101.64.0/18", "108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20",
		"197.234.240.0/22", "198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
		"104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
		"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32", "2405:b500::/32",
		"2405:8100::/32", "2a06:98c0::/29", "2c0f:f248::/32",
	}

	// Check if any IP belongs to Cloudflare
	for _, ip := range ips {
		if isIPInRanges(ip, cloudflareCIDRs) {
			fmt.Printf("The domain '%s' is using Cloudflare.\n", domain)
			return
		}
	}

	// Manually added Fastly IP ranges
	fastlyCIDRs := []string{
		"23.235.32.0/20", "43.249.72.0/22", "103.244.50.0/24", "103.245.222.0/23", "103.245.224.0/24", "104.156.80.0/20", "140.248.64.0/18", "140.248.128.0/17", "146.75.0.0/17", "151.101.0.0/16", "157.52.64.0/18", "167.82.0.0/17", "167.82.128.0/20", "167.82.160.0/20", "167.82.224.0/20", "172.111.64.0/18", "185.31.16.0/22", "199.27.72.0/21", "199.232.0.0/16", "2a04:4e40::/32", "2a04:4e42::/32",
	}

	// Check if any IP belongs to Fastly
	for _, ip := range ips {
		if isIPInRanges(ip, fastlyCIDRs) {
			fmt.Printf("The domain '%s' is using Fastly.\n", domain)
			return
		}
	}

	var azureRanges AzureIPRanges
	err = readIPRangesFromFile("azure.json", &azureRanges)
	if err != nil {
		fmt.Printf("Error reading Azure IP ranges: %v\n", err)
	} else {
		// Extract CIDR ranges
		azureCIDRs := azureRanges.AddressPrefixes

		// Check if any IP belongs to Azure
		for _, ip := range ips {
			if isIPInRanges(ip, azureCIDRs) {
				fmt.Printf("The domain '%s' is using Azure.\n", domain)
				return
			}
		}
	}

	var oracleRanges OracleIPRanges
	err = fetchIPRanges("https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json", &oracleRanges)
	if err != nil {
		fmt.Printf("Error fetching Oracle IP ranges: %v\n", err)
	} else {
		// Extract CIDR ranges
		var oracleCIDRs []string
		for _, region := range oracleRanges.Regions {
			for _, cidr := range region.CIDRs {
				oracleCIDRs = append(oracleCIDRs, cidr.CIDR)
			}
		}

		// Check if any IP belongs to Oracle
		for _, ip := range ips {
			if isIPInRanges(ip, oracleCIDRs) {
				fmt.Printf("The domain '%s' is using Oracle.\n", domain)
				return
			}
		}
	}

	fmt.Printf("The domain '%s' is not using AWS, Google, Cloudflare, Fastly, Azure, or Oracle.\n", domain)
}
