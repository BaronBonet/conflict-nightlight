import concurrent
import pathlib
from concurrent.futures import ThreadPoolExecutor  # noqa: F401

import boto3

from app.core import ports
from app.core.ports import BoundsRepository
from app.infrastructure.get_or_raise import get_or_raise


class S3BoundsRepository(BoundsRepository):
    def __init__(self, bucket_name: str, local_write_dir: pathlib.Path, logger: ports.Logger):
        logger.debug("Initiating S3BoundsRepository", bucket_name=bucket_name)
        self.local_write_dir = local_write_dir
        self.logger = logger
        self.bucket_name = bucket_name

        self.client = boto3.client("s3")

        self.local_shape_file_dir = local_write_dir / "shp"
        self.local_shape_file_dir.mkdir(parents=True, exist_ok=True)

    def download(self, key: str) -> pathlib.PosixPath | None:
        bucket_objects = self.client.list_objects(Bucket=self.bucket_name)
        if len(bucket_objects) <= 0:
            self.logger.fatal("There are not any shp files in the bucket", bucket_name=self.bucket_name)
        files_to_download = []
        for obj in bucket_objects["Contents"]:
            obj_key = get_or_raise(obj, "Key")
            if obj_key.startswith(key):
                files_to_download.append(obj_key)

        if len(files_to_download) != 4:
            self.logger.fatal(
                "The 4 shape files needed were not found",
                shape_files=files_to_download,
                key=key,
                bucket_name=self.bucket_name,
            )

        downloaded_files = self._download_concurrently(files_to_download)

        shp_file_name = None
        for file_name in downloaded_files:
            if file_name.suffix == ".shp":
                shp_file_name = file_name

        if not shp_file_name:
            self.logger.fatal("A .shp file was not found")

        return shp_file_name

    def _download_concurrently(self, files_to_download: list[str]) -> list[pathlib.PosixPath]:
        downloaded_files = []
        with concurrent.futures.ThreadPoolExecutor() as executor:
            future_to_key = {executor.submit(self._download_object, obj_key): obj_key for obj_key in files_to_download}
            for future in concurrent.futures.as_completed(future_to_key):
                obj_key = future_to_key[future]
                try:
                    downloaded_files.append(local_path := future.result())
                    self.logger.debug("Successfully downloaded object", object_key=obj_key, local_path=local_path)
                except Exception as e:
                    self.logger.error("Error downloading object", obj_key=obj_key, error=e)
        return downloaded_files

    def _download_object(self, obj_key: str) -> pathlib.Path:
        local_path = self.local_shape_file_dir / obj_key
        self.logger.info("Starting shape file download", local_path=local_path, obj_key=obj_key)
        self.client.download_file(self.bucket_name, obj_key, local_path)
        return local_path
