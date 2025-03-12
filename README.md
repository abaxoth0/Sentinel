# TECH STACK

-   Programming language: Go (1.22.3)

-   Database: PostgreSQL

-   Cache: Redis

-   Transport: HTTP (lib: echo `github.com/labstack/echo`)

-   API architecture: REST

-   Data presentation format: JSON

-   Authorization method: RBAC - Role-Based Access Control (lib: SentinelRBAC `github.com/StepanAnanin/SentinelRBAC`)

-   Password hashing method: bcrypt (ed25519)

-   Authentication: toke-based (JWT)

-   Supported OS: Linux

-   Architecture: Onion Architecture

# API

>[!IMPORTANT]
>Response will have status 401 (Unauthorized) if access token expired.
>
>Response will have status 409 (Conflict) if refresh token expired, also in this case authentication cookie will be deleted.

>[!NOTE]
> All things marked via :red_circle: are required.

## /auth
Handles user authentication. Methods:
- GET:
    - Action: get user id, login and roles
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - `NONE`
    - Response body:
      - id (string) - user id
      - login (string) - user login
      - roles (string[]) - user roles
- POST:
    - Action: generate access and refresh token, then set refresh token to refreshToken cookie
    - Required request headers:
      - `NONE`
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: login (string) - user login
      - :red_circle: password (string) - user password
    - Response body:
      - message (string) - response message
      - accessToken (string) - access token
- PUT:
    - Action: generate new authentication tokens
    - Required request headers:
      - `NONE`
    - Required request cookies:
      - :red_circle: refreshToken (string) - refresh token
    - Required request body:
      - `NONE`
    - Response body:
      - message (string) - response message
      - accessToken (string) - access token
- DELETE:
    - Action: terminate authentication by removing refreshToken cookie
    - Required request headers:
      - `NONE`
    - Required request cookies:
      - :red_circle: refreshToken (string) - refresh token
    - Required request body:
      - `NONE`
    - Response body:
      - message (string) - response message
      - accessToken (string) - access token

## /user
Handles user creation and soft deletion. Methods:
- POST:
    - Action: create new user
    - Required request headers:
      - `NONE`
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: login (string) - user login
      - :red_circle: password (string) - user password
    - Response body:
      - `NONE`
- DELETE:
    - Action: soft delete user
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: uid (string) - user id
    - Response body:
      - `NONE`
## /user/drop
Handles user hard deletion. Methods:
- DELETE:
    - Action: hard delete user
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: uid (string) - user id
    - Response body:
      - `NONE`

## /user/restore
Handles the recovery of a soft deleted user. Methods:
- POST:
    - Action: restore soft deleted user
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: uid (string) - user id
    - Response body:
      - `NONE`

## /user/password
Handles user password change. Methods:
- PATCH:
    - Action: change user password
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: uid (string) - user id
      - :red_circle: password (string) - new user password
    - Response body:
      - `NONE`

## /user/login
Handles user login change. Methods:
- PATCH:
    - Action: change user login
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: uid (string) - user id
      - :red_circle: login (string) - new user login
    - Response body:
      - `NONE`

## /user/roles
Handles user roles change. Methods:
- PATCH:
    - Action: change user roles
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: uid (string) - user id
      - :red_circle: roles (string[]) - array of new user roles
    - Response body:
      - `NONE`

## /user/login/check
Handles checking existance of user login. Methods:
- POST:
    - Action: check if login already in use 
    - Required request headers:
      - `NONE`
    - Required request cookies:
      - `NONE`
    - Request body:
      - :red_circle: login (string) - user login
    - Response body:
      - exists (bool) - `true` if login is already in use, `false` otherwise

## /roles/:serviceID
Handles getting list of all roles in service with :red_circle: `serviceID`. Methods: 
- GET:
    - Action: get list of all user roles in specified service
    - Required request headers:
      - `NONE`
    - Required request cookies:
      - `NONE`
    - Request body:
      - `NONE`
    - Response body:
      - roles (string[]) - array with all roles in specified server

## /cache
Handles cache control. Methods:
- DELETE:
    - Action: clear all cache
    - Required request headers:
      - :red_circle: Authorization (string) - access token in Bearer Token format
    - Required request cookies:
      - `NONE`
    - Request body:
      - `NONE`
    - Response body:
      - `NONE`
