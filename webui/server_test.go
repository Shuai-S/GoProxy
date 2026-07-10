package webui

import (
	"testing"

	"goproxy/config"
)

func TestValidateFixedPortsForMultiplexedLayout(t *testing.T) {
	cfg := config.DefaultConfig()

	for _, port := range []int{7779, 7780, 7781} {
		err := validateFixedPorts([]config.FixedPortBinding{{
			Port:         port,
			ProxyAddress: "127.0.0.1:1080",
		}}, cfg)
		if err != nil {
			t.Fatalf("expected port %d to be available: %v", port, err)
		}
	}

	for _, port := range []int{7776, 7777, 7778} {
		err := validateFixedPorts([]config.FixedPortBinding{{
			Port:         port,
			ProxyAddress: "127.0.0.1:1080",
		}}, cfg)
		if err == nil {
			t.Fatalf("expected reserved port %d to be rejected", port)
		}
	}
}

func TestNormalizeFixedPortBindingsClearsLegacyProtocol(t *testing.T) {
	bindings := []config.FixedPortBinding{
		{Port: 7779, Protocol: "http", ProxyAddress: "127.0.0.1:8080"},
		{Port: 7780, Protocol: "socks5", ProxyAddress: "127.0.0.1:1080"},
	}

	if err := validateFixedPorts(bindings, config.DefaultConfig()); err != nil {
		t.Fatal(err)
	}
	normalized := normalizeFixedPortBindings(bindings)
	for _, binding := range normalized {
		if binding.Protocol != "" {
			t.Fatalf("expected legacy protocol to be cleared, got %q", binding.Protocol)
		}
	}
	if bindings[0].Protocol != "http" || bindings[1].Protocol != "socks5" {
		t.Fatal("normalization mutated input bindings")
	}
}
