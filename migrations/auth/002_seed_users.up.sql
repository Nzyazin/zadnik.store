-- Insert admin user with bcrypt hashed password
-- Default password: admin123 (you should change this in production)
INSERT INTO users (username, password)
VALUES 
    ('admin', '$2a$10$aqtifq5lIBLZT8AckY.B0OU7PNW6xAgNIsZW21PU.53HFH1aci2AS')
ON CONFLICT (username) DO NOTHING;
