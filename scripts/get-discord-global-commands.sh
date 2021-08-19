#!/bin/sh

curl -vv -H 'Content-Type: application/json' -H "Authorization: Bot $BOT_TOKEN" "https://discord.com/api/v8/applications/$APPLICATION_ID/commands" | jq
