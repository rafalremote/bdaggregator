package bigquery

import (
	"bdaggregator/internal/config"
	"context"
	"fmt"
	"log"

	bg "cloud.google.com/go/bigquery"
	"google.golang.org/api/googleapi"
)

type BigQueryDB struct {
	client *bg.Client
	cfg    *config.Config
}

// Create a new instance of BigQueryDB
func NewBigQueryDB(ctx context.Context, cfg *config.Config) (*BigQueryDB, error) {
	client, err := bg.NewClient(ctx, cfg.BigQueryProject)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %v", err)
	}
	return &BigQueryDB{client: client, cfg: cfg}, nil
}

func (bq *BigQueryDB) SetupDatabase(ctx context.Context) error {
	dataset := bq.client.Dataset(bq.cfg.BigQueryDataset)
	_, err := dataset.Metadata(ctx)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			if err := dataset.Create(ctx, &bg.DatasetMetadata{Location: bq.cfg.BigQueryLocation}); err != nil {
				return fmt.Errorf("failed to create dataset: %v", err)
			}
			log.Println("Dataset created successfully.")
		} else {
			return fmt.Errorf("failed to get dataset metadata: %v", err)
		}
	}
	return nil
}

func (bq *BigQueryDB) SetupTable(ctx context.Context, tableName string) error {
	var bqSchema bg.Schema
	switch tableName {
	case "aggregation":
		bqSchema = GetAggregationSchema()
	default:
		return fmt.Errorf("unsupported table name: %s", tableName)
	}

	table := bq.client.Dataset(bq.cfg.BigQueryDataset).Table(tableName)
	_, err := table.Metadata(ctx)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			if err := table.Create(ctx, &bg.TableMetadata{Schema: bqSchema}); err != nil {
				return fmt.Errorf("failed to create table %s: %v", tableName, err)
			}
			log.Println("Table created successfully.")
		} else {
			return fmt.Errorf("failed to get table metadata for %s: %v", tableName, err)
		}
	}
	return nil
}

func (bq *BigQueryDB) Upsert(ctx context.Context, tableName string, records interface{}) error {
	table := fmt.Sprintf("`%s.%s.%s`", bq.cfg.BigQueryProject, bq.cfg.BigQueryDataset, tableName)

	query := bq.client.Query(fmt.Sprintf(`
		MERGE INTO %s AS target
		USING UNNEST(@records) AS source
		ON target.Day = DATE(source.Day) AND target.ProjectID = source.ProjectID
		WHEN MATCHED THEN
			UPDATE SET
				target.NumberOfTransactionsPerProject = source.NumberOfTransactionsPerProject,
				target.TotalVolumePerProject = CAST(source.TotalVolumePerProject AS NUMERIC),
				target.Currency = source.Currency
		WHEN NOT MATCHED THEN
			INSERT (Day, ProjectID, NumberOfTransactionsPerProject, TotalVolumePerProject, Currency)
			VALUES(DATE(source.Day), source.ProjectID, source.NumberOfTransactionsPerProject, CAST(source.TotalVolumePerProject AS NUMERIC), source.Currency)`, table))

	query.Parameters = []bg.QueryParameter{
		{Name: "records", Value: records},
	}

	job, err := query.Run(ctx)
	if err != nil {
		return fmt.Errorf("failed to run merge query: %v", err)
	}

	// Wait for the query to complete
	status, err := job.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for merge job completion: %v", err)
	}
	if err := status.Err(); err != nil {
		return fmt.Errorf("merge job failed with error: %v", err)
	}

	log.Println("Records successfully merged into BigQuery.")
	return nil
}

func (bq *BigQueryDB) Close() error {
	return bq.client.Close()
}
