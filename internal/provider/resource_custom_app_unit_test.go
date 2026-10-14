package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/ja-guerrero/terraform-provider-iru/internal/client"
)

// TestCustomAppMapToClientSelfServiceRecommended is a standard Go unit test.
// It does NOT require TF_ACC=1 or an API Token.
func TestCustomAppMapToClientSelfServiceRecommended(t *testing.T) {
	cases := []struct {
		name        string
		showInSS    types.Bool
		recommended types.Bool
		wantKey     bool
		wantValue   bool
	}{
		{"self service off, recommended unknown", types.BoolValue(false), types.BoolUnknown(), false, false},
		{"self service off, recommended null", types.BoolValue(false), types.BoolNull(), false, false},
		{"self service off, recommended true", types.BoolValue(false), types.BoolValue(true), false, false},
		{"self service unknown, recommended true", types.BoolUnknown(), types.BoolValue(true), false, false},
		{"self service on, recommended unknown", types.BoolValue(true), types.BoolUnknown(), false, false},
		{"self service on, recommended false", types.BoolValue(true), types.BoolValue(false), true, false},
		{"self service on, recommended true", types.BoolValue(true), types.BoolValue(true), true, true},
	}

	r := &customAppResource{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := &customAppResourceModel{
				Name:                   types.StringValue("app"),
				ShowInSelfService:      tc.showInSS,
				SelfServiceRecommended: tc.recommended,
			}

			body, err := json.Marshal(r.mapToClient(data))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			v, ok := got["self_service_recommended"]
			if ok != tc.wantKey {
				t.Fatalf("self_service_recommended present = %v, want %v (body: %s)", ok, tc.wantKey, body)
			}
			if ok && v != tc.wantValue {
				t.Errorf("self_service_recommended = %v, want %v", v, tc.wantValue)
			}
		})
	}
}

func TestCustomAppMapFromClientMissingRecommended(t *testing.T) {
	var app client.CustomApp
	if err := json.Unmarshal([]byte(`{"id":"1","name":"app","show_in_self_service":false}`), &app); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	var data customAppResourceModel
	(&customAppResource{}).mapFromClient(&data, &app)

	if data.SelfServiceRecommended.IsNull() || data.SelfServiceRecommended.IsUnknown() || data.SelfServiceRecommended.ValueBool() {
		t.Errorf("SelfServiceRecommended = %s, want false", data.SelfServiceRecommended)
	}
}
