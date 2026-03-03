package site

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type siteModel struct {
	ID          types.String `tfsdk:"id"`
	Site        types.String `tfsdk:"site"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
}

func (m *siteModel) GetID() string         { return m.ID.ValueString() }
func (m *siteModel) SetID(id string)       { m.ID = types.StringValue(id) }
func (m *siteModel) GetRawID() types.String { return m.ID }

func (m *siteModel) GetSite() string          { return m.Site.ValueString() }
func (m *siteModel) SetSite(site string)      { m.Site = types.StringValue(site) }
func (m *siteModel) GetRawSite() types.String { return m.Site }

func (m *siteModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	return &unifi.Site{
		ID:          m.ID.ValueString(),
		Name:        m.Name.ValueString(),
		Description: m.Description.ValueString(),
	}, diag.Diagnostics{}
}

func (m *siteModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.Site)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.Site")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	m.Description = types.StringValue(other.Description)
	return diags
}
