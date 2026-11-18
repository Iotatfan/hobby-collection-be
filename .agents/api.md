# HTTP API Guidance and Route Reference

[Agent entry point](../agents.md)

Route definitions in `internal/router/` are the source of truth. Paths below are relative to the API host, with no additional API prefix. The collection prefix is singular; preserve `/create_collection` unless a route change is explicitly requested.

## Authentication and responses

Public routes require no JWT. Protected routes require `Authorization: Bearer <token>`, validated by `internal/middleware/middleware.go`. Obtain a token from `POST /admin/token`.

CORS and rate limiting apply globally. Public routes are still rate limited. Read/write limits use separate buckets per client IP; responses expose `X-RateLimit-Limit` and `X-RateLimit-Remaining`. Rejected requests return HTTP 429 with `Retry-After`. CORS preflight requests return HTTP 204.

Collection and admin endpoints use this shared envelope:

```json
{"code":200,"message":"success","data":{},"isError":false}
```

Descriptions below refer to the contents of `data`. Errors use the same envelope with `isError: true`, an error message, and no `data`. Common statuses are 400 for recognized validation/conversion errors, 401 for authentication failures, 404 for missing records/routes, and 500 for database or unclassified errors. Not every malformed request maps to 400; consult `internal/common/error_parser.go`.

Shortlink creation returns its success DTO directly, without this envelope. Shortlink resolution redirects instead of returning JSON. Middleware may reject either request before its handler runs.

## Route index

| Method | Route | Access | Capability |
| --- | --- | --- | --- |
| POST | `/admin/token` | Public; password required | Issue an admin JWT |
| GET | `/collection` | Public | Filter, sort, and page collection summaries |
| GET | `/collection/:id` | Public | Read one collection with media and metadata |
| GET | `/collection/filter` | Public | Read options for browsing filters |
| GET | `/collection/drawer` | Public | Read catalogs for create/edit forms |
| GET | `/collection/statistics` | Public | Read collection totals and status/release counts |
| GET | `/collection/shelves` | Public | Read shelf previews or page one category |
| POST | `/collection/types` | JWT | Create a collection type |
| POST | `/collection/grades` | JWT | Create a grade linked to a collection type |
| POST | `/collection/scales` | JWT | Create a scale |
| POST | `/collection/release-types` | JWT | Create a release type |
| POST | `/collection/series` | JWT | Create a series |
| POST | `/collection/manufacturers` | JWT | Create a manufacturer |
| POST | `/collection/metadata-tags` | JWT | Create a modification or feature tag |
| POST | `/create_collection` | JWT | Create a collection and upload images |
| PATCH | `/collection/:id` | JWT | Edit fields, images, addons, and tags |
| DELETE | `/collection/:id` | JWT | Delete a collection and initiate media cleanup |
| POST | `/shortlink` | Public | Create or reuse an expiring short code |
| GET | `/shortlink/:short_code` | Public | Redirect to a shortlink destination |
| GET, HEAD | `/docs/*filepath` | Public | Serve static documentation assets, when installed |

## Admin token

### POST `/admin/token`

Accepts JSON with required `password`, checked against the configured admin password. Success is HTTP 200 with `token` and `expires_at`. An incorrect password returns 401; configuration/signing failures return 500.

```http
POST /admin/token
Content-Type: application/json

{"password":"<configured-admin-password>"}
```

Use the returned token as `Authorization: Bearer <token>`. Tokens use HS256 and the configured secret, issuer, audience, and required role. `jwt.access` is the lifetime in minutes; nonpositive values use 24 hours. Configuration requires an admin password of at least 16 characters and a JWT secret of at least 32 characters. Never put real credentials in examples.

## Collection reads

All endpoints in this section return HTTP 200 on success.

### GET `/collection`

Returns `collections` (summary items) and `total_count` (the full matching count before pagination). An empty page currently returns `total_count: 0`, including when the offset exceeds the last matching item. Each summary includes `id`, `title`, `type`, `release_type`, `status`, `series`, `built_at`, `acquired_at`, and `cover`.

