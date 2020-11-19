package pipeline_test

import (
	"github.com/The-New-Fork/email-pipeline/pkg/pipeline"
	"github.com/jmoiron/jsonq"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_GetInputVariables_Success(t *testing.T) {
	input := map[string]interface{}{
		"a": "value0",
		"b": "value1",
	}
	variables := map[string]interface{}{
		"key0": "$.a",
		"key1": "$.1",
	}
	expectedOutput := map[string]interface{}{
		"key0": "value0",
		"key1": "value1",
	}

	output := pipeline.GetInputVariables(jsonq.NewQuery(input), variables)

	require.Equal(t, expectedOutput, output)
}
