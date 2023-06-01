-- The type column groups a set of targets.
-- For instance products with target 2, 8 and 10 is of type appcat.

ALTER TABLE products
  ADD COLUMN type TEXT;
