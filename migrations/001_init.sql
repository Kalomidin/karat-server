CREATE TABLE users (
    google_id text PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE items (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users (google_id) ON DELETE CASCADE,
    name text NOT NULL,
    stock_number text NOT NULL DEFAULT '',
    kind text NOT NULL,
    metal text NOT NULL,
    purity text NOT NULL,
    weight_grams double precision NOT NULL,
    cost_cents bigint NOT NULL,
    notes text NOT NULL DEFAULT '',
    created_on date NOT NULL,
    sold_on date,
    sold_cents bigint,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT items_sale_pair CHECK ((sold_on IS NULL) = (sold_cents IS NULL)),
    CONSTRAINT items_kind CHECK (kind IN ('ring', 'necklace', 'bracelet', 'earrings', 'pendant', 'brooch', 'other')),
    CONSTRAINT items_metal CHECK (metal IN ('gold', 'silver', 'platinum', 'mixed')),
    CONSTRAINT items_weight CHECK (weight_grams > 0 AND weight_grams <= 100000),
    CONSTRAINT items_cost CHECK (cost_cents >= 0 AND cost_cents <= 10000000000),
    CONSTRAINT items_sale_price CHECK (sold_cents IS NULL OR (sold_cents >= 0 AND sold_cents <= 10000000000))
);

CREATE INDEX items_user_created_idx ON items (user_id, created_at DESC, id DESC);