| Query parameter | Capability / default |
| --- | --- |
| `collection_type_id`, `grade_id`, `scale_id` | Filter by classification |
| `release_type_id` or `release_type_ids` | Filter by one or more release-type IDs; use repeated keys for multiple IDs |
| `manufacturer_id`, `series_id` | Filter by manufacturer or series |
| `status` | Filter by code: `0` wishlist, `1` backlog, `2` owned, `3` built |
| `sort` | `latest` / `latest_built`, `name` / `name_asc`, or `name_desc`; unknown/omitted values use latest |
| `limit` | Defaults to 20 when nonpositive/omitted; capped at 100 |
| `offset` | Defaults to 0; negative values normalize to 0 |

Latest sorting uses `built_at`, falling back to `acquired_at`, descending with nulls last and ID descending as a tie-breaker. Name sorting uses title and ID in the requested direction. Omitted filters do not restrict the list.

```http
GET /collection?release_type_id=2&release_type_id=3&sort=name_asc&limit=12&offset=0
```

If both release-type aliases are present, `release_type_id` takes precedence. Prefer repeated keys as shown: the current binder/parser does not reliably accept comma-separated IDs or bracketed query keys.

### GET `/collection/:id`

Accepts an integer ID and returns a full detail object: summary fields plus `manufacturer`, `pictures`, `addons`, `modifications`, `features`, `description`, and `display_size`. `type` contains collection type, scale, and grade; addons include manufacturer details. Dates may be null. Missing records return 404; invalid integer IDs return 400.

### GET `/collection/filter`

Optional `category=gunpla|figure|other-model-kit` returns `{category,total_count,grades,scales,release_types,manufacturers,series}` in the normal envelope. Options are distinct active catalog values used by non-backlog items in the complete shelf category, independent of pagination and active selections. Arrays use `{id,name}`, with `short_name` also present on grades. Gunpla exposes grades/releases/series; Figure and Other Model Kit expose scales/manufacturers/series. Other Model Kit groups all types except Gunpla and Figure. `total_count` is the unfiltered eligible category total. Invalid or empty categories return 400. Category metadata is cached separately and invalidated by collection and catalog mutations.

Without `category`, the existing Classic View contract below remains unchanged.

Returns browsing options in `collection_types`, `release_types`, `gunpla_grades`, and `figures_scales`. No query parameters are consumed. Figure scales come from existing figure collections and exclude the repository's unknown-scale ID, so this is not the complete scale catalog. Use these IDs in list filters.

### GET `/collection/drawer`

Returns create/edit form catalogs: `grades`, `scales`, `release_types`, `manufacturers`, `series`, `modifications`, and `features`. No query parameters are consumed. Grades contain `grade_id`, `collection_type_name`, and `grade_short_name`; tags contain `id`, `slug`, `name`, and `type`. Unlike `/filter`, this includes the complete scale catalog.

### GET `/collection/statistics`

Returns `total_count`, `completed_count` (owned plus built), `backlog_count`, and `limited_count`. Limited items currently mean release-type IDs 2, 3, or 4, rather than a dynamic name-based classification. Soft-deleted collections are excluded. No query parameters are consumed.

### GET `/collection/shelves`

Returns `gunpla_shelf`, `figure_shelf`, `other_model_kit_shelf`, and `backlog_shelf`. Each has `id`, `name`, `items`, and `count`; count is the full matching total, not the page length. Items include `id`, `title`, `type`, `status`, `cover`, and `display_size`.

Without query parameters, each shelf previews up to six items. Category shelves exclude backlog items; backlog groups backlog items across types. Ordering uses `built_at`, then `acquired_at`, then `created_at` as fallbacks, descending with ID descending as a tie-breaker.

To page one category, supply `category=gunpla`, `figure`, or `other-model-kit`. Limit defaults to 12 and must be 1–100; offset defaults to 0 and must be nonnegative. Limit/offset without a category, an empty/unsupported category, or invalid paging values returns 400. Backlog is not a paging category.

```http
GET /collection/shelves?category=gunpla&limit=12&offset=12
```

Category pages also accept `collection_type_id`, `grade_id`, `scale_id`, `release_type_id` / `release_type_ids`, `manufacturer_id`, and `series_id`, using the same filter semantics as `/collection`. Filters require a supported `category`; supplying a recognized filter without one returns 400. Different fields combine with AND; repeated release-type IDs match any supplied ID, and `release_type_id` takes precedence over its alias. Use repeated keys, not comma-separated or bracketed keys. Nonpositive scalar IDs do not restrict results; malformed numeric filters return 400.

