
The backlog is kept as github issues annotated with labels and milestones, and progress is tracked in the [backlog project](https://github.com/mevansam/terraform-provider-cf/projects/1).

Please send PR against the `dev` branches.

You may reach core contributors in the [cloudfoundry slack, within the terraform channel](https://cloudfoundry.slack.com/messages/C7JRBR8CV/) (get an account from http://slack.cloudfoundry.org/). Please prefer slack channel for support requests, and github issues for qualified bugs and enhancements requests.

## Compliance

Linters are run as part of the travis build. Detected flaws will prevent the merge of the PR.

Contributors can run the same checks locally by following the procedure:

1. install dependencies 

  ```bash
  # install gometalinter tool
  go get gopkg.in/alecthomas/gometalinter.v2
  # install gometalinter internal dependencies
  gometalinter.v2 --install
  ```

2. Run the linters

  ```bash
  make check
  ```
  
Fine tuning of the linters configuration can be done in the [.gometalinter.json](.gometalinter.json) file according to the tool [specification](https://github.com/alecthomas/gometalinter#configuration-file)

## Creating a release 

* Open up an issue "cutting release 0.9.9" to gather contributors concensus on when to cut the release
* On your clone, checkout the `dev` branch, and execute `scripts/create-release.sh 0.9.9`
* travis build kicks off for this tag, and tries to publish the artifacts github, check it on [travis the branch list](https://travis-ci.org/mevansam/terraform-provider-cf/branches)
* edit the release notes in the [github release page](https://github.com/mevansam/terraform-provider-cf/releases)
