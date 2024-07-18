# Technical info

-   Programming language: Go (1.22.3)

-   Database: MongoDB

-   Cache: Redis

-   HTTP libraries: mux (`github.com/gorilla/mux`); weaver (`github.com/StepanAnanin/weaver`)

-   Hash fucntion used for passwords: bcrypt (ed25519)

-   Authentication type: toke-based (JWT)

-   Supported OS: Linux

# API

## IMPORTANT

    Response will have status 401 (Unauthorized) if access token expired.

    Response will have status 409 (Conflict) if refresh token expired, also in this case authentication cookie will be deleted.
    Refresh token used only in "/refresh" endpoint. (And mustn't be used in anywhere else)

## Endpoints

-   /login [ POST ] — Used to login. Request body must contain: login (string), password (string)

-   /logout [ DELETE ] — Used to logout. User must be authenticated.

-   /refresh [ PUT ] — Used to refresh access and refresh tokens. User must be authenticated. **IMPORTANT: Response will have status 409 (conflict) if refresh token expired**.

-   /verify [ GET ] — Used to verify user authentication, also can be used for authorization. User must be authenticated. Returns user's ID, login and role if access token is valid.

-   /user/create [ POST ] — Used to create a new user. Request body must contain: login (unique string), password (string).

-   /user/delete [ DELETE ] — Used to soft delete user with passed uid. User must be authenticated and if he want to soft delete any other user then himself he must be a moderator or administrator to do that. Request body must contain: uid (unique string). **IMPORTANT: Users with admin role cannot be deleted**.

-   /user/restore [ PUT ] — Used to restore soft deleted user with passed uid. User must be authenticated. Request body must contain: uid (unique string).

-   /user/drop [ DELETE ] — Used to hard delete user with passed uid. User must be authenticated and if he want to hard delete any other user then himself he must be a moderator or administrator to do that. Request body must contain: uid (unique string). **IMPORTANT: Users with admin role cannot be deleted**.

-   /user/drop/all-soft-deleted [ DELETE ] — Used to hard delete all soft deleted users. User must be authenticated and must be administrator to do that.

-   /user/change/login [ PATCH ] — Used to change login of user with passed uid. User must be authenticated and if he want to change login of any other user then himself he must be a moderator or administrator to do that. Request body must contain: uid (unique string), login (unique string). **IMPORTANT: Users with admin role cannot be modified by any other users than themselves**.

-   /user/change/password [ PATCH ] — Used to change password of user with passed uid. User must be authenticated and if he want to change password of any other user then himself he must be a moderator or administrator to do that. Request body must contain: uid (unique string), password (string). **IMPORTANT: Users with admin role cannot be modified by any other users than themselves**.

-   /user/change/role [ PATCH ] — Used to change role of user with passed uid. User must be authenticated and if he want to change role of any other user then himself he must be a moderator or administrator to do that. Request body must contain: uid (unique string), role (string). **IMPORTANT: Users with admin role cannot be modified by any other users than themselves**.

-   /user/check/login [ POST ] — Used to check is login free to use. Request body must contain: login (string)

-   /user/check/role [ POST ] — Used to get role of user with passed uid. User must be authenticated and be a moderator or administrator to do that. Request body must contain: uid (unique string).

-   /roles [ GET ] — Used to get list of all existing roles.

-   /cache/drop [ DELETE ] — Used to delete all cache. User must be authenticated and be an administrator to do that.
