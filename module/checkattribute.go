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

func (m checkAttribute) Run(call CallConnector) (next *flow.ModuleID, err error) {
	if m.Type != flow.ModuleCheckAttribute {
		return nil, fmt.Errorf("module of type %s being run as checkAttribute", m.Type)
	}
	pr := parameterResolver{call}
	p := checkAttributeParams{}
	err = pr.unmarshal(m.Parameters, &p)
	if err != nil {
		return
	}
	v, err := pr.get(flow.ModuleParameterNamespace(p.Namespace), p.Attribute)
	if err != nil {
		return
	}
	vs := ""
	if v != nil {
		vs = *v
	}
	return evaluateConditions(m.Branches, vs)
}

func evaluateConditions(c flow.ModuleBranchList, v string) (*flow.ModuleID, error) {
	conditions := c.List(flow.BranchEvaluate)
	vn, err := strconv.ParseFloat(v, 64)
	numeric := bool(err == nil)
	for _, c := range conditions {
		val := fmt.Sprintf("%v", c.ConditionValue)
		cvn, err := strconv.ParseFloat(val, 64)
		numeric := numeric && bool(err == nil)
		pass := false
		switch c.ConditionType {
		case flow.ConditionEquals:
			pass = bool(v == val)
		case flow.ConditionGT:
			pass = bool((numeric && vn > cvn) || (!numeric && v > val))
		case flow.ConditionGTE:
			pass = bool((numeric && vn >= cvn) || (!numeric && v >= val))
		case flow.ConditionLT:
			pass = bool((numeric && vn < cvn) || (!numeric && v < val))
		case flow.ConditionLTE:
			pass = bool((numeric && vn <= cvn) || (!numeric && v <= val))
		default:
			return nil, fmt.Errorf("unhandled condition type: %s", c.ConditionType)
		}
		if pass {
			return &c.Transition, nil
		}
	}
	return c.GetLink(flow.BranchNoMatch), nil
}
