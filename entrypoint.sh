#!/bin/sh
set -e

# Copy example config if config.toml doesn't exist
if [ ! -f /app/config.toml ]; then
  cp /app/config.toml.example /app/config.toml
fi

# Override config with environment variables if provided
if [ -n "$DB_HOST" ]; then
  sed -i '/^\[database\]/,/^\[/ s/^host[[:space:]]*=.*/host = "'"$DB_HOST"'"/' /app/config.toml
fi
if [ -n "$DB_PORT" ]; then
  sed -i '/^\[database\]/,/^\[/ s/^port[[:space:]]*=.*/port = '"$DB_PORT"'/' /app/config.toml
fi
if [ -n "$DB_USER" ]; then
  sed -i '/^\[database\]/,/^\[/ s/^user[[:space:]]*=.*/user = "'"$DB_USER"'"/' /app/config.toml
fi
if [ -n "$DB_PASSWORD" ]; then
  sed -i '/^\[database\]/,/^\[/ s/^password[[:space:]]*=.*/password = "'"$DB_PASSWORD"'"/' /app/config.toml
fi
if [ -n "$DB_NAME" ]; then
  sed -i '/^\[database\]/,/^\[/ s/^name[[:space:]]*=.*/name = "'"$DB_NAME"'"/' /app/config.toml
fi

# Wait for database if DB_HOST and DB_PORT are set
if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
    echo "Waiting for database at $DB_HOST:$DB_PORT..."
    while ! nc -z "$DB_HOST" "$DB_PORT"; do
      sleep 1
    done
    echo "Database is up!"
fi

# Run database migrations
echo "Running database migrations..."
./gorundeck migrate

# Create initial admin if requested via environment variables
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "Checking/Creating initial admin: $ADMIN_EMAIL"
    ./gorundeck admin --add-user "$ADMIN_EMAIL:$ADMIN_PASSWORD" || echo "Admin user check/creation finished."
fi

# Execute the main application
exec "$@"
