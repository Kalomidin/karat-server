# Karat server

Standalone API for the Karat stock book. It keeps two things in Postgres:

- a user, taken from Google login: Google id, email, name, and photo
- that user's jewellery pieces

The phone app sends the Google ID token to `POST /v1/session` after sign-in.

## Run

Docker Desktop has to be running.

```sh
cd karat-server
docker compose up --build
```

The API listens on port 8080. Postgres listens on port 5432. Data stays in the `karat-pg` volume. `docker compose down` stops the containers and keeps the data. `docker compose down -v` deletes it.

To run the Go process on the machine and only Postgres in Docker:

```sh
docker compose up db
cp .env.example .env
go run ./cmd/karat-server
```

## Google login

`POST /v1/session` with `Authorization: Bearer <Google ID token>` checks the token, then inserts or updates one row:

| column | source |
| --- | --- |
| `google_id` | token `sub`, the stable Google account id |
| `email` | token `email` |
| `name` | token `name`, or the given name and family name |
| `photo_url` | token `picture` |

The same Google id logs in again by updating those details. `GET /v1/me` returns the stored account.

`GOOGLE_CLIENT_IDS` lists the OAuth client ids allowed to mint that token. The iOS client id is the default. Android tokens use the web client id, so add that id to the list when it exists.

`DEV_LOGIN=1` also accepts `Bearer dev:<google id>:<email>:<name>` so the API can be tried without Google. Leave it off on any server other people can reach.

```sh
curl -H 'Authorization: Bearer dev:109820475938:ada@atelier.test:Ada' \
  http://localhost:8080/v1/session
```

## Pieces

Items belong to the Google id from the token.

- `GET /v1/items`
- `POST /v1/items`
- `GET /v1/items/{id}`
- `PUT /v1/items/{id}`
- `DELETE /v1/items/{id}`
- `POST /v1/items/{id}/sale` with `soldOn` (`YYYY-MM-DD`) and `soldCents`
- `POST /v1/items/{id}/return`

`kind` is `ring`, `necklace`, `bracelet`, `earrings`, `pendant`, `brooch`, or `other`. `metal` is `gold`, `silver`, `platinum`, or `mixed`. Prices are integer cents. A sale needs both a day and a price, and the day cannot be after today in `SHOP_TIMEZONE`.

`GET /healthz` returns `{"ok":true}` when Postgres answers.
