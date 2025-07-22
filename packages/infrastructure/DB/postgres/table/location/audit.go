package locationtable

import (
	Error "sentinel/packages/common/errors"
	"sentinel/packages/common/util"
	LocationDTO "sentinel/packages/core/location/DTO"
	"sentinel/packages/infrastructure/DB/postgres/audit"
	"sentinel/packages/infrastructure/DB/postgres/connection"
	"sentinel/packages/infrastructure/DB/postgres/query"
	"sentinel/packages/infrastructure/DB/postgres/transaction"
)

func newAuditDTO(op audit.Operation, location *LocationDTO.Full) LocationDTO.Audit {
    return LocationDTO.Audit{
		ChangedLocationID: location.ID,
		Operation: string(op),
		SessionID: location.SessionID,
		IP: location.IP,
		Country: location.Country,
		Region: location.Region,
		City: location.City,
		Latitude: location.Latitude,
		Longitude: location.Longitude,
		ISP: location.ISP,
		DeletedAt: location.DeletedAt,
		CreatedAt: location.CreatedAt,
    }
}

func newAuditQuery(dto *LocationDTO.Audit) *query.Query {
    var deletedAt = util.Ternary(dto.IsDeleted(), &dto.DeletedAt, nil)

    return query.New(
        `INSERT INTO "audit_location"
        (changed_location_id, session_id, operation, ip, country, region, city, latitude, longitude, isp, deleted_at, created_at)
        VALUES
        ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
        dto.ChangedLocationID,
		dto.SessionID,
        dto.Operation,
		dto.IP,
        dto.Country,
        dto.Region,
        dto.City,
        dto.Latitude,
		dto.Longitude,
		dto.ISP,
		deletedAt,
		dto.CreatedAt,
    )
}

func execTxWithAudit(dto *LocationDTO.Audit, queries ...*query.Query) *Error.Status {
    queries = append(queries, newAuditQuery(dto))

    return transaction.New(queries...).Exec(connection.Primary)
}

