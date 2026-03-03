package network

import (
	"context"
	"fmt"

	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource                     = &networkDatasource{}
	_ datasource.DataSourceWithConfigure        = &networkDatasource{}
	_ datasource.DataSourceWithConfigValidators = &networkDatasource{}
	_ base.Resource                             = &networkDatasource{}
)

type networkDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type networkDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`

	// Computed fields
	Purpose      types.String `tfsdk:"purpose"`
	VLANID       types.Int64  `tfsdk:"vlan_id"`
	Subnet       types.String `tfsdk:"subnet"`
	NetworkGroup types.String `tfsdk:"network_group"`
	DomainName   types.String `tfsdk:"domain_name"`

	// DHCP
	DHCPEnabled       types.Bool   `tfsdk:"dhcp_enabled"`
	DHCPStart         types.String `tfsdk:"dhcp_start"`
	DHCPStop          types.String `tfsdk:"dhcp_stop"`
	DHCPLease         types.Int64  `tfsdk:"dhcp_lease"`
	DHCPDNS           types.List   `tfsdk:"dhcp_dns"`
	DHCPDBootEnabled  types.Bool   `tfsdk:"dhcpd_boot_enabled"`
	DHCPDBootServer   types.String `tfsdk:"dhcpd_boot_server"`
	DHCPDBootFilename types.String `tfsdk:"dhcpd_boot_filename"`

	// DHCPv6
	DHCPV6DNS     types.List   `tfsdk:"dhcp_v6_dns"`
	DHCPV6DNSAuto types.Bool   `tfsdk:"dhcp_v6_dns_auto"`
	DHCPV6Enabled types.Bool   `tfsdk:"dhcp_v6_enabled"`
	DHCPV6Lease   types.Int64  `tfsdk:"dhcp_v6_lease"`
	DHCPV6Start   types.String `tfsdk:"dhcp_v6_start"`
	DHCPV6Stop    types.String `tfsdk:"dhcp_v6_stop"`

	// IPv6
	IGMPSnooping        types.Bool   `tfsdk:"igmp_snooping"`
	IPV6InterfaceType   types.String `tfsdk:"ipv6_interface_type"`
	IPV6StaticSubnet    types.String `tfsdk:"ipv6_static_subnet"`
	IPV6PDInterface     types.String `tfsdk:"ipv6_pd_interface"`
	IPV6PDPrefixID      types.String `tfsdk:"ipv6_pd_prefixid"`
	IPV6PDStart         types.String `tfsdk:"ipv6_pd_start"`
	IPV6PDStop          types.String `tfsdk:"ipv6_pd_stop"`
	IPV6RAEnable        types.Bool   `tfsdk:"ipv6_ra_enable"`
	IPV6RAPreferredLife types.Int64  `tfsdk:"ipv6_ra_preferred_lifetime"`
	IPV6RAPriority      types.String `tfsdk:"ipv6_ra_priority"`
	IPV6RAValidLife     types.Int64  `tfsdk:"ipv6_ra_valid_lifetime"`

	MulticastDNS types.Bool `tfsdk:"multicast_dns"`

	// WAN
	WANIP            types.String `tfsdk:"wan_ip"`
	WANNetmask       types.String `tfsdk:"wan_netmask"`
	WANGateway       types.String `tfsdk:"wan_gateway"`
	WANDNS           types.List   `tfsdk:"wan_dns"`
	WANType          types.String `tfsdk:"wan_type"`
	WANNetworkGroup  types.String `tfsdk:"wan_networkgroup"`
	WANEgressQOS     types.Int64  `tfsdk:"wan_egress_qos"`
	WANUsername      types.String `tfsdk:"wan_username"`
	XWANPassword     types.String `tfsdk:"x_wan_password"`
	WANTypeV6        types.String `tfsdk:"wan_type_v6"`
	WANDHCPV6PDSize  types.Int64  `tfsdk:"wan_dhcp_v6_pd_size"`
	WANIPV6          types.String `tfsdk:"wan_ipv6"`
	WANGatewayV6     types.String `tfsdk:"wan_gateway_v6"`
	WANPrefixlen     types.Int64  `tfsdk:"wan_prefixlen"`
}

func NewNetworkDatasource() datasource.DataSource {
	return &networkDatasource{}
}

func (d *networkDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *networkDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *networkDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *networkDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *networkDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_network", req.ProviderTypeName)
}

func (d *networkDatasource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("id"),
			path.MatchRoot("name"),
		),
	}
}

func (d *networkDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_network` data source can be used to retrieve settings for a network by name or ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network.",
				Optional:            true,
				Computed:            true,
			},
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the network.",
				Optional:            true,
				Computed:            true,
			},

			"purpose": schema.StringAttribute{
				MarkdownDescription: "The purpose of the network (`corporate`, `guest`, `wan`, or `vlan-only`).",
				Computed:            true,
			},
			"vlan_id": schema.Int64Attribute{
				MarkdownDescription: "The VLAN ID of the network.",
				Computed:            true,
			},
			"subnet": schema.StringAttribute{
				MarkdownDescription: "The subnet of the network (CIDR).",
				Computed:            true,
			},
			"network_group": schema.StringAttribute{
				MarkdownDescription: "The group of the network.",
				Computed:            true,
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name of this network.",
				Computed:            true,
			},
			"dhcp_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether DHCP is enabled.",
				Computed:            true,
			},
			"dhcp_start": schema.StringAttribute{
				MarkdownDescription: "The start of the DHCP range.",
				Computed:            true,
			},
			"dhcp_stop": schema.StringAttribute{
				MarkdownDescription: "The end of the DHCP range.",
				Computed:            true,
			},
			"dhcp_lease": schema.Int64Attribute{
				MarkdownDescription: "Lease time for DHCP addresses.",
				Computed:            true,
			},
			"dhcp_dns": schema.ListAttribute{
				MarkdownDescription: "IPv4 addresses for the DHCP DNS server.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"dhcpd_boot_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether DHCP boot options are enabled.",
				Computed:            true,
			},
			"dhcpd_boot_server": schema.StringAttribute{
				MarkdownDescription: "IPv4 address of a TFTP server for network boot.",
				Computed:            true,
			},
			"dhcpd_boot_filename": schema.StringAttribute{
				MarkdownDescription: "The PXE boot filename.",
				Computed:            true,
			},
			"dhcp_v6_dns": schema.ListAttribute{
				MarkdownDescription: "IPv6 addresses for the DHCPv6 DNS server.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"dhcp_v6_dns_auto": schema.BoolAttribute{
				MarkdownDescription: "Whether to use automatic DNS for DHCPv6.",
				Computed:            true,
			},
			"dhcp_v6_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether stateful DHCPv6 is enabled.",
				Computed:            true,
			},
			"dhcp_v6_lease": schema.Int64Attribute{
				MarkdownDescription: "Lease time for DHCPv6 addresses.",
				Computed:            true,
			},
			"dhcp_v6_start": schema.StringAttribute{
				MarkdownDescription: "Start address of the DHCPv6 range.",
				Computed:            true,
			},
			"dhcp_v6_stop": schema.StringAttribute{
				MarkdownDescription: "End address of the DHCPv6 range.",
				Computed:            true,
			},
			"igmp_snooping": schema.BoolAttribute{
				MarkdownDescription: "Whether IGMP snooping is enabled.",
				Computed:            true,
			},
			"ipv6_interface_type": schema.StringAttribute{
				MarkdownDescription: "IPv6 connection type (`static`, `pd`, or `none`).",
				Computed:            true,
			},
			"ipv6_static_subnet": schema.StringAttribute{
				MarkdownDescription: "The static IPv6 subnet.",
				Computed:            true,
			},
			"ipv6_pd_interface": schema.StringAttribute{
				MarkdownDescription: "WAN interface for IPv6 PD (`wan` or `wan2`).",
				Computed:            true,
			},
			"ipv6_pd_prefixid": schema.StringAttribute{
				MarkdownDescription: "The IPv6 Prefix ID.",
				Computed:            true,
			},
			"ipv6_pd_start": schema.StringAttribute{
				MarkdownDescription: "Start of the DHCPv6 PD range.",
				Computed:            true,
			},
			"ipv6_pd_stop": schema.StringAttribute{
				MarkdownDescription: "End of the DHCPv6 PD range.",
				Computed:            true,
			},
			"ipv6_ra_enable": schema.BoolAttribute{
				MarkdownDescription: "Whether router advertisements are enabled.",
				Computed:            true,
			},
			"ipv6_ra_preferred_lifetime": schema.Int64Attribute{
				MarkdownDescription: "IPv6 RA preferred lifetime.",
				Computed:            true,
			},
			"ipv6_ra_priority": schema.StringAttribute{
				MarkdownDescription: "IPv6 RA priority (`high`, `medium`, or `low`).",
				Computed:            true,
			},
			"ipv6_ra_valid_lifetime": schema.Int64Attribute{
				MarkdownDescription: "IPv6 RA valid lifetime.",
				Computed:            true,
			},
			"multicast_dns": schema.BoolAttribute{
				MarkdownDescription: "Whether mDNS is enabled.",
				Computed:            true,
			},
			"wan_ip": schema.StringAttribute{
				MarkdownDescription: "The WAN IPv4 address.",
				Computed:            true,
			},
			"wan_netmask": schema.StringAttribute{
				MarkdownDescription: "The WAN IPv4 netmask.",
				Computed:            true,
			},
			"wan_gateway": schema.StringAttribute{
				MarkdownDescription: "The WAN IPv4 gateway.",
				Computed:            true,
			},
			"wan_dns": schema.ListAttribute{
				MarkdownDescription: "WAN DNS server IPs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"wan_type": schema.StringAttribute{
				MarkdownDescription: "The WAN connection type (`disabled`, `static`, `dhcp`, or `pppoe`).",
				Computed:            true,
			},
			"wan_networkgroup": schema.StringAttribute{
				MarkdownDescription: "The WAN network group (`WAN`, `WAN2`, or `WAN_LTE_FAILOVER`).",
				Computed:            true,
			},
			"wan_egress_qos": schema.Int64Attribute{
				MarkdownDescription: "The WAN egress QoS value.",
				Computed:            true,
			},
			"wan_username": schema.StringAttribute{
				MarkdownDescription: "The WAN username.",
				Computed:            true,
			},
			"x_wan_password": schema.StringAttribute{
				MarkdownDescription: "The WAN password.",
				Computed:            true,
				Sensitive:           true,
			},
			"wan_type_v6": schema.StringAttribute{
				MarkdownDescription: "The IPv6 WAN connection type (`disabled`, `static`, or `dhcpv6`).",
				Computed:            true,
			},
			"wan_dhcp_v6_pd_size": schema.Int64Attribute{
				MarkdownDescription: "IPv6 prefix size to request from ISP (48-64).",
				Computed:            true,
			},
			"wan_ipv6": schema.StringAttribute{
				MarkdownDescription: "The WAN IPv6 address.",
				Computed:            true,
			},
			"wan_gateway_v6": schema.StringAttribute{
				MarkdownDescription: "The WAN IPv6 gateway.",
				Computed:            true,
			},
			"wan_prefixlen": schema.Int64Attribute{
				MarkdownDescription: "The WAN IPv6 prefix length.",
				Computed:            true,
			},
		},
	}
}

