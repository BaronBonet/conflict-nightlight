# ----------------------------------------------------------------------
#                       RAW TIF
# ----------------------------------------------------------------------
resource "aws_s3_bucket" "raw_tif" {
  bucket = "${var.prefix}-${var.raw_tif_bucket_name}"
}

resource "aws_s3_bucket_policy" "raw_tif_bucket_policy" {
  bucket = aws_s3_bucket.raw_tif.id
  policy = data.aws_iam_policy_document.raw_tif_bucket_policy_document.json
}

resource "aws_s3_bucket_public_access_block" "raw_tif_bucket_block_public_access" {
  bucket                  = aws_s3_bucket.raw_tif.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "raw_tif_bucket_policy_document" {
  statement {
    actions = [
      "s3:*"
    ]
    resources = [
      aws_s3_bucket.raw_tif.arn,
      "${aws_s3_bucket.raw_tif.arn}/*",
    ]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

# ----------------------------------------------------------------------
#                       PROCESSED TIF
# ----------------------------------------------------------------------
resource "aws_s3_bucket" "processed_tif" {
  bucket = "${var.prefix}-${var.processed_tif_bucket_name}"
}

resource "aws_s3_bucket_policy" "processed_tif_bucket_policy" {
  bucket = aws_s3_bucket.processed_tif.id
  policy = data.aws_iam_policy_document.processed_tif_bucket_policy_document.json
}

resource "aws_s3_bucket_public_access_block" "processed_tif_bucket_block_public_access" {
  bucket                  = aws_s3_bucket.processed_tif.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

data "aws_iam_policy_document" "processed_tif_bucket_policy_document" {
  statement {
    actions = [
      "s3:*"
    ]
    resources = [
      aws_s3_bucket.processed_tif.arn,
      "${aws_s3_bucket.processed_tif.arn}/*",
    ]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

# ----------------------------------------------------------------------
#                       SHAPE FILES
# ----------------------------------------------------------------------
resource "aws_s3_bucket" "shape_files" {
  bucket = "${var.prefix}-${var.shape_files_bucket_name}"
}

resource "aws_s3_bucket_policy" "shape_file_bucket_policy" {
  bucket = aws_s3_bucket.shape_files.id
  policy = data.aws_iam_policy_document.shape_file_bucket_policy_document.json
}

data "aws_iam_policy_document" "shape_file_bucket_policy_document" {
  statement {
    actions = [
      "s3:*"
    ]

    resources = [
      aws_s3_bucket.shape_files.arn,
      "${aws_s3_bucket.shape_files.arn}/*",
    ]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "shape_files_bucket_block_public_access" {
  bucket                  = aws_s3_bucket.shape_files.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}


# ----------------------------------------------------------------------
#                  ZIP DEPLOYABLES
# ----------------------------------------------------------------------

resource "aws_s3_bucket" "zip_deployables" {
  bucket = "${var.prefix}-${var.zip_deployables_bucket_name}"

}

resource "null_resource" "create_temp_zip" {
  provisioner "local-exec" {
    command = "cd ${path.module}/templates/fake_zip && zip -r latest.zip ."
  }

  provisioner "local-exec" {
    command = "aws s3 cp ${path.module}/templates/fake_zip/latest.zip s3://${aws_s3_bucket.zip_deployables.bucket}/map_controller/latest.zip && aws s3 cp ${path.module}/templates/fake_zip/latest.zip s3://${aws_s3_bucket.zip_deployables.bucket}/map_publisher/latest.zip"
  }

  depends_on = [aws_s3_bucket.zip_deployables]
}

resource "aws_s3_bucket_public_access_block" "zip_deployables_block_public_access" {
  bucket                  = aws_s3_bucket.zip_deployables.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ---------------------------------------------------------------
#                   Frontend
# ---------------------------------------------------------------

resource "aws_s3_bucket" "frontend_bucket" {
  bucket = "${var.prefix}-${var.frontend_bucket_name}"
}

resource "aws_s3_bucket_acl" "frontend_bucket_acl" {
  bucket = aws_s3_bucket.frontend_bucket.id
  acl    = "private"
}

data "aws_iam_policy_document" "frontend_bucket_policy_document" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.frontend_bucket.arn]
    principals {
      identifiers = [aws_cloudfront_origin_access_identity.frontend_s3_distribution.iam_arn]
      type        = "AWS"
    }
    sid = "bucket_policy_site_root"
  }
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.frontend_bucket.arn}/*"]
    principals {
      identifiers = [aws_cloudfront_origin_access_identity.frontend_s3_distribution.iam_arn]
      type        = "AWS"
    }
    sid = "bucket_policy_site_all"
  }
}

resource "aws_s3_bucket_policy" "frontend_bucket_policy" {
  depends_on = [aws_cloudfront_origin_access_identity.frontend_s3_distribution]
  bucket     = aws_s3_bucket.frontend_bucket.id
  policy     = data.aws_iam_policy_document.frontend_bucket_policy_document.json
}

resource "aws_s3_bucket_public_access_block" "frontend_bucket_block_public_access" {
  bucket                  = aws_s3_bucket.frontend_bucket.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ---------------------------------------------------------------
#                           CDN
# ---------------------------------------------------------------

resource "aws_s3_bucket" "cdn" {
  bucket = "${var.prefix}-${var.cdn_bucket_name}"
}

resource "aws_s3_bucket_acl" "cdn_bucket_acl" {
  bucket = aws_s3_bucket.cdn.id
  acl    = "private"
}

// TODO cant this be shared with the one above?
data "aws_iam_policy_document" "cdn_bucket_policy_document" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = [aws_s3_bucket.cdn.arn]
    principals {
      identifiers = [aws_cloudfront_origin_access_identity.cdn_s3_distribution.iam_arn]
      type        = "AWS"
    }
    sid = "bucket_policy_site_root"
  }
  statement {
    actions   = ["s3:*"]
    resources = [aws_s3_bucket.cdn.arn, "${aws_s3_bucket.cdn.arn}/*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
  statement {
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.cdn.arn}/*"]
    principals {
      identifiers = [aws_cloudfront_origin_access_identity.cdn_s3_distribution.iam_arn]
      type        = "AWS"
    }
    sid = "bucket_policy_site_all"
  }
}

resource "aws_s3_bucket_policy" "cdn_bucket_policy" {
  depends_on = [aws_cloudfront_origin_access_identity.cdn_s3_distribution]
  bucket     = aws_s3_bucket.cdn.id
  policy     = data.aws_iam_policy_document.cdn_bucket_policy_document.json
}

#resource "aws_s3_bucket_public_access_block" "cdn_bucket_block_public_access" {
#  bucket                  = aws_s3_bucket.cdn.id
#  block_public_acls       = true
#  block_public_policy     = true
#  ignore_public_acls      = true
#  restrict_public_buckets = true
#}

resource "aws_s3_bucket_cors_configuration" "cdn_cors" {
  bucket = aws_s3_bucket.cdn.bucket

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "HEAD"]
    allowed_origins = ["*"]
    expose_headers  = []
  }
}