Filters affect only the selected category's items and full matching `count`. Other categories and backlog keep unfiltered previews and counts. Classification filters intersect with the category, so incompatible combinations return an empty shelf. Reset `offset` when changing filters.

```http
GET /collection/shelves?category=gunpla&grade_id=2&release_type_id=2&release_type_id=3&manufacturer_id=1&limit=12&offset=0
```

All four shelves remain in the response: only the selected category uses the requested page; others retain previews. There are no separate `limit`, `offset`, or `has_more` response fields. Track pagination using request values and shelf `count`.

## Catalog creation

All seven endpoints require JWT and `Content-Type: application/json`. Success is HTTP 201 with the created entity in `data`. Names must be nonblank. These handlers return persistence entities rather than lowercase read DTOs: actual fields use Go names such as `ID`, `CollectionTypeName`, `Name`, and `ReleaseTypeName`; embedded timestamps and soft-delete fields are excluded from JSON.

| Route | Required JSON fields | Capability |
| --- | --- | --- |
| POST `/collection/types` | `name` | Add a collection classification used by grades |
| POST `/collection/grades` | `name`, `short_name`, `collection_type_id` | Add a grade linked to an existing type; type ID must be positive |
| POST `/collection/scales` | `name` | Add a reusable scale; no collection-type field is bound by the DTO |
| POST `/collection/release-types` | `name` | Add a release classification |
| POST `/collection/series` | `name` | Add a series for collections |
| POST `/collection/manufacturers` | `name` | Add a manufacturer for collections/addons |
| POST `/collection/metadata-tags` | `slug`, `name`, `type` | Add a tag; numeric type is 0 for modification or 1 for feature; explicitly supply 0 |

Duplicate names return HTTP 409 (grades are unique by name within a collection type); duplicate metadata slugs also return 409. A nonexistent grade collection type returns 400. Catalog creation refreshes filter caches. New scales need use by figure collections before appearing in `figures_scales`; new types do not add named shelf response fields. These routes do not provide catalog update/delete operations.

## Collection writes

### POST `/create_collection`

Requires JWT and `multipart/form-data`. Creates a collection, uploads images to Cloudinary, and associates addons/tags. Success is HTTP 201 with the detail shape returned by `GET /collection/:id`.

| Form field | Purpose / requirement |
| --- | --- |
| `title` | Required title |
| `grade_id`, `scale_id`, `release_type_id`, `manufacturer_id` | Required catalog IDs; collection type is derived through grade |
| `cover` | Required image file |
| `pictures` or `pictures[]` | Required gallery files; at least one, at most 20 |
| `status`, `series_id`, `description`, `display_size` | Optional metadata |
| `built_at`, `acquired_at` | Optional timestamps, e.g. `2026-11-18T10:00:00+07:00` |
| `addon_names`, `addons_manufacturer_id` | Optional names and optional corresponding manufacturer IDs; if IDs accompany names, lengths must match and IDs must be positive |
| `modifications`, `features` | Optional positive tag IDs of the corresponding metadata type |

Each image must be nonempty and at most 10 MiB; content inspection permits JPEG, PNG, GIF, and WebP. The gallery limit applies to files submitted per request. Tag IDs must exist and have the correct type; duplicates within a field are deduplicated, and IDs shared across fields are rejected.

For text arrays, use repeated keys or their `[]` variants, including numeric arrays. Do not advertise comma-separated arrays: the current parser can treat them as a single value. Internal fields such as `cover_url`, `picture_urls`, and `metadata_tag_ids` are not client inputs.

Status codes are `0` wishlist, `1` backlog, `2` owned, and `3` built; the DTO stores status as a string. Declared display sizes are `small_wide`, `small_tall`, `medium_wide`, `medium_tall`, `large_wide`, and `large_tall`. These describe intended values, not comprehensive enum validation.

PowerShell example: replace placeholder IDs with existing catalog IDs and provide local files.

```powershell
curl.exe -X POST "<api-base-url>/create_collection" -H "Authorization: Bearer <token>" -F "title=Example kit" -F "grade_id=1" -F "scale_id=1" -F "release_type_id=1" -F "manufacturer_id=1" -F "status=3" -F "cover=@cover.jpg" -F "pictures=@front.jpg"
```

### PATCH `/collection/:id`

