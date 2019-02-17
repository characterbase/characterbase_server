package characters

/*
QueryFindByID represents a database query that returns
a single character's information via their ID

$1 — Character ID
*/
const QueryFindByID = `SELECT characters.id, characters.universe_id, characters.name,
characters.tag, characters.fields, characters.meta, characters.created_at, characters.updated_at, users.id AS
"owner.id", users.email AS "owner.email", users.display_name AS "owner.display_name" FROM characters JOIN users ON
characters.owner_id = users.id WHERE characters.id = $1`

/*
QueryFindByUniversePublicNom represents a database query that returns
a list of characters that are public or owned in nominal order

$1 — Universe ID
$2 — Requesting user ID
$3 — Search query
$4 — Page limit
$5 — Offset
*/
const QueryFindByUniversePublicNom = QuerySubFindByUniverseStart + " " + QuerySubOrderNom + " " +
	QuerySubFindByUniversePublicEnd

/*
QueryFindByUniverseAllNom represents a database query that returns
all characters in nominal order

$1 — Universe ID
$2 — Search query
$3 — Page limit
$4 — Offset
*/
const QueryFindByUniverseAllNom = QuerySubFindByUniverseStart + " WHERE universe_id = $1 AND name ILIKE $2 " +
	QuerySubOrderNom + " " + QuerySubFindByUniverseAllEnd

/*
QueryFindByUniversePublicLex represents a database query that returns
a list of characters that are public or owned in lexicographical order

$1 — Universe ID
$2 — Requesting user ID
$3 — Search query
$4 — Page limit
$5 — Offset
*/
const QueryFindByUniversePublicLex = QuerySubFindByUniverseStart + " " + QuerySubOrderLex + " " +
	QuerySubFindByUniverseAllEnd

/*
QueryFindByUniverseAllLex represents a database query that returns
a list of characters that are public or owned, and in lexicographical order

$1 — Universe ID
$2 — Search query
$3 — Page limit
$4 — Offset
*/
// const QueryFindByUniverseAllLex = QuerySubFindByUniverseStart + " " + QuerySubOrderLex + " " +
// QuerySubFindByUniversePublicEnd
const QueryFindByUniverseAllLex = `SELECT id, name, tag, owner_id, created_at, updated_at, character_images.url AS 
avatar_url, meta->'hidden' AS hidden FROM characters LEFT JOIN character_images ON character_images.character_id =
characters.id `

/*
QueryFindByUniversePublicCount represents a query that returns the
total number of characters that are public or owned

$1 — Universe ID
$2 — Search query
$3 — Requesting user ID
*/
const QueryFindByUniversePublicCount = `SELECT count(*) FROM characters WHERE universe_id = $1 AND name ILIKE $2 AND 
(meta->'hidden'='false' OR owner_id=$3)`

/*
QueryFindByUniverseAllCount represents a query that returns the
total number of characters pertaining to a universe

$1 — Universe ID
$2 — Search query
*/
const QueryFindByUniverseAllCount = `SELECT count(*) FROM characters WHERE universe_id = $1 AND name ILIKE $2`

/*
QuerySubFindByUniverseStart represents a sub-query that should be prepended to queries
thet returns lists of characters pertaining to a universe
*/
const QuerySubFindByUniverseStart = `SELECT id, name, tag, owner_id, created_at, updated_at, character_images.url AS 
avatar_url, meta->'hidden' AS hidden FROM characters LEFT JOIN character_images ON character_images.character_id =
characters.id`

/*
QuerySubOrderNom represents a sub-query that sorts characters in lexographic order,
respecting resutls with hidden names
*/
const QuerySubOrderNom = `ORDER BY meta->'nameHidden'='true', name`

/*
QuerySubOrderLex represents a sub-query that sorts characters in nominal order,
respecting results with hidden names, preferred names, or empty results
*/
const QuerySubOrderLex = `ORDER BY meta->'name'->'lastName' = '' OR meta->'name'->'firstName' = '' OR meta->
'nameHidden'='true', CASE WHEN meta->'name'->'preferredName' != '' THEN meta->'name'->'preferredName' ELSE
meta->'name'->'lastName' END, meta->'name'->'lastName', meta->'name'->'firstName'`

/*
QuerySubFindByUniverseAllEnd represents a sub-query that should be appended to queries
that return lists of characters pertaining to a universe
*/
const QuerySubFindByUniverseAllEnd = `LIMIT $3 OFFSET $4`

/*
QuerySubFindByUniversePublicEnd represents a sub-query that should be appended to queries
that return lists of characters pertaining to a universe
*/
const QuerySubFindByUniversePublicEnd = `WHERE universe_id = $1 AND (meta->'hidden'='false' OR owner_id=$2) AND name
ILIKE $3 LIMIT $4 OFFSET $5`
