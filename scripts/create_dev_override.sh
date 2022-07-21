#!/bin/bash
cat << EOF > ~/.terraformrc
provider_installation {
  dev_overrides {
    "cloudfoundry" = "$HOME/.terraform.d/plugins/linux_amd64/terraform-provider-cloudfoundry"
  }

  # all the other providers, install them as usual
  direct {}
}
EOF