package bigquery

import "cloud.google.com/go/bigquery"

// GetEventSchema returns the BigQuery schema for the 'aggregation' table
func GetAggregationSchema() bigquery.Schema {
	return bigquery.Schema{
		{Name: "Day", Type: bigquery.DateFieldType, Required: true},
		{Name: "ProjectID", Type: bigquery.IntegerFieldType, Required: true},
		{Name: "NumberOfTransactionsPerProject", Type: bigquery.IntegerFieldType, Required: true},
		{Name: "TotalVolumePerProject", Type: bigquery.NumericFieldType, Required: true},
		{Name: "Currency", Type: bigquery.StringFieldType, Required: true},
	}
}
