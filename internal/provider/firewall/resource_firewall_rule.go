package firewall

import (
	"context"
	"regexp"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var firewallRuleProtocolRegexp = regexp.MustCompile("^($|all|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])|tcp_udp|ah|ax.25|dccp|ddp|egp|eigrp|encap|esp|etherip|fc|ggp|gre|hip|hmp|icmp|idpr-cmtp|idrp|igmp|igp|ip|ipcomp|ipencap|ipip|ipv6|ipv6-frag|ipv6-icmp|ipv6-nonxt|ipv6-opts|ipv6-route|isis|iso-tp4|l2tp|manet|mobility-header|mpls-in-ip|ospf|pim|pup|rdp|rohc|rspf|rsvp|sctp|shim6|skip|st|tcp|udp|udplite|vmtp|vrrp|wesp|xns-idp|xtp)$")
var firewallRuleProtocolV6Regexp = regexp.MustCompile("^($|([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])|ah|all|dccp|eigrp|esp|gre|icmpv6|ipcomp|ipv6|ipv6-frag|ipv6-icmp|ipv6-nonxt|ipv6-opts|ipv6-route|isis|l2tp|manet|mobility-header|mpls-in-ip|ospf|pim|rsvp|sctp|shim6|tcp|tcp_udp|udp|vrrp)$")
var firewallRuleICMPv6TypenameRegexp = regexp.MustCompile("^($|address-unreachable|bad-header|beyond-scope|communication-prohibited|destination-unreachable|echo-reply|echo-request|failed-policy|neighbor-advertisement|neighbor-solicitation|no-route|packet-too-big|parameter-problem|port-unreachable|redirect|reject-route|router-advertisement|router-solicitation|time-exceeded|ttl-zero-during-reassembly|ttl-zero-during-transit|unknown-header-type|unknown-option)$")

var (
	_ resource.Resource                = &firewallRuleResource{}
	_ resource.ResourceWithConfigure   = &firewallRuleResource{}
	_ resource.ResourceWithImportState = &firewallRuleResource{}
	_ base.Resource                    = &firewallRuleResource{}
)

type firewallRuleResource struct {
	*base.GenericResource[*firewallRuleModel]
}

func NewFirewallRuleResource() resource.Resource {
	return &firewallRuleResource{
		GenericResource: base.NewGenericResource(
			"unifi_firewall_rule",
			func() *firewallRuleModel { return &firewallRuleModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetFirewallRule(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateFirewallRule(ctx, site, body.(*unifi.FirewallRule))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateFirewallRule(ctx, site, body.(*unifi.FirewallRule))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteFirewallRule(ctx, site, id)
				},
			},
		),
	}
}

func (r *firewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	portRangeValidator := stringvalidator.RegexMatches(validators.PortRangeRegexp, "invalid port range")
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_firewall_rule` resource manages firewall rules.\n\n" +
			"Rules control traffic flow between network segments (WAN, LAN, Guest) for both IPv4 and IPv6. " +
			"Rules are processed in order based on `rule_index` (lower numbers first).",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the firewall rule.",
				Required:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "The action: `accept`, `drop`, or `reject`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("drop", "accept", "reject"),
				},
			},
			"ruleset": schema.StringAttribute{
				MarkdownDescription: "Traffic flow this rule applies to (e.g., `WAN_IN`, `LAN_OUT`, `GUEST_LOCAL`).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"WAN_IN", "WAN_OUT", "WAN_LOCAL",
						"LAN_IN", "LAN_OUT", "LAN_LOCAL",
						"GUEST_IN", "GUEST_OUT", "GUEST_LOCAL",
						"WANv6_IN", "WANv6_OUT", "WANv6_LOCAL",
						"LANv6_IN", "LANv6_OUT", "LANv6_LOCAL",
						"GUESTv6_IN", "GUESTv6_OUT", "GUESTv6_LOCAL",
					),
				},
			},
			"rule_index": schema.Int64Attribute{
				MarkdownDescription: "Processing order. Lower numbers processed first. Use 2000-2999 or 4000-4999 for custom rules.",
				Required:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "IPv4 protocol (e.g., `all`, `tcp`, `udp`, `tcp_udp`, `icmp`, or protocol number).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(firewallRuleProtocolRegexp, "must be a valid IPv4 protocol"),
				},
			},
			"protocol_v6": schema.StringAttribute{
				MarkdownDescription: "IPv6 protocol (e.g., `all`, `tcp`, `udp`, `tcp_udp`, `ipv6-icmp`).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(firewallRuleProtocolV6Regexp, "must be a valid IPv6 protocol"),
				},
			},
			"icmp_typename": schema.StringAttribute{
				MarkdownDescription: "ICMP type name when protocol is `icmp`.",
				Optional:            true,
			},
			"icmp_v6_typename": schema.StringAttribute{
				MarkdownDescription: "ICMPv6 type name when protocol_v6 is `ipv6-icmp`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(firewallRuleICMPv6TypenameRegexp, "must be a valid ICMPv6 type"),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether this firewall rule is active.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},

			// Sources
			"src_network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the source network.",
				Optional:            true,
			},
			"src_network_type": schema.StringAttribute{
				MarkdownDescription: "Source network address type: `ADDRv4` or `NETv4`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NETv4"),
				Validators: []validator.String{
					stringvalidator.OneOf("ADDRv4", "NETv4"),
				},
			},
			"src_firewall_group_ids": schema.SetAttribute{
				MarkdownDescription: "Firewall group IDs to use as sources.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"src_address": schema.StringAttribute{
				MarkdownDescription: "Source IPv4 address.",
				Optional:            true,
			},
			"src_address_ipv6": schema.StringAttribute{
				MarkdownDescription: "Source IPv6 address or CIDR.",
				Optional:            true,
			},
			"src_port": schema.StringAttribute{
				MarkdownDescription: "Source port(s). Single port, range (e.g., '8000:8080'), or comma-separated list.",
				Optional:            true,
				Validators: []validator.String{
					portRangeValidator,
				},
			},
			"src_mac": schema.StringAttribute{
				MarkdownDescription: "Source MAC address (format: XX:XX:XX:XX:XX:XX).",
				Optional:            true,
			},

			// Destinations
			"dst_network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the destination network.",
				Optional:            true,
			},
			"dst_network_type": schema.StringAttribute{
				MarkdownDescription: "Destination network address type: `ADDRv4` or `NETv4`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("NETv4"),
				Validators: []validator.String{
					stringvalidator.OneOf("ADDRv4", "NETv4"),
				},
			},
			"dst_firewall_group_ids": schema.SetAttribute{
				MarkdownDescription: "Firewall group IDs to use as destinations.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"dst_address": schema.StringAttribute{
				MarkdownDescription: "Destination IPv4 address or CIDR.",
				Optional:            true,
			},
			"dst_address_ipv6": schema.StringAttribute{
				MarkdownDescription: "Destination IPv6 address or CIDR.",
				Optional:            true,
			},
			"dst_port": schema.StringAttribute{
				MarkdownDescription: "Destination port(s). Single port, range, or comma-separated list.",
				Optional:            true,
				Validators: []validator.String{
					portRangeValidator,
				},
			},

			// Advanced
			"logging": schema.BoolAttribute{
				MarkdownDescription: "Enable logging for matched traffic.",
				Optional:            true,
			},
			"state_established": schema.BoolAttribute{
				MarkdownDescription: "Match established connections.",
				Optional:            true,
			},
			"state_invalid": schema.BoolAttribute{
				MarkdownDescription: "Match invalid state connections.",
				Optional:            true,
			},
			"state_new": schema.BoolAttribute{
				MarkdownDescription: "Match new connections.",
				Optional:            true,
			},
			"state_related": schema.BoolAttribute{
				MarkdownDescription: "Match related connections.",
				Optional:            true,
			},
			"ip_sec": schema.StringAttribute{
				MarkdownDescription: "IPsec matching: `match-ipsec` or `match-none`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("match-ipsec", "match-none"),
				},
			},
		},
	}
}
