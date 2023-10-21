check "health_check" {
  data "http" "terraform_io" {
    url = "https://www.terraform.io"
  }

  assert {
    condition = data.http.terraform_io.status_code == 200
    error_message = "${data.http.terraform_io.url} returned an unhealthy status code"
  }
}

check "deterministic" {
  assert {
    condition = 200 == 200
    error_message = "condition should be true"
  }
}
