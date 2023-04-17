import sys

from awslambdaric.lambda_context import LambdaContext

from app.core.domain import KillProcess
from app.handlers.proto_event_handler import handle_event
from app.infrastructure.extract_lambda_request import extract_lambda_request


def handler(event, context: LambdaContext):
    request, correlation_id = extract_lambda_request(event)
    try:
        return handle_event(request, correlation_id)
    except KillProcess:
        print("Killing process")
        # We want sqs to retry this message, so we do a sys.exit()
        sys.exit(1)
