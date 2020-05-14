package module

import (
	"fmt"
	"strconv"

	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type checkAttribute flow.Module

type checkAttributeParams struct {
	Namespace string
	Attribute string
}

func (m checkAttribute) Run(ctx CallContext) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleCheckAttribute {
		return nil, fmt.Errorf("module of type %s being run as checkAttribute", m.Type)
	}
	pr := parameterResolver{ctx}
	p := checkAttributeParams{}
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	v, err := pr.get(flow.ModuleParameterNamespace(p.Namespace), p.Attribute)
	if err != nil {
		return
	}
	vs := fmt.Sprintf("%s", v)
	return evaluateConditions(m.Branches, vs)
}

func evaluateConditions(c flow.ModuleBranchList, v string) (*flow.ModuleID, error) {
	conditions := c.List(flow.BranchEvaluate)
	vn, err := strconv.ParseFloat(v, 64)
	numeric := bool(err == nil)
	for _, c := range conditions {
		cvn, err := strconv.ParseFloat(c.ConditionValue, 64)
		numeric := numeric && bool(err == nil)
		pass := false
		switch c.ConditionType {
		case flow.ConditionEquals:
			pass = bool(v == c.ConditionValue)
		case flow.ConditionGT:
			pass = bool((numeric && vn > cvn) || (!numeric && v > c.ConditionValue))
		case flow.ConditionGTE:
			pass = bool((numeric && vn >= cvn) || (!numeric && v >= c.ConditionValue))
		case flow.ConditionLT:
			pass = bool((numeric && vn < cvn) || (!numeric && v < c.ConditionValue))
		case flow.ConditionLTE:
			pass = bool((numeric && vn <= cvn) || (!numeric && v <= c.ConditionValue))
		default:
			return nil, fmt.Errorf("unhandled condition type: %s", c.ConditionType)
		}
		if pass {
			return &c.Transition, nil
		}
	}
	return c.GetLink(flow.BranchNoMatch), nil
}
