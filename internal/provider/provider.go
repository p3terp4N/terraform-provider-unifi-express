package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/apgroup"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/base"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/device"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/dns"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/firewall"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/network"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/radius"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/routing"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/settings"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/site"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/user"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"strings"
)

const (
	ProviderUsernameDescription = "Local user name for the Unifi controller API. Can be specified with the `UNIFI_USERNAME` environment variable."
	ProviderPasswordDescription = "Password for the user accessing the API. Can be specified with the `UNIFI_PASSWORD` environment variable."
	ProviderAPIURLDescription = "URL of the UniFi Express controller API. Can be specified with the `UNIFI_API` environment variable. " +
		"Typically `https://<express-ip>:8443`. You should **NOT** supply the path (`/api`), the SDK will discover the appropriate paths."
	ProviderSiteDescription = "The site in the UniFi Express controller this provider will manage. Can be specified with the `UNIFI_SITE` environment variable. Default: `default`"
	ProviderAllowInsecureDescription = "Skip verification of TLS certificates of API requests. You may need to set this to `true` " +
		"if you are using your local API without setting up a signed certificate. Can be specified with the " +
		"`UNIFI_INSECURE` environment variable."
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown

	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		if s.Deprecated != "" {
			desc += " " + s.Deprecated
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"username": {
					Description: ProviderUsernameDescription,
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_USERNAME", ""),
				},
				"password": {
					Description: ProviderPasswordDescription,
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_PASSWORD", ""),
				},
				"api_url": {
					Description: ProviderAPIURLDescription,
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_API", ""),
				},
				"site": {
					Description: ProviderSiteDescription,
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_SITE", "default"),
				},
				"allow_insecure": {
					Description: ProviderAllowInsecureDescription,
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("UNIFI_INSECURE", false),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"unifi_ap_group":       apgroup.DataAPGroup(),
				"unifi_network":        network.DataNetwork(),
				"unifi_port_profile":   device.DataPortProfile(),
				"unifi_radius_profile": radius.DataRADIUSProfile(),
				"unifi_user_group":     user.DataUserGroup(),
				"unifi_user":           user.DataUser(),
				"unifi_account":        radius.DataAccount(),
			},
			ResourcesMap: map[string]*schema.Resource{
				// TODO: "unifi_ap_group"
				"unifi_device":         device.ResourceDevice(),
				"unifi_dynamic_dns":    dns.ResourceDynamicDNS(),
				"unifi_firewall_group": firewall.ResourceFirewallGroup(),
				"unifi_firewall_rule":  firewall.ResourceFirewallRule(),
				"unifi_network":        network.ResourceNetwork(),
				"unifi_port_forward":   routing.ResourcePortForward(),
				"unifi_static_route":   routing.ResourceStaticRoute(),
				"unifi_wlan":           network.ResourceWLAN(),
				"unifi_port_profile":   device.ResourcePortProfile(),
				"unifi_site":           site.ResourceSite(),
				"unifi_account":        radius.ResourceAccount(),
				"unifi_radius_profile": radius.ResourceRadiusProfile(),
				"unifi_setting_radius": settings.ResourceSettingRadius(),
				"unifi_user_group":     user.ResourceUserGroup(),
				"unifi_user":           user.ResourceUser(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func createHTTPTransport(insecure bool, subsystem string) http.RoundTripper {
	transport := base.CreateHttpTransport(insecure)
	t := logging.NewSubsystemLoggingHTTPTransport(subsystem, transport)
	return t
}

func configure(v string, p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		user := d.Get("username").(string)
		pass := d.Get("password").(string)
		if user == "" || pass == "" {
			return nil, diag.FromErr(errors.New("`username` and `password` must both be set"))
		}
		baseURL := d.Get("api_url").(string)
		site := d.Get("site").(string)
		insecure := d.Get("allow_insecure").(bool)

		c, err := base.NewClient(&base.ClientConfig{
			Username: user,
			Password: pass,
			Url:      baseURL,
			Site:     site,
			HttpConfigurer: func() http.RoundTripper {
				return createHTTPTransport(insecure, "unifi")
			},
		})
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, nil
	}
}
