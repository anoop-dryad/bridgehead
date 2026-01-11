data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/models", // path at which the models are kept
    "--dialect", "postgres",
  ]
}
env "gorm" {
  src = data.external_schema.gorm.url
  dev = "postgres://postgres:anoop@localhost:5432/atlas_dev?sslmode=disable"  // Atlas need an empty db to perform its tasks.
  url = "postgres://postgres:anoop@localhost:5432/bridgehead?sslmode=disable" // Its the our DB in which tables will be generated
  migration {
    dir = "file://migrations" // migration folder in our project
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}