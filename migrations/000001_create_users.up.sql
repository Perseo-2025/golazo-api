CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_type   TEXT NOT NULL,
    document_number TEXT NOT NULL,
    first_name      TEXT NOT NULL,
    last_name       TEXT NOT NULL,
    birth_date      DATE NOT NULL,
    email           TEXT NOT NULL,
    phone           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_document_type_check    CHECK (document_type IN ('DNI', 'CE', 'PASSPORT')),
    CONSTRAINT users_status_check           CHECK (status IN ('active','suspended','self_excluded')),
    CONSTRAINT users_email_lowercase        CHECK (email = lower(email)),
    CONSTRAINT users_document_unique        UNIQUE (document_type, document_number),
    CONSTRAINT users_email_unique           UNIQUE (email)

)