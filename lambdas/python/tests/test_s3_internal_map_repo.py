import datetime
import pathlib
import tempfile

import boto3
import structlog
from pytest import fixture
from moto import mock_sqs, mock_s3

from app.adapters.s3_internal_map_repository import S3InternalMapRepository, NewMessageNotificationQueue, construct_key
from app.core import domain
from app.core.domain import LocalMap
from generated.conflict_nightlight import v1
from generated.conflict_nightlight.v1 import CreateMapProductRequest

LOGGER = structlog.getLogger()

QUEUE_NAME = "example_queue"
BUCKET_NAME = "example_bucket"


@fixture
def test_map() -> domain.Map:
    return domain.Map(
        date=datetime.date(day=1, month=2, year=2023),
        map_type=v1.MapType.MAP_TYPE_DAILY,
        bounds=v1.Bounds.BOUNDS_UKRAINE_AND_AROUND,
        map_source=v1.MapSource(
            map_provider=v1.MapProvider.MAP_PROVIDER_EOGDATA,
            url="http://example.com",
        ),
    )


@fixture
def temp_example_tif():
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_dir_path = pathlib.Path(temp_dir)
        example_tif_path = temp_dir_path / "example.tif"
        with open(example_tif_path, "w") as f:
            f.write("Sample content")
        yield example_tif_path


@mock_s3
@mock_sqs
def test_s3_internal_map_repository_save(test_map, temp_example_tif):
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
        local_write_dir=pathlib.Path("/tmp"),
        new_message_notification=NewMessageNotificationQueue(
            queue_name=QUEUE_NAME, message_format=CreateMapProductRequest
        ),
    )
    repo.save(
        LocalMap(
            map=test_map,
            file_path=temp_example_tif,
        )
    )
    messages = queue.receive_messages()
    assert len(messages) == 1
    assert sum(1 for _ in bucket.objects.all()) == 1
    assert [
        f.key.split("/")[-1]
        for f in bucket.objects.filter(Prefix="MapProviderEogdata/BoundsUkraineAndAround/MapTypeDaily")
    ][0] == "2023_2_1.tif"


@mock_s3
def test_download(test_map, temp_example_tif):
    s3_client = boto3.client("s3")
    s3_client.create_bucket(Bucket=BUCKET_NAME, CreateBucketConfiguration={"LocationConstraint": "eu-central-1"})

    object_key = construct_key(test_map)
    s3_client.upload_file(Filename=str(temp_example_tif), Bucket=BUCKET_NAME, Key=object_key)
    repo = S3InternalMapRepository(
        logger=LOGGER,
        bucket_name=BUCKET_NAME,
        correlation_id="1",
        local_write_dir=temp_example_tif.parent,
        new_message_notification=NewMessageNotificationQueue(
            queue_name=QUEUE_NAME, message_format=CreateMapProductRequest
        ),
    )
    expected_local_map = domain.LocalMap(map=test_map, file_path=repo.local_write_dir / "temp.tif")

    result = repo.download(test_map)

    assert expected_local_map == result
    assert expected_local_map.file_path.exists()
    expected_local_map.file_path.unlink()  # Clean up the downloaded file
