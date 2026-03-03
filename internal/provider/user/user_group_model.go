package user

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &userGroupModel{}

type userGroupModel struct {
	base.Model
	Name           types.String `tfsdk:"name"`
	QOSRateMaxDown types.Int64  `tfsdk:"qos_rate_max_down"`
	QOSRateMaxUp   types.Int64  `tfsdk:"qos_rate_max_up"`
}

func (m *userGroupModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.UserGroup{
		ID:             m.ID.ValueString(),
		Name:           m.Name.ValueString(),
		QOSRateMaxDown: int(m.QOSRateMaxDown.ValueInt64()),
		QOSRateMaxUp:   int(m.QOSRateMaxUp.ValueInt64()),
	}, diag.Diagnostics{}
}

func (m *userGroupModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.UserGroup)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.UserGroup")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.QOSRateMaxDown = types.Int64Value(int64(other.QOSRateMaxDown))
	m.QOSRateMaxUp = types.Int64Value(int64(other.QOSRateMaxUp))
	return diags
}
