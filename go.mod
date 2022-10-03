module github.com/terraform-providers/terraform-provider-cloudfoundry

go 1.17

replace code.cloudfoundry.org/cli => github.wdf.sap.corp/intelligent-rpa/cloudfoundry-cli v1.18.6

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20180906201452-2aa6f33b730c // indirect
	code.cloudfoundry.org/cfnetworking-cli-api v0.0.0-20190103195135-4b04f26287a6
	code.cloudfoundry.org/cli v0.0.0-00010101000000-000000000000
	github.com/ArthurHlt/zipper v1.3.2
	github.com/SermoDigital/jose v0.9.2-0.20161205224733-f6df55f235c2 // indirect
	github.com/blang/semver v3.5.1+incompatible
	github.com/bmatcuk/doublestar v1.3.1 // indirect
	github.com/bmizerany/pat v0.0.0-20170815010413-6226ea591a40 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3
	github.com/charlievieth/fs v0.0.0-20170613215519-7dc373669fa1 // indirect
	github.com/cloudfoundry/bosh-cli v5.5.1+incompatible // indirect
	github.com/cloudfoundry/bosh-utils v0.0.0-20190518100210-9f9df32d41c3 // indirect
	github.com/cloudfoundry/noaa v2.1.1-0.20190110210640-5ce49363dfa6+incompatible
	github.com/cloudfoundry/sonde-go v0.0.0-20171206171820-b33733203bb4
	github.com/cppforlife/go-patch v0.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/hashicorp/go-getter v1.6.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/terraform-json v0.14.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.18.0
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/tedsuo/rata v1.0.1-0.20170830210128-07d200713958 // indirect
	github.com/vito/go-interact v1.0.0 // indirect
	github.com/zclconf/go-cty v1.10.0 // indirect
	golang.org/x/text v0.3.7 // indirect
)

require github.com/hashicorp/terraform-plugin-log v0.4.1

require (
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f // indirect
	code.cloudfoundry.org/jsonry v1.1.3 // indirect
	code.cloudfoundry.org/tlsconfig v0.0.0-20210615191307-5d92ef3894a7 // indirect
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elazarl/goproxy v0.0.0-20220529153421-8ea89ba92021 // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20220529153421-8ea89ba92021 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320 // indirect
	github.com/hashicorp/go-hclog v1.2.1 // indirect
	github.com/hashicorp/go-plugin v1.4.4 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/hc-install v0.4.0 // indirect
	github.com/hashicorp/hcl/v2 v2.13.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.17.2 // indirect
	github.com/hashicorp/terraform-plugin-go v0.10.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.0.0-20220623143253-7d51757b572c // indirect
	github.com/hashicorp/terraform-svchost v0.0.0-20200729002733-f050f53b9734 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.5-0.20181218000649-703b5e6b11ae // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.1 // indirect
	github.com/whilp/git-urls v0.0.0-20160530060445-31bac0d230fa // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	golang.org/x/crypto v0.0.0-20220517005047-85d78b3ac167 // indirect
	golang.org/x/net v0.0.0-20220531201128-c960675eff93 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220601144221-27df5f98adab // indirect
	google.golang.org/grpc v1.47.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
