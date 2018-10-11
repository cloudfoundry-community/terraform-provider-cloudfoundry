## 0.9.9 (September 27, 2018)

BREAKING CHANGES:

* the provider and all its resources have been renamed from "cf" to "cloudfoundry" (#44 thanks @mevansam ). Rename the provider distribution binary into terraform-provider-cloudfoundry when installing it. Refer to Using the provider readme for further details. Upgrading to this version requires the following procedure:
    * back up the tf files and the terraform.tfstate file
    * in all tf files replace resource/data "cf_..." with "cloudfoundry_...". For example sed 's#resource \"cf_#resource \"cloudfoundry_#g' old.tf | sed 's#data \"cf_#data \"cloudfoundry_#g'
    * in the terraform.tfstate file replace the types. For example "type": "cf_service_instance" should be replaced with "type": "cloudfoundry_service_instance"
    * in the terraform.tfstate file check the depends_on nodes and make the adjustments.
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

* cloudfoundry_app.github_release and cloudfoundry_buildpack.github_release now accept user and password attribute instead of Oauth access token attribute (#119 thanks @janosbinder and @SzucsAti)
* cloudfoundry_config was renamed into cloudfoundry_feature_flags (#66 thanks @mevansam )
* cloudfoundry_quota was split into cloudfoundry_org_quota and cloudfoundry_space_quota (thanks @psycofdj)

IMPROVEMENTS:

* cli: display workspace name in apply and destroy commands if not default ([#18253](https://github.com/hashicorp/terraform/issues/18253))
* cli: Remove error on empty outputs when `-json` is set ([#11721](https://github.com/hashicorp/terraform/issues/11721))
* helper/schema: Resources have a new `DeprecationMessage` property that can be set to a string, allowing full resources to be deprecated ([#18286](https://github.com/hashicorp/terraform/issues/18286))
* backend/s3: Allow fallback to session-derived credentials (e.g. session via `AWS_PROFILE` environment variable and shared configuration) ([#17901](https://github.com/hashicorp/terraform/issues/17901))
* backend/s3: Allow usage of `AWS_EC2_METADATA_DISABLED` environment variable ([#17901](https://github.com/hashicorp/terraform/issues/17901))

FEATURES:

* cloudfoundry_route_service_bindingsupport (#16 thanks @psycofdj )
* All resources can now be imported (#6 thanks @ArthurHlt )
* Asynchronous provisioning/deprovisioning and update of cf_service_instance (#51 thanks @samedguener)
* cloudfoundry_app docker support (#43 #84 thanks @samedguener @doktorgibson and @janosbinder )

BUG FIXES:

* Documentation spell check and fixes (#55 #66 thanks @psycofdj @gberche-orange and #112 @lixilin2301)
* Optional resource attributes now have a documented default (#59 thanks @psycofdj)
* Fixed cloudfoundry_app temp file leaks (#57 thanks @ArthurHlt and @SzucsAti)
* Crash during cloudfoundry_app async app downloading (#88 thanks @jcarrothers-sap and @janosbinder) and app restage wait
* Many reliability w.r.t. missing resources on routes, service_instances, service_key (#131 #138 #137 #133 #135 #134 #132 thanks @jcarrothers-sap)

NOTES:

* CI improvements,
    * including cleanups of leaked resource (thanks @janosbinder)
    * linter automation (#53 thanks @psycofdj)
* Acceptance test environment maintenance and upgrades (thanks @mevansam)
* improvements following #38
    * Idomatic go style: Naked return (#47 thanks @psycofdj)
    * run goimports on the code + prune unused vendor libs (#48 thanks @ArthurHlt)
    * Missing error checks (#49 thanks @ArthurHlt)
    * Clean ups: #45, #46