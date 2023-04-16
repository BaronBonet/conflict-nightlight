variable "region" {
  default = "eu-central-1"
  type    = string
}

variable "prefix" {
  description = "The prefix to use for all names"
  type        = string
  default     = "conflict-nightlight"
}

variable "raw_tif_bucket_name" {
  description = "The name of the s3 bucket used to store the raw tif files"
  type        = string
  default     = "raw-tif"
}


variable "processed_tif_bucket_name" {
  description = "The name of the s3 bucket used to store the processed tif files"
  type        = string
  default     = "processed-tif"
}

variable "shape_files_bucket_name" {
  description = "The name of the s3 bucket used to store the shape files"
  type        = string
  default     = "shape-files"
}

variable "download_and_crop_raw_tif_request_queue_name" {
  description = "The name of the sqs queue used to trigger downloading a new raw tif file"
  type        = string
  default     = "download-and-crop-raw-tif-request"
}

variable "create_map_product_request_queue_name" {
  description = "The name of the sqs queue used when a new raw tif file was uploaded to s3"
  type        = string
  default     = "create-map-product-request"
}

variable "publish_map_product_request_queue_name" {
  description = "The name of the sqs queue used when a new processed tif file was uploaded to s3"
  type        = string
  default     = "publish-map-product-request"
}

variable "map_controller" {
  description = "The name of the lambda function for the map controller golang application"
  type        = string
  default     = "map-controller"
}

variable "python_lambda" {
  description = "The name of the lambda function for the python application"
  type        = string
  default     = "python-app"
}

variable "map_publisher" {
  description = "The name of the lambda function for the map publisher golang application"
  type        = string
  default     = "map-publisher"
}

variable "zip_deployables_bucket_name" {
  description = "The bucket name that contains the deployable zip files"
  type        = string
  default     = "zip-deployables"
}

variable "correlation_id_key" {
  description = "The key passed around used to retrieve the correlation id from the context or aws messages"
  type        = string
  default     = "correlation-id"
}

variable "source_url_key" {
  description = "The key passed around to retrieve the source url of the raw tif"
  type        = string
  default     = "source-url"
}


# ---------------------------------------------------------------
#                   Frontend
# ---------------------------------------------------------------
variable "frontend_bucket_name" {
  description = "The name of the s3 bucket used to store the reactjs generated frontend"
  type        = string
  default     = "frontend"
}

variable "domain_name" {
  description = "The domain name where the frontend will be hosted"
  type        = string
  default     = "conflictnightlight.com"
}

variable "cdn_bucket_name" {
  description = "The name of the s3 bucket used for our cdn"
  type        = string
  default     = "cdn"
}
