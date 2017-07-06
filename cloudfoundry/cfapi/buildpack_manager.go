package cfapi

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/resources"
	"code.cloudfoundry.org/cli/cf/appfiles"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
)

// BuildpackManager -
type BuildpackManager struct {
	log *Logger

	config    coreconfig.Reader
	ccGateway net.Gateway

	apiEndpoint string

	bpRepo     api.BuildpackRepository
	bpBitsRepo api.BuildpackBitsRepository

	zipper appfiles.Zipper
}

// CCBuildpack -
type CCBuildpack struct {
	ID string

	Name     string `json:"name"`
	Position *int   `json:"position,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
	Locked   *bool  `json:"locked,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// CCBuildpackResource -
type CCBuildpackResource struct {
	Metadata resources.Metadata `json:"metadata"`
	Entity   CCBuildpack        `json:"entity"`
}

// toModel -
func (bp *CCBuildpack) toModel() (buildpack models.Buildpack) {
	buildpack.GUID = bp.ID
	buildpack.Name = bp.Name
	buildpack.Position = bp.Position
	buildpack.Enabled = bp.Enabled
	buildpack.Locked = bp.Locked
	return
}

// fromModel -
func (bp *CCBuildpack) fromModel(buildpack models.Buildpack) {
	bp.ID = buildpack.GUID
	bp.Name = buildpack.Name
	bp.Position = buildpack.Position
	bp.Enabled = buildpack.Enabled
	bp.Locked = buildpack.Locked
}

// newBuildpackManager -
func newBuildpackManager(config coreconfig.Reader, ccGateway net.Gateway, logger *Logger) (sm *BuildpackManager, err error) {

	zipper := appfiles.ApplicationZipper{}

	sm = &BuildpackManager{
		log: logger,

		config:    config,
		ccGateway: ccGateway,

		apiEndpoint: config.APIEndpoint(),

		bpRepo:     api.NewCloudControllerBuildpackRepository(config, ccGateway),
		bpBitsRepo: api.NewCloudControllerBuildpackBitsRepository(config, ccGateway, zipper),

		zipper: zipper,
	}

	return
}

// ReadBuildpack -
func (bpm *BuildpackManager) ReadBuildpack(buildpackID string) (bp CCBuildpack, err error) {

	resource := &CCBuildpackResource{}
	err = bpm.ccGateway.GetResource(
		fmt.Sprintf("%s/v2/buildpacks/%s", bpm.apiEndpoint, buildpackID), &resource)

	bp = resource.Entity
	bp.ID = resource.Metadata.GUID
	return
}

// CreateBuildpack -
func (bpm *BuildpackManager) CreateBuildpack(
	name string, position *int, enabled *bool, locked *bool, buildpackPath string) (bp CCBuildpack, err error) {

	var b models.Buildpack

	if b, err = bpm.bpRepo.Create(name, position, enabled, locked); err != nil {
		return
	}
	bp.fromModel(b)
	bp, err = bpm.UploadBuildpackBits(bp, buildpackPath)
	return
}

// UpdateBuildpack -
func (bpm *BuildpackManager) UpdateBuildpack(buildpackID string,
	name string, position *int, enabled *bool, locked *bool) (bp CCBuildpack, err error) {

	b := models.Buildpack{
		GUID:     buildpackID,
		Name:     name,
		Position: position,
		Enabled:  enabled,
		Locked:   locked,
	}
	if b, err = bpm.bpRepo.Update(b); err == nil {
		bp.fromModel(b)
	}
	return
}

// UploadBuildpackBits -
func (bpm *BuildpackManager) UploadBuildpackBits(bp CCBuildpack, buildpackPath string) (CCBuildpack, error) {

	var (
		zipFile *os.File
		err     error
	)

	if strings.HasPrefix(buildpackPath, "file://") {
		buildpackPath = buildpackPath[7:]
	}
	if zipFile, bp.Filename, err = bpm.bpBitsRepo.CreateBuildpackZipFile(buildpackPath); err != nil {
		return bp, err
	}
	if err = bpm.bpBitsRepo.UploadBuildpack(bp.toModel(), zipFile, bp.Filename); err != nil {
		return bp, err
	}
	return bp, nil
}

// DeleteBuildpack -
func (bpm *BuildpackManager) DeleteBuildpack(buildpackID string) error {
	return bpm.bpRepo.Delete(buildpackID)
}

// FindBuildpack -
func (bpm *BuildpackManager) FindBuildpack(buildpackName string) (bp CCBuildpack, err error) {

	var b models.Buildpack

	if b, err = bpm.bpRepo.FindByName(buildpackName); err != nil {
		return
	}
	bp.fromModel(b)
	return
}
