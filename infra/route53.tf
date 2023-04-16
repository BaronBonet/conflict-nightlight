## ---------------------------------------------------------------------------------------
##                                   Frontend
## ---------------------------------------------------------------------------------------
#
resource "aws_route53_record" "domain" {
  zone_id = data.aws_route53_zone.conflictnightlight_com.zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.frontend_distribution.domain_name
    zone_id                = aws_cloudfront_distribution.frontend_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "domain_www" {
  zone_id = data.aws_route53_zone.conflictnightlight_com.zone_id
  name    = "www.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.frontend_distribution.domain_name
    zone_id                = aws_cloudfront_distribution.frontend_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}

# ---------------------------------------------------------------------------------------
#                                   CDN
# ---------------------------------------------------------------------------------------
#
resource "aws_route53_record" "cdn" {
  zone_id = data.aws_route53_zone.conflictnightlight_com.zone_id
  name    = "cdn.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.cdn_distribution.domain_name
    zone_id                = aws_cloudfront_distribution.cdn_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "cdn_www" {
  zone_id = data.aws_route53_zone.conflictnightlight_com.zone_id
  name    = "www.cdn.${var.domain_name}"
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.cdn_distribution.domain_name
    zone_id                = aws_cloudfront_distribution.cdn_distribution.hosted_zone_id
    evaluate_target_health = false
  }
}
