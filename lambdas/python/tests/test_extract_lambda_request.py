from app.infrastructure.extract_lambda_request import extract_lambda_request


def test_extract_sqs_event():
    event = """{
  "Records": [
    {
      "messageId": "19dd0b57-b21e-4ac1-bd88-01bbb068cb78",
      "receiptHandle": "MessageReceiptHandle",
      "body": '{"message": "Hi"}',
      "attributes": {
        "ApproximateReceiveCount": "1",
        "SentTimestamp": "1523232000000",
        "SenderId": "123456789012",
        "ApproximateFirstReceiveTimestamp": "1523232000001"
      },
    "messageAttributes": {
        "correlation-id": {
            "stringValue": "603f6ae8-789b-422d-95be-9acc6e8ff238",
            "stringListValues": [],
            "binaryListValues": [],
            "dataType": "String",
        }
    },
      "md5OfBody": "{{{md5_of_body}}}",
      "eventSource": "aws:sqs",
      "eventSourceARN": "arn:aws:sqs:us-east-1:123456789012:MyQueue",
      "awsRegion": "us-east-1"
    }
  ]
}"""
    request, correlation_id = extract_lambda_request(event)
    assert request == {"message": "Hi"}
    assert correlation_id == "603f6ae8-789b-422d-95be-9acc6e8ff238"
