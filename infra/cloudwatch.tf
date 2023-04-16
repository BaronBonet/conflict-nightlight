resource "aws_iam_policy" "function_logging_policy" {
  name = "${var.prefix}-function-logging-policy"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Effect : "Allow",
        Resource : "arn:aws:logs:*:*:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "function_logging_policy_attachment" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.function_logging_policy.arn
}

# -----------------------------------------------------------------------------------
#                                Map Controller
# -----------------------------------------------------------------------------------
resource "aws_cloudwatch_log_group" "map_controller_function_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.conflict_nightlight_map_controller_lambda_function.function_name}"
  retention_in_days = 7
  lifecycle {
    prevent_destroy = false
  }
}

resource "aws_cloudwatch_event_rule" "map_controller_schedule" {
  name                = "lambda-schedule"
  description         = "Schedule for triggering lambda once per day"
  schedule_expression = "cron(0 0 * * ? *)" # Trigger at midnight UTC every day
}

resource "aws_cloudwatch_event_target" "lambda_target" {
  rule = aws_cloudwatch_event_rule.map_controller_schedule.name
  arn  = aws_lambda_function.conflict_nightlight_map_controller_lambda_function.arn
  input = jsonencode({
    "cropper" : "ukraine-and-around",
    "mapType" : "monthly",
    "selectedMonths" : [
      1,
      2,
      3,
      4,
      10,
      11,
      12
    ],
    "selectedYears" : [
      2021,
      2022,
      2023,
      2024
    ]
  })
}

# -----------------------------------------------------------------------------------
#                                Map Processor
# -----------------------------------------------------------------------------------
resource "aws_cloudwatch_log_group" "python_function_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.conflict_nightlight_python_lambda_function.function_name}"
  retention_in_days = 7
  lifecycle {
    prevent_destroy = false
  }
}

# -----------------------------------------------------------------------------------
#                                Map Publisher
# -----------------------------------------------------------------------------------
resource "aws_cloudwatch_log_group" "map_publisher_function_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.conflict_nightlight_map_publisher_lambda_function.function_name}"
  retention_in_days = 7
  lifecycle {
    prevent_destroy = false
  }
}
