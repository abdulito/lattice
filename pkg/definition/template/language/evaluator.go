package language

import (
	"fmt"
)

// OperatorEvaluator operator evaluators used by the engine to evaluate special operators

type OperatorEvaluator interface {
	eval(o interface{}, env *environment) (interface{}, error)
}

type operatorConfig struct {
	key                   string
	evaluator             OperatorEvaluator
	appendToPropertyStack bool
}

// IncludeEvaluator. evaluates $include
type IncludeEvaluator struct {
}

func (evaluator *IncludeEvaluator) eval(o interface{}, env *environment) (interface{}, error) {

	// construct the include object. We allow the include to be an object or a string.
	// string will be converted to to {url: val}
	var includeObject map[string]interface{}
	if _, isMap := o.(map[string]interface{}); isMap {
		includeObject = o.(map[string]interface{})
	} else if _, isString := o.(string); isString {
		includeObject = map[string]interface{}{
			"url": o,
		}
	} else {
		return nil, fmt.Errorf("Invalid $include %s", includeObject)
	}

	// validate include object
	if _, hasUrl := includeObject["url"]; !hasUrl {
		return nil, fmt.Errorf("$include has no url %s", includeObject)
	}

	//evaluate parameters if present

	var includeParameters map[string]interface{}
	if includeParamsVal, hasParams := includeObject["parameters"]; hasParams {
		var err error
		params, err := env.engine.eval(includeParamsVal, env)
		if err != nil {
			return nil, err
		}

		includeParameters = params.(map[string]interface{})
	}

	url := includeObject["url"].(string)

	// return the included object
	return env.engine.include(url, includeParameters, env)
}

// VariablesEvaluator. evaluates $variables
type VariablesEvaluator struct {
}

// eval
func (evaluator *VariablesEvaluator) eval(o interface{}, env *environment) (interface{}, error) {

	// validate that variables is a map
	if _, isMap := o.(map[string]interface{}); !isMap {
		return nil, fmt.Errorf("bad '$variables' value '%v'. Must be a map", o)
	}

	variables, err := env.engine.eval(o, env)
	if err != nil {
		return nil, err
	}

	// set current variables in env to result
	env.currentFrame().variables = variables.(map[string]interface{})

	// return nil as the evaluation result to indicate to the engine to continue evaluation without assigning any
	// result back to the current key being evaluated
	return nil, nil

}

// ParametersEvaluator. evaluates $parameters
type ParametersEvaluator struct {
}

// eval
func (evaluator *ParametersEvaluator) eval(o interface{}, env *environment) (interface{}, error) {
	paramMap, isMap := o.(map[string]interface{})

	// validate that parameters are passed as a map
	if !isMap {
		return nil, fmt.Errorf("bad '$parameters' value '%v'. Must be a map", o)
	}

	for name, paramDef := range paramMap {
		err := evaluator.processInputParameter(name, paramDef.(map[string]interface{}), env)
		if err != nil {
			return nil, err
		}
	}

	// return nil as the evaluation result to indicate to the engine to continue evaluation without assigning any
	// result back to the current key being evaluated
	return nil, nil

}

// processInputParameter process/validate template parameters
func (evaluator *ParametersEvaluator) processInputParameter(name string, paramDef map[string]interface{}, env *environment) error {
	parameters := env.currentFrame().parameters
	// validate required
	if isRequiredVal, requiredIsSet := paramDef["required"]; requiredIsSet {
		if isRequired, requiredIsBool := isRequiredVal.(bool); requiredIsBool {
			if _, paramIsSet := parameters[name]; isRequired && !paramIsSet {
				return fmt.Errorf("parameter %s is required", name)
			}
		} else {
			return fmt.Errorf("bad 'required' value '%v'. Must be a boolean", isRequiredVal)
		}

	}
	// default param as needed
	if defaultValue, hasDefault := paramDef["default"]; hasDefault {
		if _, paramIsSet := parameters[name]; !paramIsSet {
			parameters[name] = defaultValue
		}
	}

	return nil

}
