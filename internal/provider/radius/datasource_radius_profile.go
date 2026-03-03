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
	_ datasource.DataSource              = &radiusProfileDatasource{}
	_ datasource.DataSourceWithConfigure = &radiusProfileDatasource{}
	_ base.Resource                      = &radiusProfileDatasource{}
)

type radiusProfileDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type radiusProfileDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

func NewRadiusProfileDatasource() datasource.DataSource {
	return &radiusProfileDatasource{}
}

func (d *radiusProfileDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *radiusProfileDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *radiusProfileDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *radiusProfileDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *radiusProfileDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_radius_profile", req.ProviderTypeName)
}

func (d *radiusProfileDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_radius_profile` data source can be used to retrieve the ID for a RADIUS profile by name.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the RADIUS profile to look up. Defaults to `Default`.",
				Optional:            true,
			},
		},
	}
}

func (d *radiusProfileDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state radiusProfileDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	profiles, err := d.client.ListRADIUSProfile(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list RADIUS profiles", err.Error())
		return
	}

	name := state.Name.ValueString()
	if name == "" {
		name = "Default"
	}
	for _, p := range profiles {
		if p.Name == name {
			state.SetID(p.ID)
			state.SetSite(site)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("RADIUS profile not found", fmt.Sprintf("RADIUS profile not found with name %q", name))
}
