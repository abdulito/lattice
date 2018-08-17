package v1

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
)

const jobIdentifier = "job_id"

var jobsPath = fmt.Sprintf(v1rest.JobsPathFormat, systemIdentifierPathComponent)

var jobIdentifierPathComponent = fmt.Sprintf(":%v", jobIdentifier)
var jobPath = fmt.Sprintf(v1rest.JobPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)
var jobLogPath = fmt.Sprintf(v1rest.JobLogsPathFormat, systemIdentifierPathComponent, jobIdentifierPathComponent)

func (api *LatticeAPI) setupJobsEndpoints() {

	// run-job
	api.router.POST(jobsPath, api.handleRunJob)

	// list-jobs
	api.router.GET(jobsPath, api.handleListJobs)

	// get-job
	api.router.GET(jobPath, api.handleGetJob)

	// get-job-logs
	api.router.GET(jobLogPath, api.handleGetJobLogs)

}

// RunJob godoc
// @ID run-job
// @Summary Run job
// @Description run job
// @Router /systems/{system}/builds [post]
// @Tags jobs
// @Param system path string true "System ID"
// @Param jobRequest body rest.RunJobRequest true "Create build"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Job
// @Failure 400 {object} v1.ErrorResponse
func (api *LatticeAPI) handleRunJob(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	var req v1rest.RunJobRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	job, err := api.backend.RunJob(systemID, req.Path, req.Command, req.Environment)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, job)

}

// ListJobs godoc
// @ID list-jobs
// @Summary Lists jobs
// @Description list jobs
// @Router /systems/{system}/jobs [get]
// @Tags jobs
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Job
func (api *LatticeAPI) handleListJobs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	jobs, err := api.backend.ListJobs(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, jobs)
}

// GetJob godoc
// @ID get-job
// @Summary Get job
// @Description get job
// @Router /systems/{system}/jobs/{id} [get]
// @Tags jobs
// @Param system path string true "System ID"
// @Param id path string true "Job ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Job
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetJob(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))

	job, err := api.backend.GetJob(systemID, jobID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetJobLogs godoc
// @ID get-job-logs
// @Summary Get job logs
// @Description get job logs
// @Router /systems/{system}/jobs/{id}/logs  [get]
// @Tags jobs
// @Param system path string true "System ID"
// @Param id path string true "Job ID"
// @Param sidecar query string false "Sidecar"
// @Param follow query string bool "Follow"
// @Param previous query boolean false "Previous"
// @Param timestamps query boolean false "Timestamps"
// @Param tail query integer false "tail"
// @Param since query string false "Since"
// @Param sinceTime query string false "Since Time"
// @Accept  json
// @Produce  json
// @Success 200 {string} string "log stream"
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetJobLogs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	jobID := v1.JobID(c.Param(jobIdentifier))

	sidecarQuery, sidecarSet := c.GetQuery("sidecar")
	var sidecar *string
	if sidecarSet {
		sidecar = &sidecarQuery
	}

	logOptions, err := requestedLogOptions(c)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	log, err := api.backend.JobLogs(systemID, jobID, sidecar, logOptions)
	if err != nil {
		handleError(c, err)
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
