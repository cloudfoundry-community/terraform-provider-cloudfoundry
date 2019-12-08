Migration script for the Cloud Foundry Terraform Provider
=========================================================

Overview
--------

This migration script helps to migrate the tfstate and the tf files before <0.9.9 to make it compatible with the recent version.

Requirements
------------

-	[Python](https://www.python.org) >= 3.6 

Steps to migrate
---------------------

Breaking changes:
   * the provider and all its resources have been renamed from "cf" to "cloudfoundry". **Rename the provider distribution binary into `terraform-provider-cloudfoundry`** when installing it. Refer to [Using the provider](https://github.com/mevansam/terraform-provider-cf#using-the-provider) readme for further details. Upgrading to this version requires the following procedure:
   * back up the tf files and the `terraform.tfstate` file
   * launch the migration scripts        
```
find PATH/TO/TF_FILES/. -type f -name "*.tf" -exec python3 migrate_tf_cf.py -t tf -d -n {} \;
find PATH/TO/TF_STATE_FILES/. -type f -name "*.tfstate" -exec python3 migrate_tf_cf.py -t state -d -n {} \;
```

   * run again `terraform init` to load the new version of the provider
   * run `terraform plan` to see whether everything has been migrated. Ideally none of the artefacts should be deployed again.