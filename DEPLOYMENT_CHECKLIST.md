# SyncGuard MFA — Deployment Checklist

## Prerequisites

All the following GitHub Secrets must be set at:
https://github.com/umairakhlaque/activesyncking/settings/secrets/actions

### Backend Secrets
- [ ] `FLY_API_TOKEN` — Fly.io access token
- [ ] `NEON_DATABASE_URL` — PostgreSQL connection string (rediss:// or postgres://)
- [ ] `UPSTASH_REDIS_URL` — Redis connection string (rediss:// for TLS)
- [ ] `SYNCGUARD_JWT_SECRET` — Min 32 chars (generate: `openssl rand -base64 48`)
- [ ] `SYNCGUARD_ENCRYPTION_KEY` — 32-byte hex string (generate: `openssl rand -hex 32`)
- [ ] `SYNCGUARD_ADMIN_API_KEY` — Admin portal password (generate: `openssl rand -base64 32`)

### Frontend Secrets
- [ ] `VERCEL_TOKEN` — Vercel personal access token
- [ ] `VERCEL_ORG_ID` — Vercel team/account ID (found in Settings → General)
- [ ] `VERCEL_PROJECT_ID` — Vercel project ID (found in project Settings → General)

## Deployment Status

After pushing to `claude/enterprise-activesync-product-Sto2S`:
1. Check GitHub Actions → Deploy workflow
2. Verify all services pass health checks:
   - `https://syncguard-adminsvc.fly.dev/healthz` → 200 + `{"status":"ok"}`
   - `https://syncguard-authsvc.fly.dev/healthz` → 200 + `{"status":"ok"}`
   - `https://syncguard-gateway.fly.dev/healthz` → 200 + `{"status":"ok"}`
3. Admin portal should be live at `https://activesyncking.vercel.app/login`

## Testing the Login

1. Navigate to `https://activesyncking.vercel.app/login`
2. Enter the value of `SYNCGUARD_ADMIN_API_KEY` as the password
3. Click "Sign in"
4. On success, you should be redirected to `/dashboard`

## Debugging

If anything fails:
- Check GitHub Actions → Deploy → job logs
- Check Fly.io → Apps → `syncguard-adminsvc` → Logs
- Check Vercel → Project → Deployments → Production → Logs
- Open browser DevTools → Network → try login → check response body
# CI force-redeploy marker: Mon May 11 13:57:18 UTC 2026
