# Terraform Provider for Tanzu Data Hub

## About

This repository contains code for the Terraform Provider for Tanzu Data Hub. It supports provisioning of Clusters/Instances of Services (currently only Postgres, MySQL, RabbitMQ & Redis) and access management of Users on those resources.

## Configuration

The Terraform Provider for TDH is available via the Terraform Registry: [svc-bot-mds/tdh](https://registry.terraform.io/providers/svc-bot-mds/tdh). To be able to use it successfully, please use below snippet to set up the provider:

```hcl
terraform {
  required_providers {
    tdh = {
      source = "svc-bot-mds/tdh"
    }
  }
}

provider "tdh" {
  host      = "https://tdh-console.example.com" # (required) the URL of hosted TDH
  api_token = "XXXXXX__API_TOKEN__XXXXXX"       # (required) can be generated from CSP > Accounts page
}
```
