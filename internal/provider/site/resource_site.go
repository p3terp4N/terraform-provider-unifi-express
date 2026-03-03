package site

import (
	"context"
	"errors"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

var (
	_ resource.Resource                = &siteResource{}
	_ resource.ResourceWithConfigure   = &siteResource{}
	_ resource.ResourceWithImportState = &siteResource{}
	_ base.Resource                    = &siteResource{}
)

type siteResource struct {
	*base.GenericResource[*siteModel]
}

func NewSiteResource() resource.Resource {
	return &siteResource{
		GenericResource: base.NewGenericResource(
			"unifi_site",
			func() *siteModel { return &siteModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetSite(ctx, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					m := body.(*unifi.Site)
					resp, err := client.CreateSite(ctx, m.Description)
					if err != nil {
						return nil, err
					}
					if len(resp) == 0 {
						return nil, fmt.Errorf("CreateSite returned empty response")
					}
					return &resp[0], nil
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					m := body.(*unifi.Site)
					resp, err := client.UpdateSite(ctx, m.Name, m.Description)
					if err != nil {
						return nil, err
					}
					if len(resp) == 0 {
						return nil, fmt.Errorf("UpdateSite returned empty response")
					}
					return &resp[0], nil
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					_, err := client.DeleteSite(ctx, id)
					return err
				},
			},
		),
	}
}

func (r *siteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	client := r.GetClient()
	if client == nil {
		resp.Diagnostics.AddError("Client not configured", "The provider client is not configured")
		return
	}

	id := req.ID
	if id == "" {
		resp.Diagnostics.AddError("Invalid ID", "ID is required")
		return
	}

	// Try direct ID lookup first
	site, err := client.GetSite(ctx, id)
	if err == nil {
		state := &siteModel{}
		state.Merge(ctx, site)
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}
	if !errors.Is(err, unifi.ErrNotFound) {
		resp.Diagnostics.AddError("Error reading site", err.Error())
		return
	}

	// Fall back to name lookup
	sites, err := client.ListSites(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing sites", err.Error())
		return
	}

	for _, s := range sites {
		if s.Name == id {
			state := &siteModel{}
			state.Merge(ctx, &s)
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}

	resp.Diagnostics.AddError("Site not found", fmt.Sprintf("Unable to find site %q on controller", id))
}

func (r *siteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_site` resource manages UniFi sites, which are logical groupings of UniFi devices and their configurations.\n\n" +
			"Sites in UniFi are used to:\n" +
			"  * Organize network devices and settings for different physical locations\n" +
			"  * Isolate configurations between different networks or customers\n" +
			"  * Apply different policies and configurations to different groups of devices\n\n" +
			"Each site maintains its own:\n" +
			"  * Network configurations\n" +
			"  * Wireless networks (WLANs)\n" +
			"  * Security policies\n" +
			"  * Device configurations\n\n" +
			"A UniFi controller can manage multiple sites, making it ideal for multi-tenant or distributed network deployments.",

		Attributes: map[string]schema.Attribute{
			"id": ut.ID("The unique identifier of the site in the UniFi controller. This is automatically generated when the site is created."),
			"site": schema.StringAttribute{
				MarkdownDescription: "Not used for this resource. Sites are top-level objects.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A human-readable description of the site (e.g., 'Main Office', 'Remote Branch', 'Client A Network'). " +
					"This is used as the display name in the UniFi controller interface.",
				Required: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The site's internal name in the UniFi system. This is automatically generated based on the description " +
					"and is used in API calls and configurations. It's typically a lowercase, hyphenated version of the description.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
