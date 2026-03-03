package dns

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var (
	_ resource.Resource                = &dynamicDNSResource{}
	_ resource.ResourceWithConfigure   = &dynamicDNSResource{}
	_ resource.ResourceWithImportState = &dynamicDNSResource{}
	_ base.Resource                    = &dynamicDNSResource{}
)

type dynamicDNSResource struct {
	*base.GenericResource[*dynamicDNSModel]
}

func NewDynamicDNSResource() resource.Resource {
	return &dynamicDNSResource{
		GenericResource: base.NewGenericResource(
			"unifi_dynamic_dns",
			func() *dynamicDNSModel { return &dynamicDNSModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetDynamicDNS(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateDynamicDNS(ctx, site, body.(*unifi.DynamicDNS))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateDynamicDNS(ctx, site, body.(*unifi.DynamicDNS))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteDynamicDNS(ctx, site, id)
				},
			},
		),
	}
}

func (r *dynamicDNSResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_dynamic_dns` resource manages Dynamic DNS (DDNS) configurations.\n\n" +
			"Dynamic DNS allows you to access your network using a domain name even when your public IP address changes.",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"interface": schema.StringAttribute{
				MarkdownDescription: "The WAN interface to use for DDNS updates. Valid values: `wan`, `wan2`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("wan"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service": schema.StringAttribute{
				MarkdownDescription: "The DDNS service provider (e.g., `dyndns`, `noip`, `duckdns`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"host_name": schema.StringAttribute{
				MarkdownDescription: "The fully qualified domain name to update (e.g., 'myhouse.dyndns.org').",
				Required:            true,
			},
			"server": schema.StringAttribute{
				MarkdownDescription: "The update server hostname for the DDNS provider.",
				Optional:            true,
			},
			"login": schema.StringAttribute{
				MarkdownDescription: "The username or login for the DDNS provider account.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password or token for the DDNS provider account.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}
