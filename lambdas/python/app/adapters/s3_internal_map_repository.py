import os
import pathlib
from dataclasses import dataclass
from typing import Optional, Type

import boto3
from botocore.exceptions import ClientError

from app.core import domain, ports
from app.core.ports import InternalMapRepository
from app.infrastructure.proto_transformers import (
    bounds_to_string,
    map_provider_to_string,
    map_type_to_string,
    transform_map_domain_to_proto,
)
from generated.conflict_nightlight.v1 import CreateMapProductRequest, PublishMapProductRequest, RequestWrapper


@dataclass
class NewMessageNotificationQueue:
    queue_name: str
    message_format: Optional[Type[CreateMapProductRequest] | Type[PublishMapProductRequest]]

    def validate(self):
        if not self.message_format:
            raise Exception("Message format not provided")
        if not self.queue_name or self.queue_name == "":
            raise Exception("Queue name not provided")


class S3InternalMapRepository(InternalMapRepository):
    def __init__(
        self,
        logger: ports.Logger,
        bucket_name: str,
        correlation_id: str,
        local_write_dir: pathlib.Path,
        new_message_notification: Optional[NewMessageNotificationQueue],
        source_url_key=os.getenv("SOURCE_URL_KEY", "source-url"),
    ):
        self.logger = logger
        self.s3_client = boto3.client("s3")
        self.sqs_client = boto3.client("sqs")
        self.local_write_dir = local_write_dir
        self.new_message_notification = new_message_notification
        self.bucket_name = bucket_name
        self.correlation_id = correlation_id
        self.source_url_key = source_url_key

    def download(self, m: domain.Map) -> domain.LocalMap:
        self.logger.info("Downloading map from s3", bucket_name=self.bucket_name, map=m)
        try:
            key = construct_key(m)
        except ErrorConstructKey as e:
            self.logger.fatal(str(e))
        return domain.LocalMap(map=m, file_path=self._download_file(key))

    def save(self, m: domain.LocalMap):
        self.logger.info("Attempting to save map to s3", bucket_name=self.bucket_name, map=m)
        try:
            self.new_message_notification.validate()
        except Exception as e:
            self.logger.fatal(str(e))
        try:
            self.s3_client.upload_file(
                Filename=str(m.file_path),
                Bucket=self.bucket_name,
                Key=construct_key(m.map),
                ExtraArgs={"Metadata": {self.source_url_key: m.map.map_source.url}},
            )
        except ErrorConstructKey as e:
            self.logger.fatal(str(e))
        except Exception as e:
            self.logger.fatal("Error when trying to upload map", map=m, bucket=self.bucket_name, error=e)
        self._send_sqs_message(m.map)

    def _download_file(self, object_key: str) -> pathlib.Path:
        self.local_write_dir.mkdir(parents=True, exist_ok=True)
        local_name = self.local_write_dir / "temp.tif"
        try:
            self.s3_client.download_file(self.bucket_name, object_key, local_name)
        except ClientError as e:
            self.logger.fatal(
                "Error while trying to download a file from s3", bucket_name=self.bucket_name, key=object_key, error=e
            )
        except Exception as e:
            self.logger.fatal("Error while trying to download file from s3", error=e)
        return local_name

    def _send_sqs_message(self, m: domain.Map):
        message = self._create_message(m)
        self.sqs_client.send_message(
            QueueUrl=self.new_message_notification.queue_name,
            MessageBody=message.to_json(),
            MessageAttributes={
                os.getenv("CORRELATION_ID_KEY_NAME", "correlation-id"): {
                    "StringValue": self.correlation_id,
                    "DataType": "String",
                }
            },
        )
        self.logger.info(
            "Successfully sent message to sqs",
            message=message.to_json(),
            queue_name=self.new_message_notification.queue_name,
        )

    def _create_message(self, m: domain.Map) -> RequestWrapper:
        match self.new_message_notification.message_format():
            case CreateMapProductRequest():
                return RequestWrapper(
                    create_map_product_request=CreateMapProductRequest(transform_map_domain_to_proto(m))
                )
            case PublishMapProductRequest():
                return RequestWrapper(
                    publish_map_product_request=PublishMapProductRequest(transform_map_domain_to_proto(m))
                )
            case _:
                self.logger.fatal("Unknown message format", message_format=self.new_message_notification.message_format)


class ErrorConstructKey(Exception):
    pass


def construct_key(m: domain.Map) -> str:
    provider = map_provider_to_string(m.map_source.map_provider)
    map_type = map_type_to_string(m.map_type)
    bounds = bounds_to_string(m.bounds)
    if not provider or not map_type or not bounds:
        raise ErrorConstructKey(
            f"The key could not be constructed, provider={provider}, map_type={map_type}, bounds={bounds}"
        )
    return f"{provider}/{bounds}/{map_type}/{m.date.year}_{m.date.month}_{m.date.day}.tif"
