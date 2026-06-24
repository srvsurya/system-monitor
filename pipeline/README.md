# System Monitor — Data Pipeline
 
An end-to-end AWS data pipeline built on top of the [system-monitor](https://github.com/srvsurya/system-monitor) application. Exports live system metrics from SQLite, transforms them using PySpark on AWS Glue, and makes them queryable via Athena — orchestrated by Apache Airflow.
 
---
 
## Architecture
 
```
Local Machine                          AWS
─────────────                          ───
system-monitor (Go backend)
SQLite metrics DB
       │
       ▼
  exporter.py  ──────────────────→  S3 (raw NDJSON, partitioned by hour)
                                         │
                                         ▼
                                    Glue Crawler
                                    (schema catalog)
                                         │
                                         ▼
                                    Glue Job (PySpark)
                                    - clean + deduplicate
                                    - enrich (memory %, disk %)
                                    - flag CPU anomalies (2x rolling avg)
                                    - hourly aggregations
                                         │
                                         ▼
                                    S3 (processed Parquet, partitioned)
                                         │
                                         ▼
                                    AWS Athena
                                    (SQL analytics layer)
 
Airflow DAG orchestrates all stages hourly.
```
 
---
 
## Stack
 
| Tool | Role |
|---|---|
| Python / boto3 | Metrics exporter |
| AWS S3 | Raw and processed data storage |
| AWS Glue Crawler | Schema inference and cataloging |
| AWS Glue Job (PySpark) | Data transformation and anomaly detection |
| Apache Airflow 2.9 | Pipeline orchestration |
| AWS Athena | Queryable analytics layer |
 
---
 
## Pipeline Stages
 
### Stage 1 — Exporter (`exporter.py`)
Reads the last hour of metrics from the system-monitor SQLite database and uploads them to S3 as newline-delimited JSON (NDJSON), partitioned by `year/month/day/hour`.
 
```
s3://saurav-system-monitor-pipeline/raw/metrics/
└── year=2026/month=06/day=24/hour=11/metrics.json
```
 
### Stage 2 — Glue Crawler
Scans the raw S3 path, infers the schema (column names, types, partitions), and registers a table in the Glue Data Catalog. Only needed when schema changes — partition syncing is handled by Athena's `MSCK REPAIR TABLE` in the DAG.
 
### Stage 3 — PySpark Glue Job (`glue_job.py`)
Runs on a managed Spark cluster via AWS Glue. Performs:
- Null dropping and deduplication by `id`
- Timestamp casting to proper type
- Computed columns: `memory_usage_pct`, `disk_usage_pct`
- CPU anomaly flagging using a 10-sample rolling average window (same logic as Smart Heal)
- Hourly aggregations: avg/max CPU, avg memory %, avg disk %, anomaly count
Writes two outputs to S3 as Parquet:
```
processed/metrics/records/       ← full enriched row-level data
processed/metrics/hourly_agg/    ← rolled-up hourly summaries
```
 
### Stage 4 — Airflow DAG (`airflow/dags/system_monitor_dag.py`)
Orchestrates the full pipeline on an hourly schedule:
 
```
export_metrics
      │
data_quality_gate  (skips pipeline if no data in S3)
      │
sync_new_partitions  (MSCK REPAIR TABLE via Athena)
      │
run_glue_job
```
 
DAG is written to be MWAA-compatible — deploying to AWS Managed Airflow requires only uploading the DAG file to an S3-backed DAGs folder.

## Project Structure
 
```
pipeline/
├── exporter.py                        # Stage 1: SQLite → S3
├── glue_job.py                        # Stage 3: PySpark transformation
├── airflow/
│   └── dags/
│       └── system_monitor_dag.py      # Stage 4: Airflow orchestration
└── README.md
```
 
---
 
## IAM Setup
 
| Resource | Role / User |
|---|---|
| CLI access | `pipeline-cli-user` (S3, Glue, Athena permissions) |
| Glue Crawler + Job | `AWSGlueServiceRole-system-monitor` |
 
---
 
## Local Setup
 
```bash
# create and activate virtual environment
python3 -m venv venv
source venv/bin/activate
 
# install dependencies
pip install boto3 apache-airflow==2.9.3 apache-airflow-providers-amazon
 
# configure AWS credentials
aws configure
 
# run exporter manually
python exporter.py
 
# start Airflow
export AIRFLOW_HOME="$(pwd)/airflow"
export AIRFLOW__WEBSERVER__WEB_SERVER_PORT=8081
airflow db init
airflow standalone
```
 
---
 
## Notes
 
- Airflow is configured for local development using `standalone` mode. The DAG structure is identical to what would be deployed on AWS MWAA.
- The Glue Crawler is run once manually to establish the initial schema. Subsequent partition updates are handled cheaply via `MSCK REPAIR TABLE` in Athena.
- PySpark anomaly detection mirrors the rolling average logic used in the system-monitor Smart Heal feature.
