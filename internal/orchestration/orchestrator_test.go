package orchestration

import "testing"

func TestDueForExecution(t *testing.T) {
	now := int64(100)

	if !dueForExecution(WorkflowItem{ExecuteAfter: 0}, now) {
		t.Fatal("items without executeAfter should be considered due")
	}
	if !dueForExecution(WorkflowItem{ExecuteAfter: 100}, now) {
		t.Fatal("items at executeAfter boundary should be considered due")
	}
	if dueForExecution(WorkflowItem{ExecuteAfter: 101}, now) {
		t.Fatal("items in the future should not be considered due")
	}
}

func TestProvisioningParamsValidatesBoltOn(t *testing.T) {
	tests := []struct {
		name       string
		value      interface{}
		omit       bool
		wantError  bool
		wantBoltOn int
	}{
		{name: "missing", omit: true, wantError: true},
		{name: "nil", value: nil, wantError: true},
		{name: "string", value: "synthetic", wantError: true},
		{name: "integer", value: 1, wantError: true},
		{name: "boolean", value: true, wantError: true},
		{name: "zero", value: float64(0)},
		{name: "numeric", value: float64(1), wantBoltOn: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := map[string]interface{}{
				"msisdn":   "synthetic-line",
				"tariffId": float64(1),
			}
			if !test.omit {
				data["boltOn"] = test.value
			}
			item := WorkflowItem{Data: data}
			params, err := item.asProvisioningParams()
			if (err != nil) != test.wantError {
				t.Fatalf("error present = %t, want %t", err != nil, test.wantError)
			}
			if test.wantError {
				if params != nil {
					t.Fatal("invalid input must not return provisioning parameters")
				}
				return
			}
			if params == nil || params.boltOn != test.wantBoltOn {
				t.Fatal("valid numeric input produced incorrect parameters")
			}
		})
	}
}
