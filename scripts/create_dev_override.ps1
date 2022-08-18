cp terraform-provider-cloudfoundry.exe $env:APPDATA\terraform.d\plugins\
$data=@"
provider_installation {
  dev_overrides {
    "cloudfoundry" = "$env:APPDATA\terraform.d\plugins\terraform-provider-cloudfoundry.exe"
  }

  # all the other providers, install them as usual
  direct {}
}
"@
echo $data |Out-File $env:APPDATA\terraform.rc