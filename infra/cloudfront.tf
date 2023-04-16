locals {
  s3_frontend_origin_id = "${var.prefix}-frontend-origin-id"
  s3_cdn_origin_id      = "${var.prefix}-cdn-origin-id"
}

resource "aws_cloudfront_origin_access_identity" "frontend_s3_distribution" {
  comment = "Origin Access Identity for ${var.domain_name} s3 distribution"
}


resource "aws_cloudfront_distribution" "frontend_distribution" {
  origin {
    domain_name = aws_s3_bucket.frontend_bucket.bucket_regional_domain_name
    origin_id   = local.s3_frontend_origin_id
    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.frontend_s3_distribution.cloudfront_access_identity_path
    }

  }

  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"

  aliases = [var.domain_name, "www.${var.domain_name}"]

  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.conflictnightlight_com.arn
    ssl_support_method  = "sni-only"
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = local.s3_frontend_origin_id

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 31536000
    compress               = true
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "blacklist"
      locations = [
        "CN",
        "IN",
        "IR",
        "KP",
        "PK",
        "RU",
      ]
    }
  }

}

# ----------------------------------------------------------------------------------
#                                      CDN
# ----------------------------------------------------------------------------------
resource "aws_cloudfront_origin_access_identity" "cdn_s3_distribution" {
  comment = "Origin Access Identity for cdn.${var.domain_name} s3 distribution"
}

resource "aws_cloudfront_distribution" "cdn_distribution" {
  origin {
    domain_name = aws_s3_bucket.cdn.bucket_regional_domain_name
    origin_id   = local.s3_cdn_origin_id
    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.cdn_s3_distribution.cloudfront_access_identity_path
    }
  }
  enabled         = true
  is_ipv6_enabled = true
  aliases         = ["cdn.${var.domain_name}", "www.cdn.${var.domain_name}"]

  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.conflictnightlight_cdn.arn
    ssl_support_method  = "sni-only"
  }

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = local.s3_cdn_origin_id

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 31536000
    compress               = true
  }

  price_class = "PriceClass_200"

  restrictions {
    geo_restriction {
      restriction_type = "blacklist"
      locations = [
        "CN",
        "IN",
        "IR",
        "KP",
        "PK",
        "RU",
      ]
    }
  }

}
