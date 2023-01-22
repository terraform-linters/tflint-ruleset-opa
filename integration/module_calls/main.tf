module "local" {
  source = "./module"
}

module "remote" {
  source = "github.com/hashicorp/example"
}
