package benthos

import (
	"strings"
	"testing"

	"github.com/eleon00/hsoetlnlm/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGenerateBenthosConfig_S3ToSnowflakeWithBloblang(t *testing.T) {
	// Arrange
	sourceConn := data.Connection{
		ID:               1,
		Name:             "Test S3 Source",
		Type:             "s3",
		ConnectionString: "bucket=my-test-bucket;region=us-east-1",
	}
	targetConn := data.Connection{
		ID:               2,
		Name:             "Test Snowflake Target",
		Type:             "snowflake",
		ConnectionString: "account=myacc;user=testuser;database=testdb;schema=public;table=test_table;password=dummy", // Using dummy password for test
	}
	task := data.ReplicationTask{
		ID:                    101,
		Name:                  "S3 to Snowflake Task",
		SourceConnectionID:    sourceConn.ID,
		TargetConnectionID:    targetConn.ID,
		DataSelectionCriteria: "input-data/",                        // S3 prefix
		TransformationRules:   `root = this.map_uppercase_values()`, // Simple Bloblang
	}

	// Act
	configYAML, err := GenerateBenthosConfig(task, sourceConn, targetConn)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, configYAML)

	// Unmarshal the YAML to verify its structure and content
	var configData map[string]interface{}
	err = yaml.Unmarshal([]byte(configYAML), &configData)
	require.NoError(t, err, "Generated YAML should be valid")

	// Check top-level keys
	assert.Contains(t, configData, "http")
	assert.Contains(t, configData, "input")
	assert.Contains(t, configData, "pipeline")
	assert.Contains(t, configData, "output")
	assert.Contains(t, configData, "logger")
	assert.Contains(t, configData, "metrics")

	// Check Input (S3)
	inputMap := configData["input"].(map[string]interface{})
	inputS3Map, ok := inputMap["aws_s3"].(map[string]interface{})
	require.True(t, ok, "Input section should contain aws_s3 config")
	assert.Equal(t, "my-test-bucket", inputS3Map["bucket"])
	assert.Equal(t, "us-east-1", inputS3Map["region"])
	assert.Equal(t, "input-data/", inputS3Map["prefix"])

	// Check Pipeline (Processors)
	pipelineSection, ok := configData["pipeline"].(map[string]interface{})
	require.True(t, ok, "Pipeline section should exist")
	processors, ok := pipelineSection["processors"].([]interface{})
	require.True(t, ok, "Processors section should be a slice")
	require.Len(t, processors, 1, "Should have one processor for the transformation rule")
	bloblangProcessorMap, ok := processors[0].(map[string]interface{})
	require.True(t, ok, "Processor should be a map")
	bloblangScript, ok := bloblangProcessorMap["bloblang"]
	require.True(t, ok, "Processor map should contain bloblang key")
	assert.Equal(t, task.TransformationRules, bloblangScript)

	// Check Output (Snowflake)
	outputMap := configData["output"].(map[string]interface{})
	outputSnowflakeMap, ok := outputMap["snowflake_put"].(map[string]interface{})
	require.True(t, ok, "Output section should contain snowflake_put config")
	assert.Equal(t, "myacc", outputSnowflakeMap["account"])
	assert.Equal(t, "testuser", outputSnowflakeMap["user"])
	assert.Equal(t, "testdb", outputSnowflakeMap["database"])
	assert.Equal(t, "public", outputSnowflakeMap["schema"])
	assert.Equal(t, "test_table", outputSnowflakeMap["table"])
	assert.Equal(t, "dummy", outputSnowflakeMap["password"])
	assert.Equal(t, "BENTHOS_STAGE", outputSnowflakeMap["stage_name"])
	assert.True(t, strings.Contains(outputSnowflakeMap["file_name_format"].(string), ".json.gz"))
}

