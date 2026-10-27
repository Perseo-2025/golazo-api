CREATE TABLE wallets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL,
    type                TEXT NOT NULL DEFAULT 'real',
    currency            TEXT NOT NULL DEFAULT 'PEN',
    balance_cents       BIGINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT wallets_user_fk          FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT wallets_type_check       CHECK (type IN ('real', 'bonus')),
    CONSTRAINT wallets_currency_check   CHECK (currency IN ('PEN')),
    CONSTRAINT wallets_balance_check    CHECK (balance_cents >= 0),
    CONSTRAINT wallets_user_type_unique UNIQUE (user_id, type, currency)

)