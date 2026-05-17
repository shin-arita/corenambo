CREATE TABLE user_profiles (
                               id UUID PRIMARY KEY,
                               user_id UUID NOT NULL,
                               family_name VARCHAR(100) NOT NULL,
                               given_name VARCHAR(100) NOT NULL,
                               middle_name VARCHAR(100),
                               phonetic_family_name VARCHAR(100),
                               phonetic_given_name VARCHAR(100),
                               birth_date DATE NOT NULL,
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                               CONSTRAINT fk_user_profiles_user_id
                                   FOREIGN KEY (user_id)
                                       REFERENCES users(id)
                                       ON DELETE RESTRICT,

                               CONSTRAINT chk_user_profiles_family_name_not_blank
                                   CHECK (btrim(family_name) <> ''),

                               CONSTRAINT chk_user_profiles_given_name_not_blank
                                   CHECK (btrim(given_name) <> ''),

                               CONSTRAINT chk_user_profiles_middle_name_not_blank
                                   CHECK (middle_name IS NULL OR btrim(middle_name) <> ''),

                               CONSTRAINT chk_user_profiles_phonetic_family_name_not_blank
                                   CHECK (phonetic_family_name IS NULL OR btrim(phonetic_family_name) <> ''),

                               CONSTRAINT chk_user_profiles_phonetic_given_name_not_blank
                                   CHECK (phonetic_given_name IS NULL OR btrim(phonetic_given_name) <> ''),

                               CONSTRAINT chk_user_profiles_birth_date_not_future
                                   CHECK (birth_date <= CURRENT_DATE)
);

COMMENT ON TABLE user_profiles IS 'ユーザプロフィール';

COMMENT ON COLUMN user_profiles.id IS 'ID';
COMMENT ON COLUMN user_profiles.user_id IS 'ユーザID';
COMMENT ON COLUMN user_profiles.family_name IS '姓';
COMMENT ON COLUMN user_profiles.given_name IS '名';
COMMENT ON COLUMN user_profiles.middle_name IS 'ミドルネーム';
COMMENT ON COLUMN user_profiles.phonetic_family_name IS '姓の発音表記';
COMMENT ON COLUMN user_profiles.phonetic_given_name IS '名の発音表記';
COMMENT ON COLUMN user_profiles.birth_date IS '生年月日';
COMMENT ON COLUMN user_profiles.created_at IS '作成日時';
COMMENT ON COLUMN user_profiles.updated_at IS '更新日時';

CREATE UNIQUE INDEX uq_user_profiles_user_id
    ON user_profiles(user_id);

CREATE TABLE user_profile_details (
                                      id UUID PRIMARY KEY,
                                      user_id UUID NOT NULL,
                                      phone_country_code VARCHAR(10),
                                      phone_number VARCHAR(30),
                                      country_code CHAR(2),
                                      postal_code VARCHAR(20),
                                      region VARCHAR(100),
                                      locality VARCHAR(100),
                                      address_line1 VARCHAR(255),
                                      address_line2 VARCHAR(255),
                                      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                      CONSTRAINT fk_user_profile_details_user_id
                                          FOREIGN KEY (user_id)
                                              REFERENCES users(id)
                                              ON DELETE RESTRICT,

                                      CONSTRAINT chk_user_profile_details_phone_country_code_not_blank
                                          CHECK (phone_country_code IS NULL OR btrim(phone_country_code) <> ''),

                                      CONSTRAINT chk_user_profile_details_phone_number_not_blank
                                          CHECK (phone_number IS NULL OR btrim(phone_number) <> ''),

                                      CONSTRAINT chk_user_profile_details_country_code_not_blank
                                          CHECK (country_code IS NULL OR btrim(country_code) <> ''),

                                      CONSTRAINT chk_user_profile_details_postal_code_not_blank
                                          CHECK (postal_code IS NULL OR btrim(postal_code) <> ''),

                                      CONSTRAINT chk_user_profile_details_region_not_blank
                                          CHECK (region IS NULL OR btrim(region) <> ''),

                                      CONSTRAINT chk_user_profile_details_locality_not_blank
                                          CHECK (locality IS NULL OR btrim(locality) <> ''),

                                      CONSTRAINT chk_user_profile_details_address_line1_not_blank
                                          CHECK (address_line1 IS NULL OR btrim(address_line1) <> ''),

                                      CONSTRAINT chk_user_profile_details_address_line2_not_blank
                                          CHECK (address_line2 IS NULL OR btrim(address_line2) <> '')
);

COMMENT ON TABLE user_profile_details IS 'ユーザプロフィール詳細';

COMMENT ON COLUMN user_profile_details.id IS 'ID';
COMMENT ON COLUMN user_profile_details.user_id IS 'ユーザID';
COMMENT ON COLUMN user_profile_details.phone_country_code IS '電話番号国番号';
COMMENT ON COLUMN user_profile_details.phone_number IS '電話番号';
COMMENT ON COLUMN user_profile_details.country_code IS '国コード';
COMMENT ON COLUMN user_profile_details.postal_code IS '郵便番号';
COMMENT ON COLUMN user_profile_details.region IS '地域';
COMMENT ON COLUMN user_profile_details.locality IS '市区町村';
COMMENT ON COLUMN user_profile_details.address_line1 IS '住所1';
COMMENT ON COLUMN user_profile_details.address_line2 IS '住所2';
COMMENT ON COLUMN user_profile_details.created_at IS '作成日時';
COMMENT ON COLUMN user_profile_details.updated_at IS '更新日時';

CREATE UNIQUE INDEX uq_user_profile_details_user_id
    ON user_profile_details(user_id);

CREATE INDEX idx_user_profile_details_country_code
    ON user_profile_details(country_code);
