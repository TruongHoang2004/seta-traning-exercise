data "external_schema" "gorm" {
  program = [
    "go", "run", "-mod=mod",
    "ariga.io/atlas-provider-gorm", "load",
    "--path", "./internal/infrastructure/persistence/model",
    "--dialect", "postgres"
  ]
}

env "gorm" {
  # Nguồn schema lấy từ GORM
  src = data.external_schema.gorm.url

  # Database dev (chỉ dùng để Atlas diff, không phải DB production)
  dev = "postgres://postgres:truonghoang2004@localhost:5432/collab_service?sslmode=disable"

  url = "postgres://postgres:truonghoang2004@localhost:5432/collab_service?sslmode=disable"

  migration {
    dir = "file://internal/infrastructure/database/migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
