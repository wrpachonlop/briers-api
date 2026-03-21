-- Run this in your Supabase SQL editor or psql
-- =============================================================
-- EXTENSIONS
-- =============================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================
-- PROFILES
-- =============================================================
CREATE TABLE IF NOT EXISTS profiles (
    id          UUID PRIMARY KEY,  -- matches auth.users.id
    full_name   TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'seller'
                    CHECK (role IN ('admin', 'manager', 'seller')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_profiles_role ON profiles(role);

-- =============================================================
-- PRODUCTS
-- =============================================================
CREATE TABLE IF NOT EXISTS products (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_name   TEXT NOT NULL,
    store_name      TEXT NOT NULL,
    product_type    VARCHAR(20) NOT NULL CHECK (product_type IN ('modular', 'fixed')),
    description     TEXT,
    image_url       TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_by      UUID REFERENCES profiles(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_products_provider ON products(lower(provider_name));
CREATE INDEX IF NOT EXISTS idx_products_store    ON products(lower(store_name));
CREATE INDEX IF NOT EXISTS idx_products_type     ON products(product_type);
CREATE INDEX IF NOT EXISTS idx_products_active   ON products(is_active);

-- =============================================================
-- SECTIONS
-- =============================================================
CREATE TABLE IF NOT EXISTS sections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    width_cm        NUMERIC(8,2) NOT NULL CHECK (width_cm > 0),
    height_cm       NUMERIC(8,2) NOT NULL CHECK (height_cm > 0),
    depth_cm        NUMERIC(8,2) NOT NULL CHECK (depth_cm > 0),
    fabric_yards    NUMERIC(8,2) NOT NULL CHECK (fabric_yards > 0),
    image_url       TEXT NOT NULL,
    sort_order      INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_section_dimensions UNIQUE (product_id, width_cm, height_cm, depth_cm)
);

CREATE INDEX IF NOT EXISTS idx_sections_product ON sections(product_id);

-- =============================================================
-- FABRIC PRICES
-- =============================================================
CREATE TABLE IF NOT EXISTS fabric_prices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id      UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    grade           INTEGER NOT NULL,
    supplier_cost   NUMERIC(10,2) NOT NULL CHECK (supplier_cost > 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_grade_valid CHECK (grade IN (5, 10, 15, 20, 25, 30, 35, 40)),
    CONSTRAINT uq_fabric_price UNIQUE (product_id, grade)
);

CREATE INDEX IF NOT EXISTS idx_fabric_product ON fabric_prices(product_id);

-- =============================================================
-- EXTRA CHARGES
-- =============================================================
CREATE TABLE IF NOT EXISTS extra_charges (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT NOT NULL,
    amount      NUMERIC(10,2) NOT NULL CHECK (amount >= 0),
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- =============================================================
-- QUOTES
-- =============================================================
CREATE TABLE IF NOT EXISTS quotes (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id          UUID NOT NULL REFERENCES products(id),
    created_by          UUID NOT NULL REFERENCES profiles(id),
    customer_name       TEXT,
    fabric_grade        INTEGER NOT NULL
                            CONSTRAINT chk_quote_grade CHECK (
                                fabric_grade IN (5, 10, 15, 20, 25, 30, 35, 40)
                            ),
    supplier_cost       NUMERIC(10,2) NOT NULL,
    final_price         NUMERIC(10,2) NOT NULL,
    total_width_cm      NUMERIC(10,2) NOT NULL,
    total_depth_cm      NUMERIC(10,2) NOT NULL,
    total_fabric_yards  NUMERIC(10,2) NOT NULL,
    notes               TEXT,
    status              TEXT NOT NULL DEFAULT 'draft'
                            CHECK (status IN ('draft', 'sent', 'accepted', 'declined')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_quotes_created_by ON quotes(created_by);
CREATE INDEX IF NOT EXISTS idx_quotes_product    ON quotes(product_id);
CREATE INDEX IF NOT EXISTS idx_quotes_status     ON quotes(status);

-- =============================================================
-- QUOTE SECTIONS
-- =============================================================
CREATE TABLE IF NOT EXISTS quote_sections (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_id    UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    section_id  UUID NOT NULL REFERENCES sections(id),
    quantity    INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    rotation    INTEGER NOT NULL DEFAULT 0 CHECK (rotation IN (0, 90, 180, 270)),
    position_x  INTEGER,
    position_y  INTEGER
);

CREATE INDEX IF NOT EXISTS idx_qs_quote ON quote_sections(quote_id);

-- =============================================================
-- QUOTE EXTRA CHARGES
-- =============================================================
CREATE TABLE IF NOT EXISTS quote_extra_charges (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_id        UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    extra_charge_id UUID REFERENCES extra_charges(id),
    name            TEXT NOT NULL,
    amount          NUMERIC(10,2) NOT NULL,
    quantity        INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_qec_quote ON quote_extra_charges(quote_id);

-- =============================================================
-- TRIGGER: auto-update updated_at
-- =============================================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_profiles_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_fabric_prices_updated_at
    BEFORE UPDATE ON fabric_prices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- =============================================================
-- SEED: Extra Charges catalogue
-- =============================================================
INSERT INTO extra_charges (name, amount, description) VALUES
    ('Leg Change',      150.00, 'Upgrade to custom leg style'),
    ('Extra Cushion',    75.00, 'Additional throw cushion'),
    ('Down Fill Upgrade', 200.00, 'Upgrade seat fill to down blend'),
    ('Ottoman',         400.00, 'Matching ottoman'),
    ('Delivery & Setup', 250.00, 'White-glove delivery and setup')
ON CONFLICT DO NOTHING;

-- =============================================================
-- ROW LEVEL SECURITY (Supabase)
-- Run these if using Supabase direct client
-- =============================================================
-- ALTER TABLE profiles      ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE products      ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE sections      ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE fabric_prices ENABLE ROW LEVEL SECURITY;
-- ALTER TABLE quotes        ENABLE ROW LEVEL SECURITY;
