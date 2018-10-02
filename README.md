## radar

Keep track of links you need to get back to. It creates a GitHub issue every 24 hours (or when a given signal is sent to the process) which is a checklist of all the links you want to keep track of.

## Usage

Use of the Docker image is recommended. [Check the tags](https://hub.docker.com/r/parkr/radar/tags/) for the one you want, then run:

    docker run --rm \
      -e RADAR_MYSQL_URL=user:pass@host:port/database?parseTime=true \
      -e RADAR_ALLOWED_SENDERS=you@gmail.com \
      -e GITHUB_ACCESS_TOKEN=aaabb \
      -e RADAR_REPO=owner/name \ # where to create the new github issue
      -e RADAR_MENTION=@username \
      -e RADAR_HEALTHCHECK_URL=http://localhost:8921/health \
      -e MG_API_KEY=abcdef \
      -e MG_DOMAIN=example.com \
      -e MG_PUBLIC_API_KEY=ghijkl \
      parkr/radar:$TAG \
      radar -http=:8921 -hour=3

The `MG_` environment variables allows this server to reply to each incoming email via [Mailgun](https://mailgun.com). Other providers are not supported, but could be with very few modifications.

The only required parameters are: `RADAR_MYSQL_URL`, `RADAR_ALLOWED_SENDERS`, `RADAR_REPO`, and `GITHUB_ACCESS_TOKEN`. All others are optional.

The `-http` command line argument provides the bind address. Make sure you update `RADAR_HEALTHCHECK_URL` to match if you modify this.

The `-hour` command line argument tells the server when to generate the new radar issue.

## License

MIT, Copyright Parker Moore 2018.

