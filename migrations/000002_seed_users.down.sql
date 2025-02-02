-- Remove seeded users
DELETE FROM users WHERE username IN ('admin');
