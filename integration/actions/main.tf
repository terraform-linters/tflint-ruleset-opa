action "aws_lambda_invoke" "invalid" {
  config {
    function_name = "123456789012:function:deprecated-function:1"
    payload = jsonencode({
      key1 = "value1"
      key2 = "value2"
    })
  }
}

action "aws_lambda_invoke" "valid" {
  config {
    function_name = "123456789012:function:new-function:1"
    payload = jsonencode({
      key1 = "value1"
      key2 = "value2"
    })
  }
}
