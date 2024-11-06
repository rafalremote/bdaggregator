package bigquery

import (
	"testing"

	"cloud.google.com/go/bigquery"
)

func TestGetAggregationSchema(t *testing.T) {
	expectedSchema := bigquery.Schema{
		{Name: "Day", Type: bigquery.DateFieldType, Required: true},
		{Name: "ProjectID", Type: bigquery.IntegerFieldType, Required: true},
		{Name: "NumberOfTransactionsPerProject", Type: bigquery.IntegerFieldType, Required: true},
		{Name: "TotalVolumePerProject", Type: bigquery.NumericFieldType, Required: true},
		{Name: "Currency", Type: bigquery.StringFieldType, Required: true},
	}

	schema := GetAggregationSchema()

	if len(schema) != len(expectedSchema) {
		t.Fatalf("expected schema length %d, got %d", len(expectedSchema), len(schema))
	}

	for i, field := range schema {
		expectedField := expectedSchema[i]
		if field.Name != expectedField.Name {
			t.Errorf("expected field name %s at index %d, got %s", expectedField.Name, i, field.Name)
		}
		if field.Type != expectedField.Type {
			t.Errorf("expected field type %s for field %s, got %s", expectedField.Type, field.Name, field.Type)
		}
		if field.Required != expectedField.Required {
			t.Errorf("expected field %s to be Required=%t, got Required=%t", field.Name, expectedField.Required, field.Required)
		}
	}
}
