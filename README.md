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
  username  = "TDH_USERNAME"
  password  = "TDH_PASSWORD"
  org_id    = "TDH_ORG_ID"
}
```
## <a name="offline-use"></a> Use in an isolated environment

To use the provider without installing from Terraform's hosted registry, you can use pre-built binaries under releases. Binaries are built for different OS & architecture, follow the steps for each:

### Steps
1. Find the relevant archive under releases with name:
   - For Mac: `terraform-provider-tdh_<version>_darwin_<arch>.zip`
   - For Linux: `terraform-provider-tdh_<version>_linux_<arch>.zip`
   - For Windows: `terraform-provider-tdh_<version>_windows_<arch>.zip`
2. Download and extract it to some place on your machine
3. Onwards, the steps has to be followed on machine where terraform is installed and supposed to use TDH provider. 
4. Copy the executable file (to some **\<PATH>**):
   - For Mac & Linux, named `terraform-provider-tdh_<version>` to a common place for your executables like `/some/path/bin`
   - For Windows, named `terraform-provider-tdh_<version>.exe` to a place for this executable like `C:\Programs\TDH`
   - Make sure executable permission is granted on this file
5. Create a CLI config file
   - For Mac, named `.terraformrc` in your home directory `/Users/<username>/`
   - For Linux, named `.terraformrc` in your home directory `/home/<username>/`
   - For Windows, named `terraform.rc` in your `%APPDATA%` directory that is specific to your system
6. Put the following content in the config file created in previous step
    ```hcl
    provider_installation {
    
      dev_overrides {
          "hashicorp.com/svc-bot-mds/tdh" = "<PATH>" # CHANGE IT as per OS in step #4
      }
    
      # For all other providers, install them directly from their origin provider
      # registries as normal. If you omit this, Terraform will _only_ use
      # the dev_overrides block, and so no other providers will be available.
      direct {}
    }
    ```
7. Now, you just have to change the source in general TDH provider configuration like so:
    ```hcl
    terraform {
      required_providers {
        tdh = {
          source = "hashicorp.com/svc-bot-mds/tdh" # Changed from 'svc-bot-mds/tdh' to 'hashicorp.com/svc-bot-mds/tdh'
        }
      }
    }
    ```


## Development

Run `make hooks` after cloning or before making any change/commit.
<br>If there is any change in resource/datasource .go files, make sure to run `make generate`.