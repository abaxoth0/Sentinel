#!/bin/sh

# Wait for primary to be ready
while ! pg_isready -h postgres-primary -p 5433 -U replicator; do
    echo 'Waiting for primary database...'
    sleep 2
done

chown -R postgres:postgres /var/lib/postgresql/replica_data
chmod -R 0700 /var/lib/postgresql/replica_data

# Initialize replica if empty
if [ ! -f /var/lib/postgresql/replica_data/postgresql.conf ]; then
    echo 'Initializing replica from primary...'
    su postgres -c "PGPASSWORD=1234 pg_basebackup -h postgres-primary -p 5433 -U replicator -D /var/lib/postgresql/replica_data -Fp -Xs -P -R"
    echo "standby_mode = 'on'" > /var/lib/postgresql/replica_data/standby.signal
    chown postgres:postgres /var/lib/postgresql/replica_data/standby.signal
fi

# Start PostgreSQL
exec su postgres -c 'postgres -c config_file=/etc/postgresql/replica.conf -c port=5434 -c hba_file=/etc/postgresql/pg_hba.replica.conf'

