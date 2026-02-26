# Contributing to Svyaz

*[Русская версия ниже](#contributing-to-svyaz-ru)*

Thank you for your interest in contributing to Svyaz! This document explains how to get involved.

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js (for CSS build)
- SQLite

### Local Setup

1. Fork and clone the repository:

```bash
git clone https://github.com/<your-username>/svyaz.git
cd svyaz
```

2. Copy the environment file and fill in the required values:

```bash
cp .env.example .env
```

Required variables:
- `BOT_TOKEN` — Telegram bot token (get one from [@BotFather](https://t.me/BotFather))
- `CSRF_SECRET` — any random string for CSRF token signing

For local development without Telegram auth, set `DEV_LOGIN=1` — this enables a user picker at `/auth/dev`.

3. Install dependencies and run:

```bash
make setup   # install npm deps, copy HiQ CSS
make run     # start dev server at localhost:3000
```

## How to Contribute

### Reporting Bugs

Open a [GitHub Issue](https://github.com/azdaev/svyaz/issues) with:
- Steps to reproduce
- Expected vs actual behavior
- Browser / OS if relevant

### Suggesting Features

Open a [GitHub Issue](https://github.com/azdaev/svyaz/issues) with the `enhancement` label. Describe the problem you're solving and your proposed approach.

### Submitting Code

We use **GitHub Flow**:

1. Create a branch from `main`:
   ```bash
   git checkout -b my-feature
   ```
2. Make your changes
3. Test locally with `make run`
4. Commit with a clear message describing *what* and *why*
5. Push and open a Pull Request against `main`

### What Makes a Good PR

- One logical change per PR
- Clear title and description
- Screenshots for UI changes
- Migrations in `migrations/` if you change the schema (we use goose)

## Project Structure

See [CLAUDE.md](./CLAUDE.md) for a detailed architecture overview. Key directories:

| Directory | What's there |
|---|---|
| `cmd/server/` | Entry point |
| `internal/handler/` | HTTP handlers, routes, templates |
| `internal/repo/` | Database layer (raw SQL) |
| `internal/models/` | Domain structs |
| `migrations/` | SQL migrations (goose) |
| `templates/` | Go html/template files |
| `static/` | CSS, JS, images |

## Conventions

- **UI language is Russian.** All user-facing text should be in Russian.
- Commit messages — free format, in English. Be clear and concise.
- Go code follows standard `gofmt` formatting.
- No frontend framework — vanilla JS only.
- CSS: HiQ as base framework, custom styles in `static/css/style.css`.

## Code of Conduct

We expect all contributors to be respectful and constructive. Harassment, discrimination, and toxic behavior will not be tolerated. Maintainers reserve the right to remove comments, commits, or contributors that violate these principles.

If you experience or witness unacceptable behavior, please report it via [GitHub Issues](https://github.com/azdaev/svyaz/issues).

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](./LICENSE).

## Questions?

Open a [GitHub Issue](https://github.com/azdaev/svyaz/issues) — we'll do our best to respond promptly.

---

<a id="contributing-to-svyaz-ru"></a>

# Участие в разработке Svyaz

Спасибо за интерес к проекту! Ниже описано, как начать.

## Начало работы

### Что нужно

- Go 1.25+
- Node.js (для сборки CSS)
- SQLite

### Локальная настройка

1. Форкните и склонируйте репозиторий:

```bash
git clone https://github.com/<your-username>/svyaz.git
cd svyaz
```

2. Скопируйте файл окружения и заполните обязательные поля:

```bash
cp .env.example .env
```

Обязательные переменные:
- `BOT_TOKEN` — токен Telegram-бота (получите у [@BotFather](https://t.me/BotFather))
- `CSRF_SECRET` — любая случайная строка для подписи CSRF-токенов

Для локальной разработки без Telegram-авторизации установите `DEV_LOGIN=1` — появится выбор пользователя на `/auth/dev`.

3. Установите зависимости и запустите:

```bash
make setup   # установить npm-зависимости, скопировать HiQ CSS
make run     # запустить dev-сервер на localhost:3000
```

## Как помочь

### Баги

Создайте [Issue на GitHub](https://github.com/azdaev/svyaz/issues):
- Шаги для воспроизведения
- Ожидаемое и фактическое поведение
- Браузер / ОС, если релевантно

### Идеи и фичи

Создайте [Issue на GitHub](https://github.com/azdaev/svyaz/issues) с меткой `enhancement`. Опишите проблему и предложите решение.

### Отправка кода

Мы используем **GitHub Flow**:

1. Создайте ветку от `main`:
   ```bash
   git checkout -b my-feature
   ```
2. Внесите изменения
3. Проверьте локально через `make run`
4. Закоммитьте с понятным сообщением
5. Запушьте и откройте Pull Request в `main`

### Хороший PR — это

- Одно логическое изменение на PR
- Понятное название и описание
- Скриншоты для изменений интерфейса
- Миграции в `migrations/`, если меняете схему БД

## Соглашения

- **Язык интерфейса — русский.** Весь пользовательский текст на русском.
- Коммиты — свободный формат, на английском.
- Go-код форматируется через `gofmt`.
- Без фронтенд-фреймворков — только vanilla JS.

## Кодекс поведения

Мы ожидаем от всех участников уважительного и конструктивного общения. Оскорбления, дискриминация и токсичное поведение недопустимы. Мейнтейнеры оставляют за собой право удалять комментарии, коммиты и участников, нарушающих эти принципы.

## Лицензия

Отправляя вклад, вы соглашаетесь с тем, что он будет лицензирован на условиях [MIT License](./LICENSE).

## Вопросы?

Создайте [Issue на GitHub](https://github.com/azdaev/svyaz/issues) — постараемся ответить оперативно.
