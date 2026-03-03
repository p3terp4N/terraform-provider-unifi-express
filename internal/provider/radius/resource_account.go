package radius

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &accountResource{}
	_ resource.ResourceWithConfigure   = &accountResource{}
	_ resource.ResourceWithImportState = &accountResource{}
	_ base.Resource                    = &accountResource{}
)

type accountResource struct {
	*base.GenericResource[*accountModel]
}

func NewAccountResource() resource.Resource {
	return &accountResource{
		GenericResource: base.NewGenericResource(
			"unifi_account",
			func() *accountModel { return &accountModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetAccount(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateAccount(ctx, site, body.(*unifi.Account))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateAccount(ctx, site, body.(*unifi.Account))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteAccount(ctx, site, id)
				},
			},
		),
	}
}

func (r *accountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_account` resource manages RADIUS user accounts in the UniFi controller's built-in RADIUS server.\n\n" +
			"This resource is used for:\n" +
			"  * WPA2/WPA3-Enterprise wireless authentication\n" +
			"  * 802.1X wired authentication\n" +
			"  * MAC-based device authentication\n" +
			"  * VLAN assignment through RADIUS attributes",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The username for this RADIUS account. For MAC-based authentication, use the device's MAC address in uppercase with no separators.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for this RADIUS account.",
				Required:            true,
				Sensitive:           true,
			},
			"tunnel_type": schema.Int64Attribute{
				MarkdownDescription: "The RADIUS tunnel type attribute (RFC 2868, section 3.1). Default: `13` (VLAN).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(13),
				Validators: []validator.Int64{
					int64validator.Between(1, 13),
				},
			},
			"tunnel_medium_type": schema.Int64Attribute{
				MarkdownDescription: "The RADIUS tunnel medium type attribute (RFC 2868, section 3.2). Default: `6` (802).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(6),
				Validators: []validator.Int64{
					int64validator.Between(1, 15),
				},
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the network (VLAN) to assign to clients authenticating with this account.",
				Optional:            true,
			},
		},
	}
}
