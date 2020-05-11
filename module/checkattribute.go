package module

import (
	"fmt"

	"github.com/edwardbrowncross/amazon-connect-simulator/call"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type checkAttribute flow.Module

type checkAttributeParams struct {
	Namespace string
	Attribute string
}

func (m checkAttribute) Run(ctx *call.Context) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleCheckAttribute {
		return nil, fmt.Errorf("module of type %s being run as checkAttribute", m.Type)
	}
	p := checkAttributeParams{}
	err = ctx.UnmarshalParameters(m.Parameters, &p)
	if err != nil {
		return
	}
	v := ctx.ResolveValue(flow.ModuleParameterNamespace(p.Namespace), p.Attribute)
	vs := fmt.Sprintf("%s", v)
	return evaluateConditions(m.Branches, vs)
}

func evaluateConditions(c flow.ModuleBranchList, v string) (*flow.ModuleID, error) {
	conditions := c.List(flow.BranchEvaluate)
	for _, c := range conditions {
		pass := false
		switch c.ConditionType {
		case flow.ConditionEquals:
			pass = bool(v == c.ConditionValue)
		case flow.ConditionGT:
			pass = bool(v > c.ConditionValue)
		case flow.ConditionGTE:
			pass = bool(v >= c.ConditionValue)
		case flow.ConditionLT:
			pass = bool(v < c.ConditionValue)
		case flow.ConditionLTE:
			pass = bool(v <= c.ConditionValue)
		default:
			return nil, fmt.Errorf("unhandled condition type: %s", c.ConditionType)
		}
		if pass {
			return &c.Transition, nil
		}
	}
	return c.GetLink(flow.BranchNoMatch), nil
}
