package v1

import (
	// TODO: feels a little weird to have to import this here. should type definitions under pkg/system be moved into pkg/types?
	"time"

	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type (
	BuildID    string
	BuildState string
)

const (
	BuildStatePending   BuildState = "pending"
	BuildStateRunning   BuildState = "running"
	BuildStateSucceeded BuildState = "succeeded"
	BuildStateFailed    BuildState = "failed"
)

type Build struct {
	// ID
	ID BuildID `json:"id"`
	// State
	State BuildState `json:"state"`
	// Start timestamp
	StartTimestamp *time.Time `json:"startTimestamp,omitempty"`
	// Completion timestamp
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`
	// Version
	Version SystemVersion `json:"version"`

	// Services maps service paths (e.g. /foo/bar/buzz) to the
	// status of the build for that service in the Build.
	Services map[tree.Path]ServiceBuild `json:"services"`
}

type ServiceBuild struct {
	// Container Build
	ContainerBuild
	// Sidecars
	Sidecars map[string]ContainerBuild `json:"sidecars"`
}

type (
	ContainerBuildState string
	ContainerBuildPhase string
	ContainerBuildID    string
)

const (
	ContainerBuildPhasePullingGitRepository ContainerBuildPhase = "pulling git repository"
	ContainerBuildPhasePullingDockerImage   ContainerBuildPhase = "pulling docker image"
	ContainerBuildPhaseBuildingDockerImage  ContainerBuildPhase = "building docker image"
	ContainerBuildPhasePushingDockerImage   ContainerBuildPhase = "pushing docker image"

	ContainerBuildStatePending   ContainerBuildState = "pending"
	ContainerBuildStateQueued    ContainerBuildState = "queued"
	ContainerBuildStateRunning   ContainerBuildState = "running"
	ContainerBuildStateSucceeded ContainerBuildState = "succeeded"
	ContainerBuildStateFailed    ContainerBuildState = "failed"
)

type ContainerBuild struct {
	ID    ContainerBuildID    `json:"id"`
	State ContainerBuildState `json:"state"`

	StartTimestamp      *time.Time `json:"startTimestamp,omitempty"`
	CompletionTimestamp *time.Time `json:"completionTimestamp,omitempty"`

	LastObservedPhase *ContainerBuildPhase `json:"lastObservedPhase,omitempty"`
	FailureMessage    *string              `json:"failureMessage,omitempty"`
}

type ContainerBuildFailureInfo struct {
	Message  string `json:"message"`
	Internal bool   `json:"internal"`
}
