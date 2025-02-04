# Berry

Berry is a searchable glossary bot for Discord.

## Requirements

- PostgreSQL (only 12.5 tested)
- Go 1.16 or later
- For the site and API: a reverse proxy (such as Caddy or nginx)

## Configuration

Each executable has its own configuration file, located next to the executable.

## Running

This is still very much a WIP. Although we've done our best to not hardcode names and links into the source code, they still show up in some places (notably the site), so we recommend going through the source and fixing all those references before you run it.

All of these executables are standalone and can be run independently from each other.

### Bot

1. `go build` in the cmd/bot directory
2. Copy `config.sample.json` to `config.json` and fill it in
3. Run the executable

#### Required configuration keys

```
- auth:
  - token (string): Discord bot token
  - database_url (string): dsn for the Postgres database
- bot:
  - prefixes: ([]string): default prefixes used
  - bot_owners: ([]int): bot owner IDs, these users can use all commands including admin commands
```

All other keys are optional but strongly recommended.

### API

1. `go build` in the cmd/bot directory
2. Copy `config.sample.yaml` to `config.yaml` and fill it in
3. Run the executable

### Site

1. `go build` in the cmd/bot directory
2. Copy `config.sample.yaml` to `config.yaml` and fill it in
3. Run the executable

## License

Copyright (C) 2021, Starshine System

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.