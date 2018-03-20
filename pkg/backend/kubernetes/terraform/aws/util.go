package aws

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/types"
)

func GetS3BackendSystemStatePathRoot(clusterID types.LatticeID, systemID types.SystemID) string {
	return fmt.Sprintf("%v/system/%v/terraform/state", GetS3BackendStatePathRoot(clusterID), systemID)
}

func GetS3BackendNodePoolPathRoot(clusterID types.LatticeID, nodePoolID string) string {
	return fmt.Sprintf("%v/node-pool/terraform/state", GetS3BackendStatePathRoot(clusterID))
}

func GetS3BackendStatePathRoot(clusterID types.LatticeID) string {
	return fmt.Sprintf("lattice/%v", clusterID)
}
