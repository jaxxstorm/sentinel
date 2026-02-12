package source

import "testing"

func TestDecodePeersFromStatusJSON(t *testing.T) {
	input := []byte(`{
		"Peer": {
			"nodekey:abc": {
				"StableID": "n1",
				"HostName": "ai",
				"DNSName": "ai.tail.test.",
				"Online": true,
				"OS": "linux",
				"UserID": 123
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
				"Hostinfo": {"OS":"macOS","Hostname":"sentinel"}
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
}
