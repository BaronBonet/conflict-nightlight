import ast

import boto3
from botocore.exceptions import ClientError

from app.infrastructure.get_or_raise import get_or_raise


def get_secrets_from_aws_secrets_manager(secret_id: str) -> dict[str, str]:
    client = boto3.client("secretsmanager")
    try:
        secret_data = client.get_secret_value(SecretId=secret_id)
    except ClientError as e:
        raise Exception(f"The secret value {secret_id} was not found. Exception raised: {e}")

    secret = get_or_raise(secret_data, "SecretString")

    return ast.literal_eval(secret)
