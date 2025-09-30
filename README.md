# Telegram Connector

This plugin provides Telegram authentication for Apache Answer

## Environment Variables

The connector uses the following environment variables:

- `TELEGRAM_BOT_TOKEN` - Your Telegram Bot token (required)
- `TELEGRAM_BOT_USERNAME` - Your Telegram Bot username WITHOUT @ symbol (required)
- `TELEGRAM_REDIRECT_PATH` - Custom redirect path (optional)

## Build

```bash
# Build for AMD64 architecture
docker build --platform linux/amd64 -t ghcr.io/neermilov/apache-answer-telegram-connector:1.6.0 .
```