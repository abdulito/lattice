package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition"
	"github.com/mlab-lattice/lattice/pkg/definition/template/language"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

// SystemResolver resolves system definitions from different sources such as git
type SystemResolver struct {
	gitResolver *git.Resolver
}

type resolveContext struct {
	gitURI            string
	gitResolveOptions *git.Options
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
		gitResolver: gitResolver,
	}
	return sr, nil
}

// resolves the definition
func (resolver *SystemResolver) ResolveDefinition(uri string, gitResolveOptions *git.Options) (tree.Node, error) {

	if gitResolveOptions == nil {
		gitResolveOptions = &git.Options{}
	}
	ctx := &resolveContext{
		gitURI:            uri,
		gitResolveOptions: gitResolveOptions,
	}

	return resolver.readNodeFromFile(ctx)
}

// lists the versions of the specified definition's uri
func (resolver *SystemResolver) ListDefinitionVersions(uri string, gitResolveOptions *git.Options) ([]string, error) {
	if gitResolveOptions == nil {
		gitResolveOptions = &git.Options{}
	}
	ctx := &resolveContext{
		gitURI:            uri,
		gitResolveOptions: gitResolveOptions,
	}
	return resolver.listRepoVersionTags(ctx)

}

// readNodeFromFile reads a definition node from a file
func (resolver *SystemResolver) readNodeFromFile(ctx *resolveContext) (tree.Node, error) {
	engine := language.NewEngine()

	options, err := language.CreateOptions(resolver.gitResolver.WorkDirectory, ctx.gitResolveOptions)
	if err != nil {
		return nil, err
	}

	result, err := engine.EvalFromURL(ctx.gitURI, make(map[string]interface{}), options)
	if err != nil {
		return nil, err
	}

	jsonBytes, err := json.Marshal(result.ValueAsMap())
	if err != nil {
		return nil, err
	}

	def, err := definition.NewFromJSON(jsonBytes)
	if err != nil {
		return nil, err
	}

	return tree.NewNode(def, nil)
}

// lists the tags in a repo
func (resolver *SystemResolver) listRepoVersionTags(ctx *resolveContext) ([]string, error) {
	gitResolverContext := &git.Context{
		URI:     ctx.gitURI,
		Options: ctx.gitResolveOptions,
	}
	return resolver.gitResolver.GetTagNames(gitResolverContext)
}
