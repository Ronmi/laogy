Simple tool to host your own url shorten service.

# Synopsis

```bash
# build binary
go get ./... && go build

# set environment variables and run
export MYSQL_DSN='user:password@(127.0.0.1:3306)/test?parseTime=true'
export BIND_ADDRESS=':8000'
export SHARED_SECRET='hackme'
./laogy

# or with docker
docker run -d --name laogy -e 'MYSQL_DSN=user:password@(mysql.server:3306)/test?parseTime=true' -e TOTP_SECRET=feed5eed0fdead70beef -p 8000:80 -v `pwd`/laogy:/laogy debian:stable-slim /laogy
```

You might want to take a llok at `docker-compose.yml`, `nginx.conf` and `index.html` as examples.

# Configurations

To best fit in container, configurations are passed through envvar.

- `MYSQL_DSN`: DSN string for connecting to MySQL server, see [DSN format](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for detail. YOU MUST SET `parseTime` TO `true`.
- `BIND_ADDRESS`: Ip and port to bind, default to ":80".
- `SHARED_SECRET`: Authenticates user with this shared password. Leave it unset to disable.
- `TOTP_SECRET`: A 10 bytes, hexdecimal encoded data as secret to authenticates user with TOTP password (use with Google Authenticator). Leave it unset to disable. This takes precedence over shared password.
- `TOTP_USER`: Username displayed in your TOTP app. Default to "admin".
- `TOTP_ISSUER`: Issuer displayed in your TOTP app. Default to "my URL shorter".

# Shorten an url

Submit a request (can be `GET` or `POST`) to `http://your.server/s` with following parameters:

- `url`: The url to be processed.
- `secret`: Your password setted in `SHARED_SERET` or app generated one if you're using `TOTP_SECRET`. You can bypass this parameter if authenticating is not enabled.

If success, it returns a JSON-encoded object with `{"data":{"code":"mycode"}}`. Otherwise a 4xx or 5xx HTTP status code is returned.
