resource "aws_instance" "invalid" {
  provisioner "local-exec" {
    command = "echo 'This is an invalid provisioner'"
  }
}

resource "aws_instance" "valid" {
  provisioner "file" {
    source      = "test.txt"
    destination = "/tmp/test.txt"
  }
}

resource "aws_instance" "undefined" {
}

