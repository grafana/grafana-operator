// Package keystone provides a go http middleware for authentication incoming
// http request against Openstack Keystone. It it modelled after the original
// keystone middleware:
// http://docs.openstack.org/developer/keystonemiddleware/middlewarearchitecture.html
//
// The middleware authenticates incoming requests by validating the `X-Auth-Token` header
// and adding additional headers to the incoming request containing the validation result.
// The final authentication/authorization decision is delegated to subsequent http handlers.
package keystone

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var Log func(string, ...interface{}) = func(format string, a ...interface{}) {
	log.Printf(format, a...)
}

// Cache provides the interface for cache implementations.
type Cache interface {
	//Set stores a value with the given ttl
	Set(key string, value interface{}, ttl time.Duration)
	//Get retrieves a value previously stored in the cache.
	//value has to be a pointer to a data structure that matches the type previously given to Set
	//The return value indicates if a value was found
	Get(key string, value interface{}) bool
}

//Auth is the entrypoint for creating the middlware
type Auth struct {
	//Keystone v3 endpoint url for validating tokens ( e.g https://some.where:5000/v3)
	Endpoint string
	//User-Agent used for all http request by the middlware. Defaults to go-keystone-middlware/1.0
	UserAgent string
	//A cache implementation the middleware should use for caching tokens. By default no caching is performed.
	TokenCache Cache
	//How long to cache tokens. Defaults to 5 minutes.
	CacheTime time.Duration

	//http client to use for requests, default to  &http.Client{ Timeout: 5 * time.Second }
	Client *http.Client
}

// New returns a new Auth object initialized with default values
func New(endpoint string) *Auth {
	auth := &Auth{Endpoint: endpoint}
	auth.ensureDefaults()
	return auth
}

//Handler returns a http handler for use in a middleware chain.
func (a *Auth) Handler(h http.Handler) http.Handler {
	a.ensureDefaults()
	return &handler{Auth: a, handler: h}
}

//Validate a token.
//This is useful if you don't want to use the http middleware
func (a *Auth) Validate(authToken string) (*Token, error) {

	if a.TokenCache != nil {
		var cachedToken Token
		if ok := a.TokenCache.Get(authToken, &cachedToken); ok && cachedToken.Valid() {
			Log("Found valid token in cache")
			return &cachedToken, nil
		}
	}

	req, err := http.NewRequest("GET", a.Endpoint+"/auth/tokens?nocatalog", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", authToken)
	req.Header.Set("X-Subject-Token", authToken)
	req.Header.Set("User-Agent", a.UserAgent)

	r, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		return nil, errors.New(r.Status)
	}

	var resp authResponse
	if err = json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	if e := resp.Error; e != nil {
		return nil, fmt.Errorf("%s : %s", r.Status, e.Message)
	}
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", r.Status)
	}
	if resp.Token == nil {
		return nil, errors.New("Response didn't contain token context")
	}
	if !resp.Token.Valid() {
		return nil, errors.New("Returned token is not valid")
	}

	if a.TokenCache != nil {
		ttl := a.CacheTime
		//The expiry date of the token provides an upper bound on the cache time
		if expiresIn := resp.Token.ExpiresAt.Sub(time.Now()); expiresIn < a.CacheTime {
			ttl = expiresIn
		}
		a.TokenCache.Set(authToken, *resp.Token, ttl)
	}

	return resp.Token, nil
}

func (a *Auth) ensureDefaults() {

	if a.UserAgent == "" {
		a.UserAgent = "go-keystone-middleware/1.0"
	}

	if a.CacheTime == 0 {
		a.CacheTime = 5 * time.Minute
	}

	if a.Client == nil {
		a.Client = &http.Client{
			Timeout: 5 * time.Second,
		}
	}

}

