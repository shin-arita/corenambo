#!/bin/bash
set -e

if [ "$(id -u)" = "0" ]; then
  mkdir -p "$PGDATA"
  chown -R postgres:postgres /var/lib/postgresql
  exec gosu postgres "$0" "$@"
fi

if [ ! -s "$PGDATA/PG_VERSION" ]; then
  /usr/lib/postgresql/16/bin/initdb -D "$PGDATA"

  /usr/lib/postgresql/16/bin/pg_ctl -D "$PGDATA" -o "-c listen_addresses='localhost'" -w start

  for f in /docker-entrypoint-initdb.d/*; do
    case "$f" in
      *.sql)
        psql -v ON_ERROR_STOP=1 -U postgres -f "$f"
        ;;
      *.sh)
        bash "$f"
        ;;
    esac
  done

  /usr/lib/postgresql/16/bin/pg_ctl -D "$PGDATA" -m fast -w stop
fi

exec "$@"