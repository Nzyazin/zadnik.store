# 🛍 Zadnik.Store

Современная система управления интернет-магазином с административной панелью.

## 🚀 Технологии

### Бэкенд
- Go
- gRPC
- PostgreSQL
- Docker

### Фронтенд
- Pug
- SASS
- JavaScript
- Gulp

## 📦 Структура проекта

```
zadnik.store/
├── api/                    # gRPC API определения и сгенерированный код
├── cmd/                    # Точки входа приложения
├── internal/               # Внутренняя логика приложения
├── static/                 # Собранная статика
│   ├── admin/             # Статика админ-панели
│   │   ├── css/          # Стили
│   │   ├── js/           # Скрипты
│   │   ├── images/       # Изображения
│   │   ├── fonts/        # Шрифты
│   │   └── templates/    # HTML шаблоны
│   └── client/           # Статика клиентской части
│       ├── css/
│       ├── js/
│       ├── images/
│       ├── fonts/
│       └── templates/
├── public/                # Публичная статика
│   ├── admin/             # Статика админ-панели
│   │   ├── css/          # Стили
│   │   ├── js/           # Скрипты
│   │   ├── images/       # Изображения
│   │   ├── fonts/        # Шрифты
│   │   └── templates/    # HTML шаблоны
│   └── client/           # Статика клиентской части
│       ├── css/
│       ├── js/
│       ├── images/
│       ├── fonts/
│       └── templates/
└── web/                   # Исходники фронтенда
    ├── html-css-js-admin/ # Исходники админ-панели
    │   ├── assets/       # Исходные файлы
    │   │   ├── fonts/
    │   │   ├── images/
    │   │   ├── scripts/
    │   │   ├── styles/
    │   │   └── views/
    │   └── tasks/        # Gulp задачи
    └── html-css-js-client/ # Исходники клиентской части
        ├── assets/
        │   ├── fonts/
        │   ├── images/
        │   ├── scripts/
        │   ├── styles/
        │   └── views/
        └── tasks/
```

## 🛠 Установка и запуск

### Требования
- Go 1.21+
- Node.js 18+
- Docker
- Make

### Бэкенд
```bash
# Запуск базы данных
make up-local

# Запуск сервера
make run
```

### Сборка фронтенда
```bash
# Сборка админ-панели
cd web/html-css-js-admin
npm install
npm run build  # Соберет в static/admin/

# Сборка клиентской части
cd web/html-css-js-client
npm install
npm run build  # Соберет в static/client/
```

## 🔑 Функционал

- 👤 Авторизация администраторов
- 📦 Управление товарами
  - Просмотр списка товаров
  - Добавление новых товаров
  - Редактирование существующих товаров
- 🎨 Современный адаптивный дизайн
- 🚀 Быстрая и отзывчивая админ-панель

## 📝 Лицензия

MIT License © 2024 Zadnik.Store