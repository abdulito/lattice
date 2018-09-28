package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/authenticator"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/authenticator/apikey"
	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/authenticator/bearertoken"

	"github.com/mlab-lattice/lattice/pkg/api/server/backend"
	restv1 "github.com/mlab-lattice/lattice/pkg/api/server/rest/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

const (
	currentUserContextKey = "CURRENT_USER"
)

type restServer struct {
	router         *gin.Engine
	backend        backend.Interface
	resolver       resolver.Interface
	authenticators []authenticator.Request
}

func RunNewRestServer(backend backend.Interface, resolver resolver.Interface, port int32, options *ServerOptions) {
	router := gin.Default()
	// Some of our paths use URL encoded paths, so don't have
	// gin decode those
	router.UseRawPath = true
	s := restServer{
		router:   router,
		backend:  backend,
		resolver: resolver,
	}
	s.initAuthenticators(options)

	s.mountHandlers(options)
	s.router.Run(fmt.Sprintf(":%v", port))
}
func (r *restServer) initAuthenticators(options *ServerOptions) {

	authenticators := make([]authenticator.Request, 0)

	// setup legacy authentication as needed
	if options.AuthOptions.LegacyAPIAuthKey != "" {
		fmt.Println("Setting up authentication with legacy api key header")
		authenticators = append(authenticators, apikey.New(options.AuthOptions.LegacyAPIAuthKey))
	}

	// setup bearer token auth as needed
	if options.AuthOptions.Token != nil {
		bearerAuthenticator, err := bearertoken.New(options.AuthOptions.Token)
		if err != nil {
			panic(err)
		}
		authenticators = append(authenticators, bearerAuthenticator)
	}
	r.authenticators = authenticators
}
func (r *restServer) mountHandlers(options *ServerOptions) {
	// Status
	r.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	routerGroup := r.router.Group("/")
	r.setupAuthentication(routerGroup)

	restv1.MountHandlers(routerGroup, r.backend.V1(), r.resolver)
}

func (r *restServer) setupAuthentication(router *gin.RouterGroup) {
	if len(r.authenticators) == 0 {
		fmt.Println("WARNING: No authenticators configured.")
	} else {
		router.Use(r.authenticateRequest())
	}

}

// authenticateRequest authenticates the request against the configured authentication api key
func (r *restServer) authenticateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, a := range r.authenticators {
			userObject, ok, err := a.AuthenticateRequest(c)

			if err != nil {
				fmt.Printf("Failed to authenticated. Got error %v\n", err)
				abortUnauthorized(c)
				return
			} else if ok { // Auth Success!
				fmt.Printf("User %v successfully authenticated\n", userObject.Name())
				// Attach user to current context
				c.Set(currentUserContextKey, userObject)
				return
			}

		}

		// No authentication provided
		abortUnauthorized(c)
	}
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
}
