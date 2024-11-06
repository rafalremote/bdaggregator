# Blockchain Data Aggregator for Marketplace Analytics

## Introduction

This project is a blockchain data aggregation pipeline designed to extract, normalize, and transform transaction data from a CSV file into a daily analytics table in **BigQuery**. 

The goal is to generate daily metrics for marketplace analytics, such as total volume in USD and the number of transactions per project. 

The data is extracted from a **Google Cloud Storage** bucket, transformed with **CoinGecko API** data, and stored in **BigQuery**, making it suitable for visualization and further analysis.

## Setup Instructions

### 1. Set Up Google Cloud Platform (GCP) Access

#### 1.1 Download and Install Google Cloud SDK
   - [Download the Google Cloud SDK](https://cloud.google.com/sdk/docs/install) for your operating system and follow the installation instructions.
   - After installation, initialize the SDK:
     ```bash
     gcloud init
     ```
   - Follow the prompts to log in to your Google account and set your default project.

#### 1.2 Create a Google Service Account with Required Permissions
   - Go to the [Google Cloud Console](https://console.cloud.google.com/).
   - Navigate to **IAM & Admin** > **Service Accounts**.
   - Click **Create Service Account** and provide a name for the account.
   - Grant the service account the following roles:
     - **BigQuery Data Editor** – for creating datasets and tables in BigQuery.
     - **Storage Object Viewer** – for accessing objects in Google Cloud Storage (GCS).
   - Once the account is created, go to **Keys**, click **Add Key** > **Create New Key**, and download the JSON key file.

#### 1.3 Authenticate the Service Account Locally
   - Set up the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to point to the path of your downloaded JSON key file:
     ```bash
     export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account-file.json"
     ```

#### 1.4 Enable Required APIs
   - Ensure the following APIs are enabled in your Google Cloud project:
     - **BigQuery API**
     - **Cloud Storage API**

### 2. Set Up CoinGecko API Access

#### 2.1 Create a CoinGecko Developer Account
   - Visit the [CoinGecko Developer Portal](https://www.coingecko.com/en/api) and sign up for an account if you don't already have one.
   - After logging in, you’ll receive an API key, which you’ll need to add to the `.env` file for currency exchange data.

### 3. Configure the `.env` File

Create a `.env` file in the root of the project and add the following variables. This file configures paths, database settings, and API endpoints for the data pipeline.

```bash
# Path to your Google service account credentials file
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account-file.json"

# Default currency for CoinGecko API
export SEQUENCE_DEFAULT_CURRENCY="usd"

# Database type and BigQuery settings
export SEQUENCE_DB_TYPE="BigQuery"
export SEQUENCE_BIGQUERY_PROJECT="<your-project-id>"
export SEQUENCE_BIGQUERY_DATASET="sequence"
export SEQUENCE_BIGQUERY_LOCATION="EU"

# Storage type and settings for Google Cloud Storage
export SEQUENCE_STORAGE_TYPE="GCS"
export SEQUENCE_GOOGLE_CLOUD_STORAGE_URL="https://storage.cloud.google.com/"
export SEQUENCE_GCS_BUCKET="<your-bucket-name>"
export SEQUENCE_GCS_OBJECT="sample_data.csv"

# For testing and demo purposes, specify a local file path
export SEQUENCE_LOCAL_STORAGE_PATH="sample_data.csv"

# Path to a file containing a list of CoinGecko currency IDs
export SEQUENCE_COINS_FILE_PATH="coins.json"

# CoinGecko API key and endpoint
export SEQUENCE_COINGECKO_API_KEY="<your-coingecko-api-key>"
export SEQUENCE_COINGECKO_API_URL="https://api.coingecko.com/api/v3/"
```

After creating and editing the .env file, load the environment variables:

```bash
source .env
```

### 4. Run the Project

#### Install Dependencies

```bash
go mod tidy
```

To run the project, use:

```bash
go run ./cmd
```

Alternatively, you can build an executable and then run it:

```bash
go build -o bdagg ./cmd
./bdagg
```

If you want to deploy the executable, please deploy along with `coins.json` (path to the file is set in ENV).

It's also possible to load CSV data from a local path, just set SEQUENCE_STORAGE_TYPE to `local` and point to file:

 ```bash
 export SEQUENCE_STORAGE_TYPE="local"
 export SEQUENCE_GCS_OBJECT="sample_data.csv"
 ```

### 5. Implementation details

#### Data Extraction 

To showcase the flexibility of this solution, we can choose between using `local` storage (to extract data from a locally stored file) or Google Cloud Storage. It’s also easy to add new storage types by implementing NewStorage in `internal/storage/factory.go` to meet the requirements of the Storage interface.

While extracting data, I build a list of events:

```go
type Event struct {
	Ts                   time.Time
	TsUnix               int64
	Event                string
	ProjectID            int
	CurrencySymbol       string
	CoinID               string
	CurrencyExchangeRate decimal.Decimal
	CurrencyValueDecimal decimal.Decimal
}
```

You may notice that I store additional fields not directly extracted from the source file:

- `TsUnix`: the event timestamp in Unix format.
- `CoinID`: a placeholder for the CoinID retrieved from coins.json.
- `CurrencyExchangeRate`: a placeholder for the currency rate retrieved from the CoinGecko API.

During data extraction, I also build a list of currencies used, along with the date range during which each currency was utilized. This approach saves time since we don’t need to reprocess the list of events, and to make it even faster, I run concurrent workers for this task.

As a result of data extraction, I have a list of events and a map of currency usage:

```go
type CurrencyUsage struct {
	From int64
	To   int64
}

type CurrencyUsageMap map[string]CurrencyUsage
```

`From` and `To` are in `int64` format as they store timestamps.

#### Normalization

To prepare the final list of events before aggregation, I need to fill in missing information:

- `CoinID`
- `Exchange rate` of the coin for each specific day

There are two reasons why we need CoinID:

- The currency symbol is not unique (e.g., "SFL" is used by multiple coins).
- To query the CoinGecko API, we need the CoinID, not the symbol.

A list of currency symbols used by coins rarely changes, so I downloaded that list from CoinGecko and stored it in `coins.json`. This approach saves CoinGecko API credits and reduces bandwidth usage. 

To ensure the correct `CoinID` for a given `currencySymbol`, I also double-check the result with `currencyAddress` when available. Additionally, there is an exception for **MATIC** where I verify with the `chainId`, representing the network (**137** for **Polygon**).

To further reduce the number of calls to CoinGecko, I use the `CurrencyUsageMap`built in the previous step. Since CoinGecko provides historical data within a time range, I don’t need to make a call per day but can request data for an entire time range.

To improve efficiency, I run concurrent workers to fetch exchange rates. Storing timestamps in Unix format also allows me to easily locate the nearest timestamp in the CoinGecko API results.

#### Calculations and Aggregation

The goal is to flatten the file into the following table structure:

- date: Aggregated at the day level from ts.
- project_id
- Number of transactions (aggregated from the sample data).
- Total volume in USD.

Sample result:

```
Day          ProjectID    Numbe rOfTransactionsPerProject    TotalVolumePerProject    Currency
2024-04-01   4974         100                               20492271.890000001       usd
2024-04-01   1609         13                                24.03                    usd
2024-04-01   0            102                               410.98                   usd
2024-04-02   1609         9                                 22.23                    usd
2024-04-02   4974         97                                1089939.62               usd
2024-04-02   0            104                               470.29                   usd
...

```

I assumed we should calculate the total volume and number of transactions for each `project_id`per day. This assumption is safe since calculating total volume and number of transactions per day in BigQuery is straightforward:


```sql
SELECT
  Day,
  SUM(NumberOfTransactionsPerProject) AS TotalTransactions,
  SUM(TotalVolumePerProject) AS TotalVolume,
  Currency
FROM
  `sequence-take-home-exercise.sequence.aggregation`
GROUP BY
  Day, Currency
ORDER BY
  Day;
```

Example result:

```
Day          TotalTransactions    TotalVolume           Currency
2024-04-01   215                 20492706.900000001    usd
2024-04-02   210                 1090432.14            usd
2024-04-15   466                 621.4                 usd
2024-04-16   109                 68.48                 usd
```

There’s a reason I include the currency in the table. By changing the exchange rate currency in the **configuration file**, we can easily generate results in **EUR**, **PLN**, or any other supported currency. This approach enhances the solution’s flexibility and configurability.

While multiplying the exchange rate by `CurrencyValueDecimal`, I apply a correction specifically for MATIC. This is necessary because MATIC values are stored in wei (the smallest unit), so I divide by 10^18 to convert it back into standard MATIC units.

To avoid data duplication, I check for existing data for the given day and project_id, updating it if the data already exists.

### The entire pipeline consistently finishes in less than 5 seconds on my laptop. ###


#### Visualization

BigQuery makes it easy to visualize the stored data. I successfully tested this visualization immediately after the results were upserted to BigQuery.