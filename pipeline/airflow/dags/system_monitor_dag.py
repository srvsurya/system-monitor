from datetime import datetime, timedelta
from airflow import DAG
from airflow.operators.python import PythonOperator, ShortCircuitOperator
from airflow.providers.amazon.aws.operators.glue import GlueJobOperator
from airflow.providers.amazon.aws.operators.athena import AthenaOperator
import subprocess
import boto3

default_args = {
    'owner': 'airflow',
    'start_date': datetime(2026, 1, 1),
    'retries': 1,
    'retry_delay': timedelta(minutes=2),
    'execution_timeout': timedelta(minutes=15),
}

def run_exporter():
    """Run the exporter script to push metrics to S3."""
    result = subprocess.run(
        ["/home/user/Desktop/Projects/resourcemonitor/system-monitor/venv/bin/python",
         "/home/user/Desktop/Projects/resourcemonitor/system-monitor/pipeline/exporter.py"],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        raise Exception(f"Exporter failed: {result.stderr}")
    print(result.stdout)

def check_data_quality():
    """Gate: check S3 has data before running Glue."""
    s3 = boto3.client('s3')
    response = s3.list_objects_v2(
        Bucket='saurav-system-monitor-pipeline',
        Prefix='raw/metrics/'
    )
    count = response.get('KeyCount', 0)
    if count == 0:
        print("No data in S3 — skipping pipeline.")
        return False
    print(f"Found {count} objects in S3 — proceeding.")
    return True

with DAG(
    dag_id='system_monitor_pipeline',
    default_args=default_args,
    schedule=None,
    catchup=False,
    description='Export metrics → S3 → Glue Crawler → PySpark transform',
) as dag:

    export_metrics = PythonOperator(
        task_id='export_metrics',
        python_callable=run_exporter,
    )

    data_quality_gate = ShortCircuitOperator(
        task_id='data_quality_gate',
        python_callable=check_data_quality,
    )

    sync_partitions = AthenaOperator(
    task_id='sync_new_partitions',
    query='MSCK REPAIR TABLE metrics;',
    database='monitor-catalog-db',
    output_location='s3://saurav-system-monitor-pipeline/athena-queries/',
    aws_conn_id='aws_default',
    )

    run_glue_job = GlueJobOperator(
        task_id='run_glue_job',
        job_name='system-metrics-job',
        aws_conn_id='aws_default',
    )

    export_metrics >> data_quality_gate >> sync_partitions >> run_glue_job