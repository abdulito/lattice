package constants

const (
	LabelKeyLatticeID = "lattice.mlab.com/id"

	LabelKeyControlPlane        = "control-plane.lattice.mlab.com"
	LabelKeyControlPlaneService = LabelKeyControlPlane + "/service"

	LabelKeyNodeRoleLattice  = "node-role.lattice.mlab.com"
	LabelKeyMasterNode       = LabelKeyNodeRoleLattice + "/master"
	LabelKeyBuildNode        = LabelKeyNodeRoleLattice + "/build"
	LabelKeyNodeRoleNodePool = LabelKeyNodeRoleLattice + "/node-pool"

	LabelKeyComponentBuildID = "component.build.lattice.mlab.com/id"

	LabelKeySystemRolloutVersion = "rollout.system.lattice.mlab.com/version"
	LabelKeySystemRolloutBuildID = "rollout.system.lattice.mlab.com/build"

	LabelKeyServiceID   = "service.lattice.mlab.com/id"
	LabelKeyServicePath = "service.lattice.mlab.com/path"

	LabelKeySystemBuildID = "system.build.lattice.mlab.com/id"
	LabelKeySystemVersion = "system.lattice.mlab.com/version"

	LabelKeySecret = "secret.lattice.mlab.com"
)
