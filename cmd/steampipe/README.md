# Steampipe

This agent uses [steampipe](https://github.com/turbot/steampipe) to retrieve resources.

1. Install steampipe.
1. Install steampipe AWS plugin: `steampipe plugin install aws`.
1. Run steampipe in server mode: `steampipe service start`.
1. Run agent: `STEAMPIPE_DB=<connection string> go run .`.