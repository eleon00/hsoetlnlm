package benthos

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3" // We'll use YAML for Benthos configs

	"github.com/eleon00/hsoetlnlm/internal/data" // Assuming models are in internal/data
)

// parseConnectionString splits a key-value string (e.g., "key1=val1;key2=val2") into a map.
func parseConnectionString(connStr string) map[string]string {
	params := make(map[string]string)
	pairs := strings.Split(connStr, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if key != "" {
				params[key] = value
			}
		}
	}
	return params
}

// GenerateBenthosConfig dynamically creates a Benthos configuration YAML string
// based on the replication task and connection details.
func GenerateBenthosConfig(task data.ReplicationTask, sourceConn data.Connection, targetConn data.Connection) (string, error) {
	// Basic Benthos config structure
	config := map[string]interface{}{
		"http": map[string]interface{}{
			"address": "0.0.0.0:4195", // Default Benthos metrics/API port
			"enabled": true,
		},
		"input": map[string]interface{}{}, // To be filled based on sourceConn
		"pipeline": map[string]interface{}{ // Pipeline section structure
			"processors": []interface{}{}, // Start with empty processors
			"threads":    -1,              // Use default thread count initially
		},
		"output": map[string]interface{}{}, // To be filled based on targetConn
		"logger": map[string]interface{}{
			"level": "INFO", // Default logging level
		},
		"metrics": map[string]interface{}{ // Use standard prometheus metrics
			"prometheus": map[string]interface{}{},
		},
		"tracer": map[string]interface{}{ // Disable tracing by default
			"none": map[string]interface{}{},
		},
	}

	// --- Input Configuration ---
	inputConfig, err := generateInputConfig(sourceConn, task)
	if err != nil {
		return "", fmt.Errorf("failed to generate input config: %w", err)
	}
	config["input"] = inputConfig

	// --- Output Configuration ---
	outputConfig, err := generateOutputConfig(targetConn, task)
	if err != nil {
		return "", fmt.Errorf("failed to generate output config: %w", err)
	}
	config["output"] = outputConfig

	// --- Processor Configuration ---
	if task.TransformationRules != "" {
		// Assuming TransformationRules contains Bloblang script
		pipeline := config["pipeline"].(map[string]interface{})
		processors := pipeline["processors"].([]interface{})
		processors = append(processors, map[string]interface{}{
			"bloblang": task.TransformationRules,
		})
		pipeline["processors"] = processors
	}

	// Marshal the map into a YAML string
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// generateInputConfig creates the Benthos input section based on the source connection.
func generateInputConfig(conn data.Connection, task data.ReplicationTask) (map[string]interface{}, error) {
	inputConf := map[string]interface{}{}
	params := parseConnectionString(conn.ConnectionString)

	// Example: Add logic based on conn.Type
	switch conn.Type {
	case "sqlserver", "oracle":
		driver := conn.Type
		if driver == "sqlserver" {
			driver = "mssql" // Benthos uses 'mssql'
		}
		dsn, ok := params["dsn"]
		if !ok {
			return nil, fmt.Errorf("'dsn' not found in connection string for %s", conn.Type)
		}
		// For simplicity, assume DataSelectionCriteria IS the query. Real world might need more parsing.
		query := task.DataSelectionCriteria
		if query == "" {
			return nil, fmt.Errorf("DataSelectionCriteria (query) cannot be empty for %s input", conn.Type)
		}

		inputConf["sql_select"] = map[string]interface{}{ // Use sql_select for pulling data
			"driver":       driver,
			"dsn":          dsn,
			"query":        query,
			"args_mapping": "", // No args for now
		}
	case "s3":
		bucket, ok := params["bucket"]
		if !ok {
			return nil, fmt.Errorf("'bucket' not found in connection string for S3")
		}
		prefix := task.DataSelectionCriteria // Use criteria as the prefix
		region, ok := params["region"]
		if !ok {
			// Attempt to infer region or use a default; Benthos might handle this too
			region = "" // Let Benthos SDK try to determine
		}

		s3Input := map[string]interface{}{ // Use aws_s3 input
			"bucket": bucket,
			"prefix": prefix,
		}
		if region != "" {
			s3Input["region"] = region
		}
		// Credentials should ideally be handled via environment vars or AWS SDK defaults
		// Avoid putting explicit keys in the config if possible.
		// Example: s3Input["credentials"] = map[string]string{"profile": params["profile"]} if profile is specified
		inputConf["aws_s3"] = s3Input

	case "localfile":
		paths := strings.Split(task.DataSelectionCriteria, ",") // Allow comma-separated paths
		if len(paths) == 0 || paths[0] == "" {
			return nil, fmt.Errorf("DataSelectionCriteria (file paths) cannot be empty for localfile input")
		}
		codec, ok := params["codec"]
		if !ok {
			codec = "lines" // Default codec
		}
		inputConf["file"] = map[string]interface{}{ // Use file input
			"paths": paths,
			"codec": codec,
		}

	case "bigquery":
		project, ok := params["project"]
		if !ok {
			return nil, fmt.Errorf("'project' not found in connection string for BigQuery")
		}
		query := task.DataSelectionCriteria
		if query == "" {
			return nil, fmt.Errorf("DataSelectionCriteria (query) cannot be empty for BigQuery input")
		}

		inputConf["gcp_bigquery_select"] = map[string]interface{}{ // Use gcp_bigquery_select input
			"project": project,
			"query":   query,
			// Credentials should ideally be handled via environment vars or GCP SDK defaults
		}
	default:
		return nil, fmt.Errorf("unsupported source connection type: %s", conn.Type)
	}

	return inputConf, nil
}

// generateOutputConfig creates the Benthos output section based on the target connection.
func generateOutputConfig(conn data.Connection, task data.ReplicationTask) (map[string]interface{}, error) {
	outputConf := map[string]interface{}{}
	params := parseConnectionString(conn.ConnectionString)

	switch conn.Type {
	case "snowflake":
		account, _ := params["account"]
		user, _ := params["user"]
		password, _ := params["password"] // WARNING: Password in connection string is insecure!
		database, _ := params["database"]
		schema, _ := params["schema"]
		table, ok := params["table"] // Expect target table in connection string for now
		if !ok {
			return nil, fmt.Errorf("'table' not found in connection string for Snowflake output")
		}

		snowflakeOutput := map[string]interface{}{ // Use snowflake_put output
			"account":          account,
			"user":             user,
			"database":         database,
			"schema":           schema,
			"table":            table,
			"stage_name":       "BENTHOS_STAGE", // Default stage name
			"file_name_format": `${!count("files")}-${!timestamp_unix_nano()}.json.gz`,
		}
		// Use password only if provided - prefer key pair auth or other methods
		if password != "" {
			snowflakeOutput["password"] = password
		}
		// TODO: Add role, warehouse, key pair auth details if needed based on params
		outputConf["snowflake_put"] = snowflakeOutput

	case "s3":
		bucket, ok := params["bucket"]
		if !ok {
			return nil, fmt.Errorf("'bucket' not found in connection string for S3 output")
		}
		pathPrefix, ok := params["path_prefix"]
		if !ok {
			pathPrefix = "output/" // Default prefix
		}
		region, ok := params["region"]
		if !ok {
			region = "" // Let Benthos SDK try to determine
		}

		s3Output := map[string]interface{}{ // Use aws_s3 output
			"bucket": bucket,
			"path":   fmt.Sprintf(`%s${!count("files")}-${!timestamp_unix_nano()}.json`, pathPrefix),
			"batching": map[string]interface{}{ // Enable batching for S3 efficiency
				"count":  100,
				"period": "1s",
			},
		}
		if region != "" {
			s3Output["region"] = region
		}
		// Handle credentials like in input
		outputConf["aws_s3"] = s3Output

	default:
		return nil, fmt.Errorf("unsupported target connection type: %s", conn.Type)
	}
	return outputConf, nil
}
