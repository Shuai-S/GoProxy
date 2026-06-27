package custom

import "testing"

func TestBuildOutboundHysteria2Obfs(t *testing.T) {
	node := ParsedNode{
		Name:   "hy2",
		Type:   "hysteria2",
		Server: "example.com",
		Port:   443,
		Raw: map[string]interface{}{
			"type":          "hysteria2",
			"server":        "example.com",
			"port":          443,
			"password":      "secret",
			"obfs":          "salamander",
			"obfs-password": "obfs-secret",
			"tls":           true,
		},
	}

	outbound := buildOutbound(node, "node-0")
	obfs, ok := outbound["obfs"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected hysteria2 obfs config, got %#v", outbound["obfs"])
	}
	if obfs["type"] != "salamander" {
		t.Fatalf("expected salamander obfs type, got %#v", obfs["type"])
	}
	if obfs["password"] != "obfs-secret" {
		t.Fatalf("expected obfs password, got %#v", obfs["password"])
	}
}

func TestParseAnyTLSLink(t *testing.T) {
	node, err := parseProxyLink("anytls://secret@example.com:443?sni=example.com#any")
	if err != nil {
		t.Fatalf("parse anytls link: %v", err)
	}
	if node.Type != "anytls" {
		t.Fatalf("expected anytls type, got %s", node.Type)
	}
	if node.Raw["password"] != "secret" {
		t.Fatalf("expected anytls password, got %#v", node.Raw["password"])
	}
}

func TestBuildOutboundTUICForcesTLS(t *testing.T) {
	node := ParsedNode{
		Name:   "tuic",
		Type:   "tuic",
		Server: "example.com",
		Port:   443,
		Raw: map[string]interface{}{
			"type":     "tuic",
			"server":   "example.com",
			"port":     443,
			"uuid":     "00000000-0000-0000-0000-000000000000",
			"password": "secret",
		},
	}

	outbound := buildOutbound(node, "node-0")
	tlsConfig, ok := outbound["tls"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tuic tls config, got %#v", outbound["tls"])
	}
	if tlsConfig["enabled"] != true {
		t.Fatalf("expected tuic tls enabled, got %#v", tlsConfig["enabled"])
	}
}
