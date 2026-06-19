env "local" {
  url = getenv("DB_DSN")
  dev = "docker://postgres/15/dev"
  migration {
    dir = "file://migrations/versioned"
  }
}

env "production" {
  url = getenv("DB_DSN")
  migration {
    dir    = "file://migrations/versioned"
    format = atlas
  }
}