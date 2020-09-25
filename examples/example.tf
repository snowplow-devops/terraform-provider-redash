// Requires the following env var to be set:
// export REDASH_API_KEY="<personal api token>"

provider "redash" {
  redash_uri = "https://com.acme.redash"
}

resource "redash_data_source" "acme_corp" {
  name = "ACME Corporation Product Database"
  type = "redshift"

  options {
    host     = "newproducts.acme.com"
    port     = 5439
    dbname   = "products"
    user     = "wcoyote"
    password = "eth3LbeRt"
  }
}

resource "redash_data_source" "gps" {
  name = "GPS Coordinates"
  type = "pg"

  options {
    host     = "coordinates.gps.com"
    port     = 5432
    dbname   = "escape_routes"
    user     = "rrunner"
    password = "G3oc0ccyX"
  }
}

resource "redash_group" "geniuses" {
  name = "ACME Users"
}

resource "redash_group" "runners" {
  name = "Beep Beep"
}

resource "redash_user" "wcoyote" {
  name   = "Wile E. Coyote"
  email  = "wcoyote@acme.com"
  groups = [redash_group.geniuses.id]

  depends_on = [
    redash_group.geniuses,
  ]
}

resource "redash_user" "rrunner" {
  name   = "Road Runner"
  email  = "rrunner@acme.com"
  groups = [redash_group.runners.id]

  depends_on = [
    redash_group.runners,
  ]
}

resource "redash_group_data_source_attachment" "wcoyote_acme" {
  group_id       = redash_group.geniuses.id
  data_source_id = redash_data_source.acme_corp.id

    depends_on = [
    redash_group.geniuses,
    redash_data_source.acme_corp,
  ]
}

resource "redash_group_data_source_attachment" "rrunner_gps" {
  group_id       = redash_group.runners.id
  data_source_id = redash_data_source.gps.id

    depends_on = [
    redash_data_source.gps,
    redash_group.runners,
  ]
}
