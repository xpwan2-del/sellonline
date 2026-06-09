package models

import "testing"

func TestJSONScanAcceptsStringValue(t *testing.T) {
	var value JSON
	if err := value.Scan(`{"enabled":true,"host":"smtp.gmail.com","port":465}`); err != nil {
		t.Fatalf("scan string json failed: %v", err)
	}

	if value["enabled"] != true {
		t.Fatalf("enabled should be true, got %v", value["enabled"])
	}
	if value["host"] != "smtp.gmail.com" {
		t.Fatalf("host mismatch: %v", value["host"])
	}
	if value["port"] != float64(465) {
		t.Fatalf("port mismatch: %v", value["port"])
	}
}

func TestJSONScanAcceptsByteValue(t *testing.T) {
	var value JSON
	if err := value.Scan([]byte(`{"enabled":true,"host":"smtp.gmail.com"}`)); err != nil {
		t.Fatalf("scan byte json failed: %v", err)
	}

	if value["enabled"] != true {
		t.Fatalf("enabled should be true, got %v", value["enabled"])
	}
	if value["host"] != "smtp.gmail.com" {
		t.Fatalf("host mismatch: %v", value["host"])
	}
}
