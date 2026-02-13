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
