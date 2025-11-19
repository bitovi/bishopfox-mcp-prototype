-- Contains a copy of all asset data.
CREATE TABLE assets (
    -- Asset ID
    id UUID NOT NULL,
    -- Organization ID this asset belongs to.
    org_id UUID NOT NULL,
    -- Asset type
    type TEXT NOT NULL,
    -- Parent asset
    parent_id UUID,
    parent_type TEXT,
    -- JSON data with asset details
    details JSONB NOT NULL,
    tags TEXT[] NOT NULL,
    link TEXT GENERATED ALWAYS AS (
        CASE 
            WHEN type = 'domain' THEN 'https://ui.api.non.usea2.bf9.io/' || org_id || '/assets/domains/' || (details->>'name')
            WHEN type = 'subdomain' THEN 'https://ui.api.non.usea2.bf9.io/' || org_id || '/assets/subdomains/' || (details->>'name')
            WHEN type = 'ip' THEN 'https://ui.api.non.usea2.bf9.io/' || org_id || '/assets/ips/' || (details->>'ip')
            WHEN type = 'service' THEN 'https://ui.api.non.usea2.bf9.io/' || org_id || '/assets/services/' || id
            ELSE NULL
        END
    ) STORED,
    PRIMARY KEY (org_id, id)
) PARTITION BY LIST (org_id);

-- Our approach here is no indexes other than the org partition. Each org should have a
-- reasonable number of assets. Some bigname orgs might have a LOT of assets, which MIGHT
-- need additional indexes, but, for the most part, this table's traffic is MANUAL. The
-- user needs to directly type something in to query it. It's not part of UI
-- functionality, so high performance is less of a factor. It should also scale
-- horizontally across additional readers.
-- 
-- A table scan even over a million asset records is not very bad if it's done sparingly.
-- If we do start to see slowdowns due to high volume assets like IP addresses, those can
-- be handled separately.
--
-- The cost of running AI inference should outweigh the cost of the database queries.

-- Create 5 org partitions.
CREATE TABLE assets_org_111111111111 PARTITION OF assets FOR VALUES IN ('11111111-1111-1111-1111-111111111111');
CREATE TABLE assets_org_222222222222 PARTITION OF assets FOR VALUES IN ('22222222-2222-2222-2222-222222222222');
CREATE TABLE assets_org_333333333333 PARTITION OF assets FOR VALUES IN ('33333333-3333-3333-3333-333333333333');
CREATE TABLE assets_org_444444444444 PARTITION OF assets FOR VALUES IN ('44444444-4444-4444-4444-444444444444');
CREATE TABLE assets_org_555555555555 PARTITION OF assets FOR VALUES IN ('55555555-5555-5555-5555-555555555555');

-- These roles only have access to SELECT queries on their own tables.
CREATE ROLE customer_query_role_111111111111 NOLOGIN NOINHERIT;
CREATE ROLE customer_query_role_222222222222 NOLOGIN NOINHERIT;
CREATE ROLE customer_query_role_333333333333 NOLOGIN NOINHERIT;
CREATE ROLE customer_query_role_444444444444 NOLOGIN NOINHERIT;
CREATE ROLE customer_query_role_555555555555 NOLOGIN NOINHERIT;
GRANT SELECT ON assets_org_111111111111 TO customer_query_role_111111111111;
GRANT SELECT ON assets_org_222222222222 TO customer_query_role_222222222222;
GRANT SELECT ON assets_org_333333333333 TO customer_query_role_333333333333;
GRANT SELECT ON assets_org_444444444444 TO customer_query_role_444444444444;
GRANT SELECT ON assets_org_555555555555 TO customer_query_role_555555555555;
