package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"

	policy "github.com/databus23/goslo.policy"
	runtime "github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware/denco"
	"github.com/integr8ly/grafana-operator/pkg/api/models"
	flag "github.com/spf13/pflag"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	DefaultPolicyFile string
	pathConverter     = regexp.MustCompile(`{(.+?)}`)
	log               = logf.Log.WithName("api_authorizer")
)

func init() {
	flag.StringVar(&DefaultPolicyFile, "policy", "etc/policy.json", "API authorization policy file")
}

type osloPolicyAuthorizer struct {
	routers  map[string]*denco.Router
	enforcer *policy.Enforcer
}

func LoadPolicy(policyFile string) (map[string]string, error) {
	file, err := os.Open(policyFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var rules map[string]string
	err = json.NewDecoder(file).Decode(&rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func NewOsloPolicyAuthorizer() (runtime.Authorizer, error) {

	recordMap := make(map[string][]denco.Record)

	routers := make(map[string]*denco.Router, len(recordMap))

	for method, routes := range recordMap {
		routers[method] = denco.New()
		if err := routers[method].Build(routes); err != nil {
			return nil, err
		}
	}

	return &osloPolicyAuthorizer{routers: routers}, nil
}

func (o *osloPolicyAuthorizer) Authorize(req *http.Request, principal interface{}) error {
	authUser := principal.(*models.Principal)
	router, ok := o.routers[req.Method]
	if !ok {
		return fmt.Errorf("No router found for method %s", req.Method)
	}
	operation, params, found := router.Lookup(path.Clean(req.URL.EscapedPath()))
	if !found {
		return fmt.Errorf("Operation not found for %s %s", req.Method, req.URL.Path)
	}
	operationID := operation.(string)
	requestVars := make(map[string]string, len(params))
	for _, param := range params {
		requestVars[param.Name] = param.Value
	}
	allowed := o.enforcer.Enforce(operationID, policy.Context{
		Auth:    map[string]string{"user_id": authUser.ID, "project_id": authUser.Account, "project_name": authUser.AccountName, "domain_name": authUser.Domain},
		Roles:   authUser.Roles,
		Request: requestVars,
	})
	if !allowed {
		return fmt.Errorf("Authorization failed for user %s for operation %s", authUser.Name, operationID)
	}

	return nil
}
