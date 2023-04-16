resource "aws_iam_role" "lambda" {
  name = "${var.prefix}-lambda-role"
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : "sts:AssumeRole",
        Effect : "Allow",
        Principal : {
          "Service" : "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# -----------------------------------------------------------------------------------
#                                Map Controller
# -----------------------------------------------------------------------------------
resource "aws_lambda_function" "conflict_nightlight_map_controller_lambda_function" {
  ephemeral_storage {
    size = 512 # Min 512 MB and the Max 10240 MB
  }
  function_name = "${var.prefix}-${var.map_controller}-function"
  environment {
    variables = {
      USES_DEBUG_LOGGER      = "true"
      DOWNLOAD_RAW_TIF_QUEUE = aws_sqs_queue.download_and_crop_raw_tif_request_queue.name
      RAW_TIF_BUCKET         = aws_s3_bucket.raw_tif.bucket
      CORRELATION_ID_KEY     = var.correlation_id_key
      SOURCE_KEY_URL         = var.source_url_key
      SHAPE_FILE_BUCKET      = aws_s3_bucket.shape_files.bucket
    }
  }
  s3_bucket     = aws_s3_bucket.zip_deployables.bucket
  s3_key        = "map_controller/latest.zip"
  runtime       = "provided.al2"
  architectures = ["arm64"]
  handler       = "handler"
  role          = aws_iam_role.lambda.arn
  timeout       = 100
}

# -----------------------------------------------------------------------------------
#                                Python
# -----------------------------------------------------------------------------------

resource "aws_lambda_function" "conflict_nightlight_python_lambda_function" {
  // The lambda function will not be created until the dummy image exists on ecr
  depends_on = [
    null_resource.ecr_image_fake_downloader
  ]
  ephemeral_storage {
    size = 10240 # Min 512 MB and the Max 10240 MB
  }
  memory_size = 3000 # Prevents the lambda from crashing when cropping the tif
  environment {
    variables = {
      WRITE_DIR                         = "/tmp"
      WRITE_BUCKET                      = var.raw_tif_bucket_name
      CORRELATION_ID_KEY_NAME           = var.correlation_id_key
      CREATE_MAP_PRODUCT_REQUEST_QUEUE  = aws_sqs_queue.create_map_product_request_queue.name
      PUBLISH_MAP_PRODUCT_REQUEST_QUEUE = aws_sqs_queue.publish_map_product_request_queue.name
      PROCESSED_TIF_BUCKET_NAME         = aws_s3_bucket.processed_tif.bucket
      CONFLICT_NIGHTLIGHT_SECRETS_KEY   = aws_secretsmanager_secret.conflict_nightlight-secrets.name
      USES_DEBUG_LOGGER                 = "true"
    }
  }
  function_name = "${var.prefix}-${var.python_lambda}-function"
  role          = aws_iam_role.lambda.arn
  timeout       = 200
  image_uri     = "${aws_ecr_repository.conflict_nightlight_python_lambda_repo.repository_url}:latest"
  package_type  = "Image"
  publish       = true
}

resource "aws_lambda_event_source_mapping" "conflict_nightlight_raw_tif_downloader_lambda_event_source" {
  event_source_arn = aws_sqs_queue.download_and_crop_raw_tif_request_queue.arn
  function_name    = aws_lambda_function.conflict_nightlight_python_lambda_function.arn
  batch_size       = 1
}

resource "aws_lambda_event_source_mapping" "conflict_nightlight_processor_lambda_event_source" {
  event_source_arn = aws_sqs_queue.create_map_product_request_queue.arn
  function_name    = aws_lambda_function.conflict_nightlight_python_lambda_function.arn
  batch_size       = 1
}

# -----------------------------------------------------------------------------------
#                                Map Publisher
# -----------------------------------------------------------------------------------
resource "aws_lambda_function" "conflict_nightlight_map_publisher_lambda_function" {
  ephemeral_storage {
    size = 512 # Min 512 MB and the Max 10240 MB
  }
  function_name                  = "${var.prefix}-${var.map_publisher}-function"
  reserved_concurrent_executions = 1
  environment {
    variables = {
      WRITE_DIR                 = "/tmp"
      FRONTEND_MAP_OPTIONS_JSON = "conflict-nightlight-map-options.json"
      PROCESSED_TIF_BUCKET_NAME = aws_s3_bucket.processed_tif.bucket
      CORRELATION_ID_KEY        = var.correlation_id_key
      SOURCE_KEY_URL            = var.source_url_key
      CDN_BUCKET_NAME           = aws_s3_bucket.cdn.bucket
    }
  }
  s3_bucket     = aws_s3_bucket.zip_deployables.bucket
  s3_key        = "map_publisher/latest.zip"
  runtime       = "provided.al2"
  architectures = ["arm64"]
  handler       = "handler"
  role          = aws_iam_role.lambda.arn
  timeout       = 50
}

resource "aws_lambda_event_source_mapping" "conflict_nightlight_publisher_lambda_event_source" {
  event_source_arn = aws_sqs_queue.publish_map_product_request_queue.arn
  function_name    = aws_lambda_function.conflict_nightlight_map_publisher_lambda_function.arn
  batch_size       = 1
}