func TestGenerateBenthosConfig_SQLServerToS3_NoTransforms(t *testing.T) {
	// Arrange
	sourceConn := data.Connection{
		ID:               3,
		Name:             "Test SQL Server Source",
		Type:             "sqlserver",
		ConnectionString: "dsn=sqlserver://user:pass@host:1433?database=sourcedb",
	}
	targetConn := data.Connection{
		ID:               4,
		Name:             "Test S3 Target",
		Type:             "s3",
		ConnectionString: "bucket=target-bucket;region=ap-southeast-2;path_prefix=sqlout/",
	}
	task := data.ReplicationTask{
		ID:                    102,
		Name:                  "SQL to S3 Task",
		SourceConnectionID:    sourceConn.ID,
		TargetConnectionID:    targetConn.ID,
		DataSelectionCriteria: "SELECT col1, col2 FROM source_table WHERE updated_at > ?", // Example query
		TransformationRules:   "",                                                         // No transformation
	}

	// Act
	configYAML, err := GenerateBenthosConfig(task, sourceConn, targetConn)

	// Assert
	require.NoError(t, err)
	require.NotEmpty(t, configYAML)

	var configData map[string]interface{}
	err = yaml.Unmarshal([]byte(configYAML), &configData)
	require.NoError(t, err, "Generated YAML should be valid")

	// Check Input (SQL Server)
	inputMap := configData["input"].(map[string]interface{})
	inputSQLMap, ok := inputMap["sql_select"].(map[string]interface{})
	require.True(t, ok, "Input section should contain sql_select config")
	assert.Equal(t, "mssql", inputSQLMap["driver"])
	assert.Equal(t, "sqlserver://user:pass@host:1433?database=sourcedb", inputSQLMap["dsn"])
	assert.Equal(t, task.DataSelectionCriteria, inputSQLMap["query"])

	// Check Pipeline (Processors)
	pipelineSection, ok := configData["pipeline"].(map[string]interface{})
	require.True(t, ok, "Pipeline section should exist")
	processors, ok := pipelineSection["processors"].([]interface{})
	require.True(t, ok, "Processors section should be a slice")
	assert.Empty(t, processors, "Should have no processors when TransformationRules is empty")

	// Check Output (S3)
	outputMap := configData["output"].(map[string]interface{})
	outputS3Map, ok := outputMap["aws_s3"].(map[string]interface{})
	require.True(t, ok, "Output section should contain aws_s3 config")
	assert.Equal(t, "target-bucket", outputS3Map["bucket"])
	assert.Equal(t, "ap-southeast-2", outputS3Map["region"])
	assert.True(t, strings.HasPrefix(outputS3Map["path"].(string), "sqlout/"))
}

func TestGenerateBenthosConfig_Error_MissingDSN(t *testing.T) {
	// Arrange
	sourceConn := data.Connection{
		ID:               5,
		Name:             "Bad SQL Server Source",
		Type:             "sqlserver",
		ConnectionString: "nodsn=here", // Missing dsn parameter
	}
	targetConn := data.Connection{ID: 4} // Minimal target needed
	task := data.ReplicationTask{
		ID:                    103,
		SourceConnectionID:    sourceConn.ID,
		TargetConnectionID:    targetConn.ID,
		DataSelectionCriteria: "SELECT 1",
	}

	// Act
	_, err := GenerateBenthosConfig(task, sourceConn, targetConn)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'dsn' not found")
}

func TestGenerateBenthosConfig_Error_UnsupportedSource(t *testing.T) {
	// Arrange
	sourceConn := data.Connection{
		ID:               6,
		Name:             "Unsupported Source",
		Type:             "some_future_db",
		ConnectionString: "foo=bar",
	}
	targetConn := data.Connection{ID: 4} // Minimal target needed
	task := data.ReplicationTask{
		ID:                 104,
		SourceConnectionID: sourceConn.ID,
		TargetConnectionID: targetConn.ID,
	}

	// Act
	_, err := GenerateBenthosConfig(task, sourceConn, targetConn)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported source connection type: some_future_db")
}

// TODO: Add more test cases:
// - Different source/target combinations (File -> Snowflake, etc.)
// - Missing or invalid connection string parameters for other types (S3 bucket, Snowflake table)
// - Empty DataSelectionCriteria where required
// - More complex TransformationRules (if applicable)
// - Cases where target connection types are unsupported
