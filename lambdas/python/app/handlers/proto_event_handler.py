import os
import pathlib

import betterproto

from app.adapters.eogdata_external_map_repository import EogdataMapRepository
from app.adapters.s3_bounds_repository import S3BoundsRepository
from app.adapters.s3_internal_map_repository import NewMessageNotificationQueue, S3InternalMapRepository
from app.adapters.struct_logger import StructLogger
from app.core import ports
from app.core.services.product import MapProductService
from app.core.services.raw_processor import RawMapProcessorService
from app.infrastructure.get_secrets_from_aws_secret_manager import get_secrets_from_aws_secrets_manager
from app.infrastructure.proto_transformers import proto_map_to_domain
from generated.conflict_nightlight.v1 import CreateMapProductRequest, PublishMapProductRequest, RequestWrapper


def handle_event(event: dict[str, str], correlation_id: str):
    write_dir = pathlib.Path(os.getenv("LOCAL_WRITE_DIRECTORY", "/tmp"))
    # TODO: Use hexalog
    logger = StructLogger(
        version=os.getenv("VERSION", "unknown"),
        use_debug=os.getenv("USE_DEBUG_LOGGER", "true") == "true",
        noisy_logs=["boto3", "botocore", "urllib3", "s3transfer", "rasterio", "fiona"],
        correlation_id=correlation_id,
    )
    request = RequestWrapper().from_dict(event)
    if betterproto.serialized_on_wire(request.download_and_crop_raw_tif_request):
        download_and_crop_raw_tif(logger, request, correlation_id, write_dir)
    elif betterproto.serialized_on_wire(request.create_map_product_request):
        create_new_map_product(logger=logger, request=request, correlation_id=correlation_id, write_dir=write_dir)
    else:
        logger.fatal("unknown event", **event)


def create_new_map_product(logger: ports.Logger, request: RequestWrapper, correlation_id: str, write_dir: pathlib.Path):
    logger.debug("Configuring service to handle new map product request", request=request.to_json())
    message_notification_queue = NewMessageNotificationQueue(
        queue_name=os.getenv("PUBLISH_MAP_PRODUCT_REQUEST_QUEUE", "conflict-nightlight-publish-map-product-request"),
        message_format=PublishMapProductRequest,
    )
    message_notification_queue.validate()
    service = MapProductService(
        raw_map_repository=S3InternalMapRepository(
            bucket_name=os.getenv("RAW_TIF_BUCKET", "conflict-nightlight-raw-tif"),
            new_message_notification=None,
            logger=logger,
            correlation_id=correlation_id,
            local_write_dir=write_dir,
        ),
        processed_map_repository=S3InternalMapRepository(
            logger=logger,
            bucket_name=os.getenv("PROCESSED_TIF_BUCKET_NAME", "conflict-nightlight-processed-tif"),
            correlation_id=correlation_id,
            new_message_notification=message_notification_queue,
            local_write_dir=write_dir,
        ),
    )
    service.process_save(proto_map_to_domain(request.create_map_product_request.map))


def download_and_crop_raw_tif(
    logger: ports.Logger, request: RequestWrapper, correlation_id: str, write_dir: pathlib.Path
):
    logger.debug("Configuring service to handle a download raw tif request", request=request.to_json())
    message_notification_queue = NewMessageNotificationQueue(
        queue_name=os.getenv("CREATE_MAP_PRODUCT_REQUEST_QUEUE", "conflict-nightlight-create-map-product-request"),
        message_format=CreateMapProductRequest,
    )
    message_notification_queue.validate()
    service = RawMapProcessorService(
        logger=logger,
        external_map_repository=EogdataMapRepository(
            write_dir,
            logger=logger,
            secrets=get_secrets_from_aws_secrets_manager(
                os.environ.get("CONFLICT_NIGHTLIGHT_SECRETS_KEY", "conflict-nightlight-secrets")
            ),
        ),
        internal_map_repository=S3InternalMapRepository(
            logger=logger,
            new_message_notification=message_notification_queue,
            bucket_name=os.getenv("RAW_TIF_BUCKET", "conflict-nightlight-raw-tif"),
            correlation_id=correlation_id,
            local_write_dir=write_dir,
        ),
        bounds_repository=S3BoundsRepository(
            bucket_name=os.getenv("SHAPE_FILE_BUCKET", "conflict-nightlight-shape-files"),
            local_write_dir=write_dir,
            logger=logger,
        ),
    )
    m = proto_map_to_domain(request.download_and_crop_raw_tif_request.map)
    logger.debug("Converted proto map to domain map", map=m)
    service.download_crop_save(m)
