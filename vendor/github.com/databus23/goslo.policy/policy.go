//Package policy provides RBAC policy enforcement similar to the OpenStack oslo.policy library.
package policy

import "fmt"

//go:generate go tool yacc -v "" -o parser.go parser.y
//go:generate sed -i .tmp -e s/yyEofCode/yyEOFCode/ parser.go
//go:generate sed -i .tmp -e 1d parser.go
//go:generate rm parser.go.tmp

type rule func(c Context) bool

//Check is the interface for checks
type Check func(c Context, key, match string) bool

//Enforcer is responsible for loading and enforcing rules.
type Enforcer struct {
	rules  map[string]rule
	checks map[string]Check
}

//Enforce checks authorization of a rule for the given Context
func (p *Enforcer) Enforce(rule string, c Context) bool {
	c.rules = &p.rules
	c.checks = p.checks
	r, ok := p.rules[rule]
	return ok && r(c)
}

//AddCheck registers a custom check for the given name.
//A custom check can by used by specifing the name as the left side of the check.
//E.g. mycheck:valueformycheck
func (p *Enforcer) AddCheck(name string, c Check) {
	p.checks[name] = c
}

//NewEnforcer parses the provided rule set and returns a policy enforcer
//By default the Enforcer registers the following checks
// "rule": RuleCheck
// "role": RoleCheck
// "http": HttpCheck
// "default": DefaultCheck
func NewEnforcer(rules map[string]string) (*Enforcer, error) {
	p := Enforcer{
		rules:  make(map[string]rule, len(rules)),
		checks: make(map[string]Check, 4),
	}
	p.AddCheck("rule", RuleCheck)
	p.AddCheck("role", RoleCheck)
	p.AddCheck("http", HTTPCheck)
	p.AddCheck("default", DefaultCheck)

	for name, str := range rules {
		lexer := newLexer(str)
		if yyParse(lexer) != 0 {
			return nil, fmt.Errorf("Failed to parse rule %s: %s", name, lexer.parseResult.(string))
		}
		p.rules[name] = lexer.parseResult.(rule)
	}
	return &p, nil
}

//Context encapsulates the external data required for enforcing a rules. Populating a Context object is left to the application using the policy engine.
type Context struct {
	//Authentication context information from the keystone token, e.g. user_id, user_domain_id...
	Auth map[string]string
	//Roles assigned to the user for the current scope
	Roles []string
	//Request variables that are referenced in policy rules
	Request map[string]string
	//Logger can be used to enable debug logging for this context.
	Logger func(msg string, args ...interface{})

	rules  *map[string]rule
	checks map[string]Check
}

//RuleCheck provides the standard rule:... check
func RuleCheck(c Context, key, match string) bool {
	rule, ok := (*c.rules)[match]
	if !ok {
		return false
	}
	return rule(c)
}

//RoleCheck provides the standard role:... check.
func RoleCheck(c Context, key, match string) bool {
	for _, r := range c.Roles {
		if r == match {
			return true
		}
	}
	return false
}

//HTTPCheck implements the http:... check
func HTTPCheck(c Context, key, match string) bool {
	return false // not implemented yet
}

//DefaultCheck is used whenever there is no specific check registered for the left hand side.
//It simply tries to match the right side if the check to the authentication credential given by
//the left side. E.g. user_id:%(target.user_id)
func DefaultCheck(c Context, key, match string) bool {
	cred, ok := c.Auth[key]
	if !ok {
		return false
	}
	return cred == match
}

func (c Context) genericCheck(key, match string, isVariable bool) bool {
	if c.Logger != nil {
		c.Logger("executing %s:%s", key, match)
	}
	if isVariable {
		m, ok := c.Request[match]
		if !ok {
			if c.Logger != nil {
				c.Logger("request variable %s not present in context. failing\n", match)
			}
			return false
		}
		match = m
	}

	if check, ok := c.checks[key]; ok {
		result := check(c, key, match)
		if c.Logger != nil {
			c.Logger("check %s returned: %v\n", key, result)
		}
		return result
	}
	result := c.checks["default"](c, key, match)
	if c.Logger != nil {
		c.Logger("default check result: %s:%s = %v\n", key, match, result)
	}
	return result

}

func (c Context) checkVariable(variable string, match interface{}) bool {
	if c.Logger != nil {
		c.Logger("executing '%s':%s\n", match, variable)
	}
	val, ok := c.Request[variable]
	if !ok {
		if c.Logger != nil {
			c.Logger("variable %s not present in context. failing\n", variable)
		}
		return false
	}
	if c.Logger != nil {
		c.Logger("'%s':%s = %v\n", match, val, val == match)
	}
	return val == match
}
