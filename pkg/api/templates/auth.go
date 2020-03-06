package templates

import (
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	"github.com/gophercloud/gophercloud/pagination"
)

func getServiceClient(projectName, domainName string) (sc gophercloud.ServiceClient, err error) {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return
	}
	opts.AllowReauth = true
	opts.Scope = &gophercloud.AuthScope{
		ProjectName: projectName,
		DomainName:  domainName,
	}

	pc, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return
	}

	sc = gophercloud.ServiceClient{
		ProviderClient: pc,
		Endpoint:       opts.IdentityEndpoint,
		ResourceBase:   opts.IdentityEndpoint,
	}

	return
}

func createApplicationCredentials(oc *gophercloud.ServiceClient, domainID string, roles []applicationcredentials.Role) (ac *applicationcredentials.ApplicationCredential, err error) {
	userID := os.Getenv("OS_USER_ID")
	createOpts := applicationcredentials.CreateOpts{
		Name:  domainID,
		Roles: roles,
		/*
			AccessRules: []applicationcredentials.AccessRule{
				{
					Method:  "GET",
					Service: "maia",
				},
			},
		*/
		Unrestricted: false,
	}

	//listOpts := applicationcredentials.ListOpts{Name: domainID}

	ac, err = applicationcredentials.Create(oc, userID, createOpts).Extract()
	if err != nil {
		if _, ok := err.(gophercloud.ErrDefault409); ok {
			r := applicationcredentials.List(oc, userID, nil)
			if err = r.EachPage(func(page pagination.Page) (bool, error) {
				acs, err := applicationcredentials.ExtractApplicationCredentials(page)
				if err != nil {
					return false, err
				}
				for _, a := range acs {
					if a.Name == domainID {
						if err = applicationcredentials.Delete(oc, userID, a.ID).ExtractErr(); err != nil {
							return false, err
						}
						return false, nil
					}
				}
				return false, nil
			}); err == nil {
				return createApplicationCredentials(oc, domainID, roles)
			}
			return
		}
	}
	return
}
