#!/bin/bash

# Configuration - Set this paths to existing directories with postgres DBs
PRIMARY_DATA="/home/$USER/postgres/pg_primary_data"
REPLICA_DATA="/home/$USER/postgres/pg_replica_data"

PG_CTL="pg_ctl" # Can be either:
                # 1. Full path to binary: "/usr/bin/pg_ctl"
                # 2. Command name if in PATH: "pg_ctl" (like in my case)

validate_paths() {
    local errors=0

    # Check database directories
    if [ ! -d "$PRIMARY_DATA" ]; then
        echo "ERROR: Primary data directory not found: $PRIMARY_DATA" >&2
        errors=$((errors + 1))
    fi
    if [ ! -d "$REPLICA_DATA" ]; then
        echo "ERROR: Replica data directory not found: $REPLICA_DATA" >&2
        errors=$((errors + 1))
    fi

    # Check pg_ctl accessibility - supports both path and command name
    if [[ "$PG_CTL" == */* ]]; then
        # Path specified - check if exists and is executable
        if [ ! -f "$PG_CTL" ]; then
            echo "ERROR: pg_ctl binary not found at specified path: $PG_CTL" >&2
            errors=$((errors + 1))
        elif [ ! -x "$PG_CTL" ]; then
            echo "ERROR: pg_ctl binary is not executable: $PG_CTL" >&2
            errors=$((errors + 1))
        fi
    else
        # Command name specified - check if available in PATH
        if ! command -v "$PG_CTL" &> /dev/null; then
            echo "ERROR: pg_ctl command not found. Ensure it's in your PATH" >&2
            errors=$((errors + 1))
        fi
    fi

    return $errors
}

start_db() {
    echo "Starting PostgreSQL in: $1"
    "$PG_CTL" -D "$1" -l "$1/startup.log" start
    local status=$?
    if [ $status -eq 0 ]; then
        echo "Successfully started PostgreSQL"
    else
        echo "ERROR: Failed to start PostgreSQL (exit code $status)" >&2
    fi
    return $status
}

stop_db() {
    echo "Stopping PostgreSQL in: $1"
    "$PG_CTL" -D "$1" stop -m fast
    local status=$?
    if [ $status -eq 0 ]; then
        echo "Successfully stopped PostgreSQL"
    else
        echo "ERROR: Failed to stop PostgreSQL (exit code $status)" >&2
    fi
    return $status
}

# Handle single database instance
manage_db() {
    case $1 in
        primary)
            case $2 in
                start) start_db "$PRIMARY_DATA" ;;
                stop) stop_db "$PRIMARY_DATA" ;;
            esac ;;
        replica)
            case $2 in
                start) start_db "$REPLICA_DATA" ;;
                stop) stop_db "$REPLICA_DATA" ;;
            esac ;;
    esac
}

# Main script logic
case $1 in
    start|stop)
        action=$1
        target=${2:-both}  # Default to 'both' if no target specified

        # Validate paths and capture errors
        validate_paths
        errors=$?

        if [ $errors -gt 0 ]; then
            echo "Aborting due to $errors configuration errors" >&2
            exit 1
        fi

        case $target in
            primary|replica)
                echo "Action: $action"
                echo "Target: $target"
                echo ""

                manage_db "$target" "$action"
                ;;
            both)
                echo "Action: $action"
                echo "Target: both instances"
                echo ""

                manage_db "primary" "$action"
                echo ""
                manage_db "replica" "$action"
                ;;
            *)
                echo "ERROR: Invalid target: $target. Use primary/replica/both" >&2
                exit 1
                ;;
        esac
        ;;
    status)
        echo "Checking PostgreSQL status:"
        echo ""
        echo "Primary:"
        "$PG_CTL" -D "$PRIMARY_DATA" status
        echo ""
        echo "Replica:"
        "$PG_CTL" -D "$REPLICA_DATA" status
        ;;
    *)
        echo "Usage: $0 {start|stop|status} [primary|replica|both]" >&2
        echo "Example:" >&2
        echo "  $0 start both" >&2
        echo "  $0 stop replica" >&2
        echo "  $0 status" >&2
        exit 1
        ;;
esac

