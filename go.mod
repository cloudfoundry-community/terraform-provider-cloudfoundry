module github.com/terraform-providers/terraform-provider-cloudfoundry

go 1.12

replace code.cloudfoundry.org/cli => github.com/orange-cloudfoundry/cloudfoundry-cli v0.0.0-complete-api

require (
	code.cloudfoundry.org/bytefmt v0.0.0-20180906201452-2aa6f33b730c // indirect
	code.cloudfoundry.org/cli v6.44.0+incompatible
	code.cloudfoundry.org/diego-ssh v0.0.0-20190419150756-fd9db4fe28e9 // indirect
	code.cloudfoundry.org/gofileutils v0.0.0-20170111115228-4d0c80011a0f // indirect
	code.cloudfoundry.org/inigo v0.0.0-20190424163605-414d9ec47804 // indirect
	code.cloudfoundry.org/lager v2.0.0+incompatible // indirect
	code.cloudfoundry.org/ykk v0.0.0-20170424192843-e4df4ce2fd4d // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/SermoDigital/jose v0.0.0-20161205225155-2bd9b81ac51d // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/alcortesm/tgz v0.0.0-20161220082320-9c5fe88206d7 // indirect
	github.com/apoydence/eachers v0.0.0-20181020210610-23942921fe77 // indirect
	github.com/apparentlymart/go-cidr v1.0.0 // indirect
	github.com/apparentlymart/go-textseg v0.0.0-20170531203952-b836f5c4d331 // indirect
	github.com/armon/go-radix v0.0.0-20170727155443-1fca145dffbc // indirect
	github.com/bgentry/speakeasy v0.1.0 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bmatcuk/doublestar v1.1.1 // indirect
	github.com/bmizerany/pat v0.0.0-20170815010413-6226ea591a40 // indirect
	github.com/charlievieth/fs v0.0.0-20170613215519-7dc373669fa1 // indirect
	github.com/cloudfoundry/bosh-cli v5.5.0+incompatible // indirect
	github.com/cloudfoundry/bosh-utils v0.0.0-20190504100202-2830420c6f51 // indirect
	github.com/cloudfoundry/cli-plugin-repo v0.0.0-20190506165029-1123b00f89cb // indirect
	github.com/cloudfoundry/noaa v2.1.0+incompatible
	github.com/cloudfoundry/sonde-go v0.0.0-20171206171820-b33733203bb4
	github.com/cppforlife/go-patch v0.2.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/docker v0.0.0-20171120205147-9de84a78d76e // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/gorilla/websocket v1.2.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-getter v1.2.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/go-plugin v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.1
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/hcl2 v0.0.0-20180114202535-883a81b4902e // indirect
	github.com/hashicorp/hil v0.0.0-20190212132231-97b3a9cdfa93 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform v0.11.2
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/kr/pretty v0.1.0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lunixbochs/vtclean v0.0.0-20170504063817-d14193dfc626 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/mattn/go-colorable v0.1.0 // indirect
	github.com/mitchellh/cli v0.0.0-20180117155440-518dc677a1e1 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/moby/moby v1.13.1 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/posener/complete v0.0.0-20180110201102-22fe9ceed3cf // indirect
	github.com/poy/eachers v0.0.0-20181020210610-23942921fe77 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.0.5 // indirect
	github.com/src-d/gcfg v1.3.0 // indirect
	github.com/src-d/go-git-fixtures v3.5.0+incompatible // indirect
	github.com/tedsuo/ifrit v0.0.0-20180802180643-bea94bb476cc // indirect
	github.com/tedsuo/rata v1.0.0 // indirect
	github.com/vito/go-interact v0.0.0-20171111012221-fa338ed9e9ec // indirect
	github.com/xanzy/ssh-agent v0.1.0 // indirect
	github.com/zclconf/go-cty v0.0.0-20180106055834-709e4033eeb0 // indirect
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c // indirect
	golang.org/x/net v0.0.0-20190326090315-15845e8f865b // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/src-d/go-billy.v3 v3.1.0 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.0.0-rc11
	gopkg.in/warnings.v0 v0.1.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)
