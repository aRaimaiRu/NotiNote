# NotiNoteApp Bruno API Collection

Complete API testing collection for NotiNoteApp OAuth social login (frontend-initiated flow).

## 🚀 Quick Start

### 1. Open Collection in Bruno

1. Open Bruno application
2. Click "Open Collection"
3. Navigate to `g:\NotiNoteApp\bruno\NotiNote`
4. Collection should load automatically

### 2. Select Environment

1. In Bruno, select the **Local** environment from the dropdown
2. This sets `base_url` to `http://localhost:8080/api/v1`

### 3. Start Backend Server

```bash
# From NotiNoteApp directory
go run cmd/server/main.go
```

You should see:
```
Server listening on :8080
Google OAuth provider registered
Facebook OAuth provider registered
```

---

## 📋 Available Endpoints

### Frontend-Initiated OAuth (Social Login)

#### Google Verify Token
**Endpoint:** `POST /auth/google/verify`

**Purpose:** Verify Google ID token from frontend SDK and authenticate user

**How to test:**

1. **Option A: Use Google OAuth Playground**
   - Go to https://developers.google.com/oauthplayground/
   - Select "Google OAuth2 API v2" → userinfo.email & userinfo.profile
   - Click "Authorize APIs"
   - Sign in with your Google account
   - Exchange authorization code for tokens
   - Copy the ID token
   - Paste in Bruno request body

2. **Option B: Implement in test frontend**
   ```jsx
   import { GoogleLogin } from '@react-oauth/google';

   <GoogleLogin
     onSuccess={(response) => {
       console.log('ID Token:', response.credential);
       // Copy this token to Bruno
     }}
   />
   ```

3. **Paste token in Bruno**
   - Open `Authen` → `Google Verify Token`
   - Replace `"paste_your_google_id_token_here"` with actual token
   - Click "Send"

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "email": "your-email@gmail.com",
      "name": "Your Name",
      "provider": "google",
      "avatar_url": "https://lh3.googleusercontent.com/...",
      "created_at": "2026-01-02T14:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

#### Facebook Verify Token
**Endpoint:** `POST /auth/facebook/verify`

**Purpose:** Verify Facebook access token from frontend SDK and authenticate user

**How to test:**

1. **Option A: Use Facebook Graph API Explorer**
   - Go to https://developers.facebook.com/tools/explorer/
   - Select your app from dropdown
   - Click "Generate Access Token"
   - Grant permissions: email, public_profile
   - Copy the **User Access Token**
   - Paste in Bruno request body

2. **Option B: Implement in test frontend**
   ```jsx
   import FacebookLogin from 'react-facebook-login';

   <FacebookLogin
     appId="YOUR_APP_ID"
     callback={(response) => {
       console.log('Access Token:', response.accessToken);
       // Copy this token to Bruno
     }}
   />
   ```

3. **Paste token in Bruno**
   - Open `Authen` → `Facebook Verify Token`
   - Replace `"paste_your_facebook_access_token_here"` with actual token
   - Click "Send"

**Expected Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 2,
      "email": "your-email@facebook.com",
      "name": "Your Name",
      "provider": "facebook",
      "avatar_url": "https://graph.facebook.com/.../picture",
      "created_at": "2026-01-02T14:00:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 86400
  }
}
```

---

## 🔧 Environment Variables

### Local Environment

The **Local** environment is pre-configured with:

```
base_url = http://localhost:8080/api/v1
```

### Create Additional Environments

You can create additional environments for staging/production:

1. In Bruno, go to Environments
2. Click "+ New Environment"
3. Name it (e.g., "Production")
4. Add variables:
   ```
   base_url = https://api.yourapp.com/api/v1
   ```

---

## 📝 Testing Workflow

### First Time Setup

1. ✅ Configure OAuth apps (Google & Facebook)
2. ✅ Set environment variables in `.env`
3. ✅ Start PostgreSQL database
4. ✅ Run backend server
5. ✅ Open Bruno collection
6. ✅ Select "Local" environment

### Testing Google Login

1. Get test token from [OAuth Playground](https://developers.google.com/oauthplayground/)
2. Open `Authen` → `Google Verify Token`
3. Paste token in request body
4. Send request
5. Should receive JWT tokens ✅

### Testing Facebook Login

1. Get test token from [Graph API Explorer](https://developers.facebook.com/tools/explorer/)
2. Open `Authen` → `Facebook Verify Token`
3. Paste token in request body
4. Send request
5. Should receive JWT tokens ✅

### Using JWT Tokens

After successful login, you receive `access_token`. Use it for protected endpoints:

```
Authorization: Bearer <your_access_token>
```

---

## ❌ Common Errors

### "google OAuth provider not registered"

**Cause:** Backend not configured with Google credentials

**Solution:**
```bash
# Add to .env
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
```
Restart server.

### "token audience mismatch"

**Cause:** ID token from different Google Client ID

**Solution:** Ensure frontend and backend use the same `GOOGLE_CLIENT_ID`

### "email not verified"

**Cause:** Google account email not verified

**Solution:** Verify email in Google account settings or use a different account

### "token is not valid"

**Cause:** Token expired or invalid

**Solution:** Get a fresh token (they expire quickly!)

### "email not provided by Facebook"

**Cause:** User didn't grant email permission

**Solution:** Request email scope when generating token in Graph API Explorer

---

## 🎯 Tips

### Token Expiration

- Google ID tokens expire in ~1 hour
- Facebook access tokens expire in ~1-2 hours (depending on settings)
- Get fresh tokens if requests fail

### Testing Multiple Accounts

You can test with multiple Google/Facebook accounts:
1. Get token from Account A
2. Test in Bruno → Creates User 1
3. Get token from Account B
4. Test in Bruno → Creates User 2

### Inspect Responses

Bruno has great response visualization:
- **Preview** tab for formatted JSON
- **Headers** tab for response headers
- **Timeline** tab for performance metrics

---

## 📚 Additional Resources

**Backend Documentation:**
- [OAUTH_IMPLEMENTATION.md](../../OAUTH_IMPLEMENTATION.md) - Main OAuth guide
- [FRONTEND_OAUTH_INTEGRATION.md](../../FRONTEND_OAUTH_INTEGRATION.md) - Frontend integration
- [QUICK_START_SOCIAL_LOGIN.md](../../QUICK_START_SOCIAL_LOGIN.md) - Quick start guide

**OAuth Providers:**
- [Google OAuth Playground](https://developers.google.com/oauthplayground/)
- [Facebook Graph API Explorer](https://developers.facebook.com/tools/explorer/)
- [Google Cloud Console](https://console.cloud.google.com/)
- [Facebook Developers](https://developers.facebook.com/)

---

## ✅ Testing Checklist

Before considering OAuth implementation complete:

- [ ] Google token verification works in Bruno
- [ ] Facebook token verification works in Bruno
- [ ] User created in database on first login
- [ ] User info updated on subsequent logins
- [ ] JWT tokens generated correctly
- [ ] Access token can be used for protected endpoints
- [ ] Error messages are clear and helpful
- [ ] Multiple Google accounts can register
- [ ] Multiple Facebook accounts can register
- [ ] Email conflicts handled correctly (same email, different provider)

---

**Last Updated:** January 2, 2026
**Collection Version:** 1.0
**Backend API Version:** v1
