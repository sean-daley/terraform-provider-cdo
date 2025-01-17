---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cdo_ios_device Data Source - cdo"
subcategory: ""
description: |-
  IOS data source
---

# cdo_ios_device (Data Source)

IOS data source



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The human-readable name of the device. This is the name displayed on the CDO Inventory page. Device names are unique across a CDO tenant.

### Read-Only

- `connector_name` (String) The name of the Secure Device Connector (SDC) that is used by CDO to communicate with the device.
- `host` (String) The host used to connect to the device.
- `id` (String) Universally unique identifier of the device.
- `ignore_certificate` (Boolean) This attribute indicates whether certificates were ignored when onboarding this device.
- `port` (Number) The port used to connect to the device.
- `socket_address` (String) The address of the device to onboard, specified in the format `host:port`.
