package device

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
)

var _ base.ResourceModel = &portProfileModel{}

type portProfileModel struct {
	base.Model

	Autoneg                    types.Bool   `tfsdk:"autoneg"`
	Dot1XCtrl                  types.String `tfsdk:"dot1x_ctrl"`
	Dot1XIdleTimeout           types.Int64  `tfsdk:"dot1x_idle_timeout"`
	EgressRateLimitKbps        types.Int64  `tfsdk:"egress_rate_limit_kbps"`
	EgressRateLimitKbpsEnabled types.Bool   `tfsdk:"egress_rate_limit_kbps_enabled"`
	ExcludedNetworkIDs         types.Set    `tfsdk:"excluded_network_ids"`
	Forward                    types.String `tfsdk:"forward"`
	FullDuplex                 types.Bool   `tfsdk:"full_duplex"`
	Isolation                  types.Bool   `tfsdk:"isolation"`
	LldpmedEnabled             types.Bool   `tfsdk:"lldpmed_enabled"`
	LldpmedNotifyEnabled       types.Bool   `tfsdk:"lldpmed_notify_enabled"`
	NativeNetworkconfID        types.String `tfsdk:"native_networkconf_id"`
	Name                       types.String `tfsdk:"name"`
	OpMode                     types.String `tfsdk:"op_mode"`
	PoeMode                    types.String `tfsdk:"poe_mode"`
	PortSecurityEnabled        types.Bool   `tfsdk:"port_security_enabled"`
	PortSecurityMACAddress     types.Set    `tfsdk:"port_security_mac_address"`
	PriorityQueue1Level        types.Int64  `tfsdk:"priority_queue1_level"`
	PriorityQueue2Level        types.Int64  `tfsdk:"priority_queue2_level"`
	PriorityQueue3Level        types.Int64  `tfsdk:"priority_queue3_level"`
	PriorityQueue4Level        types.Int64  `tfsdk:"priority_queue4_level"`
	Speed                      types.Int64  `tfsdk:"speed"`
	StormctrlBcastEnabled      types.Bool   `tfsdk:"stormctrl_bcast_enabled"`
	StormctrlBcastLevel        types.Int64  `tfsdk:"stormctrl_bcast_level"`
	StormctrlBcastRate         types.Int64  `tfsdk:"stormctrl_bcast_rate"`
	StormctrlMcastEnabled      types.Bool   `tfsdk:"stormctrl_mcast_enabled"`
	StormctrlMcastLevel        types.Int64  `tfsdk:"stormctrl_mcast_level"`
	StormctrlMcastRate         types.Int64  `tfsdk:"stormctrl_mcast_rate"`
	StormctrlType              types.String `tfsdk:"stormctrl_type"`
	StormctrlUcastEnabled      types.Bool   `tfsdk:"stormctrl_ucast_enabled"`
	StormctrlUcastLevel        types.Int64  `tfsdk:"stormctrl_ucast_level"`
	StormctrlUcastRate         types.Int64  `tfsdk:"stormctrl_ucast_rate"`
	StpPortMode                types.Bool   `tfsdk:"stp_port_mode"`
	TaggedVlanMgmt             types.String `tfsdk:"tagged_vlan_mgmt"`
	VoiceNetworkconfID         types.String `tfsdk:"voice_networkconf_id"`
}

func (m *portProfileModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var excludedNetworkIDs []string
	diags.Append(m.ExcludedNetworkIDs.ElementsAs(ctx, &excludedNetworkIDs, false)...)

	var portSecurityMACAddress []string
	diags.Append(m.PortSecurityMACAddress.ElementsAs(ctx, &portSecurityMACAddress, false)...)

	if diags.HasError() {
		return nil, diags
	}

	return &unifi.PortProfile{
		ID:                           m.ID.ValueString(),
		Autoneg:                      m.Autoneg.ValueBool(),
		Dot1XCtrl:                    m.Dot1XCtrl.ValueString(),
		Dot1XIDleTimeout:             int(m.Dot1XIdleTimeout.ValueInt64()),
		EgressRateLimitKbps:          int(m.EgressRateLimitKbps.ValueInt64()),
		EgressRateLimitKbpsEnabled:   m.EgressRateLimitKbpsEnabled.ValueBool(),
		ExcludedNetworkIDs:           excludedNetworkIDs,
		Forward:                      m.Forward.ValueString(),
		FullDuplex:                   m.FullDuplex.ValueBool(),
		Isolation:                    m.Isolation.ValueBool(),
		LldpmedEnabled:               m.LldpmedEnabled.ValueBool(),
		LldpmedNotifyEnabled:         m.LldpmedNotifyEnabled.ValueBool(),
		NATiveNetworkID:              m.NativeNetworkconfID.ValueString(),
		Name:                         m.Name.ValueString(),
		OpMode:                       m.OpMode.ValueString(),
		PoeMode:                      m.PoeMode.ValueString(),
		PortSecurityEnabled:          m.PortSecurityEnabled.ValueBool(),
		PortSecurityMACAddress:       portSecurityMACAddress,
		PriorityQueue1Level:          int(m.PriorityQueue1Level.ValueInt64()),
		PriorityQueue2Level:          int(m.PriorityQueue2Level.ValueInt64()),
		PriorityQueue3Level:          int(m.PriorityQueue3Level.ValueInt64()),
		PriorityQueue4Level:          int(m.PriorityQueue4Level.ValueInt64()),
		Speed:                        int(m.Speed.ValueInt64()),
		StormctrlBroadcastastEnabled: m.StormctrlBcastEnabled.ValueBool(),
		StormctrlBroadcastastLevel:   int(m.StormctrlBcastLevel.ValueInt64()),
		StormctrlBroadcastastRate:    int(m.StormctrlBcastRate.ValueInt64()),
		StormctrlMcastEnabled:        m.StormctrlMcastEnabled.ValueBool(),
		StormctrlMcastLevel:          int(m.StormctrlMcastLevel.ValueInt64()),
		StormctrlMcastRate:           int(m.StormctrlMcastRate.ValueInt64()),
		StormctrlType:                m.StormctrlType.ValueString(),
		StormctrlUcastEnabled:        m.StormctrlUcastEnabled.ValueBool(),
		StormctrlUcastLevel:          int(m.StormctrlUcastLevel.ValueInt64()),
		StormctrlUcastRate:           int(m.StormctrlUcastRate.ValueInt64()),
		StpPortMode:                  m.StpPortMode.ValueBool(),
		TaggedVLANMgmt:               m.TaggedVlanMgmt.ValueString(),
		VoiceNetworkID:               m.VoiceNetworkconfID.ValueString(),
	}, diags
}

