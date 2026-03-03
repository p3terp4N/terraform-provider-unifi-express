package device

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
	_ datasource.DataSource              = &portProfileDatasource{}
	_ datasource.DataSourceWithConfigure = &portProfileDatasource{}
	_ base.Resource                      = &portProfileDatasource{}
)

type portProfileDatasource struct {
	base.ControllerVersionValidator
	base.FeatureValidator
	client *base.Client
}

type portProfileDatasourceModel struct {
	base.Model
	Name types.String `tfsdk:"name"`
}

func NewPortProfileDatasource() datasource.DataSource {
	return &portProfileDatasource{}
}

func (d *portProfileDatasource) SetClient(client *base.Client) {
	d.client = client
}

func (d *portProfileDatasource) SetVersionValidator(validator base.ControllerVersionValidator) {
	d.ControllerVersionValidator = validator
}

func (d *portProfileDatasource) SetFeatureValidator(validator base.FeatureValidator) {
	d.FeatureValidator = validator
}

func (d *portProfileDatasource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	base.ConfigureDatasource(d, req, resp)
}

func (d *portProfileDatasource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_port_profile", req.ProviderTypeName)
}

func (d *portProfileDatasource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "`unifi_port_profile` data source can be used to retrieve port profile configurations " +
			"from your UniFi network by name.",
		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the port profile to look up. Defaults to `All`.",
				Optional:            true,
			},
		},
	}
}

func (d *portProfileDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state portProfileDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	site := d.client.ResolveSite(&state)

	profiles, err := d.client.ListPortProfile(ctx, site)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list port profiles", err.Error())
		return
	}

	name := state.Name.ValueString()
	if name == "" {
		name = "All"
	}
	for _, p := range profiles {
		if p.Name == name {
			state.SetID(p.ID)
			state.SetSite(site)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("Port profile not found", fmt.Sprintf("Port profile not found with name %q", name))
}
