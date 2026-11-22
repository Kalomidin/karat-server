# Karat server

API for the Karat stock book. A Google login stores the account, and that account owns the jewellery pieces.

The code is split the same way as ketalk-api-go:

- `pkg/handler` reads HTTP and writes JSON
- `pkg/manager` holds the stock-book rules
- `pkg/manager/*/repository` talks to Postgres through Gorm
- `pkg/provider/google` checks the Google ID token

## Run

Docker Desktop has to be running.

```sh
cd karat-server
docker compose up --build
```

Postgres data stays in the `karat-pg` volume. `docker compose down` keeps it. `docker compose down -v` deletes it.

To run the API on the machine, start Postgres and then the server. `go run` reads `.env` in this directory. That file is gitignored; `.env.example` is the template.

```sh
docker compose up db
go run ./cmd/karat-server
```

## Google login

`POST /auth/session` with `{ "idToken": "<Google ID token>" }` checks the token and inserts or updates one user:

| column | source |
| --- | --- |
| `google_id` | token `sub` |
| `email` | token `email` |
| `name` | token `name`, or the given name and family name |
| `photo_url` | token `picture` |

Later calls send `Authorization: Bearer <same Google ID token>`. `GET /user/me` returns the stored account.

`DEV_LOGIN=true` also accepts `dev:<google id>:<email>:<name>` so the API can be tried without Google. Leave it off on any server other people can reach.

```sh
curl -H 'Content-Type: application/json' \
  -d '{"idToken":"dev:109820475938:ada@atelier.test:Ada"}' \
  http://localhost:8080/auth/session
```

## Pieces

- `GET /item`
- `POST /item`
- `GET /item/{id}`
- `PUT /item/{id}`
- `DELETE /item/{id}`
- `POST /item/upload-url` with `{ "contentType": "image/jpeg" }` returns a 15-minute R2 PUT URL
- `GET /currency`
- `POST /item/{id}/sale` with `soldOn` (`YYYY-MM-DD`) and `soldAmount`
- `POST /item/{id}/return`

`kind` is `ring`, `necklace`, `bracelet`, `earrings`, `pendant`, `brooch`, or `other`. `metal` is `gold`, `silver`, `platinum`, or `mixed`. `GET /currency` returns the currency rows (`id`, `code`, `name`) for dollar, ruble, and somoni. A piece stores that row’s `currencyId`. Prices are whole units of that currency. A sale needs both a day and a price, and the day cannot be after today in `SHOP_TIMEZONE`.

A piece can carry `imageKey`. Photos are stored in the `karat` bucket under `items/<google id>/`. The key has to be one this account received from `POST /item/upload-url`. Item responses add `imageUrl`, a one-hour read-only presigned link. The bucket stays private. Leave the R2 variables empty and the API still starts; the upload route then answers `image_unavailable`.

`GET /healthz` returns `{"ok":true}` when Postgres answers.