func (d *networkDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state networkDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	var nameFilter, idFilter string
	if !state.Name.IsNull() && !state.Name.IsUnknown() {
		nameFilter = state.Name.ValueString()
	}
	if !state.ID.IsNull() && !state.ID.IsUnknown() {
		idFilter = state.ID.ValueString()
	}

	networks, err := d.client.ListNetwork(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list networks", err.Error())
		return
	}

	for _, n := range networks {
		if (nameFilter != "" && n.Name == nameFilter) || (idFilter != "" && n.ID == idFilter) {
			// Collect DHCP DNS entries
			dhcpDNS := collectNonEmpty(n.DHCPDDNS1, n.DHCPDDNS2, n.DHCPDDNS3, n.DHCPDDNS4)
			wanDNS := collectNonEmpty(n.WANDNS1, n.WANDNS2, n.WANDNS3, n.WANDNS4)

			state.SetID(n.ID)
			state.SetSite(site)
			state.Name = types.StringValue(n.Name)
			state.Purpose = ut.StringOrNull(n.Purpose)
			state.VLANID = types.Int64Value(int64(n.VLAN))
			if n.IPSubnet != "" {
				subnet, err := utils.CidrZeroBased(n.IPSubnet)
				if err != nil {
					resp.Diagnostics.AddError("Invalid subnet from controller", err.Error())
					return
				}
				state.Subnet = ut.StringOrNull(subnet)
			} else {
				state.Subnet = types.StringNull()
			}
			state.NetworkGroup = ut.StringOrNull(n.NetworkGroup)
			state.DomainName = ut.StringOrNull(n.DomainName)

			state.DHCPEnabled = types.BoolValue(n.DHCPDEnabled)
			state.DHCPStart = ut.StringOrNull(n.DHCPDStart)
			state.DHCPStop = ut.StringOrNull(n.DHCPDStop)
			state.DHCPLease = types.Int64Value(int64(n.DHCPDLeaseTime))
			dhcpDNSList, diags := types.ListValueFrom(ctx, types.StringType, dhcpDNS)
			resp.Diagnostics.Append(diags...)
			state.DHCPDNS = dhcpDNSList

			state.DHCPDBootEnabled = types.BoolValue(n.DHCPDBootEnabled)
			state.DHCPDBootServer = ut.StringOrNull(n.DHCPDBootServer)
			state.DHCPDBootFilename = ut.StringOrNull(n.DHCPDBootFilename)

			dhcpV6DNS := collectNonEmpty(n.DHCPDV6DNS1, n.DHCPDV6DNS2, n.DHCPDV6DNS3, n.DHCPDV6DNS4)
			dhcpV6DNSList, diags := types.ListValueFrom(ctx, types.StringType, dhcpV6DNS)
			resp.Diagnostics.Append(diags...)
			state.DHCPV6DNS = dhcpV6DNSList
			state.DHCPV6DNSAuto = types.BoolValue(n.DHCPDV6DNSAuto)
			state.DHCPV6Enabled = types.BoolValue(n.DHCPDV6Enabled)
			state.DHCPV6Lease = types.Int64Value(int64(n.DHCPDV6LeaseTime))
			state.DHCPV6Start = ut.StringOrNull(n.DHCPDV6Start)
			state.DHCPV6Stop = ut.StringOrNull(n.DHCPDV6Stop)

			state.IGMPSnooping = types.BoolValue(n.IGMPSnooping)
			state.IPV6InterfaceType = ut.StringOrNull(n.IPV6InterfaceType)
			state.IPV6StaticSubnet = ut.StringOrNull(n.IPV6Subnet)
			state.IPV6PDInterface = ut.StringOrNull(n.IPV6PDInterface)
			state.IPV6PDPrefixID = ut.StringOrNull(n.IPV6PDPrefixid)
			state.IPV6PDStart = ut.StringOrNull(n.IPV6PDStart)
			state.IPV6PDStop = ut.StringOrNull(n.IPV6PDStop)
			state.IPV6RAEnable = types.BoolValue(n.IPV6RaEnabled)
			state.IPV6RAPreferredLife = types.Int64Value(int64(n.IPV6RaPreferredLifetime))
			state.IPV6RAPriority = ut.StringOrNull(n.IPV6RaPriority)
			state.IPV6RAValidLife = types.Int64Value(int64(n.IPV6RaValidLifetime))

			state.MulticastDNS = types.BoolValue(n.MdnsEnabled)

			state.WANIP = ut.StringOrNull(n.WANIP)
			state.WANNetmask = ut.StringOrNull(n.WANNetmask)
			state.WANGateway = ut.StringOrNull(n.WANGateway)
			wanDNSList, diags := types.ListValueFrom(ctx, types.StringType, wanDNS)
			resp.Diagnostics.Append(diags...)
			state.WANDNS = wanDNSList
			state.WANType = ut.StringOrNull(n.WANType)
			state.WANNetworkGroup = ut.StringOrNull(n.WANNetworkGroup)
			state.WANEgressQOS = types.Int64Value(int64(n.WANEgressQOS))
			state.WANUsername = ut.StringOrNull(n.WANUsername)
			state.XWANPassword = ut.StringOrNull(n.XWANPassword)
			state.WANTypeV6 = ut.StringOrNull(n.WANTypeV6)
			state.WANDHCPV6PDSize = types.Int64Value(int64(n.WANDHCPv6PDSize))
			state.WANIPV6 = ut.StringOrNull(n.WANIPV6)
			state.WANGatewayV6 = ut.StringOrNull(n.WANGatewayV6)
			state.WANPrefixlen = types.Int64Value(int64(n.WANPrefixlen))

			if resp.Diagnostics.HasError() {
				return
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	filter := nameFilter
	if filter == "" {
		filter = idFilter
	}
	resp.Diagnostics.AddError("Network not found", fmt.Sprintf("No network found matching %q", filter))
}

func collectNonEmpty(values ...string) []string {
	var result []string
	for _, v := range values {
		if v != "" {
			result = append(result, v)
		}
	}
	if result == nil {
		result = []string{}
	}
	return result
}
