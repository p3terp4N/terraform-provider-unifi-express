package routing

import (
	"context"

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
)

var (
	_ resource.Resource                = &portForwardResource{}
	_ resource.ResourceWithConfigure   = &portForwardResource{}
	_ resource.ResourceWithImportState = &portForwardResource{}
	_ base.Resource                    = &portForwardResource{}
)

type portForwardResource struct {
	*base.GenericResource[*portForwardModel]
}

func NewPortForwardResource() resource.Resource {
	return &portForwardResource{
		GenericResource: base.NewGenericResource(
			"unifi_port_forward",
			func() *portForwardModel { return &portForwardModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetPortForward(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreatePortForward(ctx, site, body.(*unifi.PortForward))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdatePortForward(ctx, site, body.(*unifi.PortForward))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeletePortForward(ctx, site, id)
				},
			},
		),
	}
}

func (r *portForwardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	portRangeValidator := stringvalidator.RegexMatches(validators.PortRangeRegexp, "invalid port range")
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_port_forward` resource manages port forwarding rules on UniFi controllers.\n\n" +
			"Port forwarding allows external traffic to reach services hosted on your internal network.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"dst_port": schema.StringAttribute{
				MarkdownDescription: "The external port(s) to forward. Single port (e.g., '80') or range (e.g., '8080:8090').",
				Optional:            true,
				Validators: []validator.String{
					portRangeValidator,
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the port forwarding rule is enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				DeprecationMessage:  "This attribute will be removed in a future release. Instead of disabling a port forwarding rule you can remove it from your configuration.",
			},
			"fwd_ip": schema.StringAttribute{
				MarkdownDescription: "The internal IPv4 address to forward traffic to.",
				Optional:            true,
				Validators: []validator.String{
					validators.IPv4(),
				},
			},
			"fwd_port": schema.StringAttribute{
				MarkdownDescription: "The internal port(s) to forward traffic to.",
				Optional:            true,
				Validators: []validator.String{
					portRangeValidator,
				},
			},
			"log": schema.BoolAttribute{
				MarkdownDescription: "Enable logging of traffic matching this rule.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the port forwarding rule.",
				Optional:            true,
			},
			"port_forward_interface": schema.StringAttribute{
				MarkdownDescription: "The WAN interface to apply the rule to. Valid values: `wan`, `wan2`, `both`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("wan", "wan2", "both"),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol(s) this rule applies to. Valid values: `tcp_udp`, `tcp`, `udp`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("tcp_udp"),
				Validators: []validator.String{
					stringvalidator.OneOf("tcp_udp", "tcp", "udp"),
				},
			},
			"src_ip": schema.StringAttribute{
				MarkdownDescription: "Source IP, network, or 'any' allowed to use this forward.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("any"),
				Validators: []validator.String{
					stringvalidator.Any(
						stringvalidator.OneOf("any"),
						validators.IPv4(),
						validators.CIDR(),
					),
				},
			},
		},
	}
}
