package network

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
)

var _ base.ResourceModel = &networkModel{}

type networkModel struct {
	base.Model

	// Core
	Name         types.String `tfsdk:"name"`
	Purpose      types.String `tfsdk:"purpose"`
	VLANID       types.Int64  `tfsdk:"vlan_id"`
	Subnet       types.String `tfsdk:"subnet"`
	NetworkGroup types.String `tfsdk:"network_group"`
	DomainName   types.String `tfsdk:"domain_name"`
	Enabled      types.Bool   `tfsdk:"enabled"`

	// DHCP v4
	DHCPStart        types.String `tfsdk:"dhcp_start"`
	DHCPStop         types.String `tfsdk:"dhcp_stop"`
	DHCPEnabled      types.Bool   `tfsdk:"dhcp_enabled"`
	DHCPLease        types.Int64  `tfsdk:"dhcp_lease"`
	DHCPDNS          types.List   `tfsdk:"dhcp_dns"`
	DHCPRelayEnabled types.Bool   `tfsdk:"dhcp_relay_enabled"`

	// DHCP boot
	DHCPDBootEnabled  types.Bool   `tfsdk:"dhcpd_boot_enabled"`
	DHCPDBootServer   types.String `tfsdk:"dhcpd_boot_server"`
	DHCPDBootFilename types.String `tfsdk:"dhcpd_boot_filename"`

	// DHCP v6
	DHCPV6DNS     types.List   `tfsdk:"dhcp_v6_dns"`
	DHCPV6DNSAuto types.Bool   `tfsdk:"dhcp_v6_dns_auto"`
	DHCPV6Enabled types.Bool   `tfsdk:"dhcp_v6_enabled"`
	DHCPV6Lease   types.Int64  `tfsdk:"dhcp_v6_lease"`
	DHCPV6Start   types.String `tfsdk:"dhcp_v6_start"`
	DHCPV6Stop    types.String `tfsdk:"dhcp_v6_stop"`

	// Network features
	IGMPSnooping            types.Bool `tfsdk:"igmp_snooping"`
	MulticastDNS            types.Bool `tfsdk:"multicast_dns"`
	InternetAccessEnabled   types.Bool `tfsdk:"internet_access_enabled"`
	NetworkIsolationEnabled types.Bool `tfsdk:"network_isolation_enabled"`

	// IPv6
	IPV6InterfaceType       types.String `tfsdk:"ipv6_interface_type"`
	IPV6StaticSubnet        types.String `tfsdk:"ipv6_static_subnet"`
	IPV6PDInterface         types.String `tfsdk:"ipv6_pd_interface"`
	IPV6PDPrefixID          types.String `tfsdk:"ipv6_pd_prefixid"`
	IPV6PDStart             types.String `tfsdk:"ipv6_pd_start"`
	IPV6PDStop              types.String `tfsdk:"ipv6_pd_stop"`
	IPV6RAEnable            types.Bool   `tfsdk:"ipv6_ra_enable"`
	IPV6RAPreferredLifetime types.Int64  `tfsdk:"ipv6_ra_preferred_lifetime"`
	IPV6RAPriority          types.String `tfsdk:"ipv6_ra_priority"`
	IPV6RAValidLifetime     types.Int64  `tfsdk:"ipv6_ra_valid_lifetime"`

	// WAN IPv4
	WANIP           types.String `tfsdk:"wan_ip"`
	WANNetmask      types.String `tfsdk:"wan_netmask"`
	WANGateway      types.String `tfsdk:"wan_gateway"`
	WANDNS          types.List   `tfsdk:"wan_dns"`
	WANType         types.String `tfsdk:"wan_type"`
	WANNetworkGroup types.String `tfsdk:"wan_networkgroup"`
	WANEgressQOS    types.Int64  `tfsdk:"wan_egress_qos"`
	WANUsername     types.String `tfsdk:"wan_username"`
	XWANPassword    types.String `tfsdk:"x_wan_password"`

	// WAN IPv6
	WANTypeV6       types.String `tfsdk:"wan_type_v6"`
	WANDHCPv6PDSize types.Int64  `tfsdk:"wan_dhcp_v6_pd_size"`
	WANIPV6         types.String `tfsdk:"wan_ipv6"`
	WANGatewayV6    types.String `tfsdk:"wan_gateway_v6"`
	WANPrefixlen    types.Int64  `tfsdk:"wan_prefixlen"`
}

