plugin "terraform" {
  enabled = false
}

plugin "opa" {
  enabled = true

  policy_dirs = ["policies"]
}
