package network

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/validators"
)

var (
	wanUsernameRegexp       = regexp.MustCompile(`^([^"' ]+|)$`)
	wanPasswordRegexp       = regexp.MustCompile(`[^"' ]+`)
	wanNetworkGroupRegexp   = regexp.MustCompile(`^(WAN2?|WAN_LTE_FAILOVER)$`)
	wanV6NetworkGroupRegexp = regexp.MustCompile(`^wan2?$`)
)

var (
	_ resource.Resource                = &networkResource{}
	_ resource.ResourceWithConfigure   = &networkResource{}
	_ resource.ResourceWithImportState = &networkResource{}
	_ base.Resource                    = &networkResource{}
)

type networkResource struct {
	*base.GenericResource[*networkModel]
}

func NewNetworkResource() resource.Resource {
	return &networkResource{
		GenericResource: base.NewGenericResource(
			"unifi_network",
			func() *networkModel { return &networkModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetNetwork(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateNetwork(ctx, site, body.(*unifi.Network))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateNetwork(ctx, site, body.(*unifi.Network))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteNetwork(ctx, site, id)
				},
			},
		),
	}
}

func (r *networkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_network` resource manages networks in your UniFi environment, including WAN, LAN, and VLAN networks.\n\n" +
			"This resource enables you to:\n" +
			"* Create and manage different types of networks (corporate, guest, WAN, VLAN-only)\n" +
			"* Configure network addressing and DHCP settings\n" +
			"* Set up IPv6 networking features\n" +
			"* Manage DHCP relay and DNS settings\n" +
			"* Configure network groups and VLANs\n\n" +
			"Common use cases include:\n" +
			"* Setting up corporate and guest networks with different security policies\n" +
			"* Configuring WAN connectivity with various authentication methods\n" +
			"* Creating VLANs for network segmentation\n" +
			"* Managing DHCP and DNS services for network clients",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),

			// Core
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the network. This should be a descriptive name that helps identify the network's purpose, " +
					"such as 'Corporate-Main', 'Guest-Network', or 'IoT-VLAN'.",
				Required: true,
			},
			"purpose": schema.StringAttribute{
				MarkdownDescription: "The purpose/type of the network. Must be one of:\n" +
					"* `corporate` - Standard network for corporate use with full access\n" +
					"* `guest` - Isolated network for guest access with limited permissions\n" +
					"* `wan` - External network connection (WAN uplink)\n" +
					"* `vlan-only` - VLAN network without DHCP services",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("corporate", "guest", "wan", "vlan-only"),
				},
			},
			"vlan_id": schema.Int64Attribute{
				MarkdownDescription: "The VLAN ID for this network. Valid range is 0-4096.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.Between(0, 4096),
				},
			},
			"subnet": schema.StringAttribute{
				MarkdownDescription: "The IPv4 subnet for this network in CIDR notation (e.g., '192.168.1.0/24').",
				Optional:            true,
				Validators: []validator.String{
					validators.CIDR(),
				},
				PlanModifiers: []planmodifier.String{
					ut.CIDRNormalization(),
				},
			},
			"network_group": schema.StringAttribute{
				MarkdownDescription: "The network group for this network. Default is 'LAN'. For WAN networks, use 'WAN' or 'WAN2'.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("LAN"),
			},
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name for this network (e.g., 'corp.example.com').",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Controls whether this network is active.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},

			// DHCP v4
			"dhcp_start": schema.StringAttribute{
				MarkdownDescription: "The starting IPv4 address of the DHCP range.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"dhcp_stop": schema.StringAttribute{
				MarkdownDescription: "The ending IPv4 address of the DHCP range.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"dhcp_enabled": schema.BoolAttribute{
				MarkdownDescription: "Controls whether DHCP server is enabled for this network.",
				Optional:            true,
			},
			"dhcp_lease": schema.Int64Attribute{
				MarkdownDescription: "The DHCP lease time in seconds. Default: 86400 (1 day).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(86400),
			},
			"dhcp_dns": schema.ListAttribute{
				MarkdownDescription: "List of IPv4 DNS server addresses to be provided to DHCP clients. Maximum 4 servers.",
				Optional:            true,
				ElementType:         types.StringType,
				// Note: MaxItems is not directly supported on ListAttribute in Framework;
				// individual element validation is applied via validators on the elements.
			},
			"dhcp_relay_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables DHCP relay for this network.",
				Optional:            true,
			},

			// DHCP boot
			"dhcpd_boot_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables DHCP boot options for PXE boot or network boot configurations.",
				Optional:            true,
			},
			"dhcpd_boot_server": schema.StringAttribute{
				MarkdownDescription: "The IPv4 address of the TFTP server for network boot.",
				Optional:            true,
			},
			"dhcpd_boot_filename": schema.StringAttribute{
				MarkdownDescription: "The boot filename to be loaded from the TFTP server (e.g., 'pxelinux.0').",
				Optional:            true,
			},

			// DHCP v6
			"dhcp_v6_dns": schema.ListAttribute{
				MarkdownDescription: "List of IPv6 DNS server addresses for DHCPv6 clients. Maximum 4 addresses.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"dhcp_v6_dns_auto": schema.BoolAttribute{
				MarkdownDescription: "Controls DNS server source for DHCPv6 clients. `true` uses upstream DNS servers.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"dhcp_v6_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables stateful DHCPv6 for IPv6 address assignment.",
				Optional:            true,
			},
			"dhcp_v6_lease": schema.Int64Attribute{
				MarkdownDescription: "The DHCPv6 lease time in seconds. Default: 86400 (1 day).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(86400),
			},
			"dhcp_v6_start": schema.StringAttribute{
				MarkdownDescription: "The starting IPv6 address for the DHCPv6 range.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv6(),
				},
			},
			"dhcp_v6_stop": schema.StringAttribute{
				MarkdownDescription: "The ending IPv6 address for the DHCPv6 range.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv6(),
				},
			},

			// Network features
			"igmp_snooping": schema.BoolAttribute{
				MarkdownDescription: "Enables IGMP snooping to optimize multicast traffic flow.",
				Optional:            true,
			},
			"multicast_dns": schema.BoolAttribute{
				MarkdownDescription: "Enables Multicast DNS (mDNS/Bonjour/Avahi) on the network.",
				Optional:            true,
			},
			"internet_access_enabled": schema.BoolAttribute{
				MarkdownDescription: "Controls internet access for this network.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"network_isolation_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enables network isolation, preventing communication between clients on this network.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},

			// IPv6
			"ipv6_interface_type": schema.StringAttribute{
				MarkdownDescription: "Specifies the IPv6 connection type: `none`, `static`, or `pd`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "pd", "static"),
				},
			},
			"ipv6_static_subnet": schema.StringAttribute{
				MarkdownDescription: "The static IPv6 subnet in CIDR notation (e.g., '2001:db8::/64'). Only applicable when `ipv6_interface_type` is 'static'.",
				Optional:            true,
			},
			"ipv6_pd_interface": schema.StringAttribute{
				MarkdownDescription: "The WAN interface for IPv6 Prefix Delegation: `wan` or `wan2`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanV6NetworkGroupRegexp, "must be 'wan' or 'wan2'"),
				},
			},
			"ipv6_pd_prefixid": schema.StringAttribute{
				MarkdownDescription: "The IPv6 Prefix ID for Prefix Delegation (hexadecimal value, e.g., '0', '1', 'a1').",
				Optional:            true,
			},
			"ipv6_pd_start": schema.StringAttribute{
				MarkdownDescription: "The starting IPv6 address for Prefix Delegation range.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv6(),
				},
			},
			"ipv6_pd_stop": schema.StringAttribute{
				MarkdownDescription: "The ending IPv6 address for Prefix Delegation range.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv6(),
				},
			},
			"ipv6_ra_enable": schema.BoolAttribute{
				MarkdownDescription: "Enables IPv6 Router Advertisements (RA).",
				Optional:            true,
			},
			"ipv6_ra_preferred_lifetime": schema.Int64Attribute{
				MarkdownDescription: "The preferred lifetime (in seconds) for IPv6 addresses in Router Advertisements. Default: 14400 (4 hours).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(14400),
			},
			"ipv6_ra_priority": schema.StringAttribute{
				MarkdownDescription: "The priority for IPv6 Router Advertisements: `high`, `medium`, or `low`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("high", "medium", "low"),
				},
			},
			"ipv6_ra_valid_lifetime": schema.Int64Attribute{
				MarkdownDescription: "The valid lifetime (in seconds) for IPv6 addresses in Router Advertisements. Default: 86400 (24 hours).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(86400),
			},

			// WAN IPv4
			"wan_ip": schema.StringAttribute{
				MarkdownDescription: "The static IPv4 address for WAN interface. Required when `wan_type` is 'static'.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"wan_netmask": schema.StringAttribute{
				MarkdownDescription: "The IPv4 netmask for WAN interface (e.g., '255.255.255.0').",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"wan_gateway": schema.StringAttribute{
				MarkdownDescription: "The IPv4 gateway address for WAN interface.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"wan_dns": schema.ListAttribute{
				MarkdownDescription: "List of IPv4 DNS servers for WAN interface. Maximum 4 servers.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"wan_type": schema.StringAttribute{
				MarkdownDescription: "The IPv4 WAN connection type: `disabled`, `dhcp`, `static`, or `pppoe`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "dhcp", "static", "pppoe"),
				},
			},
			"wan_networkgroup": schema.StringAttribute{
				MarkdownDescription: "The WAN interface group: `WAN`, `WAN2`, or `WAN_LTE_FAILOVER`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanNetworkGroupRegexp, "must be 'WAN', 'WAN2', or 'WAN_LTE_FAILOVER'"),
				},
			},
			"wan_egress_qos": schema.Int64Attribute{
				MarkdownDescription: "QoS priority for WAN egress traffic (0-7). Default: 0.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"wan_username": schema.StringAttribute{
				MarkdownDescription: "Username for WAN authentication (PPPoE).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanUsernameRegexp, "invalid WAN username"),
				},
			},
			"x_wan_password": schema.StringAttribute{
				MarkdownDescription: "Password for WAN authentication (PPPoE).",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(wanPasswordRegexp, "invalid WAN password"),
				},
			},

			// WAN IPv6
			"wan_type_v6": schema.StringAttribute{
				MarkdownDescription: "The IPv6 WAN connection type: `disabled`, `static`, or `dhcpv6`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "dhcpv6", "static"),
				},
			},
			"wan_dhcp_v6_pd_size": schema.Int64Attribute{
				MarkdownDescription: "The IPv6 prefix size to request from ISP (48-64). Only for `wan_type_v6` = 'dhcpv6'.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(48, 64),
				},
			},
			"wan_ipv6": schema.StringAttribute{
				MarkdownDescription: "The static IPv6 address for WAN interface.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv6(),
				},
			},
			"wan_gateway_v6": schema.StringAttribute{
				MarkdownDescription: "The IPv6 gateway address for WAN interface.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv6(),
				},
			},
			"wan_prefixlen": schema.Int64Attribute{
				MarkdownDescription: "The IPv6 prefix length for WAN interface (1-128).",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 128),
				},
			},
		},
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError("Client Not Configured", "Expected configured client. Please report this issue to the provider developers.")
		return
	}

	id := req.ID
	site := client.Site

	// Support site:id and site:name=xxx format
	if strings.Contains(id, ":") {
		importParts := strings.SplitN(id, ":", 2)
		site = importParts[0]
		id = importParts[1]
	}

	// Support import-by-name: "name=MyNetwork" or "site:name=MyNetwork"
	if strings.HasPrefix(id, "name=") {
		targetName := strings.TrimPrefix(id, "name=")
		resolvedID, err := getNetworkIDByName(ctx, client.Client, targetName, site)
		if err != nil {
			resp.Diagnostics.AddError("Error finding network by name", err.Error())
			return
		}
		id = resolvedID
	}

	// Read the network to populate state
	state := &networkModel{}
	state.SetID(id)
	state.SetSite(site)

	network, err := client.GetNetwork(ctx, site, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading network during import", err.Error())
		return
	}

	resp.Diagnostics.Append(state.Merge(ctx, network)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SetSite(site)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func getNetworkIDByName(ctx context.Context, client unifi.Client, networkName, site string) (string, error) {
	networks, err := client.ListNetwork(ctx, site)
	if err != nil {
		return "", err
	}

	idMatchingName := ""
	var allNames []string
	for _, network := range networks {
		allNames = append(allNames, network.Name)
		if network.Name != networkName {
			continue
		}
		if idMatchingName != "" {
			return "", fmt.Errorf("found multiple networks with name '%s'", networkName)
		}
		idMatchingName = network.ID
	}
	if idMatchingName == "" {
		return "", fmt.Errorf("found no networks with name '%s', found: %s", networkName, strings.Join(allNames, ", "))
	}
	return idMatchingName, nil
}
