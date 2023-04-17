import os
import pathlib
import tempfile

import boto3
import structlog

from app.adapters.s3_bounds_repository import S3BoundsRepository
from moto import mock_s3

LOGGER = structlog.getLogger()
BUCKET_NAME = "example_bucket"
local_write_dir = pathlib.Path("./tmp")
os.environ["AWS_DEFAULT_REGION"] = "eu-central-1"


TEST_KEY = "test_key"


@mock_s3
def test_download():
    s3_client = boto3.client("s3")
    s3_client.create_bucket(Bucket=BUCKET_NAME, CreateBucketConfiguration={"LocationConstraint": "eu-central-1"})

    shape_files = [f"{TEST_KEY}.shp", f"{TEST_KEY}.shx", f"{TEST_KEY}.dbf", f"{TEST_KEY}.prj"]

    with tempfile.TemporaryDirectory() as temp_dir:
        for shape_file in shape_files:
            file_path = pathlib.Path(temp_dir) / shape_file
            with open(file_path, "w") as f:
                f.write("Sample content")
            s3_client.upload_file(str(file_path), BUCKET_NAME, shape_file)

        repo = S3BoundsRepository(
            bucket_name=BUCKET_NAME,
            local_write_dir=pathlib.Path(temp_dir),
            logger=LOGGER,
        )

        expected_local_path = repo.local_shape_file_dir / f"{TEST_KEY}.shp"

        result = repo.download(TEST_KEY)

        assert expected_local_path == result
        assert expected_local_path.exists()
        for file in shape_files:
            local_path = repo.local_shape_file_dir / file
            assert local_path.exists()
            local_path.unlink()  # Clean up the downloaded files
