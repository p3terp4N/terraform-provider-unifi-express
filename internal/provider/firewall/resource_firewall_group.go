package firewall

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &firewallGroupResource{}
	_ resource.ResourceWithConfigure   = &firewallGroupResource{}
	_ resource.ResourceWithImportState = &firewallGroupResource{}
	_ base.Resource                    = &firewallGroupResource{}
)

type firewallGroupResource struct {
	*base.GenericResource[*firewallGroupModel]
}

func NewFirewallGroupResource() resource.Resource {
	return &firewallGroupResource{
		GenericResource: base.NewGenericResource(
			"unifi_firewall_group",
			func() *firewallGroupModel { return &firewallGroupModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetFirewallGroup(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateFirewallGroup(ctx, site, body.(*unifi.FirewallGroup))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateFirewallGroup(ctx, site, body.(*unifi.FirewallGroup))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteFirewallGroup(ctx, site, id)
				},
			},
		),
	}
}

func (r *firewallGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_firewall_group` resource manages reusable groups of addresses or ports that can be referenced in firewall rules (`unifi_firewall_rule`).\n\n" +
			"Firewall groups help organize and simplify firewall rule management by allowing you to:\n" +
			"  * Create collections of IP addresses or networks\n" +
			"  * Define sets of ports for specific services\n" +
			"  * Group IPv6 addresses for IPv6-specific rules",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the firewall group (e.g., 'Trusted IPs' or 'Web Server Ports'). Must be unique within the site.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of firewall group. Valid values are:\n" +
					"  * `address-group` - Group of IPv4 addresses and/or networks\n" +
					"  * `port-group` - Group of ports or port ranges\n" +
					"  * `ipv6-address-group` - Group of IPv6 addresses and/or networks",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("address-group", "port-group", "ipv6-address-group"),
				},
			},
			"members": schema.SetAttribute{
				MarkdownDescription: "The members of the group. Format depends on the group type:\n" +
					"  * For address-group: IPv4 addresses or CIDR notation\n" +
					"  * For port-group: Port numbers or ranges (e.g., '80', '8000-8080')\n" +
					"  * For ipv6-address-group: IPv6 addresses or CIDR notation",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}
