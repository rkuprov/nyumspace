-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS nyum_appliance
(
    id                BIGSERIAL PRIMARY KEY,
    name              VARCHAR(255),
    type              VARCHAR(50),
    brand             VARCHAR(100),
    model             VARCHAR(100),
    manufacturer      VARCHAR(100),
    serial_number     VARCHAR(100),
    part_number       VARCHAR(100),
    manual_url        VARCHAR(255),
    default_image_url VARCHAR(255),
    description       TEXT,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_nyum_appliance_name ON nyum_appliance(name);
CREATE INDEX IF NOT EXISTS idx_nyum_appliance_type ON nyum_appliance(type);
CREATE INDEX IF NOT EXISTS idx_nyum_part_number ON nyum_appliance(part_number);

CREATE TABLE IF NOT EXISTS nyum_appliance_metadata
(
    house_id UUID REFERENCES homes(id) ON DELETE CASCADE,
    appliance_id BIGINT REFERENCES nyum_appliance(id),
    location VARCHAR(255),
    date_acquired TIMESTAMP WITH TIME ZONE,
    date_installed TIMESTAMP WITH TIME ZONE,
    note TEXT,
    warranty_expiration TIMESTAMP WITH TIME ZONE,
    warranty_information TEXT,
    image_url VARCHAR(255),
    purchase_price DECIMAL(10, 2),
    purchase_receipt_url VARCHAR(255)
);
CREATE INDEX IF NOT EXISTS idx_nyum_appliance_metadata_house_id ON nyum_appliance_metadata(house_id);
CREATE INDEX IF NOT EXISTS idx_nyum_appliance_metadata_appliance_id ON nyum_appliance_metadata(appliance_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE nyum_appliance_metadata;
DROP TABLE nyum_appliance;
-- +goose StatementEnd
