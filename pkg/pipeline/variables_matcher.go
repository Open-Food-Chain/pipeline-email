package pipeline

import (
	"github.com/jmoiron/jsonq"
	"github.com/The-New-Fork/pipeline/pkg/xos"
	"github.com/unchainio/pkg/errors"
	"strings"
)

func GetInputVariables(jq *jsonq.JsonQuery, variables map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})

	for k, v := range variables {
		res[k] = getVariableInterface(jq, v)
	}

	return res
}

func getVariableInterface(jq *jsonq.JsonQuery, variable interface{}) interface{} {
	switch cfg := variable.(type) {
	case map[string]interface{}:
		return GetInputVariables(jq, cfg)
	case []map[string]interface{}:
		res := make([]map[string]interface{}, len(cfg))
		for i, vv := range cfg {
			res[i] = GetInputVariables(jq, vv)
		}
		return res
	case []interface{}:
		res := make([]interface{}, len(cfg))
		for i, vv := range cfg {
			res[i] = getVariableInterface(jq, vv)
		}
		return res
	case string:
		return getVariableString(jq, cfg)
	default:
		return cfg
	}

}

const ExpansionStarterJSONPath = "$."

func getVariableString(jq *jsonq.JsonQuery, v string) interface{} {
	// The whole thing is a json path -> return whatever that path leads to as interface{}. This works for most things (slices, maps, etc.)
	if strings.HasPrefix(v, "$.") {
		return expandJSONPathToInterface(v[2:], jq)
	}

	// Expand combinations of constants, env-vars and json paths. This only works for json paths that lead to strings (does not work for slices, maps, etc.)
	v = xos.MultiExpand(v, []*xos.Expander{
		{
			StartString: ExpansionStarterJSONPath,
			Fn:          expandJSONPathToStringFn(jq),
		},
		{
			StartString: xos.ExpansionStarterEnvironmentVariable,
			Fn:          xos.EscapedGetEnv,
		},
	})

	return v
}

func expandJSONPathToInterface(v string, jq *jsonq.JsonQuery) interface{} {
	v = strings.TrimPrefix(v, "{")
	v = strings.TrimSuffix(v, "}")

	path := strings.Split(v, ".")
	// discard the $
	msg, err := jq.Interface(path...)
	if err != nil {
		err = errors.Wrap(err, "")
	}

	return msg
}

// expandJSONPaths expands the json path
func expandJSONPathToStringFn(jq *jsonq.JsonQuery) func(string) string {
	return func(s string) string {
		path := strings.Split(s, ".")

		msg, err := jq.String(path...)
		if err != nil {
			err = errors.Wrap(err, "")
		}

		return msg
	}
}
