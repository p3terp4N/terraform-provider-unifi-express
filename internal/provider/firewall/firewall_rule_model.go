package firewall

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &firewallRuleModel{}

type firewallRuleModel struct {
	base.Model
	Name             types.String `tfsdk:"name"`
	Action           types.String `tfsdk:"action"`
	Ruleset          types.String `tfsdk:"ruleset"`
	RuleIndex        types.Int64  `tfsdk:"rule_index"`
	Protocol         types.String `tfsdk:"protocol"`
	ProtocolV6       types.String `tfsdk:"protocol_v6"`
	ICMPTypename     types.String `tfsdk:"icmp_typename"`
	ICMPv6Typename   types.String `tfsdk:"icmp_v6_typename"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Logging          types.Bool   `tfsdk:"logging"`
	IPSec            types.String `tfsdk:"ip_sec"`
	StateEstablished types.Bool   `tfsdk:"state_established"`
	StateInvalid     types.Bool   `tfsdk:"state_invalid"`
	StateNew         types.Bool   `tfsdk:"state_new"`
	StateRelated     types.Bool   `tfsdk:"state_related"`

	// Sources
	SrcNetworkID        types.String `tfsdk:"src_network_id"`
	SrcNetworkType      types.String `tfsdk:"src_network_type"`
	SrcFirewallGroupIDs types.Set    `tfsdk:"src_firewall_group_ids"`
	SrcAddress          types.String `tfsdk:"src_address"`
	SrcAddressIPv6      types.String `tfsdk:"src_address_ipv6"`
	SrcPort             types.String `tfsdk:"src_port"`
	SrcMAC              types.String `tfsdk:"src_mac"`

	// Destinations
	DstNetworkID        types.String `tfsdk:"dst_network_id"`
	DstNetworkType      types.String `tfsdk:"dst_network_type"`
	DstFirewallGroupIDs types.Set    `tfsdk:"dst_firewall_group_ids"`
	DstAddress          types.String `tfsdk:"dst_address"`
	DstAddressIPv6      types.String `tfsdk:"dst_address_ipv6"`
	DstPort             types.String `tfsdk:"dst_port"`
}

func (m *firewallRuleModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var srcGroupIDs, dstGroupIDs []string
	diags.Append(m.SrcFirewallGroupIDs.ElementsAs(ctx, &srcGroupIDs, false)...)
	diags.Append(m.DstFirewallGroupIDs.ElementsAs(ctx, &dstGroupIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	return &unifi.FirewallRule{
		ID:               m.ID.ValueString(),
		Enabled:          m.Enabled.ValueBool(),
		Name:             m.Name.ValueString(),
		Action:           m.Action.ValueString(),
		Ruleset:          m.Ruleset.ValueString(),
		RuleIndex:        int(m.RuleIndex.ValueInt64()),
		Protocol:         m.Protocol.ValueString(),
		ProtocolV6:       m.ProtocolV6.ValueString(),
		ICMPTypename:     m.ICMPTypename.ValueString(),
		ICMPv6Typename:   m.ICMPv6Typename.ValueString(),
		Logging:          m.Logging.ValueBool(),
		IPSec:            m.IPSec.ValueString(),
		StateEstablished: m.StateEstablished.ValueBool(),
		StateInvalid:     m.StateInvalid.ValueBool(),
		StateNew:         m.StateNew.ValueBool(),
		StateRelated:     m.StateRelated.ValueBool(),

		SrcNetworkType:      m.SrcNetworkType.ValueString(),
		SrcMACAddress:       m.SrcMAC.ValueString(),
		SrcAddress:          m.SrcAddress.ValueString(),
		SrcAddressIPV6:      m.SrcAddressIPv6.ValueString(),
		SrcPort:             m.SrcPort.ValueString(),
		SrcNetworkID:        m.SrcNetworkID.ValueString(),
		SrcFirewallGroupIDs: srcGroupIDs,

		DstNetworkType:      m.DstNetworkType.ValueString(),
		DstAddress:          m.DstAddress.ValueString(),
		DstAddressIPV6:      m.DstAddressIPv6.ValueString(),
		DstPort:             m.DstPort.ValueString(),
		DstNetworkID:        m.DstNetworkID.ValueString(),
		DstFirewallGroupIDs: dstGroupIDs,
	}, diags
}

func (m *firewallRuleModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.FirewallRule)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.FirewallRule")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.Enabled = types.BoolValue(other.Enabled)
	m.Action = types.StringValue(other.Action)
	m.Ruleset = types.StringValue(other.Ruleset)
	m.RuleIndex = types.Int64Value(int64(other.RuleIndex))
	m.Protocol = ut.StringOrNull(other.Protocol)
	m.ProtocolV6 = ut.StringOrNull(other.ProtocolV6)
	m.ICMPTypename = ut.StringOrNull(other.ICMPTypename)
	m.ICMPv6Typename = ut.StringOrNull(other.ICMPv6Typename)
	m.Logging = types.BoolValue(other.Logging)
	m.IPSec = ut.StringOrNull(other.IPSec)
	m.StateEstablished = types.BoolValue(other.StateEstablished)
	m.StateInvalid = types.BoolValue(other.StateInvalid)
	m.StateNew = types.BoolValue(other.StateNew)
	m.StateRelated = types.BoolValue(other.StateRelated)

	m.SrcNetworkType = types.StringValue(other.SrcNetworkType)
	m.SrcMAC = ut.StringOrNull(other.SrcMACAddress)
	m.SrcAddress = ut.StringOrNull(other.SrcAddress)
	m.SrcAddressIPv6 = ut.StringOrNull(other.SrcAddressIPV6)
	m.SrcPort = ut.StringOrNull(other.SrcPort)
	m.SrcNetworkID = ut.StringOrNull(other.SrcNetworkID)

	srcSet, d := types.SetValueFrom(ctx, types.StringType, other.SrcFirewallGroupIDs)
	diags.Append(d...)
	if !diags.HasError() {
		m.SrcFirewallGroupIDs = srcSet
	}

	m.DstNetworkType = types.StringValue(other.DstNetworkType)
	m.DstAddress = ut.StringOrNull(other.DstAddress)
	m.DstAddressIPv6 = ut.StringOrNull(other.DstAddressIPV6)
	m.DstPort = ut.StringOrNull(other.DstPort)
	m.DstNetworkID = ut.StringOrNull(other.DstNetworkID)

	dstSet, d := types.SetValueFrom(ctx, types.StringType, other.DstFirewallGroupIDs)
	diags.Append(d...)
	if !diags.HasError() {
		m.DstFirewallGroupIDs = dstSet
	}

	return diags
}
