package user

import (
	"context"

	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	ut "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
)

var (
	_ resource.Resource                = &userGroupResource{}
	_ resource.ResourceWithConfigure   = &userGroupResource{}
	_ resource.ResourceWithImportState = &userGroupResource{}
	_ base.Resource                    = &userGroupResource{}
)

type userGroupResource struct {
	*base.GenericResource[*userGroupModel]
}

func NewUserGroupResource() resource.Resource {
	return &userGroupResource{
		GenericResource: base.NewGenericResource(
			"unifi_user_group",
			func() *userGroupModel { return &userGroupModel{} },
			base.ResourceFunctions{
				Read: func(ctx context.Context, client *base.Client, site, id string) (interface{}, error) {
					return client.GetUserGroup(ctx, site, id)
				},
				Create: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.CreateUserGroup(ctx, site, body.(*unifi.UserGroup))
				},
				Update: func(ctx context.Context, client *base.Client, site string, body interface{}) (interface{}, error) {
					return client.UpdateUserGroup(ctx, site, body.(*unifi.UserGroup))
				},
				Delete: func(ctx context.Context, client *base.Client, site, id string) error {
					return client.DeleteUserGroup(ctx, site, id)
				},
			},
		),
	}
}

func (r *userGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `unifi_user_group` resource manages client groups in the UniFi controller, which allow you to apply " +
			"common settings and restrictions to multiple network clients.\n\n" +
			"User groups are primarily used for:\n" +
			"  * Implementing Quality of Service (QoS) policies\n" +
			"  * Setting bandwidth limits for different types of users\n" +
			"  * Organizing clients into logical groups (e.g., Staff, Guests, IoT devices)",

		Attributes: map[string]schema.Attribute{
			"id":   ut.ID(),
			"site": ut.SiteAttribute(),
			"name": schema.StringAttribute{
				MarkdownDescription: "A descriptive name for the user group (e.g., 'Staff', 'Guests', 'IoT Devices').",
				Required:            true,
			},
			"qos_rate_max_down": schema.Int64Attribute{
				MarkdownDescription: "The maximum allowed download speed in Kbps for clients in this group. Set to -1 for unlimited.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(-1),
			},
			"qos_rate_max_up": schema.Int64Attribute{
				MarkdownDescription: "The maximum allowed upload speed in Kbps for clients in this group. Set to -1 for unlimited.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(-1),
			},
		},
	}
}
