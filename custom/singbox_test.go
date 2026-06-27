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
