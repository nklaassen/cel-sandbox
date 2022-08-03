package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func eval(traits map[string][]string, expr string) []string {
	env, err := cel.NewEnv(
		cel.Variable("external", cel.MapType(cel.StringType, cel.ListType(cel.StringType))),
		ext.Strings(),
	)
	check(err)

	ast, issues := env.Compile(expr)
	check(issues.Err())

	prg, err := env.Program(ast)
	check(err)

	out, _, err := prg.Eval(map[string]interface{}{
		"external": traits,
	})
	check(err)

	if val, err := out.ConvertToNative(reflect.TypeOf("")); err == nil {
		return []string{val.(string)}
	}

	list, err := out.ConvertToNative(reflect.TypeOf([]string{}))
	check(err)
	return list.([]string)
}

func main() {
	traits := map[string][]string{
		"username": {"my-username"},
		"email":    {"nic@goteleport.com"},
		"groups":   {"env-staging", "env-qa", "devs"},
	}

	logins := eval(traits, `
		['ubuntu'] +
		external.username.map(username, username.replace('-', '_')) +
		('nic@goteleport.com' in external.email ? ['root'] : []) +
		external.email.map(email, email.matches('^[^@]+@goteleport.com$'), email.replace('@goteleport.com', '', 1))
	`)
	fmt.Printf("logins: %v\n", logins)

	allow_env := eval(traits, `
		external.groups.map(group, group.matches('^env-\\w+$'), group.replace('env-', '', 1)) +
		('contractors' in external.groups ? [] : 'devs' in external.groups ? ['dev'] : [])
	`)
	fmt.Printf("allow-env: %v\n", allow_env)

}
