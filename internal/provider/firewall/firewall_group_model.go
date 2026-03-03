package firewall

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &firewallGroupModel{}

type firewallGroupModel struct {
	base.Model
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	Members types.Set    `tfsdk:"members"`
}

func (m *firewallGroupModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var members []string
	diags.Append(m.Members.ElementsAs(ctx, &members, false)...)
	if diags.HasError() {
		return nil, diags
	}
	return &unifi.FirewallGroup{
		ID:           m.ID.ValueString(),
		Name:         m.Name.ValueString(),
		GroupType:    m.Type.ValueString(),
		GroupMembers: members,
	}, diags
}

func (m *firewallGroupModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.FirewallGroup)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.FirewallGroup")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.Type = types.StringValue(other.GroupType)

	memberSet, d := types.SetValueFrom(ctx, types.StringType, other.GroupMembers)
	diags.Append(d...)
	if !diags.HasError() {
		m.Members = memberSet
	}
	return diags
}
