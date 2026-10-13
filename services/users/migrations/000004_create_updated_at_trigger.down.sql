DROP TRIGGER IF EXISTS trigger_buyer_address_updated_at ON buyer_address;
DROP TRIGGER IF EXISTS trigger_seller_profile_updated_at ON seller_profile;
DROP TRIGGER IF EXISTS trigger_seller_storefront_updated_at ON seller_storefront;
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at();
