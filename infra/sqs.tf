# ----------------------------------------------------------------------
#                        NEW RAW TIF FILE DOWNLOAD REQUEST
# ----------------------------------------------------------------------
resource "aws_sqs_queue" "download_and_crop_raw_tif_request_queue" {
  name                      = "${var.prefix}-${var.download_and_crop_raw_tif_request_queue_name}"
  delay_seconds             = 1
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
  # because the lambda can run for 200 seconds
  visibility_timeout_seconds = 201

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.download_raw_tif_request_dlq.arn
    maxReceiveCount     = 2
  })
}

resource "aws_iam_policy" "download_raw_tif_request_queue_message_policy" {
  name = "${var.prefix}-download-tif-request-sqs-message-policy"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl",
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
        ],
        Effect : "Allow",
        Resource : aws_sqs_queue.download_and_crop_raw_tif_request_queue.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "download_tif_request_sqs_policy_document" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.download_raw_tif_request_queue_message_policy.arn
}

resource "aws_sqs_queue" "download_raw_tif_request_dlq" {
  name                       = "${var.prefix}-${var.download_and_crop_raw_tif_request_queue_name}-dlq"
  message_retention_seconds  = 86400
  visibility_timeout_seconds = 201
  max_message_size           = 2048
}

# ----------------------------------------------------------------------
#                        NEW RAW TIF FILE NOTIFICATION
# ----------------------------------------------------------------------

resource "aws_sqs_queue" "create_map_product_request_queue" {
  name                       = "${var.prefix}-${var.create_map_product_request_queue_name}"
  delay_seconds              = 1
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 201

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.new_raw_tif_file_notification_dlq.arn
    maxReceiveCount     = 2
  })
}

resource "aws_iam_policy" "new_raw_tif_file_notification_queue_message_policy" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl",
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
        ],
        Effect : "Allow",
        Resource : aws_sqs_queue.create_map_product_request_queue.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "new_raw_tif_file_notification_sqs_policy_document" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.new_raw_tif_file_notification_queue_message_policy.arn
}

resource "aws_sqs_queue" "new_raw_tif_file_notification_dlq" {
  name                       = "${var.prefix}-${var.create_map_product_request_queue_name}-dlq"
  message_retention_seconds  = 86400
  visibility_timeout_seconds = 201
  max_message_size           = 2048
}

# ----------------------------------------------------------------------
#                        Publish map product request queue
# ----------------------------------------------------------------------

resource "aws_sqs_queue" "publish_map_product_request_queue" {
  name                       = "${var.prefix}-${var.publish_map_product_request_queue_name}"
  delay_seconds              = 1
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.new_processed_tif_file_notification_dlq.arn
    maxReceiveCount     = 2
  })
}

resource "aws_iam_policy" "new_processed_tif_file_notification_queue_message_policy" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl",
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
        ],
        Effect : "Allow",
        Resource : aws_sqs_queue.publish_map_product_request_queue.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "new_processed_tif_file_notification_sqs_policy_document" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.new_processed_tif_file_notification_queue_message_policy.arn
}

resource "aws_sqs_queue" "new_processed_tif_file_notification_dlq" {
  name                       = "${var.prefix}-${var.publish_map_product_request_queue_name}-dlq"
  message_retention_seconds  = 86400
  visibility_timeout_seconds = 201
  max_message_size           = 2048
}
