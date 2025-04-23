-- Insert admin user with bcrypt hashed password
-- Default password: admin123 (you should change this in production)
INSERT INTO users (username, password_hash)
VALUES 
    ('admin', '$2a$10$RoWGjii1WP1hyhY1Su4XROcLkMwVKYpb1kCTFzY7JEivQrNcau9N6')
ON CONFLICT (username) DO NOTHING;