Requires JWT, an integer ID, and multipart fields. Success is HTTP 200 with updated details. Scalar fields from creation are optional: omitted fields remain unchanged. Supply `cover` to replace the cover, and `new_pictures` / `new_pictures[]` to append gallery images, with the same file validation as creation.

| Additional form fields | Capability |
| --- | --- |
| `deleted_picture_urls` | Remove existing gallery images belonging to this collection |
| `new_addon_names`, `new_addons_manufacturer_id` | Append addons; supplied manufacturer IDs must align with names |
| `existing_addon_ids` | Keep only these existing addons; schedule others for deletion |
| `deleted_addon_ids` | Explicitly remove existing addons |
| `update_addon_ids`, `update_addon_names`, `update_addons_manufacturer_id` | Edit addons; supplied name/manufacturer arrays must match the ID array length |
| `modifications`, `features` | Replace the combined tag associations when either field is present |

Addon IDs must belong to the collection; edits cannot target addons scheduled for deletion. Omit `existing_addon_ids` to retain addons unless explicitly deleted; an explicitly empty value removes all existing addons. Omitted tags preserve associations only when neither tag field is supplied: when updating either category, send both desired sets to retain both. An empty field such as `modifications=` participates in replacement and can clear associations. Internal presence flags are not client inputs.

### DELETE `/collection/:id`

Requires JWT and an integer ID, with no body. Soft-deletes the collection through the repository, initiates associated Cloudinary cleanup, and invalidates collection caches. Success is HTTP 200 with `data: {"message":"Collection deleted successfully"}`; missing records return 404. Media cleanup is best effort, so success does not guarantee all external assets were removed.

## Shortlinks

### POST `/shortlink`

Accepts public JSON with required `original_url` (HTTP/HTTPS URL with a host) and `expiry_days` (1–365). Optional `custom_url` selects an alphanumeric code of at most ten characters; otherwise the service generates seven characters. A matching URL/duration may reuse an existing link when no custom code is requested. A custom-code conflict reported as a duplicate key is a validation error.

```http
POST /shortlink
Content-Type: application/json

{"original_url":"https://example.com/collection/42","expiry_days":30,"custom_url":"kit42"}
```

Success is HTTP 201 with a direct payload containing `short_code` and `expired_at`, including when reusing a link. Clients construct `<api-base-url>/shortlink/<short_code>`; no full short URL is returned. Errors use the shared envelope. Threat detection rejects malicious destinations when enabled; Web Risk screening depends on `WEBRISK_API_KEY` and is not always active.

### GET `/shortlink/:short_code`

Returns HTTP 302 with `Location` set to the original URL. Missing, expired, invalid, or rejected links, and any other service error, redirect to `shortlink.fe_url + "/notfound"` with HTTP 302 instead. Destination validation and enabled threat checks also apply during resolution, including cache hits. The handler does not return a JSON destination or JSON 404.

## Static docs and unmatched routes

The default router mounts `/docs` from `./dist` using Gin's static handler (GET/HEAD). That directory is absent from this checkout: no generated API specification or documentation UI is bundled here. Unmatched application routes use the shared 404 envelope; static-file handling may produce its own file-serving response.

## Implementation and maintenance rules

- Use `ShouldBindJSON` for catalog, admin-token, and shortlink JSON; `ShouldBindQuery` for list filters; and existing `ShouldBind` plus multipart parsing for collection uploads/updates. Return binding errors through the endpoint's error path.
- Preserve request tags, update pointers, and presence flags, which distinguish omission from explicit edits/empty arrays. Preserve release-type aliases and repository pagination/sort normalization without promising unsupported encodings.
- Collection/admin responses use `common.SuccessResponse` and `common.ErrorResponse` from `internal/common/response.go`. Domain read payloads generally live in `entity/response.go`; admin DTOs live in the handler. Preserve JSON tags, persistence-entity exceptions, and contractual zero-value fields.
- Preserve shortlink's direct HTTP 201 payload and HTTP 302 resolution. New ordinary JSON endpoints should use the common envelope and named DTOs unless the feature requires another contract.
- Check status mapping in `internal/common/error_parser.go` and cover behavior changes with focused handler tests. Review write cache invalidation. See the [architecture guide](architecture.md) and [development guide](development.md) for layer boundaries and validation.
