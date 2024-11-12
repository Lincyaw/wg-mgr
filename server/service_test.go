package main

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"testing"
)

func TestWgClient(t *testing.T) {
	cli, err := wgctrl.New()
	if err != nil {
		t.Fatal(err)
	}
	devs, err := cli.Devices()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(devs)

	cli.ConfigureDevice("test", wgtypes.Config{
		PrivateKey:   nil,
		ListenPort:   nil,
		FirewallMark: nil,
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   wgtypes.Key{},
				Remove:                      false,
				UpdateOnly:                  false,
				PresharedKey:                nil,
				Endpoint:                    nil,
				PersistentKeepaliveInterval: nil,
				ReplaceAllowedIPs:           false,
				AllowedIPs:                  nil,
			},
		},
	})

}