func (m *networkModel) AsUnifiModel(ctx context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Flatten DHCP DNS list (up to 4 entries) into individual fields
	var dhcpDNS []string
	if !m.DHCPDNS.IsNull() && !m.DHCPDNS.IsUnknown() {
		diags.Append(m.DHCPDNS.ElementsAs(ctx, &dhcpDNS, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Flatten DHCPv6 DNS list
	var dhcpV6DNS []string
	if !m.DHCPV6DNS.IsNull() && !m.DHCPV6DNS.IsUnknown() {
		diags.Append(m.DHCPV6DNS.ElementsAs(ctx, &dhcpV6DNS, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Flatten WAN DNS list
	var wanDNS []string
	if !m.WANDNS.IsNull() && !m.WANDNS.IsUnknown() {
		diags.Append(m.WANDNS.ElementsAs(ctx, &wanDNS, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}

	vlan := int(m.VLANID.ValueInt64())

	var ipSubnet string
	if !m.Subnet.IsNull() && !m.Subnet.IsUnknown() && m.Subnet.ValueString() != "" {
		var err error
		ipSubnet, err = utils.CidrOneBased(m.Subnet.ValueString())
		if err != nil {
			diags.AddError("Invalid subnet CIDR", err.Error())
			return nil, diags
		}
	}

	return &unifi.Network{
		ID:   m.ID.ValueString(),
		Name: m.Name.ValueString(),

		Purpose:      m.Purpose.ValueString(),
		VLAN:         vlan,
		VLANEnabled:  vlan != 0 && vlan != 1,
		IPSubnet:     ipSubnet,
		NetworkGroup: m.NetworkGroup.ValueString(),
		DomainName:   m.DomainName.ValueString(),
		Enabled:      m.Enabled.ValueBool(),

		// DHCP v4
		DHCPDStart:       m.DHCPStart.ValueString(),
		DHCPDStop:        m.DHCPStop.ValueString(),
		DHCPDEnabled:     m.DHCPEnabled.ValueBool(),
		DHCPDLeaseTime:   int(m.DHCPLease.ValueInt64()),
		DHCPRelayEnabled: m.DHCPRelayEnabled.ValueBool(),

		DHCPDDNSEnabled: len(dhcpDNS) > 0,
		DHCPDDNS1:       safeIndex(dhcpDNS, 0),
		DHCPDDNS2:       safeIndex(dhcpDNS, 1),
		DHCPDDNS3:       safeIndex(dhcpDNS, 2),
		DHCPDDNS4:       safeIndex(dhcpDNS, 3),

		// DHCP boot
		DHCPDBootEnabled:  m.DHCPDBootEnabled.ValueBool(),
		DHCPDBootServer:   m.DHCPDBootServer.ValueString(),
		DHCPDBootFilename: m.DHCPDBootFilename.ValueString(),

		// DHCP v6
		DHCPDV6DNS1:      safeIndex(dhcpV6DNS, 0),
		DHCPDV6DNS2:      safeIndex(dhcpV6DNS, 1),
		DHCPDV6DNS3:      safeIndex(dhcpV6DNS, 2),
		DHCPDV6DNS4:      safeIndex(dhcpV6DNS, 3),
		DHCPDV6DNSAuto:   m.DHCPV6DNSAuto.ValueBool(),
		DHCPDV6Enabled:   m.DHCPV6Enabled.ValueBool(),
		DHCPDV6LeaseTime: int(m.DHCPV6Lease.ValueInt64()),
		DHCPDV6Start:     m.DHCPV6Start.ValueString(),
		DHCPDV6Stop:      m.DHCPV6Stop.ValueString(),

		// Network features
		IGMPSnooping:            m.IGMPSnooping.ValueBool(),
		MdnsEnabled:             m.MulticastDNS.ValueBool(),
		InternetAccessEnabled:   m.InternetAccessEnabled.ValueBool(),
		NetworkIsolationEnabled: m.NetworkIsolationEnabled.ValueBool(),

		// IPv6
		IPV6InterfaceType:       m.IPV6InterfaceType.ValueString(),
		IPV6Subnet:              m.IPV6StaticSubnet.ValueString(),
		IPV6PDInterface:         m.IPV6PDInterface.ValueString(),
		IPV6PDPrefixid:          m.IPV6PDPrefixID.ValueString(),
		IPV6PDStart:             m.IPV6PDStart.ValueString(),
		IPV6PDStop:              m.IPV6PDStop.ValueString(),
		IPV6RaEnabled:           m.IPV6RAEnable.ValueBool(),
		IPV6RaPreferredLifetime: int(m.IPV6RAPreferredLifetime.ValueInt64()),
		IPV6RaPriority:          m.IPV6RAPriority.ValueString(),
		IPV6RaValidLifetime:     int(m.IPV6RAValidLifetime.ValueInt64()),

		// WAN IPv4
		WANIP:           m.WANIP.ValueString(),
		WANNetmask:      m.WANNetmask.ValueString(),
		WANGateway:      m.WANGateway.ValueString(),
		WANType:         m.WANType.ValueString(),
		WANNetworkGroup: m.WANNetworkGroup.ValueString(),
		WANEgressQOS:    int(m.WANEgressQOS.ValueInt64()),
		WANUsername:      m.WANUsername.ValueString(),
		XWANPassword:    m.XWANPassword.ValueString(),

		WANDNS1: safeIndex(wanDNS, 0),
		WANDNS2: safeIndex(wanDNS, 1),
		WANDNS3: safeIndex(wanDNS, 2),
		WANDNS4: safeIndex(wanDNS, 3),

		// WAN IPv6
		WANTypeV6:       m.WANTypeV6.ValueString(),
		WANDHCPv6PDSize: int(m.WANDHCPv6PDSize.ValueInt64()),
		WANIPV6:         m.WANIPV6.ValueString(),
		WANGatewayV6:    m.WANGatewayV6.ValueString(),
		WANPrefixlen:    int(m.WANPrefixlen.ValueInt64()),
	}, diags
}

func (m *networkModel) Merge(ctx context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.Network)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.Network")
		return diags
	}

	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.Purpose = types.StringValue(other.Purpose)
	m.NetworkGroup = ut.StringOrNull(other.NetworkGroup)
	m.DomainName = ut.StringOrNull(other.DomainName)
	m.Enabled = types.BoolValue(other.Enabled)

	// VLAN: only set if VLANEnabled, otherwise 0
	if other.VLANEnabled {
		m.VLANID = types.Int64Value(int64(other.VLAN))
	} else {
		m.VLANID = types.Int64Value(0)
	}

	// Subnet: normalize to zero-based CIDR
	if other.IPSubnet != "" {
		subnet, err := utils.CidrZeroBased(other.IPSubnet)
		if err != nil {
			diags.AddError("Invalid subnet from controller", err.Error())
			return diags
		}
		m.Subnet = ut.StringOrNull(subnet)
	} else {
		m.Subnet = types.StringNull()
	}

	// DHCP v4
	m.DHCPStart = ut.StringOrNull(other.DHCPDStart)
	m.DHCPStop = ut.StringOrNull(other.DHCPDStop)
	m.DHCPEnabled = types.BoolValue(other.DHCPDEnabled)
	m.DHCPRelayEnabled = types.BoolValue(other.DHCPRelayEnabled)

	// DHCP lease: default to 86400 when enabled and API returns 0
	dhcpLease := other.DHCPDLeaseTime
	if other.DHCPDEnabled && dhcpLease == 0 {
		dhcpLease = 86400
	}
	m.DHCPLease = types.Int64Value(int64(dhcpLease))

	// DHCP DNS: collect non-empty entries into list
	if other.DHCPDDNSEnabled {
		dhcpDNS := collectNonEmpty(other.DHCPDDNS1, other.DHCPDDNS2, other.DHCPDDNS3, other.DHCPDDNS4)
		list, d := types.ListValueFrom(ctx, types.StringType, dhcpDNS)
		diags.Append(d...)
		if !diags.HasError() {
			m.DHCPDNS = list
		}
	} else {
		m.DHCPDNS = types.ListNull(types.StringType)
	}

	// DHCP boot
	m.DHCPDBootEnabled = types.BoolValue(other.DHCPDBootEnabled)
	m.DHCPDBootServer = ut.StringOrNull(other.DHCPDBootServer)
	m.DHCPDBootFilename = ut.StringOrNull(other.DHCPDBootFilename)

	// DHCP v6
	dhcpV6DNS := collectNonEmpty(other.DHCPDV6DNS1, other.DHCPDV6DNS2, other.DHCPDV6DNS3, other.DHCPDV6DNS4)
	if len(dhcpV6DNS) > 0 {
		list, d := types.ListValueFrom(ctx, types.StringType, dhcpV6DNS)
		diags.Append(d...)
		if !diags.HasError() {
			m.DHCPV6DNS = list
		}
	} else {
		m.DHCPV6DNS = types.ListNull(types.StringType)
	}
	m.DHCPV6DNSAuto = types.BoolValue(other.DHCPDV6DNSAuto)
	m.DHCPV6Enabled = types.BoolValue(other.DHCPDV6Enabled)
	m.DHCPV6Lease = types.Int64Value(int64(other.DHCPDV6LeaseTime))
	m.DHCPV6Start = ut.StringOrNull(other.DHCPDV6Start)
	m.DHCPV6Stop = ut.StringOrNull(other.DHCPDV6Stop)

	// Network features
	m.IGMPSnooping = types.BoolValue(other.IGMPSnooping)
	m.MulticastDNS = types.BoolValue(other.MdnsEnabled)
	m.InternetAccessEnabled = types.BoolValue(other.InternetAccessEnabled)
	m.NetworkIsolationEnabled = types.BoolValue(other.NetworkIsolationEnabled)

	// IPv6
	m.IPV6InterfaceType = types.StringValue(other.IPV6InterfaceType)
	m.IPV6StaticSubnet = ut.StringOrNull(other.IPV6Subnet)
	m.IPV6PDInterface = ut.StringOrNull(other.IPV6PDInterface)
	m.IPV6PDPrefixID = ut.StringOrNull(other.IPV6PDPrefixid)
	m.IPV6PDStart = ut.StringOrNull(other.IPV6PDStart)
	m.IPV6PDStop = ut.StringOrNull(other.IPV6PDStop)
	m.IPV6RAEnable = types.BoolValue(other.IPV6RaEnabled)
	m.IPV6RAPreferredLifetime = types.Int64Value(int64(other.IPV6RaPreferredLifetime))
	m.IPV6RAPriority = ut.StringOrNull(other.IPV6RaPriority)
	m.IPV6RAValidLifetime = types.Int64Value(int64(other.IPV6RaValidLifetime))

	// WAN fields: only populate when purpose is "wan"
	if other.Purpose == "wan" {
		m.WANType = ut.StringOrNull(other.WANType)

		// WAN DNS
		wanDNS := collectNonEmpty(other.WANDNS1, other.WANDNS2, other.WANDNS3, other.WANDNS4)
		if len(wanDNS) > 0 {
			list, d := types.ListValueFrom(ctx, types.StringType, wanDNS)
			diags.Append(d...)
			if !diags.HasError() {
				m.WANDNS = list
			}
		} else {
			m.WANDNS = types.ListNull(types.StringType)
		}

		// WAN IP/netmask/gateway: only set for non-DHCP types
		if other.WANType != "dhcp" {
			m.WANIP = ut.StringOrNull(other.WANIP)
			m.WANNetmask = ut.StringOrNull(other.WANNetmask)
			m.WANGateway = ut.StringOrNull(other.WANGateway)
		} else {
			m.WANIP = ut.StringOrNull("")
			m.WANNetmask = ut.StringOrNull("")
			m.WANGateway = ut.StringOrNull("")
		}
	} else {
		m.WANType = ut.StringOrNull("")
		m.WANDNS = types.ListNull(types.StringType)
		m.WANIP = ut.StringOrNull("")
		m.WANNetmask = ut.StringOrNull("")
		m.WANGateway = ut.StringOrNull("")
	}

	m.WANNetworkGroup = ut.StringOrNull(other.WANNetworkGroup)
	m.WANEgressQOS = types.Int64Value(int64(other.WANEgressQOS))
	m.WANUsername = ut.StringOrNull(other.WANUsername)
	m.XWANPassword = ut.StringOrNull(other.XWANPassword)

	// WAN IPv6
	m.WANTypeV6 = ut.StringOrNull(other.WANTypeV6)
	m.WANDHCPv6PDSize = ut.Int64OrNull(other.WANDHCPv6PDSize)
	m.WANIPV6 = ut.StringOrNull(other.WANIPV6)
	m.WANGatewayV6 = ut.StringOrNull(other.WANGatewayV6)
	m.WANPrefixlen = ut.Int64OrNull(other.WANPrefixlen)

	return diags
}

// safeIndex returns the element at index i from the slice, or "" if out of bounds.
func safeIndex(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}

// collectNonEmpty is defined in datasource_network.go (shared within the package).
