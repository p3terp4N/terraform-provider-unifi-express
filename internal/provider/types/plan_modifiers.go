package types

import (
	"context"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// CIDRNormalization returns a plan modifier that normalizes CIDR notation
// so that "192.168.1.10/24" and "192.168.1.0/24" are treated as equivalent.
// This replaces the V1 SDK DiffSuppressFunc for CIDR fields.
func CIDRNormalization() planmodifier.String {
	return cidrNormalizationModifier{}
}

type cidrNormalizationModifier struct{}

func (m cidrNormalizationModifier) Description(_ context.Context) string {
	return "Normalizes CIDR notation to its canonical network address form."
}

func (m cidrNormalizationModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m cidrNormalizationModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	planVal := req.PlanValue.ValueString()
	stateVal := req.StateValue.ValueString()

	_, planNet, planErr := net.ParseCIDR(planVal)
	_, stateNet, stateErr := net.ParseCIDR(stateVal)

	if planErr != nil || stateErr != nil {
		return
	}

	if planNet.String() == stateNet.String() {
		resp.PlanValue = req.StateValue
	}
}

// MACNormalization returns a plan modifier that normalizes MAC addresses
// so that case and separator differences are ignored.
// "00-11-22-33-44-55" and "00:11:22:33:44:55" are treated as equivalent.
func MACNormalization() planmodifier.String {
	return macNormalizationModifier{}
}

type macNormalizationModifier struct{}

func (m macNormalizationModifier) Description(_ context.Context) string {
	return "Normalizes MAC addresses to lowercase colon-separated form."
}

func (m macNormalizationModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m macNormalizationModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	planClean := cleanMAC(req.PlanValue.ValueString())
	stateClean := cleanMAC(req.StateValue.ValueString())

	if planClean == stateClean {
		resp.PlanValue = req.StateValue
	}
}

func cleanMAC(mac string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ToLower(mac), "-", ":"))
}
