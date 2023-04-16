from awslambdaric.lambda_context import LambdaContext

from app.handlers.proto_event_handler import handle_event
from app.infrastructure.extract_lambda_request import extract_lambda_request


def handler(event, context: LambdaContext):
    request, correlation_id = extract_lambda_request(event)
    return handle_event(request, correlation_id)
