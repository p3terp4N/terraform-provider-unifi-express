package apgroup

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
	_ datasource.DataSource              = &apGroupDatasource{}
	_ datasource.DataSourceWithConfigure = &apGroupDatasource{}
	_ base.Resource                      = &apGroupDatasource{}
)

type apGroupDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type apGroupDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

func NewAPGroupDatasource() datasource.DataSource {
	return &apGroupDatasource{}
}

func (d *apGroupDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *apGroupDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *apGroupDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *apGroupDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *apGroupDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_ap_group", req.ProviderTypeName)
}

func (d *apGroupDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_ap_group` data source can be used to retrieve the ID for an AP group by name.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the AP group to look up. Leave blank to look up the default AP group.",
				Optional:            true,
			},
		},
	}
}

func (d *apGroupDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apGroupDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	groups, err := d.client.ListAPGroup(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list AP groups", err.Error())
		return
	}

	name := state.Name.ValueString()
	for _, g := range groups {
		if (name == "" && g.HiddenID == "default") || g.Name == name {
			state.SetID(g.ID)
			state.SetSite(site)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("AP group not found", fmt.Sprintf("AP group not found with name %q", name))
}
