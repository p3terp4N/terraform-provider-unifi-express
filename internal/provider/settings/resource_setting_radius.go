package settings

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ resource.Resource                = &settingRadiusResource{}
	_ resource.ResourceWithConfigure   = &settingRadiusResource{}
	_ resource.ResourceWithImportState = &settingRadiusResource{}
	_ base.Resource                    = &settingRadiusResource{}
)

type settingRadiusResource struct {
	*base.GenericResource[*settingRadiusModel]
}

func NewSettingRadiusResource() resource.Resource {
	r := &settingRadiusResource{}
	r.GenericResource = NewSettingResource(
		"unifi_setting_radius",
		func() *settingRadiusModel { return &settingRadiusModel{} },
		func(ctx context.Context, client *base.Client, site string) (interface{}, error) {
			return client.GetSettingRadius(ctx, site)
		},
		func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
			return client.UpdateSettingRadius(ctx, site, body.(*unifi.SettingRadius))
		},
	)
	return r
}

func (r *settingRadiusResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_setting_radius` resource manages the built-in RADIUS server configuration in the UniFi controller.\n\n" +
			"The RADIUS server is commonly used for:\n" +
			"  * Enterprise WPA2/WPA3-Enterprise wireless networks\n" +
			"  * 802.1X port-based network access control\n" +
			"  * Centralized user authentication and accounting",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable or disable the built-in RADIUS server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"accounting_enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable RADIUS accounting to track user sessions.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"accounting_port": schema.Int64Attribute{
				MarkdownDescription: "The UDP port number for RADIUS accounting. Default: `1813`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1813),
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"auth_port": schema.Int64Attribute{
				MarkdownDescription: "The UDP port number for RADIUS authentication. Default: `1812`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1812),
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"interim_update_interval": schema.Int64Attribute{
				MarkdownDescription: "The interval (in seconds) for collecting client statistics updates. Default: `3600`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3600),
			},
			"tunneled_reply": schema.BoolAttribute{
				MarkdownDescription: "Enable encrypted RADIUS tunneling for attribute protection.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"secret": schema.StringAttribute{
				MarkdownDescription: "The shared secret for RADIUS client authentication.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				Default:             stringdefault.StaticString(""),
			},
		},
	}
}
