package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/databus23/keystone"
	"github.com/databus23/keystone/cache/memory"
	errors "github.com/go-openapi/errors"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	flag "github.com/spf13/pflag"
)

var authURL string

func init() {
	flag.StringVar(&authURL, "auth-url", "https://identity-3.qa-de-1.cloud.sap/v3", "Openstack identity v3 auth url")
}

func Keystone() func(token string) (*models.Principal, error) {
	if !(strings.HasSuffix(authURL, "/v3") || strings.HasSuffix(authURL, "/v3/")) {
		authURL = fmt.Sprintf("%s/%s", strings.TrimRight(authURL, "/"), "/v3")
	}

	keystone.Log = func(format string, a ...interface{}) {
		log.Info("library", "keystone", "msg", fmt.Sprintf(format, a...))
	}
	auth := keystone.New(authURL)
	auth.TokenCache = memory.New(10 * time.Minute)

	return func(token string) (*models.Principal, error) {
		//return &models.Principal{AccountName: "project123"}, nil
		t, err := auth.Validate(token)
		if err != nil {
			return nil, errors.New(401, fmt.Sprintf("Authentication failed: %s", err))
		}

		if t.Project == nil {
			return nil, errors.New(401, "Auth token isn't project scoped")
		}
		roles := make([]string, 0, len(t.Roles))
		for _, role := range t.Roles {
			roles = append(roles, role.Name)
		}
		return &models.Principal{AuthURL: authURL, ID: t.User.ID, Name: t.User.Name, Domain: t.User.Domain.Name, Account: t.Project.ID, AccountName: t.Project.Name, Roles: roles}, nil
	}
}
