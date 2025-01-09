-- testsuites
ALTER TABLE testsuites
DROP COLUMN IF EXISTS properties
;

ALTER TABLE testsuites_in
DROP COLUMN IF EXISTS properties
;

-- testcases
ALTER TABLE testcases
DROP COLUMN IF EXISTS properties
;

ALTER TABLE testcases_in
DROP COLUMN IF EXISTS properties
;