func (m *portProfileModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.PortProfile)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.PortProfile")
		return diags
	}

	m.ID = types.StringValue(other.ID)
	m.Autoneg = types.BoolValue(other.Autoneg)
	m.Dot1XCtrl = ut.StringOrNull(other.Dot1XCtrl)
	m.Dot1XIdleTimeout = types.Int64Value(int64(other.Dot1XIDleTimeout))
	m.EgressRateLimitKbps = ut.Int64OrNull(other.EgressRateLimitKbps)
	m.EgressRateLimitKbpsEnabled = types.BoolValue(other.EgressRateLimitKbpsEnabled)
	m.Forward = ut.StringOrNull(other.Forward)
	m.FullDuplex = types.BoolValue(other.FullDuplex)
	m.Isolation = types.BoolValue(other.Isolation)
	m.LldpmedEnabled = types.BoolValue(other.LldpmedEnabled)
	m.LldpmedNotifyEnabled = types.BoolValue(other.LldpmedNotifyEnabled)
	m.NativeNetworkconfID = ut.StringOrNull(other.NATiveNetworkID)
	m.Name = ut.StringOrNull(other.Name)
	m.OpMode = ut.StringOrNull(other.OpMode)
	m.PoeMode = ut.StringOrNull(other.PoeMode)
	m.PortSecurityEnabled = types.BoolValue(other.PortSecurityEnabled)
	m.PriorityQueue1Level = ut.Int64OrNull(other.PriorityQueue1Level)
	m.PriorityQueue2Level = ut.Int64OrNull(other.PriorityQueue2Level)
	m.PriorityQueue3Level = ut.Int64OrNull(other.PriorityQueue3Level)
	m.PriorityQueue4Level = ut.Int64OrNull(other.PriorityQueue4Level)
	m.Speed = ut.Int64OrNull(other.Speed)
	m.StormctrlBcastEnabled = types.BoolValue(other.StormctrlBroadcastastEnabled)
	m.StormctrlBcastLevel = ut.Int64OrNull(other.StormctrlBroadcastastLevel)
	m.StormctrlBcastRate = ut.Int64OrNull(other.StormctrlBroadcastastRate)
	m.StormctrlMcastEnabled = types.BoolValue(other.StormctrlMcastEnabled)
	m.StormctrlMcastLevel = ut.Int64OrNull(other.StormctrlMcastLevel)
	m.StormctrlMcastRate = ut.Int64OrNull(other.StormctrlMcastRate)
	m.StormctrlType = ut.StringOrNull(other.StormctrlType)
	m.StormctrlUcastEnabled = types.BoolValue(other.StormctrlUcastEnabled)
	m.StormctrlUcastLevel = ut.Int64OrNull(other.StormctrlUcastLevel)
	m.StormctrlUcastRate = ut.Int64OrNull(other.StormctrlUcastRate)
	m.StpPortMode = types.BoolValue(other.StpPortMode)
	m.TaggedVlanMgmt = ut.StringOrNull(other.TaggedVLANMgmt)
	m.VoiceNetworkconfID = ut.StringOrNull(other.VoiceNetworkID)

	excludedSet, d := types.SetValueFrom(ctx, types.StringType, other.ExcludedNetworkIDs)
	diags.Append(d...)
	if !diags.HasError() {
		m.ExcludedNetworkIDs = excludedSet
	}

	macSet, d := types.SetValueFrom(ctx, types.StringType, other.PortSecurityMACAddress)
	diags.Append(d...)
	if !diags.HasError() {
		m.PortSecurityMACAddress = macSet
	}

	return diags
}
