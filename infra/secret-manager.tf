resource "aws_secretsmanager_secret" "conflict_nightlight-secrets" {
  name        = "${var.prefix}-secrets"
  description = "all of the secrets necessary for the conflict-nightlight application, not the best in terms of security, but cheaper"
}

resource "aws_secretsmanager_secret_version" "conflict_nightlight_secret_version" {
  secret_id = aws_secretsmanager_secret.conflict_nightlight-secrets.id
  secret_string = jsonencode({
    eogdataUsername   = var.eogdata_auth_username,
    eogdataPassword   = var.eogdata_auth_password,
    mapboxPublicToken = var.mapbox_public_token,
    mapboxUsername    = var.mapbox_username
  })
}

variable "eogdata_auth_password" {
  default   = "create an account on eogdata to obtain"
  type      = string
  sensitive = true
}
variable "eogdata_auth_username" {
  default = "username used when creating the eogdata account"
  type    = string
}
variable "mapbox_public_token" {
  default   = "Obtained from mapbox console"
  type      = string
  sensitive = true
}
variable "mapbox_username" {
  default = "the username for your mapbox account"
  type    = string
}

resource "aws_iam_policy" "secret_manager_policy" {
  name = "${var.prefix}-secret-manager-policy"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        Action : [
          "secretsmanager:GetSecretValue",
        ],
        Effect : "Allow",
        Resource : aws_secretsmanager_secret.conflict_nightlight-secrets.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "secret_manager_policy_document" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.secret_manager_policy.arn
}
