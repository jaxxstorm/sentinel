package source

import "testing"

func TestDecodePeersFromStatusJSON(t *testing.T) {
	input := []byte(`{
		"BackendState":"Running",
		"Peer": {
			"nodekey:abc": {
				"StableID": "n1",
				"HostName": "ai",
				"DNSName": "ai.tail.test.",
				"Online": true,
				"OS": "linux",
				"UserID": 123,
				"TailscaleIPs": ["100.64.0.10", "fd7a:115c:a1e0::10"],
				"PrimaryRoutes": ["10.0.0.0/24"],
				"MachineAuthorized": true,
				"Expired": false,
				"KeyExpiry": "2026-03-01T00:00:00Z",
				"Hostinfo": {"OS":"linux","Hostname":"ai"}
			}
		}
	}`)
	peers, err := decodePeersFromStatusJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].ID != "n1" {
		t.Fatalf("expected stable id n1, got %q", peers[0].ID)
	}
	if peers[0].Name != "ai" {
		t.Fatalf("expected name ai, got %q", peers[0].Name)
	}
	if !peers[0].Online {
		t.Fatal("expected peer to be online")
	}
	if peers[0].Metadata["os"] != "linux" {
		t.Fatalf("expected os metadata linux, got %q", peers[0].Metadata["os"])
	}
	if got := len(peers[0].Owners); got != 1 || peers[0].Owners[0] != "123" {
		t.Fatalf("expected owners [123], got %#v", peers[0].Owners)
	}
	if got := len(peers[0].IPs); got != 2 || peers[0].IPs[0] != "100.64.0.10" {
		t.Fatalf("expected canonical identity IPs, got %#v", peers[0].IPs)
	}
	if got := len(peers[0].Routes); got != 1 || peers[0].Routes[0] != "10.0.0.0/24" {
		t.Fatalf("expected primary route in decoded peer, got %#v", peers[0].Routes)
	}
	if !peers[0].MachineAuthorized {
		t.Fatal("expected machine_authorized=true")
	}
	if peers[0].HostinfoHash == "" {
		t.Fatal("expected hostinfo hash to be populated")
	}
}

func TestDecodePeersFromNetMapJSON(t *testing.T) {
	input := []byte(`{
		"Peers": [
			{
				"ID": 1694798792745323,
				"StableID": "ncwSkSVaEE11CNTRL",
				"Name": "sentinel.tail4cf751.ts.net.",
				"ComputedName": "sentinel",
				"Online": false,
				"Tags": ["tag:dev"],
				"Addresses": ["100.64.0.20/32", "fd7a:115c:a1e0::20/128"],
				"PrimaryRoutes": ["10.42.0.0/24"],
				"MachineAuthorized": true,
				"Expired": true,
				"KeyExpiry": "2026-08-11T19:20:24Z",
				"Hostinfo": {"OS":"macOS","Hostname":"sentinel","Services":[{"Proto":"peerapi4","Port":50626}]}
			}
		]
	}`)
	peers, err := decodePeersFromNetMapJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].ID != "ncwSkSVaEE11CNTRL" {
		t.Fatalf("expected stable id, got %q", peers[0].ID)
	}
	if peers[0].Name != "sentinel" {
		t.Fatalf("expected computed name sentinel, got %q", peers[0].Name)
	}
	if peers[0].Online {
		t.Fatal("expected peer to be offline")
	}
	if peers[0].Metadata["os"] != "macOS" {
		t.Fatalf("expected os metadata macOS, got %q", peers[0].Metadata["os"])
	}
	if got := len(peers[0].Owners); got != 0 {
		t.Fatalf("expected no owners in fixture, got %#v", peers[0].Owners)
	}
	if got := len(peers[0].IPs); got != 2 || peers[0].IPs[0] != "100.64.0.20" {
		t.Fatalf("expected canonical netmap addresses in identity IPs, got %#v", peers[0].IPs)
	}
	if got := len(peers[0].Routes); got != 1 || peers[0].Routes[0] != "10.42.0.0/24" {
		t.Fatalf("expected route to be decoded, got %#v", peers[0].Routes)
	}
	if !peers[0].MachineAuthorized {
		t.Fatal("expected machine_authorized=true")
	}
	if !peers[0].Expired {
		t.Fatal("expected expired=true")
	}
	if peers[0].KeyExpiry == "" {
		t.Fatal("expected key expiry to be populated")
	}
	if peers[0].HostinfoHash == "" {
		t.Fatal("expected hostinfo hash to be populated")
	}
}

func TestDecodeStatusAndNetmapIdentityParity(t *testing.T) {
	statusInput := []byte(`{
		"Peer": {
			"nodekey:abc": {
				"StableID": "peer-identity",
				"HostName": "peer-identity",
				"Online": true,
				"UserID": 456,
				"TailscaleIPs": ["100.64.0.30", "fd7a:115c:a1e0::30"]
			}
		}
	}`)
	netmapInput := []byte(`{
		"Peers": [
			{
				"StableID": "peer-identity",
				"ComputedName": "peer-identity",
				"Online": true,
				"User": 456,
				"Addresses": ["100.64.0.30/32", "fd7a:115c:a1e0::30/128"]
			}
		]
	}`)

	statusPeers, err := decodePeersFromStatusJSON(statusInput)
	if err != nil {
		t.Fatal(err)
	}
	netmapPeers, err := decodePeersFromNetMapJSON(netmapInput)
	if err != nil {
		t.Fatal(err)
	}
	if len(statusPeers) != 1 || len(netmapPeers) != 1 {
		t.Fatalf("expected one peer from each decode path, got status=%d netmap=%d", len(statusPeers), len(netmapPeers))
	}
	if len(statusPeers[0].Owners) == 0 || len(netmapPeers[0].Owners) == 0 {
		t.Fatalf("expected non-empty owner identities, status=%#v netmap=%#v", statusPeers[0].Owners, netmapPeers[0].Owners)
	}
	if len(statusPeers[0].IPs) != 2 || len(netmapPeers[0].IPs) != 2 {
		t.Fatalf("expected two identity IPs from each decode path, status=%#v netmap=%#v", statusPeers[0].IPs, netmapPeers[0].IPs)
	}
	if got, want := statusPeers[0].Owners, netmapPeers[0].Owners; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("expected owner identity parity, status=%#v netmap=%#v", got, want)
	}
	if got, want := statusPeers[0].IPs, netmapPeers[0].IPs; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("expected IP identity parity, status=%#v netmap=%#v", got, want)
	}
}
