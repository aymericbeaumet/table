# coffee

## Commands

### Development

```bash
cd www && npm i && npm run dev -- --open
go run -race .
```

### Update

```bash
cd www && npm i -D @sveltejs/kit@next @sveltejs/adapter-static@next && npx ncu -u
go get -u && go mod tidy
```

### Deploy

Automatically deployed when pushed to master.

- Website: https://chill-room-production.netlify.app
- Netlify: https://app.netlify.com/sites/chill-room-production
- Heroku: https://dashboard.heroku.com/apps/chill-room
