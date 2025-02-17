-- Insert admin user with bcrypt hashed password
-- Default password: admin123 (you should change this in production)
INSERT INTO products (name, slug, description, price)
VALUES 
    ('Первый', 'prevyi', 'Описание первого товара', 10000),
    ('Второй', 'vtoroi', 'Описание второго товара', 20000),
    ('Третий', 'tretii', 'Описание третьего товара', 30000),
    ('Четвёртый', 'chetvyoryi', 'Описание четвёртого товара', 40000),
    ('Пятый', 'pyatyi', 'Описание пятого товара', 50000),
    ('Шестый', 'shestyi', 'Описание шестого товара', 60000),
    ('Седьмый', 'sedmyi', 'Описание седьмого товара', 70000),
    ('Восьмый', 'vosemyi', 'Описание восьмого товара', 80000),
    ('Девятый', 'devyatyi', 'Описание девятого товара', 90000),
    ('Десятый', 'desyatyi', 'Описание десятого товара', 100000);
