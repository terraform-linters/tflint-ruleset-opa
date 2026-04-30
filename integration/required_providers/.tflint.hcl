plugin "terraform" {
  enabled = false
}

plugin "opa" {
  enabled = true

  policy_dir = "policies"
}
