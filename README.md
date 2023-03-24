# go-chatgpt-telegram-bot

## Usage

```sh
export TELEGRAM_BOT_TOKEN=
export OPENAI_API_KEY=

# whitelist
export BOT_WHITELIST=

# use redis to store data
export STORE_TYPE=redis
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0

go run ./cmd
```

![](https://i.imgur.com/PKEobN7.png)
