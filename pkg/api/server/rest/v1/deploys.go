package v1

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

const deployIdentifier = "deploy_id"

var (
	deployIdentifierPathComponent = fmt.Sprintf(":%v", deployIdentifier)
	deployPath                    = fmt.Sprintf(v1rest.DeployPathFormat, systemIdentifierPathComponent, deployIdentifierPathComponent)
)

func (api *LatticeAPI) setupDeployEndpoints() {
	deploysPath := fmt.Sprintf(v1rest.DeploysPathFormat, systemIdentifierPathComponent)
	// deploy
	api.router.POST(deploysPath, api.handleDeploySystem)

	// list-deploys
	api.router.GET(deploysPath, api.handleListDeploys)

	// get-deploy
	api.router.GET(deployPath, api.handleGetDeploy)

}

// handleDeploySystem handler for deploy-system
// @ID deploy-system
// @Summary Deploy system
// @Description Deploys the system
// @Router /systems/{system}/deploys [post]
// @Security ApiKeyAuth
// @Tags deploys
// @Param system path string true "System ID"
// @Param deployRequest body rest.DeployRequest true "Create deploy"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Deploy
// @Failure 400 {object} v1.ErrorResponse
func (api *LatticeAPI) handleDeploySystem(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	var req v1rest.DeployRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	if req.Version != nil && req.BuildID != nil {
		c.String(http.StatusBadRequest, "can only specify version or buildId")
		return
	}

	if req.Version == nil && req.BuildID == nil {
		c.String(http.StatusBadRequest, "must specify version or buildId")
		return
	}

	var deploy *v1.Deploy
	var err error
	if req.Version != nil {
		path := tree.RootPath()
		if req.Version.Path != nil {
			path = *req.Version.Path
		}
		deploy, err = api.backend.Systems().Deploys(systemID).CreateFromVersion(req.Version.Version, path)
	} else {
		deploy, err = api.backend.Systems().Deploys(systemID).CreateFromBuild(*req.BuildID)
	}

	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidSystemVersion, v1.ErrorCodeInvalidBuildID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, deploy)
}

// handleListDeploys handler for list-deploys
// @ID list-deploys
// @Summary List deploys
// @Description Lists all deploys of the system
// @Router /systems/{system}/deploys [get]
// @Security ApiKeyAuth
// @Tags deploys
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Deploy
func (api *LatticeAPI) handleListDeploys(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	deploys, err := api.backend.Systems().Deploys(systemID).List()
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, deploys)
}

// handleGetDeploy handler for get-deploy
// @ID get-deploy
// @Summary Get deploy
// @Description Gets the deploy object
// @Router /systems/{system}/deploys/{id} [get]
// @Security ApiKeyAuth
// @Tags deploys
// @Param system path string true "System ID"
// @Param id path string true "Deploy ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Deploy
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetDeploy(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	deployID := v1.DeployID(c.Param(deployIdentifier))

	deploy, err := api.backend.Systems().Deploys(systemID).Get(deployID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidDeployID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, deploy)
}
