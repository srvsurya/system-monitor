import sys
from awsglue.transforms import *
from awsglue.utils import getResolvedOptions
from pyspark.context import SparkContext
from awsglue.context import GlueContext
from awsglue.job import Job
from pyspark.sql import functions as F
from pyspark.sql.window import Window

# --- Init ---
args = getResolvedOptions(sys.argv, ['JOB_NAME'])
sc = SparkContext()
glueContext = GlueContext(sc)
spark = glueContext.spark_session
job = Job(glueContext)
job.init(args['JOB_NAME'], args)

# --- Config ---
RAW_S3_PATH = "s3://saurav-system-monitor-pipeline/raw/metrics/"
PROCESSED_S3_PATH = "s3://saurav-system-monitor-pipeline/processed/metrics/"

# --- Step 1: Read raw data from S3 ---
df = spark.read.json(RAW_S3_PATH)

# --- Step 2: Clean ---
# drop rows where critical fields are null
df = df.dropna(subset=["cpu_usage", "memory_used", "timestamp"])

# drop duplicates
df = df.dropDuplicates(["id"])

# cast timestamp string to actual timestamp type
df = df.withColumn("timestamp", F.to_timestamp("timestamp", "yyyy-MM-dd HH:mm:ss"))

# --- Step 3: Enrich ---
# memory usage as percentage
df = df.withColumn(
    "memory_usage_pct",
    F.round((F.col("memory_used") / F.col("memory_total")) * 100, 2)
)

# disk usage as percentage
df = df.withColumn(
    "disk_usage_pct",
    F.round((F.col("disk_used") / F.col("disk_total")) * 100, 2)
)

# extract hour from timestamp for partitioning
df = df.withColumn("hour", F.hour("timestamp"))
df = df.withColumn("day", F.dayofmonth("timestamp"))
df = df.withColumn("month", F.month("timestamp"))
df = df.withColumn("year", F.year("timestamp"))

# --- Step 4: Anomaly flagging ---
# same logic as your Smart Heal — flag if cpu > 2x rolling average
window = Window.orderBy("timestamp").rowsBetween(-10, 0)

df = df.withColumn("cpu_rolling_avg", F.avg("cpu_usage").over(window))
df = df.withColumn(
    "cpu_anomaly",
    F.when(F.col("cpu_usage") > F.col("cpu_rolling_avg") * 2, True).otherwise(False)
)

# --- Step 5: Hourly aggregations ---
hourly = df.groupBy("year", "month", "day", "hour").agg(
    F.round(F.avg("cpu_usage"), 2).alias("avg_cpu"),
    F.round(F.max("cpu_usage"), 2).alias("max_cpu"),
    F.round(F.avg("memory_usage_pct"), 2).alias("avg_memory_pct"),
    F.round(F.avg("disk_usage_pct"), 2).alias("avg_disk_pct"),
    F.round(F.avg("net_upload"), 4).alias("avg_net_upload"),
    F.round(F.avg("net_download"), 4).alias("avg_net_download"),
    F.sum(F.col("cpu_anomaly").cast("int")).alias("anomaly_count"),
    F.count("id").alias("record_count")
)

# --- Step 6: Write output ---
# write full enriched records
df.write.mode("overwrite") \
    .partitionBy("year", "month", "day", "hour") \
    .parquet(PROCESSED_S3_PATH + "records/")

# write hourly aggregations separately
hourly.write.mode("overwrite") \
    .partitionBy("year", "month", "day") \
    .parquet(PROCESSED_S3_PATH + "hourly_agg/")

job.commit()