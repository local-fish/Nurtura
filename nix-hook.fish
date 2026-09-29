export PGDATA=$PWD/.pgdata
export PGHOST=$PWD/.pgsock
export PGUSER=postgres
export PGDATABASE=postgres
mkdir -p "$PGHOST"

if [ ! -d "$PGDATA" ]; then
initdb -U postgres --auth=trust -E UTF8 >/dev/null
fi

if ! pg_isready -q; then
pg_ctl -l "$PGDATA/server.log" -o "-c listen_addresses= -c unix_socket_directories=$PGHOST -c ssl=off" start
fi

tmux -L "$1" new-session -A -s "$1"
