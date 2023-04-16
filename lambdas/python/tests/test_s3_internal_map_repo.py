import datetime
import os

import boto3
import structlog
from moto import mock_sqs, mock_s3

from app import BASE_DIR
from app.adapters.s3_internal_map_repository import S3InternalMapRepository, NewMessageNotificationQueue
from app.core.domain import LocalMap, Map
from generated.conflict_nightlight import v1
from generated.conflict_nightlight.v1 import CreateMapProductRequest

DATA_DIR = BASE_DIR / "tests" / "data"
LOGGER = structlog.getLogger()
QUEUE_NAME = "example_queue"
BUCKET_NAME = "example_bucket"
os.environ["AWS_DEFAULT_REGION"] = "eu-central-1"


@mock_s3
@mock_sqs
def test_s3_internal_map_repository_save():
    sqs_client = boto3.resource("sqs")
    queue = sqs_client.create_queue(QueueName=QUEUE_NAME)
    s3_client = boto3.resource("s3")
    bucket = s3_client.create_bucket(
        Bucket=BUCKET_NAME, CreateBucketConfiguration={"LocationConstraint": "eu-central-1"}
    )
    repo = S3InternalMapRepository(
        logger=LOGGER,
        bucket_name=BUCKET_NAME,
        correlation_id="1",
        local_write_dir=DATA_DIR,
        new_message_notification=NewMessageNotificationQueue(
            queue_name=QUEUE_NAME, message_format=CreateMapProductRequest
        ),
    )
    repo.save(
        LocalMap(
            map=Map(
                map_source=v1.MapSource(map_provider=v1.MapProvider.MAP_PROVIDER_EOGDATA, url="example"),
                map_type=v1.MapType.MAP_TYPE_MONTHLY,
                bounds=v1.Bounds.BOUNDS_UKRAINE_AND_AROUND,
                date=datetime.date(year=2021, month=1, day=1),
            ),
            file_path=DATA_DIR / "example.tif",
        )
    )
    messages = queue.receive_messages()
    assert len(messages) == 1
    assert sum(1 for _ in bucket.objects.all()) == 1
    assert [f.key.split("/")[-1] for f in bucket.objects.filter(Prefix="Eogdata/UkraineAndAround/Monthly")][
        0
    ] == "2021_1_1.tif"
