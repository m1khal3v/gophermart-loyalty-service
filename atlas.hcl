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
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://migrations/pgsql?format=goose"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
