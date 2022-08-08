#!/bin/bash
cat << EOF > ~/.terraformrc
provider_installation {
  dev_overrides {
    "cloudfoundry" = "$HOME/.terraform.d/plugins/linux_amd64"
  }

  # all the other providers, install them as usual
  direct {}
}
EOF
