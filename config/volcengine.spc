connection "volcengine" {
  plugin = "volcengine"

  # You may connect to one or more regions. If `regions` is not specified,
  # Steampipe will use a single default region using the below resolution
  # order:
  # The `VOLCENGINE_REGION` or `VOLCENGINE_REGION_ID` environment variable
  # regions = ["cn-beijing", "cn-shanghai"]

  # Static credentials
  # access_key  = "your-volcengine-access-key"
  # secret_key  = "your-volcengine-secret-key"

  # Alternatively, set credentials via environment variables:
  # VOLCENGINE_ACCESS_KEY_ID and VOLCENGINE_SECRET_ACCESS_KEY

  # Timeout for API requests in seconds. Defaults to 60 seconds.
  # timeout = 60

  # List of additional Volcengine error codes to ignore for all queries.
  # ignore_error_codes = ["AccessDenied", "Forbidden"]
}
