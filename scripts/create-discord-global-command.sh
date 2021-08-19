#!/bin/sh

# https://discord.com/developers/docs/interactions/application-commands#making-a-global-command
# https://discord.com/developers/docs/interactions/application-commands#application-command-object

curl -vv -X POST -d @- -H 'Content-Type: application/json' -H "Authorization: Bot $BOT_TOKEN" "https://discord.com/api/v8/applications/$APPLICATION_ID/commands" <<EOF
{
    "type": 1,
    "application_id": "$APPLICATION_ID",
    "name": "table",
    "description": "Join a table"
}
EOF
