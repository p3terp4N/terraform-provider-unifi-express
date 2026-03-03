package radius

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
	_ datasource.DataSource              = &accountDatasource{}
	_ datasource.DataSourceWithConfigure = &accountDatasource{}
	_ base.Resource                      = &accountDatasource{}
)

type accountDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type accountDatasourceModel struct {
	base.Model
	Name             types.String `tfsdk:"name"`
	Password         types.String `tfsdk:"password"`
	TunnelType       types.Int64  `tfsdk:"tunnel_type"`
	TunnelMediumType types.Int64  `tfsdk:"tunnel_medium_type"`
	NetworkID        types.String `tfsdk:"network_id"`
}

func NewAccountDatasource() datasource.DataSource {
	return &accountDatasource{}
}

func (d *accountDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *accountDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *accountDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *accountDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *accountDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_account", req.ProviderTypeName)
}

func (d *accountDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_account` data source can be used to retrieve RADIUS user accounts.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the account to look up.",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password of the account.",
				Computed:            true,
				Sensitive:           true,
			},
			"tunnel_type": schema.Int64Attribute{
				MarkdownDescription: "See RFC2868 section 3.1.",
				Computed:            true,
			},
			"tunnel_medium_type": schema.Int64Attribute{
				MarkdownDescription: "See RFC2868 section 3.2.",
				Computed:            true,
			},
			"network_id": schema.StringAttribute{
				MarkdownDescription: "ID of the network for this account.",
				Computed:            true,
			},
		},
	}
}

func (d *accountDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state accountDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	accounts, err := d.client.ListAccount(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list accounts", err.Error())
		return
	}

	name := state.Name.ValueString()
	for _, a := range accounts {
		if a.Name == name {
			state.SetID(a.ID)
			state.SetSite(site)
			state.Name = types.StringValue(a.Name)
			state.Password = types.StringValue(a.XPassword)
			state.TunnelType = types.Int64Value(int64(a.TunnelType))
			state.TunnelMediumType = types.Int64Value(int64(a.TunnelMediumType))
			state.NetworkID = types.StringValue(a.NetworkID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("Account not found", fmt.Sprintf("Account not found with name %q", name))
}
