# Installation

To run Airstation on your machine, there are two ways: using [Docker](https://docs.docker.com/) (recommended) or building it yourself using the [Go](https://go.dev/) compiler for server and [Node.js](https://nodejs.org/) with [npm](https://www.npmjs.com/) for web clients.

## Docker

1.  Clone Airstation repository

    ```sh
    git clone https://github.com/cheatsnake/airstation.git
    ```

    ```sh
    cd ./airstation
    ```

2.  Setup environment variables

    Next you need an `.env` file with secret keys

    ```sh
    touch .env
    ```

    Inside this file you must define 2 variables:

    ```
    AIRSTATION_SECRET_KEY=
    AIRSTATION_JWT_SIGN=
    ```

    > `AIRSTATION_SECRET_KEY` - the secret key you need to log in to the station control panel <br> `AIRSTATION_JWT_SIGN` - the key to sign the JWT session

    > Use [random string generator](https://it-tools.tech/token-generator?length=20) with a length of at least 10 characters for these variables!

3.  Build a docker image and start a new container

    ```sh
    docker compose up -d
    ```

And finally you can see:

- Control panel on [http://localhost:7331/studio/](http://localhost:7331/studio/) (extra slash matters!)
- Radio player on [http://localhost:7331](http://localhost:7331)

To stop the container, just type:

```sh
docker compose down
```

### Docker Compose

You can get pre-built image from [Docker Hub](https://hub.docker.com/r/cheatsnake/airstation) and run it quickly with custom `docker-compose.yml` file as shown bellow:

```yml
# docker-compose.yml
services:
  airstation:
    image: cheatsnake/airstation:latest
    ports:
      - "7331:7331"
    volumes:
      - airstation-data:/app/storage
      - ./static:/app/static
    restart: unless-stopped
    environment:
      AIRSTATION_SECRET_KEY: ${AIRSTATION_SECRET_KEY:-PASTE_YOUR_OWN_KEY}
      AIRSTATION_JWT_SIGN: ${AIRSTATION_JWT_SIGN:-PASTE_RANDOM_STRING}
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:7331/"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 10s

volumes:
  airstation-data:
```

> Don't forget to modify environment variables inside this file or via your own `.env` file in the same directory as the `docker-compose.yml`

## Build from source

1. Follow steps 1 and 2 from the previous section

2. Install dependencies

```sh
npm ci --prefix ./web/player
```

```sh
npm ci --prefix ./web/studio
```

3. Build web clients

```sh
npm run build --prefix ./web/player
```

```sh
npm run build --prefix ./web/studio
```

4. Build server

```sh
go build ./cmd/main.go
```

5. Run app

```sh
./main
```

See the result on [http://localhost:7331](http://localhost:7331) and [http://localhost:7331/studio/](http://localhost:7331/studio/) (extra slash matters!)

## Development mode

To run the application in development mode, start each part of the application using the commands below:

```sh
npm run dev --prefix ./web/player
```

```sh
npm run dev --prefix ./web/studio
```

```sh
go run ./cmd/main.go
```
