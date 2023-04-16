locals {
  python_lambda_repo_name = "${var.prefix}-${var.python_lambda}-repo"
  image_tag               = "latest"
}

data "aws_ecr_authorization_token" "token" {}

resource "aws_ecr_repository" "conflict_nightlight_python_lambda_repo" {
  name                 = local.python_lambda_repo_name
  image_tag_mutability = "MUTABLE"
  force_delete         = true # allows for deletion of repository even if it has images
  image_scanning_configuration {
    scan_on_push = true
  }
  lifecycle {
    ignore_changes = all
  }
}

resource "null_resource" "ecr_image_fake_downloader" {
  # The local-exec provisioner invokes a local executable after a resource is created.
  # This invokes a process on the machine running Terraform, not on the resource.
  #
  # This is a 1-time execution to put a dummy image into the ECR repo, so
  #    terraform provisioning works on the lambda function. Otherwise there is
  #    a chicken-egg scenario where the lambda can't be provisioned because no
  #    image exists in the ECR
  provisioner "local-exec" {
    command = <<EOF
      docker login ${data.aws_ecr_authorization_token.token.proxy_endpoint} -u AWS -p ${data.aws_ecr_authorization_token.token.password}
      docker pull alpine
      docker tag alpine ${aws_ecr_repository.conflict_nightlight_python_lambda_repo.repository_url}:${local.image_tag}
      docker push ${aws_ecr_repository.conflict_nightlight_python_lambda_repo.repository_url}:${local.image_tag}
      EOF
  }
}
