data "aws_caller_identity" "current" {}

data "aws_iam_openid_connect_provider" "github_actions" {
  url = "https://token.actions.githubusercontent.com"
}

locals {
  github_actions_role_name   = "${var.prefix}-github-actions-deploy"
  github_actions_repo_branch = "repo:${var.github_repository}:ref:refs/heads/${var.github_deploy_branch}"
  github_frontend_bucket     = "${var.prefix}-${var.frontend_bucket_name}"
  github_zip_bucket          = "${var.prefix}-${var.zip_deployables_bucket_name}"
  github_python_repo         = "${var.prefix}-${var.python_lambda}-repo"
}

data "aws_iam_policy_document" "github_actions_assume_role" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.github_actions.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values   = [local.github_actions_repo_branch]
    }
  }
}

resource "aws_iam_role" "github_actions_deploy" {
  name               = local.github_actions_role_name
  assume_role_policy = data.aws_iam_policy_document.github_actions_assume_role.json
}

data "aws_iam_policy_document" "github_actions_deploy" {
  statement {
    sid = "FrontendBucketDeploy"
    actions = [
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:PutObject",
    ]
    resources = [
      "arn:aws:s3:::${local.github_frontend_bucket}",
      "arn:aws:s3:::${local.github_frontend_bucket}/*",
    ]
  }

  statement {
    sid = "GoZipDeploy"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "arn:aws:s3:::${local.github_zip_bucket}/*",
    ]
  }

  statement {
    sid = "EcrAuth"
    actions = [
      "ecr:GetAuthorizationToken",
    ]
    resources = ["*"]
  }

  statement {
    sid = "PythonImageDeploy"
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeRepositories",
      "ecr:InitiateLayerUpload",
      "ecr:PutImage",
      "ecr:UploadLayerPart",
    ]
    resources = [
      "arn:aws:ecr:${var.region}:${data.aws_caller_identity.current.account_id}:repository/${local.github_python_repo}",
    ]
  }

  statement {
    sid = "LambdaCodeDeploy"
    actions = [
      "lambda:GetFunction",
      "lambda:GetFunctionConfiguration",
      "lambda:UpdateFunctionCode",
    ]
    resources = [
      "arn:aws:lambda:${var.region}:${data.aws_caller_identity.current.account_id}:function:${var.prefix}-${var.map_controller}-function",
      "arn:aws:lambda:${var.region}:${data.aws_caller_identity.current.account_id}:function:${var.prefix}-${var.map_publisher}-function",
      "arn:aws:lambda:${var.region}:${data.aws_caller_identity.current.account_id}:function:${var.prefix}-${var.python_lambda}-function",
    ]
  }

  statement {
    sid = "CloudFrontInvalidation"
    actions = [
      "cloudfront:CreateInvalidation",
      "cloudfront:GetInvalidation",
      "cloudfront:ListDistributions",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "github_actions_deploy" {
  name   = local.github_actions_role_name
  policy = data.aws_iam_policy_document.github_actions_deploy.json
}

resource "aws_iam_role_policy_attachment" "github_actions_deploy" {
  role       = aws_iam_role.github_actions_deploy.name
  policy_arn = aws_iam_policy.github_actions_deploy.arn
}

output "github_actions_role_arn" {
  value = aws_iam_role.github_actions_deploy.arn
}
