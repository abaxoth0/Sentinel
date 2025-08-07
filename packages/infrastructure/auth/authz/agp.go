package authz

import rbac "github.com/abaxoth0/SentinelRBAC"

func initAGP() {
	log.Info("Initializing Action Gate Policy...", nil)

	agp := rbac.NewActionGatePolicy()

	requireAdminRoleToDropCache := rbac.NewActionGateRule(
		&userDropCacheContext,
		rbac.RequireActionGateEffect,
		// Only role name matters
		[]rbac.Role{rbac.NewRole("admin", 0)},
	)

	if err := agp.AddRule(requireAdminRoleToDropCache); err != nil {
		log.Panic("Failed to initialize Action Gate Policy", err.Error(), nil)
	}

	Schema.ActionGatePolicy = agp

	log.Info("Initializing Action Gate Policy: OK", nil)
}

