## 0.10.0  (Unreleased)

ENHANCEMENTS:
 * `cloudfoundry_app.route.default_route` deprecated in favor or `cloudfoundry_app.routes` array [#150](https://github.com/mevansam/terraform-provider-cf/pull/150) following cleaned up of blue/green unimplemented support. Thanks [@jcarrothers-sap](https://github.com/jcarrothers-sap).
 * Migration script to migrate from 0.9.8 to 0.9.9 syntax. [#165](https://github.com/mevansam/terraform-provider-cf/pull/165), [#171](https://github.com/mevansam/terraform-provider-cf/pull/171). Thanks [@janosbinder](https://github.com/janosbinder), [@lixilin2301](https://github.com/lixilin2301)
 * Added a flag support the recursive deletion of service bindings, service keys, and routes associated with the service instance. [#174](https://github.com/mevansam/terraform-provider-cf/pull/174). Thanks [@samedguener](https://github.com/samedguener), [@lixilin2301](https://github.com/lixilin2301)


BUG FIXES:
  * changes to app health check required an app restart [#168](https://github.com/mevansam/terraform-provider-cf/pull/168). Thanks [@jcarrothers-sap](https://github.com/jcarrothers-sap)
  * fix a crash when using `syslog_drain_url` services [#175](https://github.com/mevansam/terraform-provider-cf/pull/175). Thanks [@psycofdj](https://github.com/psycofdj)
  * fix git http(s) authentication on cloudfoundry_app resource [#169](https://github.com/mevansam/terraform-provider-cf/pull/169). Thanks [@psycofdj](https://github.com/psycofdj)
  * fix edge case crash with null app port [#159](https://github.com/mevansam/terraform-provider-cf/pull/159). Thanks [@loafoe](https://github.com/loafoe)

## 0.9.9 (September 27, 2018)

BREAKING CHANGES:

* the provider and all its resources have been renamed from "cf" to "cloudfoundry" [#44](https://github.com/mevansam/terraform-provider-cf/issues/44) thanks [@mevansam](https://github.com/mevansam) . Rename the provider distribution binary into terraform-provider-cloudfoundry when installing it. Refer to Using the provider readme for further details. Upgrading to this version requires the following procedure:
    * back up the tf files and the terraform.tfstate file
    * in all tf files replace resource/data "cf_..." with "cloudfoundry_...". For example `sed 's#resource \"cf_#resource \"cloudfoundry_#g' old.tf | sed 's#data \"cf_#data \"cloudfoundry_#g`
    * in the `terraform.tfstate` file replace the types. For example `"type": "cf_service_instance"` should be replaced with `"type": "cloudfoundry_service_instance"`
    * in the `terraform.tfstate` file check the depends_on nodes and make the adjustments.
        * For example:
        ```
        "depends_on": [
        "data.cf_service.abc"
        "cf_service_instance.xyz"
        "cf_route.example"
        ],
        ```
        will become:
        ```
        "depends_on": [
        "data.cloudfoundry_service.abc"
        "cloudfoundry_service_instance.xyz"
        "cloudfoundry_route.example"
        ],
        ```

if there are no errors after calling terraform plan, and no changes are planned to be applied then the migration was successful.

* `cloudfoundry_app.github_release` and `cloudfoundry_buildpack.github_release` now accept `user` and `password` attribute instead of Oauth access token attribute [#119](https://github.com/mevansam/terraform-provider-cf/issues/119) thanks  [@janosbinder](https://github.com/janosbinder) and
[@SzucsAti](https://github.com/SzucsAti)
* `cloudfoundry_config` was renamed into cloudfoundry_feature_flags [#6](https://github.com/mevansam/terraform-provider-cf/issues/#6) [#66](https://github.com/mevansam/terraform-provider-cf/issues/66) thanks [@SzucsAti](https://github.com/mevansam)
* `cloudfoundry_quota` was split into `cloudfoundry_org_quota` and `cloudfoundry_space_quota` thanks [@psycofdj](https://github.com/psycofdj)

IMPROVEMENTS:

* cli: display workspace name in apply and destroy commands if not default [#18253](https://github.com/hashicorp/terraform/issues/18253)
* cli: Remove error on empty outputs when `-json` is set [#11721](https://github.com/hashicorp/terraform/issues/11721)
* helper/schema: Resources have a new `DeprecationMessage` property that can be set to a string, allowing full resources to be deprecated [#18286](https://github.com/hashicorp/terraform/issues/18286)
* backend/s3: Allow fallback to session-derived credentials (e.g. session via `AWS_PROFILE` environment variable and shared configuration) [#17901](https://github.com/hashicorp/terraform/issues/17901)
* backend/s3: Allow usage of `AWS_EC2_METADATA_DISABLED` environment variable [#17901](https://github.com/hashicorp/terraform/issues/17901)

FEATURES:

* `cloudfoundry_route_service_binding` support [#16](https://github.com/mevansam/terraform-provider-cf/issues/16) thanks [@psycofdj](https://github.com/psycofdj)
* All resources can now be imported [#6](https://github.com/mevansam/terraform-provider-cf/issues/#6) thanks [@ArthurHlt](https://github.com/ArthurHlt)
* Asynchronous provisioning/deprovisioning and update of cf_service_instance ([#51](https://github.com/mevansam/terraform-provider-cf/issues/51) thanks [@samedguener](https://github.com/samedguener)
* `cloudfoundry_app` docker support [#43](https://github.com/mevansam/terraform-provider-cf/issues/#43) [#84](https://github.com/mevansam/terraform-provider-cf/issues/#84) thanks [@samedguener](https://github.com/samedguener) [@doktorgibson](https://github.com/doktorgibson) and [@janosbinder](https://github.com/janosbinder)

BUG FIXES:

* Documentation spell check and fixes [#55](https://github.com/mevansam/terraform-provider-cf/issues/55) [#6](https://github.com/mevansam/terraform-provider-cf/issues/#6) [#66](https://github.com/mevansam/terraform-provider-cf/issues/66) thanks [@psycofdj](https://github.com/psycofdj) [@gberche-orange](https://github.com/gberche-orange) and [#112](https://github.com/mevansam/terraform-provider-cf/issues/#112)  [@lixilin2301](https://github.com/lixilin2301)

* Optional resource attributes now have a documented default [#59](https://github.com/mevansam/terraform-provider-cf/issues/59) thanks [@psycofdj](https://github.com/psycofdj)
* Fixed `cloudfoundry_app` temp file leaks [#57](https://github.com/mevansam/terraform-provider-cf/issues/57) thanks [@ArthurHlt](https://github.com/ArthurHlt) and [@SzucsAti](https://github.com/SzucsAti)
* Crash during `cloudfoundry_app` async app downloading [#88](https://github.com/mevansam/terraform-provider-cf/issues/88) thanks [@jcarrothers-sap](https://github.com/jcarrothers-sap) and [@janosbinder](https://github.com/janosbinder) and app restage wait
* Many reliability w.r.t. missing resources on routes, service_instances, service_key [#131](https://github.com/mevansam/terraform-provider-cf/issues/131) [#138](https://github.com/mevansam/terraform-provider-cf/issues/138) [#137](https://github.com/mevansam/terraform-provider-cf/issues/137) [#133](https://github.com/mevansam/terraform-provider-cf/issues/133) [#135](https://github.com/mevansam/terraform-provider-cf/issues/135) [#134](https://github.com/mevansam/terraform-provider-cf/issues/134) [#132](https://github.com/mevansam/terraform-provider-cf/issues/132) thanks [@jcarrothers-sap](https://github.com/jcarrothers-sap)

NOTES:

* CI improvements,
    * including cleanups of leaked resource (thanks [@janosbinder](https://github.com/janosbinder)
    * linter automation [#53](https://github.com/mevansam/terraform-provider-cf/issues/53) thanks [@psycofdj](https://github.com/psycofdj)
* Acceptance test environment maintenance and upgrades (thanks [@mevansam](https://github.com/mevansam)
* improvements following [#38](https://github.com/mevansam/terraform-provider-cf/issues/38)
    * Idomatic go style: Naked return [#47](https://github.com/mevansam/terraform-provider-cf/issues/47) thanks [@psycofdj](https://github.com/psycofdj)
    * run `goimports` on the code + prune unused vendor libs (#48 thanks [@ArthurHlt](https://github.com/ArthurHlt)
    * Missing error checks [#49](https://github.com/mevansam/terraform-provider-cf/issues/49) thanks [@ArthurHlt](https://github.com/ArthurHlt)
    * Clean ups: [#45](https://github.com/mevansam/terraform-provider-cf/issues/45), [#46](https://github.com/mevansam/terraform-provider-cf/issues/46)

<!-- Local Variables: -->
<!-- End: -->
