# OAuth 2.0 Authorization Code Flow with PKCE

OAuth 2.0 Authorization Code Grant with Proof Key for Code Exchange (PKCE) as defined in RFC 7636

## Roles

1. **Resource Owner**: End user who owns the protected resources
2. **User Agent**: User's web browser or native app
3. **Public Client**: Public client (SPA or native app) that cannot securely store secrets
4. **Authorization Server**: Server that authenticates the user and issues tokens
5. **Resource Server**: Server hosting the protected resources

## Scenarios

### 1. OAuth 2.0 Authorization Code Flow with PKCE

Participants: Resource Owner, User Agent, Public Client, Authorization Server, Resource Server

1. Resource Owner → Public Client: Click Login
2. Public Client → Public Client: Generate code_verifier + code_challenge
3. Public Client → User Agent: Redirect with code_challenge
4. User Agent → Authorization Server: GET /authorize?code_challenge=...&code_challenge_method=S256
5. Authorization Server → Resource Owner: Login + Consent Form
6. Resource Owner → Authorization Server: Submit Credentials + Grant Consent
7. Authorization Server → User Agent: Redirect with code (code_challenge stored server-side)
8. User Agent → Public Client: Callback with Authorization Code
9. Public Client → Authorization Server: POST /token (code + code_verifier)
10. Authorization Server ⇢ Public Client: Access Token + Refresh Token
11. Public Client → Resource Server: API Request (Authorization: Bearer <token>)
12. Resource Server ⇢ Public Client: Protected Resource
