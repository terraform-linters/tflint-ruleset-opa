resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "arn:aws:elasticloadbalancing:us-east-1:187416307283:loadbalancer/app/my-load-balancer/50dc6c495c6c9a29"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "arn:aws:iam::187416307283:server-certificate/test_cert_rab3wuqwgja25ct3n4jdj2tzu4"

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = "arn:aws:elasticloadbalancing:us-east-1:187416307283:targetgroup/my-target-group/7f26b4fad4f0118b"
        weight = 100
      }
      target_group {
        arn    = "arn:aws:elasticloadbalancing:us-east-1:187416307283:targetgroup/my-other-target-group/8a37c5bfe5e129c"
        weight = 0
      }
    }
  }
}

resource "aws_lb_listener" "valid" {
  load_balancer_arn = "arn:aws:elasticloadbalancing:us-east-1:187416307283:loadbalancer/app/valid-lb/61dc7d506d7d0b3a"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-Res-PQ-2025-09"
  certificate_arn   = "arn:aws:iam::187416307283:server-certificate/test_cert_rab3wuqwgja25ct3n4jdj2tzu4"

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = "arn:aws:elasticloadbalancing:us-east-1:187416307283:targetgroup/valid-tg/8g37c5gfe6f230d"
        weight = 100
      }
      target_group {
        arn    = "arn:aws:elasticloadbalancing:us-east-1:187416307283:targetgroup/backup-tg/9h48d6hgf7g341e"
        weight = 0
      }
    }
  }
}

resource "aws_lb_listener" "undefined" {
  load_balancer_arn = "arn:aws:elasticloadbalancing:us-east-1:187416307283:loadbalancer/app/undefined-lb/72ed8e617e8e1c4b"
  port              = "443"
  protocol          = "HTTPS"

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = "arn:aws:elasticloadbalancing:us-east-1:187416307283:targetgroup/undefined-tg/ai59e7ihg8h452f"
        weight = 100
      }
      target_group {
        arn    = "arn:aws:elasticloadbalancing:us-east-1:187416307283:targetgroup/undefined-tg2/bj6af8jih9i563g"
        weight = 0
      }
    }
  }
}
