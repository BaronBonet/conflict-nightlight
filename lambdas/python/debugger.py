from app.handlers.proto_event_handler import handle_event
from app.infrastructure.extract_lambda_request import extract_lambda_request
from generated.conflict_nightlight.v1 import (
    DownloadAndCropRawTifRequest,
    CreateMapProductRequest,
    Map,
    Date,
    MapType,
    MapProvider,
    MapSource,
    Bounds,
    RequestWrapper,
)

example_message_download_and_crop = RequestWrapper(
    download_and_crop_raw_tif_request=DownloadAndCropRawTifRequest(
        map=Map(
            date=Date(month=1, year=2021),
            map_type=MapType.MAP_TYPE_MONTHLY,
            map_source=MapSource(
                map_provider=MapProvider.MAP_PROVIDER_EOGDATA,
                url=(
                    "https://eogdata.mines.edu/nighttime_light/monthly/v10/"
                    "2021/202101/vcmcfg/SVDNB_npp_20210101-20210131_75N060W_vcmcfg_v10_c202102062300.tgz"
                ),
            ),
            bounds=Bounds.BOUNDS_UKRAINE_AND_AROUND,
        )
    )
)

example_message_create_new_product = RequestWrapper(
    create_map_product_request=CreateMapProductRequest(
        map=Map(
            date=Date(month=1, year=2021),
            map_type=MapType.MAP_TYPE_MONTHLY,
            map_source=MapSource(
                map_provider=MapProvider.MAP_PROVIDER_EOGDATA,
                url=(
                    "https://eogdata.mines.edu/nighttime_light/monthly/v10/"
                    "2021/202101/vcmcfg/SVDNB_npp_20210101-20210131_75N060W_vcmcfg_v10_c202102062300.tgz"
                ),
            ),
            bounds=Bounds.BOUNDS_UKRAINE_AND_AROUND,
        )
    )
)

event = (
    """{
      "Records": [
        {
          "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
          "receiptHandle": "MessageReceiptHandle",
          "body": """
    + example_message_create_new_product.to_json()
    + """,
    "attributes": {
        "ApproximateReceiveCount": "1",
        "SentTimestamp": "1679738941300",
        "SenderId": "AROAQNEQMQH5SQYGSWSE6:conflict-nightlight-map-controller-function",
        "ApproximateFirstReceiveTimestamp": "1679738942300",
    },
    "messageAttributes": {
        "correlation-id": {
            "stringValue": "603f6ae8-789b-422d-95be-9acc6e8ff238",
            "stringListValues": [],
            "binaryListValues": [],
            "dataType": "String",
        }
    },
      "md5OfBody": "5d449930cf4aae378741cbbee9d1851a",
      "eventSource": "aws:sqs",
      "eventSourceARN": "arn:aws:sqs:eu-central-1:028220097019:conflict-nightlight-download-raw-tif-request",
      "awsRegion": "eu-central-1"
    }
  ]
}"""
)

if __name__ == "__main__":
    request, correlation_id = extract_lambda_request(event)
    handle_event(request, correlation_id)