type handler struct {
	*Auth
	handler http.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filterIncomingHeaders(req)
	req.Header.Set("X-Identity-Status", "Invalid")
	defer h.handler.ServeHTTP(w, req)
	authToken := req.Header.Get("X-Auth-Token")
	if authToken == "" {
		return
	}

	context, err := h.Auth.Validate(authToken)
	if err != nil {
		//ToDo: How to handle logging, printing to stdout isn't the best thing
		Log("Failed to validate token: %v", err)
		return
	}

	req.Header.Set("X-Identity-Status", "Confirmed")
	for k, v := range context.headers() {
		req.Header.Set(k, v)
	}
}

//Domain holds information about the scope of a token
type Domain struct {
	ID      string
	Name    string
	Enabled bool
}

//Project contains information about the scope of a token
type Project struct {
	ID      string
	Name    string
	Enabled bool
	Domain  Domain
}

//Token describes the scope of a validated token
type Token struct {
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
	User      struct {
		ID      string
		Name    string
		Email   string
		Enabled bool
		Domain  struct {
			ID   string
			Name string
		}
	}
	Project *Project
	Domain  *Domain
	Roles   []struct {
		ID   string
		Name string
	}
}

// Valid returns if the token is valid based on the expiration and issue date
func (t Token) Valid() bool {
	now := time.Now().Unix()
	return t.IssuedAt.Unix() <= now && now < t.ExpiresAt.Unix()
}

type authResponse struct {
	Error *struct {
		Code    int
		Message string
		Title   string
	}
	Token *Token
}

func (t Token) headers() map[string]string {
	headers := make(map[string]string)
	headers["X-User-Id"] = t.User.ID
	headers["X-User-Name"] = t.User.Name
	headers["X-User-Domain-Id"] = t.User.Domain.ID
	headers["X-User-Domain-Name"] = t.User.Domain.Name

	if project := t.Project; project != nil {
		headers["X-Project-Name"] = project.Name
		headers["X-Project-Id"] = project.ID
		headers["X-Project-Domain-Name"] = project.Domain.Name
		headers["X-Project-Domain-Id"] = project.Domain.ID

	}

	if domain := t.Domain; domain != nil {
		headers["X-Domain-Id"] = domain.ID
		headers["X-Domain-Name"] = domain.Name
	}

	if roles := t.Roles; roles != nil {
		roleNames := []string{}
		for _, role := range t.Roles {
			roleNames = append(roleNames, role.Name)
		}
		headers["X-Roles"] = strings.Join(roleNames, ",")

	}

	return headers
}

func filterIncomingHeaders(req *http.Request) {
	req.Header.Del("X-Identity-Status")
	req.Header.Del("X-Service-Identity-Status")

	req.Header.Del("X-Domain-Id")
	req.Header.Del("X-Service-Domain-Id")

	req.Header.Del("X-Domain-Name")
	req.Header.Del("X-Service-Domain-Name")

	req.Header.Del("X-Project-Id")
	req.Header.Del("X-Service-Project-Id")

	req.Header.Del("X-Project-Name")
	req.Header.Del("X-Service-Project-Name")

	req.Header.Del("X-Project-Domain-Id")
	req.Header.Del("X-Service-Project-Domain-Id")

	req.Header.Del("X-Project-Domain-Name")
	req.Header.Del("X-Service-Project-Domain-Name")

	req.Header.Del("X-User-Id")
	req.Header.Del("X-Service-User-Id")

	req.Header.Del("X-User-Name")
	req.Header.Del("X-Service-User-Name")

	req.Header.Del("X-User-Domain-Id")
	req.Header.Del("X-Service-User-Domain-Id")

	req.Header.Del("X-User-Domain-Name")
	req.Header.Del("X-Service-User-Domain-Name")

	req.Header.Del("X-Roles")
	req.Header.Del("X-Service-Roles")

	req.Header.Del("X-Servie-Catalog")

	//deprecated Headers
	req.Header.Del("X-Tenant-Id")
	req.Header.Del("X-Tenant")
	req.Header.Del("X-User")
	req.Header.Del("X-Role")
}
