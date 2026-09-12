package provider

import "testing"

func TestPurchaseOrderMetadata(t *testing.T) {
	tests := []struct {
		name string
		meta *ProvisionOrderMeta
		want string
	}{
		{"nil", nil, ""},
		{"empty", &ProvisionOrderMeta{}, ""},
		{"name only", &ProvisionOrderMeta{CustomerName: "Demo Subscriber"}, "Demo Subscriber"},
		{"email only", &ProvisionOrderMeta{CustomerEmail: "subscriber@example.invalid"}, "subscriber@example.invalid"},
		{"both", &ProvisionOrderMeta{CustomerName: "Demo Subscriber", CustomerEmail: "subscriber@example.invalid"}, "Demo Subscriber,subscriber@example.invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.meta.toPurchaseOrderNumber(); got != tt.want {
				t.Fatal("unexpected synthetic metadata result")
			}
		})
	}
}
