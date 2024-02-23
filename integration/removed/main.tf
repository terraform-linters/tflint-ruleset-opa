removed {
  from = aws_instance.example

  lifecycle {
    destroy = false
  }
}
