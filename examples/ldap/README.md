# Cloud Foundry Org and User Configuration

This example configures a local PCFDev environment with users queried from an LDAP directory. Along with the Cloud Foundry provider this example requires the [LDAP provider](https://github.com/mevansam/terraform-provider-ldap).

In order to run this example you will need to first launch the test LDAP server in a local Docker container via the `scripts/ldap-up.sh` script, which needs to be run form within the repository root. Then start an PCF Dev via `cf dev start`. Once the environment is run `cd` to this folder and run `terraform apply`. 
