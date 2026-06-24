import sqlite3
import json
import boto3
import os
from datetime import datetime, timezone, timedelta

# --- Config ---
DB_PATH = "/home/user/Desktop/Projects/resourcemonitor/system-monitor/monitor.db"
S3_BUCKET = "saurav-system-monitor-pipeline"      
S3_PREFIX = "raw/metrics"

def get_last_hour_metrics(db_path: str) -> list[dict]:
    """Read last hour of metrics from SQLite."""
    one_hour_ago = (datetime.now(timezone.utc) - timedelta(hours=1)).strftime("%Y-%m-%d %H:%M:%S")

    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)  # read-only
    conn.row_factory = sqlite3.Row  # rows behave like dicts
    cur = conn.cursor()

    cur.execute("""
        SELECT *
        FROM system_metrics
        WHERE timestamp >= ?
        ORDER BY timestamp ASC
    """, (one_hour_ago,))

    rows = [dict(row) for row in cur.fetchall()]
    conn.close()
    return rows

def upload_to_s3(data: list[dict], bucket: str, prefix: str):
    """Upload metrics as a JSON file to S3, partitioned by date/hour."""
    if not data:
        print("No data found for the last hour. Skipping upload.")
        return

    now = datetime.now(timezone.utc)
    # Partition path: raw/metrics/year=2026/month=06/day=22/hour=10/metrics.json
    s3_key = (
        f"{prefix}/"
        f"year={now.strftime('%Y')}/"
        f"month={now.strftime('%m')}/"
        f"day={now.strftime('%d')}/"
        f"hour={now.strftime('%H')}/"
        f"metrics.json"
    )

    s3 = boto3.client("s3")
    s3.put_object(
        Bucket=bucket,
        Key=s3_key,
        Body="\n".join(json.dumps(row) for row in data),
        ContentType="application/json"
    )
    print(f"Uploaded {len(data)} records to s3://{bucket}/{s3_key}")

if __name__ == "__main__": # Runs only if dunder is main, ran from the terminal or something.
    rows = get_last_hour_metrics(DB_PATH)
    upload_to_s3(rows, S3_BUCKET, S3_PREFIX)