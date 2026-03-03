package radius

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &accountModel{}

type accountModel struct {
	base.Model
	Name             types.String `tfsdk:"name"`
	Password         types.String `tfsdk:"password"`
	TunnelType       types.Int64  `tfsdk:"tunnel_type"`
	TunnelMediumType types.Int64  `tfsdk:"tunnel_medium_type"`
	NetworkID        types.String `tfsdk:"network_id"`
}

func (m *accountModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.Account{
		ID:               m.ID.ValueString(),
		Name:             m.Name.ValueString(),
		XPassword:        m.Password.ValueString(),
		TunnelType:       int(m.TunnelType.ValueInt64()),
		TunnelMediumType: int(m.TunnelMediumType.ValueInt64()),
		NetworkID:        m.NetworkID.ValueString(),
	}, diag.Diagnostics{}
}

func (m *accountModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.Account)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.Account")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.Password = types.StringValue(other.XPassword)
	m.TunnelType = types.Int64Value(int64(other.TunnelType))
	m.TunnelMediumType = types.Int64Value(int64(other.TunnelMediumType))
	m.NetworkID = ut.StringOrNull(other.NetworkID)
	return diags
}
