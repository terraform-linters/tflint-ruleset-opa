package main

import (
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-opa/opa"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &opa.RuleSet{
			BuiltinRuleSet: tflint.BuiltinRuleSet{
				Name:       "opa",
				Version:    "0.9.0",
				Constraint: ">= 0.43.0",
			},
		},
	})
}
