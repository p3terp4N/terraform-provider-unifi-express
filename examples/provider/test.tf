# Phase 1: Authentication smoke test
data "unifi_ap_group" "default" {}

output "ap_group" {
  value = data.unifi_ap_group.default
}

# Phase 2: Read-only data sources
data "unifi_network" "default" {
  name = "Default"
}

output "default_network_id" {
  value = data.unifi_network.default.id
}

output "default_network_subnet" {
  value = data.unifi_network.default.subnet
}
