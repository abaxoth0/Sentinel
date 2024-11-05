## Authozation

Sentinel uses RBAC (Role-Based Access Control) system for authorization (defining what user can do).
All roles and services are defined into **RBAC.json** file, which must be at the same directory with Sentinel app.

### ROLES

Roles can be specified globaly, also services can have their own roles, with their own permissions. Service specific roles will overwrite global roles.

### PERMISSIONS

Permissions are specified for CRUD operations. There are 9 possible permissions:

-   C (Create) - can any entities

-   SC (Self Create) - can create entities, which will belong to this user

-   R (Read) - can read any entity

-   SR (Self Read) - can read entities, that was created by this user (also can read himself)

-   U (Update) - can modify any entity

-   SU (Self Update) - can modify entities, that was created by this user (also can update himself)

-   D (Delete) - can delete any entity

-   SD (Delete) - can delete entities, that was created by this user

-   M (Moderator) - can do moderator-specific actions. Also prevent some actions to be performed on user with this role (for example moderator can't delete another moderator, this can do only an admin)

-   A (Admin) - pretty same as M, but for administrators

Like roles, permissions can be specified globaly or individually for each service. And like roles, individual permissions will overwrite global permissions.
