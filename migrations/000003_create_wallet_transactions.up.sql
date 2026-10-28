CREATE TABLE wallet_transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id           UUID NOT NULL,
    type                TEXT NOT NULL,
    amount_cents        BIGINT NOT NULL,
    balance_after_cents BIGINT NOT NULL,
    reference_id        UUID,
    idempotency_key     TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT wallet_transactions_wallet_fk            FOREIGN KEY (wallet_id) REFERENCES wallets (id) ON DELETE RESTRICT,
    CONSTRAINT wallet_transactions_type_check           CHECK (type IN ('deposit','withdrawal','bet_stake','bet_win','bet_refund','bonus','adjustment')),
    CONSTRAINT wallet_transactions_amount_check         CHECK (amount_cents <> 0),
    CONSTRAINT wallet_transactions_balance_after_check  CHECK (balance_after_cents >= 0),
    CONSTRAINT wallet_transactions_idempotency_unique   UNIQUE (wallet_id, idempotency_key)
);

CREATE INDEX wallet_transactions_wallet_created_idx
    ON wallet_transactions (wallet_id, created_at DESC);