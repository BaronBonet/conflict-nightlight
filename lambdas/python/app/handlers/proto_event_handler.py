import os
import pathlib

from app.adapters.eogdata_external_map_repository import EogdataMapRepository
from app.adapters.s3_bounds_repository import S3BoundsRepository
from app.adapters.s3_internal_map_repository import S3InternalMapRepository, NewMessageNotificationQueue
from app.adapters.struct_logger import StructLogger
from app.core import ports
from app.core.services.product import MapProductService
from app.core.services.raw_processor import RawMapProcessorService
from app.infrastructure.get_secrets_from_aws_secret_manager import get_secrets_from_aws_secrets_manager
from generated.conflict_nightlight.v1 import (
    DownloadAndCropRawTifRequest,
    CreateMapProductRequest,
    PublishMapProductRequest,
    RequestWrapper,
)


def is_used(p: DownloadAndCropRawTifRequest | CreateMapProductRequest) -> bool:
    return p.__dict__["_serialized_on_wire"]


def handle_event(event: dict[str, str], correlation_id: str):
    write_dir = pathlib.Path(os.getenv("LOCAL_WRITE_DIRECTORY", "/tmp"))
    logger = StructLogger(
        version=os.getenv("VERSION", "unknown"),
        use_debug=os.getenv("USE_DEBUG_LOGGER", "true") == "true",
        noisy_logs=["boto3", "botocore", "urllib3", "s3transfer", "rasterio", "fiona"],
        correlation_id=correlation_id,
    )
    request = RequestWrapper().from_dict(event)
    if is_used(request.download_and_crop_raw_tif_request):
        download_and_crop_raw_tif_request(logger, request, correlation_id, write_dir)

    elif is_used(request.create_map_product_request):
        logger.debug("Configuring service to handle a new raw tif file request", request=event)
        create_new_map_product_request(
            logger=logger, request=request, correlation_id=correlation_id, write_dir=write_dir
        )
    else:
        logger.fatal("unknown event", **event)


def create_new_map_product_request(
    logger: ports.Logger, request: RequestWrapper, correlation_id: str, write_dir: pathlib.Path
):
    logger.debug("Configuring service to handle new map product request", request=request.to_json())
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
            new_message_notification=NewMessageNotificationQueue(
                queue_name=os.getenv(
                    "PUBLISH_MAP_PRODUCT_REQUEST_QUEUE", "conflict-nightlight-publish-map-product-request"
                ),
                message_format=PublishMapProductRequest,
            ),
            local_write_dir=write_dir,
        ),
    )
    service.process_save(request.create_map_product_request.map)


def download_and_crop_raw_tif_request(
    logger: ports.Logger, request: RequestWrapper, correlation_id: str, write_dir: pathlib.Path
):
    logger.debug("Configuring service to handle a download raw tif request", request=request.to_json())
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
            new_message_notification=NewMessageNotificationQueue(
                queue_name=os.getenv(
                    "CREATE_MAP_PRODUCT_REQUEST_QUEUE", "conflict-nightlight-create-map-product-request"
                ),
                message_format=CreateMapProductRequest,
            ),
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
    service.download_crop_save(m=request.download_and_crop_raw_tif_request.map)
