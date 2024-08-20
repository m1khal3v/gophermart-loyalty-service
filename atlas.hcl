data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "ariga.io/atlas-provider-gorm",
    "load",
    "--path", "./internal/entity",
    "--dialect", "postgres",
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  dev = "postgres://postgres:postgres@pgsql:5432/postgres?search_path=public&sslmode=disable"
  migration {
    dir = "file://migrations/pgsql?format=goose"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
