package api

import (
	"encoding/json"
	"testing"
)

func TestRequestMarshal(t *testing.T) {
	request := Request{
		Command: CommandEnsureService,
		Service: &ServiceRequest{
			Name: "nginx",
			Type: "systemd",
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	expected := `{"command":"services.ensure","service":{"name":"nginx","type":"systemd"}}`

	if string(data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data))
	}
}

func TestRequestMarshalReload(t *testing.T) {
	request := Request{
		Command: CommandReload,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	expected := `{"command":"reload"}`

	if string(data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data))
	}
}

func TestResponseMarshalEnsure(t *testing.T) {
	added := true

	response := Response{
		Success: true,
		Added:   &added,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	expected := `{"success":true,"added":true}`

	if string(data) != expected {
		t.Fatalf("expected %s, got %s", expected, string(data))
	}
}
