---
page_title: "Provider: CDO"
description: |-
  The Cisco Defense Orchestrator provider is used to manage devices and other security resources on Cisco Defense Orchestrator using Terraform.
---

# CDO Provider

Use the Cisco Defense Orchestrator (CDO) provider to onboard and manage the many devices and other resources supported by CDO. 

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `base_url` (String) The base CDO URL. This is the URL you enter when logging into your CDO account.

### Optional

- `api_token` (String, Sensitive) The API token used to authenticate with CDO. [See here](https://docs.defenseorchestrator.com/c_api-tokens.html#!t-generatean-api-token.html) to learn how to generate an API token.