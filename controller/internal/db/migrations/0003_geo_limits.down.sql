-- +migrate Down
ALTER TABLE nodes DROP COLUMN IF EXISTS city;
ALTER TABLE nodes DROP COLUMN IF EXISTS country;
ALTER TABLE node_metrics DROP COLUMN IF EXISTS cpu_limit_pct;
ALTER TABLE node_metrics DROP COLUMN IF EXISTS ram_limit_mb;
ALTER TABLE node_metrics DROP COLUMN IF EXISTS battery_pct;
ALTER TABLE node_metrics DROP COLUMN IF EXISTS charging;
ALTER TABLE node_metrics DROP COLUMN IF EXISTS network_type;
