# TECH STACK

-   Programming language: Go (1.22.3)

-   Database: PostgreSQL

-   Cache: Redis

-   Transport: HTTP (lib: echo `github.com/labstack/echo`)

-   API architecture: REST

-   Data presentation format: JSON

-   Authorization method: RBAC - Role-Based Access Control (lib: SentinelRBAC `github.com/abaxoth0/SentinelRBAC`)

-   Password hashing method: bcrypt (ed25519)

-   Authentication: toke-based (JWT)

-   Supported OS: Linux

-   Architecture: Onion Architecture

# License

This project is licensed under the GNU Affero General Public License (AGPL).
Please see the [LICENSE](LICENSE) file for the full license text.

### Additional Terms
The [NOTICE](NOTICE) file is considered part of the license. By using, modifying,
or distributing this software, you agree to comply with the terms outlined in the NOTICE file.

# API

You can find OpenAPI spec in [docs](docs). If you want to see visualized version - install this app and then access /docs/index.html.

>[!IMPORTANT]
>Endpoints matching /docs/* requires authentication and you also must have read permission for docs resource to access them.
>
>To avoid this, you can enable debug mode, then anyone will be able access these endpoints.

