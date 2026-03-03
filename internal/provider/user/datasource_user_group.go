package user

import (
	"context"
	"fmt"

	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &userGroupDatasource{}
	_ datasource.DataSourceWithConfigure = &userGroupDatasource{}
	_ base.Resource                      = &userGroupDatasource{}
)

type userGroupDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type userGroupDatasourceModel struct {
	base.Model
	Name           types.String `tfsdk:"name"`
	QOSRateMaxDown types.Int64  `tfsdk:"qos_rate_max_down"`
	QOSRateMaxUp   types.Int64  `tfsdk:"qos_rate_max_up"`
}

func NewUserGroupDatasource() datasource.DataSource {
	return &userGroupDatasource{}
}

func (d *userGroupDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *userGroupDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *userGroupDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *userGroupDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *userGroupDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_user_group", req.ProviderTypeName)
}

func (d *userGroupDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_user_group` data source can be used to retrieve the ID for a user group by name.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user group to look up. Defaults to `Default`.",
				Optional:            true,
			},
			"qos_rate_max_down": schema.Int64Attribute{
				MarkdownDescription: "The QoS maximum download rate.",
				Computed:            true,
			},
			"qos_rate_max_up": schema.Int64Attribute{
				MarkdownDescription: "The QoS maximum upload rate.",
				Computed:            true,
			},
		},
	}
}

func (d *userGroupDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userGroupDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	groups, err := d.client.ListUserGroup(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list user groups", err.Error())
		return
	}

	name := state.Name.ValueString()
	if name == "" {
		name = "Default"
	}
	for _, g := range groups {
		if g.Name == name {
			state.SetID(g.ID)
			state.SetSite(site)
			state.QOSRateMaxDown = types.Int64Value(int64(g.QOSRateMaxDown))
			state.QOSRateMaxUp = types.Int64Value(int64(g.QOSRateMaxUp))
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("User group not found", fmt.Sprintf("User group not found with name %q", name))
}
