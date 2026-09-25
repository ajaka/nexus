CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();


CREATE TRIGGER trigger_buyer_address_updated_at
BEFORE UPDATE ON buyer_address
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_seller_profile_updated_at
BEFORE UPDATE ON seller_profile
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_seller_storefront_updated_at
BEFORE UPDATE ON seller_storefront
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();
