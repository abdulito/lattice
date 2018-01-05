package language

import (
	"fmt"
	"regexp"
	"strings"
)

// Contains functions needed for evaluating string expressions

// used for matching a single parameter reference expression. i.e. ${var}
var varDefRegex = regexp.MustCompile(`\$\{([a-zA-Z_$][a-zA-Z_$.0-9]*)\}`)
var singleVarRegex = regexp.MustCompile(`^\$\{([a-zA-Z_$][a-zA-Z_$.0-9]*)\}$`)

// evalStringExpression
func evalStringExpression(expression string, parameters map[string]interface{}) (interface{}, error) {

	if isSingleParameterExpression(expression) {
		return evalSingleParameterExpression(expression, parameters)
	}

	// otherwise just return the expression as is
	return replaceAllParameters(expression, parameters), nil
}

// isSingleParameterExpression
func isSingleParameterExpression(expression string) bool {
	return singleVarRegex.MatchString(expression)
}

// evalSingleParameterExpression
func evalSingleParameterExpression(expression string, parameters map[string]interface{}) (interface{}, error) {
	parts := singleVarRegex.FindAllStringSubmatch(expression, -1)
	parameterName := parts[0][1]
	return parameters[parameterName], nil
}

// replaceAllParameters
func replaceAllParameters(expression string, parameters map[string]interface{}) string {
	varDefMatches := varDefRegex.FindAllStringSubmatch(expression, -1)
	result := expression
	for _, group := range varDefMatches {
		parameterName := group[1]
		result = replaceParameter(result, parameterName, parameters)

	}

	return result
}

// replaceParameter
func replaceParameter(expression string, parameterName string, parameters map[string]interface{}) string {
	varDef := fmt.Sprintf("${%s}", parameterName)
	val := getParameterStringValue(parameterName, parameters)
	return strings.Replace(expression, varDef, val, -1)
}

// getParameterStringValue
func getParameterStringValue(parameterName string, parameters map[string]interface{}) string {

	if val, exists := parameters[parameterName]; exists {
		return fmt.Sprintf("%v", val)
	}

	if strings.Contains(parameterName, ".") {
		parts := strings.Split(parameterName, ".")
		firstVar := parts[0]
		last := strings.Join(parts[1:], ".")

		if newParameters, exists := parameters[firstVar]; exists {
			if newParametersMap, isVarMap := newParameters.(map[string]interface{}); isVarMap {
				return getParameterStringValue(last, newParametersMap)
			}
		}
	}

	// Unable to determine parameter value
	return ""
}
