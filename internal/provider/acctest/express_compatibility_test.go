package acctest

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExpressCompatibility_ControllerVersion(t *testing.T) {
	v, err := version.NewVersion(testClient.Version())
	if err != nil {
		t.Fatalf("failed to parse controller version: %s", err)
	}
	minVersion := version.Must(version.NewVersion("8.0.0"))
	maxVersion := version.Must(version.NewVersion("9.0.0"))

	if v.LessThan(minVersion) || v.GreaterThanOrEqual(maxVersion) {
		t.Fatalf("expected controller version 8.x, got %s", v)
	}
	t.Logf("controller version: %s (OK)", v)
}

func TestAccExpressCompatibility_NetworkCRUD(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: Steps{
			{
				Config: `
resource "unifi_network" "express_test" {
  name    = "express-compat-test"
  purpose = "corporate"

  subnet       = "10.99.0.0/24"
  vlan_id      = 99
  dhcp_start   = "10.99.0.6"
  dhcp_stop    = "10.99.0.254"
  dhcp_enabled = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_network.express_test", "name", "express-compat-test"),
					resource.TestCheckResourceAttr("unifi_network.express_test", "purpose", "corporate"),
				),
			},
		},
	})
}

func TestAccExpressCompatibility_WLAN(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: Steps{
			{
				Config: `
resource "unifi_network" "express_wlan_test" {
  name    = "express-wlan-net"
  purpose = "corporate"

  subnet       = "10.98.0.0/24"
  vlan_id      = 98
  dhcp_start   = "10.98.0.6"
  dhcp_stop    = "10.98.0.254"
  dhcp_enabled = true
}

resource "unifi_wlan" "express_test" {
  name       = "express-compat-wlan"
  passphrase = "expresstestpass123"
  security   = "wpapsk"

  network_id = unifi_network.express_wlan_test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_wlan.express_test", "name", "express-compat-wlan"),
					resource.TestCheckResourceAttr("unifi_wlan.express_test", "security", "wpapsk"),
				),
			},
		},
	})
}

func TestAccExpressCompatibility_FirewallRule(t *testing.T) {
	AcceptanceTest(t, AcceptanceTestCase{
		Steps: Steps{
			{
				Config: `
resource "unifi_firewall_group" "express_test" {
  name    = "express-compat-fw-group"
  type    = "address-group"
  members = ["10.0.0.1"]
}

resource "unifi_firewall_rule" "express_test" {
  name    = "express-compat-fw-rule"
  action  = "drop"
  ruleset = "LAN_IN"

  rule_index = 2010

  protocol = "all"

  dst_firewall_group_ids = [unifi_firewall_group.express_test.id]
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("unifi_firewall_rule.express_test", "name", "express-compat-fw-rule"),
					resource.TestCheckResourceAttr("unifi_firewall_rule.express_test", "action", "drop"),
				),
			},
		},
	})
}
