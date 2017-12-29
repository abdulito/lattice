package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/template/language"
	"github.com/mlab-lattice/system/pkg/util/git"
)

// SystemResolver resolves system definitions from different sources such as git
type SystemResolver struct {
	GitResolver   *git.Resolver
}

// GitResolveOptions allows options for resolution
type GitResolveOptions struct {
	SSHKey []byte
}

type resolveContext struct {
	gitURI            string
	gitResolveOptions *GitResolveOptions
}

func NewSystemResolver(workDirectory string) (*SystemResolver, error) {
	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	gitResolver, err := git.NewResolver(workDirectory + "/git")
	if err != nil {
		return nil, err
	}

	sr := &SystemResolver{
		GitResolver:   gitResolver,
	}
	return sr, nil
}

// resolves the definition
func (resolver *SystemResolver) ResolveDefinition(uri string) (tree.Node, error) {

	return resolver.readNodeFromFile(uri)
}

// lists the versions of the specified definition's uri
func (resolver *SystemResolver) ListDefinitionVersions(uri string, gitResolveOptions *GitResolveOptions) ([]string, error) {
	if gitResolveOptions == nil {
		gitResolveOptions = &GitResolveOptions{}
	}
	ctx := &resolveContext{
		gitURI:            uri,
		gitResolveOptions: gitResolveOptions,
	}
	return resolver.listRepoVersionTags(ctx)

}

// readNodeFromFile reads a definition node from a file
func (resolver *SystemResolver) readNodeFromFile(uri string) (tree.Node, error) {

	engine := language.NewEngine()
	template, err := engine.ParseTemplate(uri, make(map[string]interface{}))

	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(template.Value)
	if err != nil {
		return nil, err
	}

	defInterface, err := definition.UnmarshalJSON(jsonBytes)

	if err != nil {
		return nil, err
	}

	return tree.NewNode(defInterface, nil)
}

// lists the tags in a repo
func (resolver *SystemResolver) listRepoVersionTags(ctx *resolveContext) ([]string, error) {
	gitResolverContext := &git.Context{
		URI:    ctx.gitURI,
		SSHKey: ctx.gitResolveOptions.SSHKey,
	}
	return resolver.GitResolver.GetTagNames(gitResolverContext)
}
