package storage

import (
	"github.com/brocaar/lorawan"
	"github.com/jmoiron/sqlx"
)

// SearchResult defines a search result.
type SearchResult struct {
	Kind             string         `db:"kind"`
	Score            float64        `db:"score"`
	OrganizationID   *int64         `db:"organization_id"`
	OrganizationName *string        `db:"organization_name"`
	ApplicationID    *int64         `db:"application_id"`
	ApplicationName  *string        `db:"application_name"`
	DeviceDevEUI     *lorawan.EUI64 `db:"device_dev_eui"`
	DeviceName       *string        `db:"device_name"`
	GatewayMAC       *lorawan.EUI64 `db:"gateway_mac"`
	GatewayName      *string        `db:"gateway_name"`
}

// GlobalSearch performs a search on organizations, applications, gateways
// and devices.
func GlobalSearch(db sqlx.Queryer, username string, globalAdmin bool, search string, limit, offset int) ([]SearchResult, error) {
	var result []SearchResult
	query := "%" + search + "%"

	err := sqlx.Select(db, &result, `
		select
			'device' as kind,
			greatest(similarity(d.name, $1), similarity(encode(d.dev_eui, 'hex'), $1)) as score,
			o.id as organization_id,
			o.name as organization_name,
			a.id as application_id,
			a.name as application_name,
			d.dev_eui as device_dev_eui,
			d.name as device_name,
			null as gateway_mac,
			null as gateway_name
		from device d
		inner join application a
			on a.id = d.application_id
		inner join organization o
			on o.id = a.organization_id
		left join organization_user ou
			on ou.organization_id = o.id
		left join "user" u
			on u.id = ou.user_id
		where
			($3 = true or u.username = $4)
			and (d.name ilike $2 or encode(d.dev_eui, 'hex') ilike $2)
		union
		select
			'gateway' as kind,
			greatest(similarity(g.name, $1), similarity(encode(g.mac, 'hex'), $1)) as score,
			o.id as organization_id,
			o.name as organization_name,
			null as application_id,
			null as application_name,
			null as device_dev_eui,
			null as device_name,
			g.mac as gateway_mac,
			g.name as gateway_name
		from
			gateway g
		inner join organization o
			on o.id = g.organization_id
		left join organization_user ou
			on ou.organization_id = o.id
		left join "user" u
			on u.id = ou.user_id
		where
			($3 = true or u.username = $4)
			and (g.name ilike $2 or encode(g.mac, 'hex') ilike $2)
		union
		select
			'organization' as kind,
			similarity(o.name, $1) as score,
			o.id as organization_id,
			o.name as organization_name,
			null as application_id,
			null as application_name,
			null as device_dev_eui,
			null as device_name,
			null as gateway_mac,
			null as gateway_name
		from
			organization o
		left join organization_user ou
			on ou.organization_id = o.id
		left join "user" u
			on u.id = ou.user_id
		where
			($3 = true or u.username = $4)
			and o.name ilike $2
		union
		select
			'application' as kind,
			similarity(a.name, $1) as score,
			o.id as organization_id,
			o.name as organization_name,
			a.id as application_id,
			a.name as application_name,
			null as device_dev_eui,
			null as device_name,
			null as gateway_mac,
			null as gateway_name
		from
			application a
		inner join organization o
			on o.id = a.organization_id
		left join organization_user ou
			on ou.organization_id = o.id
		left join "user" u
			on u.id = ou.user_id
		where
			($3 = true or u.username = $4)
			and a.name ilike $2
		order by
			score desc
		limit $5
		offset $6`,
		search,
		query,
		globalAdmin,
		username,
		limit,
		offset,
	)
	if err != nil {
		return nil, handlePSQLError(Select, err, "select error")
	}

	return result, nil
}
