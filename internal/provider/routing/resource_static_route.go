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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &staticRouteResource{}
	_ resource.ResourceWithConfigure   = &staticRouteResource{}
	_ resource.ResourceWithImportState = &staticRouteResource{}
	_ base.Resource                    = &staticRouteResource{}
)

type staticRouteResource struct {
	*base.GenericResource[*staticRouteModel]
}

func NewStaticRouteResource() resource.Resource {
	return &staticRouteResource{
		GenericResource: base.NewGenericResource(
			"unifi_static_route",
			func() *staticRouteModel { return &staticRouteModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetRouting(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateRouting(ctx, site, body.(*unifi.Routing))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateRouting(ctx, site, body.(*unifi.Routing))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteRouting(ctx, site, id)
				},
			},
		),
	}
}

func (r *staticRouteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_static_route` resource manages static routes on UniFi gateways.\n\n" +
			"Routes can be configured to use a next-hop IP address, a specific interface, or as a blackhole route.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "A friendly name for the static route.",
				Required:            true,
			},
			"network": schema.StringAttribute{
				MarkdownDescription: "The destination network in CIDR notation (e.g., '10.0.0.0/16').",
				Required:            true,
				Validators: []validator.String{
					validators.CIDR(),
				},
				PlanModifiers: []planmodifier.String{
					ut.CIDRNormalization(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The route type: `interface-route`, `nexthop-route`, or `blackhole`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("interface-route", "nexthop-route", "blackhole"),
				},
			},
			"distance": schema.Int64Attribute{
				MarkdownDescription: "The administrative distance for route selection. Lower values are preferred.",
				Required:            true,
			},
			"next_hop": schema.StringAttribute{
				MarkdownDescription: "The next-hop router IP address. Used when type is `nexthop-route`.",
				Optional:            true,
			},
			"interface": schema.StringAttribute{
				MarkdownDescription: "The outbound interface (e.g., `WAN1`, `WAN2`, or a network ID). Used when type is `interface-route`.",
				Optional:            true,
			},
		},
	}
}
