import ast
import os

from app.infrastructure.get_or_raise import get_or_raise


def extract_lambda_request(event: str | dict[str, str]) -> tuple[dict[str, str], str]:
    if type(event) == str:
        event = ast.literal_eval(event)
    records = get_or_raise(event, "Records")
    if len(records) > 1:
        raise Exception("More than one record was provided.")
    record = records[0]
    request = get_or_raise(record, "body")
    if type(request) == str:
        request = ast.literal_eval(request)
    return request, _get_correlation_id(record)


def _get_correlation_id(record: dict[str, str]) -> str:
    correlation_id = "unknown"
    attributes = record.get("messageAttributes")
    if not attributes or not isinstance(attributes, dict):
        return correlation_id
    c_id = attributes.get(os.getenv("CORRELATION_ID_KEY_NAME", "correlation-id"))
    if not c_id or not isinstance(c_id, dict):
        return correlation_id
    return c_id.get("stringValue", "unknown")
