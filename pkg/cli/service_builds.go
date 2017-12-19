package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getServiceBuildRenderMap(sb types.ServiceBuild) RenderMap {
	return RenderMap{
		"ID":    string(sb.ID),
		"State": string(sb.State),
	}
}

func ShowServiceBuild(build types.ServiceBuild) {
	renderMap := getServiceBuildRenderMap(build)
	showResource(renderMap)
}

func ShowServiceBuilds(builds []types.ServiceBuild) {
	renderMaps := make([]RenderMap, len(builds))
	for i, b := range builds {
		renderMaps[i] = getServiceBuildRenderMap(b)
	}
	listResources(renderMaps)
}
