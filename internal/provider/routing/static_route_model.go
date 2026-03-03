package routing

import (
	"context"
	"fmt"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ base.ResourceModel = &staticRouteModel{}

type staticRouteModel struct {
	base.Model
	Name      types.String `tfsdk:"name"`
	Network   types.String `tfsdk:"network"`
	Type      types.String `tfsdk:"type"`
	Distance  types.Int64  `tfsdk:"distance"`
	NextHop   types.String `tfsdk:"next_hop"`
	Interface types.String `tfsdk:"interface"`
}

func (m *staticRouteModel) AsUnifiModel(_ context.Context) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	t := m.Type.ValueString()

	network, err := utils.CidrZeroBased(m.Network.ValueString())
	if err != nil {
		diags.AddError("Invalid route network CIDR", err.Error())
		return nil, diags
	}

	r := &unifi.Routing{
		ID:                  m.ID.ValueString(),
		Enabled:             true,
		Type:                "static-route",
		Name:                m.Name.ValueString(),
		StaticRouteNetwork:  network,
		StaticRouteDistance: int(m.Distance.ValueInt64()),
		StaticRouteType:     t,
	}

	switch t {
	case "interface-route":
		r.StaticRouteInterface = m.Interface.ValueString()
	case "nexthop-route":
		r.StaticRouteNexthop = m.NextHop.ValueString()
	case "blackhole":
	default:
		diags.AddError("Invalid route type", fmt.Sprintf("Unexpected route type: %q", t))
	}

	return r, diags
}

func (m *staticRouteModel) Merge(_ context.Context, i interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}
	other, ok := i.(*unifi.Routing)
	if !ok {
		diags.AddError("Invalid model type", "Expected *unifi.Routing")
		return diags
	}
	m.ID = types.StringValue(other.ID)
	m.Name = types.StringValue(other.Name)
	network, err := utils.CidrZeroBased(other.StaticRouteNetwork)
	if err != nil {
		diags.AddError("Invalid route network from controller", err.Error())
		return diags
	}
	m.Network = types.StringValue(network)
	m.Distance = types.Int64Value(int64(other.StaticRouteDistance))
	m.Type = types.StringValue(other.StaticRouteType)

	switch other.StaticRouteType {
	case "interface-route":
		m.Interface = types.StringValue(other.StaticRouteInterface)
		m.NextHop = types.StringNull()
	case "nexthop-route":
		m.NextHop = types.StringValue(other.StaticRouteNexthop)
		m.Interface = types.StringNull()
	case "blackhole":
		m.NextHop = types.StringNull()
		m.Interface = types.StringNull()
	default:
		m.NextHop = ut.StringOrNull(other.StaticRouteNexthop)
		m.Interface = ut.StringOrNull(other.StaticRouteInterface)
	}
	return diags
}
