# Modami Core Service — FE/Mobile Integration Guide

> Base URL: `http://<host>/v1/core-services`
> Swagger UI: `http://<host>/swagger/index.html`
> Last updated: 2026-04-02

---

## Global Contract

### Request Headers

| Header | Value | When |
|---|---|---|
| `Content-Type` | `application/json` | All POST / PUT requests |
| `Authorization` | `Bearer <access_token>` | All authenticated endpoints |
| `X-Request-ID` | `<uuid>` | Optional — trace correlation |

### Response Envelope

**Success**
```json
{
  "success": true,
  "data": { ... },
  "meta": { }
}
```

**Success with cursor pagination**
```json
{
  "success": true,
  "data": [ ... ],
  "meta": {
    "next_cursor": "eyJ...",
    "has_more": true
  }
}
```

**Error**
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "...",
    "details": [
      { "field": "price", "message": "must be at least 1000" }
    ]
  }
}
```

### HTTP Status Reference

| Status | FE action |
|---|---|
| `200` | Render data |
| `201` | Created — update local state / list |
| `204` | No body — refresh list / pop screen |
| `400` | Validation error — map `details` to form fields |
| `401` | Token expired / missing — redirect to login or refresh |
| `403` | Insufficient permission — show permission error |
| `404` | Not found — show empty / not-found screen |
| `500` | Server error — show retry + fallback UI |

### Cursor Pagination

Send `cursor` + `limit` as query params. Use `meta.next_cursor` from the response as `cursor` in the next request. Stop when `meta.has_more` is `false`.

```
GET /products/feed?cursor=&limit=20
GET /products/feed?cursor=eyJ...&limit=20
```

Default page size is **20**; max is **100**.

---

## Domain Models

### Product

```json
{
  "id": "64f1a2b3c4d5e6f7a8b9c0d1",
  "seller_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "active",
  "title": "Áo khoác vintage Burberry",
  "slug": "ao-khoac-vintage-burberry-64f1",
  "description": "...",
  "price": 1500000,
  "category": {
    "id": "64f1a2b3c4d5e6f7a8b9c0d2",
    "name": "Outerwear",
    "name_vi": "Áo khoác",
    "slug": "outerwear",
    "icon": "https://..."
  },
  "condition": "like_new",
  "size": "M",
  "brand": "Burberry",
  "color": "beige",
  "material": "wool",
  "images": [
    { "url": "https://...", "position": 0, "width": 800, "height": 1000 }
  ],
  "is_verified": true,
  "is_featured": false,
  "is_select": false,
  "credit_cost": 0,
  "hashtags": ["vintage", "burberry"],
  "created_at": "2025-12-01T10:00:00Z",
  "updated_at": "2025-12-01T10:00:00Z",
  "published_at": "2025-12-01T10:05:00Z"
}
```

**Product statuses**: `draft` → `pending` → `active` → `sold` / `archived`

**Condition values**: `new` · `like_new` · `good` · `fair`

### Category

```json
{
  "id": "64f1a2b3c4d5e6f7a8b9c0d2",
  "name": "Outerwear",
  "name_vi": "Áo khoác",
  "slug": "outerwear",
  "icon": "https://...",
  "parent_id": null,
  "sort_order": 1,
  "is_active": true,
  "product_count": 142,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-12-01T00:00:00Z"
}
```

### BlogPost

```json
{
  "id": "64f1a2b3c4d5e6f7a8b9c0d3",
  "slug": "xu-huong-thu-dong-2025",
  "series_name": "MODAMI INSIGHT",
  "series_no": 12,
  "series_quarter": "Q4/2025",
  "post_type": "XU HƯỚNG TIÊU ĐIỂM",
  "depth": "deep",
  "title": "Xu hướng Thu Đông 2025",
  "subtitle": "...",
  "body": "...",
  "cover_image": "https://...",
  "cover_caption": "...",
  "read_time_min": 8,
  "word_count": 1200,
  "author": { "name": "...", "title": "...", "bio": "..." },
  "key_points": ["..."],
  "references": ["..."],
  "hashtags": ["thuong-hieu", "vintage"],
  "cta_link": "https://...",
  "is_featured": true,
  "status": "published",
  "published_at": "2025-12-01T10:00:00Z",
  "updated_at": "2025-12-01T10:00:00Z",
  "created_at": "2025-11-28T10:00:00Z"
}
```

**Post depth values**: `quick` · `deep`

---

## Public APIs (no auth required)

### Home Feed

#### `GET /home-feeds`

Returns 4 sections for the home screen in a single request (fetched concurrently server-side). Sections are always present; they return empty arrays on partial failure.

**Response `data`:**
```json
{
  "news": [ /* up to 10 newest active products */ ],
  "categories": [ /* up to 4 top-level active categories */ ],
  "near": [ /* up to 4 featured active products */ ],
  "blogs": [ /* up to 10 latest published posts */ ]
}
```

---

### Products

#### `GET /products/feed`

Cursor-paginated feed of active products (newest first).

| Param | Type | Default | Notes |
|---|---|---|---|
| `cursor` | string | `""` | Omit or empty for first page |
| `limit` | int | 20 | Max 100 |

---

#### `GET /products/featured`

Cursor-paginated list of featured products.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

---

#### `GET /products/select`

Cursor-paginated list of curated "Select" products.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

---

#### `GET /products/search`

Full-text search with optional filters (Elasticsearch-backed).

| Param | Type | Notes |
|---|---|---|
| `q` | string | Search query |
| `category_id` | string | MongoDB ObjectID |
| `condition` | string | `new` / `like_new` / `good` / `fair` |
| `brand` | string | |
| `min_price` | int | VND |
| `max_price` | int | VND |
| `cursor` | string | |
| `limit` | int | Default 20 |

Also available at `GET /search` (alias).

---

#### `GET /products/slug/:slug`

Get product detail by URL slug.

**Response `data`:**
```json
{
  "product": { /* Product object */ },
  "stats": {
    "totalView": 123,
    "totalLike": 45,
    "totalComment": 12
  }
}
```

---

#### `GET /products/:id`

Get product detail by MongoDB ID.

Same response shape as `/products/slug/:slug`.

---

#### `GET /products/:id/similar`

| Param | Type | Default | Max |
|---|---|---|---|
| `limit` | int | 10 | 50 |

**Response `data`:** array of Product objects.

---

#### `GET /products/:id/moderation`

Moderation history for a product (admin / seller visibility).

**Response `data`:** array of:
```json
{
  "id": "...",
  "product_id": "...",
  "round": 1,
  "action": "rejected",
  "reject_code": "POOR_IMAGES",
  "reason": "...",
  "suggestion": "...",
  "moderator_id": "uuid",
  "created_at": "..."
}
```

---

#### `POST /products/:id/view`

Track a product view (fire-and-forget). Returns `204`.

---

#### `GET /hashtags/:tag/products`

Products tagged with a specific hashtag.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

---

### Search

#### `GET /search`

Alias for `GET /products/search` — same params and response.

#### `GET /search/suggest`

Hashtag autocomplete.

| Param | Type | Required |
|---|---|---|
| `q` | string | Yes — prefix string |

**Response `data`:** array of `{ "tag": "vintage", "usage_count": 412 }`

#### `GET /search/trending`

| Param | Type | Default |
|---|---|---|
| `limit` | int | 20 |

**Response `data`:** array of `{ "tag": "vintage", "usage_count": 412 }`

---

### Categories

#### `GET /categories`

Full category tree (all active + inactive, ordered by `sort_order`).

**Response `data`:** array of Category objects.

#### `GET /categories/:slug`

Single category by slug.

**Response `data`:** Category object.

#### `GET /categories/:slug/children`

Direct children of a category.

**Response `data`:** array of Category objects.

---

### Hashtags

#### `GET /hashtags/trending`

| Param | Type | Default |
|---|---|---|
| `limit` | int | 20 |

#### `GET /hashtags/suggest`

| Param | Type | Required |
|---|---|---|
| `q` | string | Yes |

---

### Sellers

#### `GET /sellers/:id`

Public seller profile. `:id` is the seller's UUID from auth-service.

**Response `data`:**
```json
{
  "seller_id": "550e8400-...",
  "product_count": 0,
  "follower_count": 120,
  "following_count": 45,
  "avg_rating": 0,
  "review_count": 0
}
```

#### `GET /sellers/:id/products`

Active products listed by seller.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

#### `GET /sellers/:id/reviews`

Reviews for a seller.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

**Response `data`:** array of:
```json
{
  "id": "...",
  "order_id": "...",
  "product_id": "...",
  "buyer_id": "uuid",
  "seller_id": "uuid",
  "rating": 5,
  "comment": "...",
  "images": ["https://..."],
  "created_at": "..."
}
```

#### `GET /sellers/:id/stats`

Aggregated seller statistics.

**Response `data`:**
```json
{
  "total_products": 24,
  "total_sold": 10,
  "avg_rating": 4.8,
  "review_count": 32,
  "follower_count": 120
}
```

---

### Community & Blog

#### `GET /community`

Hero featured post + recent published posts for the community screen.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

**Response `data`:**
```json
{
  "featured": { /* BlogPost or null */ },
  "posts": [ /* array of BlogPost */ ],
  "next_cursor": "...",
  "has_more": false
}
```

#### `GET /blog/posts`

Paginated list of published posts, optionally filtered by type.

| Param | Type | Notes |
|---|---|---|
| `post_type` | string | e.g. `"XU HƯỚNG TIÊU ĐIỂM"` |
| `cursor` | string | |
| `limit` | int | Default 20 |

#### `GET /blog/posts/:slug`

Single post by slug.

**Response `data`:** BlogPost object.

#### `GET /blog/reports`

Trend report posts (`post_type = "trend_report"`).

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

#### `GET /blog/hashtags/:tag`

Posts that carry a specific hashtag.

| Param | Type | Default |
|---|---|---|
| `cursor` | string | `""` |
| `limit` | int | 20 |

---

## Authenticated APIs (seller — requires `Authorization: Bearer`)

> All endpoints below require a valid Keycloak access token.
> The user ID is extracted server-side from the JWT — do **not** pass it in the request body.

### My Products

#### `GET /products/me`

List the authenticated seller's own products.

| Param | Type | Notes |
|---|---|---|
| `status` | string | Filter: `draft` / `pending` / `active` / `sold` / `archived` |
| `cursor` | string | |
| `limit` | int | Default 20 |

---

### Product Lifecycle

#### `POST /products`

Create a new product (starts in `draft` status).

**Request body:**
```json
{
  "title": "string (required, 5–200)",
  "description": "string (required, 20–5000)",
  "price": 1500000,
  "category_id": "64f1a2b3c4d5e6f7a8b9c0d2",
  "condition": "like_new",
  "size": "M",
  "brand": "Burberry",
  "color": "beige",
  "material": "wool",
  "images": [
    { "url": "https://...", "position": 0, "width": 800, "height": 1000 }
  ],
  "hashtags": ["vintage", "burberry"],
  "credit_cost": 0
}
```

| Field | Required | Validation |
|---|---|---|
| `title` | Yes | 5–200 chars |
| `description` | Yes | 20–5000 chars |
| `price` | Yes | min 1000 (VND) |
| `category_id` | Yes | valid category ObjectID |
| `condition` | Yes | `new` / `like_new` / `good` / `fair` |
| `size` | Yes | |
| `images` | Yes | 1–6 items; each `url` is required |
| `hashtags` | No | max 10 |

Returns `201` with the created Product object.

---

#### `PUT /products/:id`

Update an existing product (partial update — only send changed fields).

**Request body:** same fields as create but all optional:
```json
{
  "title": "New title",
  "price": 1800000,
  "images": [ ... ]
}
```

Returns `200` with the updated Product object.

---

#### `DELETE /products/:id`

Hard delete a product. Only the owning seller can delete. Returns `204`.

---

#### `POST /products/:id/submit`

Submit a `draft` product for moderation review. Transitions status: `draft` → `pending`.

No request body. Returns `200` with updated Product.

---

#### `POST /products/:id/resubmit`

Resubmit a `rejected` (back to `draft`) product after making corrections.

**Request body:** same optional fields as update, plus:
```json
{
  "note": "Đã thay ảnh rõ hơn theo yêu cầu"
}
```

Returns `200` with updated Product.

---

#### `POST /products/:id/archive`

Move an `active` product to `archived`. Returns `200` with updated Product.

---

#### `POST /products/:id/unarchive`

Move an `archived` product back to `active`. Returns `200` with updated Product.

---

## Admin APIs (requires auth + permission claim in JWT)

> These endpoints require both a valid Bearer token **and** a specific Keycloak role/permission in the JWT claims.
> The required permission is noted for each endpoint.

---

### Product Administration

| Method | Path | Permission | Description |
|---|---|---|---|
| — | — | — | No product admin endpoints yet — moderation is handled by a separate admin service |

---

### Category Administration

#### `POST /categories` — permission: `category.create`

Create a new category.

**Request body:**
```json
{
  "name": "Outerwear",
  "name_vi": "Áo khoác",
  "slug": "outerwear",
  "icon": "https://cdn.../icon.svg",
  "parent_id": null,
  "sort_order": 1
}
```

| Field | Required | Notes |
|---|---|---|
| `name` | Yes | English display name |
| `name_vi` | Yes | Vietnamese display name |
| `slug` | Yes | URL-safe string, unique |
| `icon` | No | Image URL |
| `parent_id` | No | ObjectID string of parent category; omit for root |
| `sort_order` | No | Integer, ascending |

Returns `201` with the created Category object.

---

#### `PUT /categories/:id` — permission: `category.update`

Update category fields (partial update).

**Request body:**
```json
{
  "name": "New name",
  "name_vi": "Tên mới",
  "slug": "new-slug",
  "icon": "https://...",
  "sort_order": 2
}
```

All fields optional. Returns `200` with updated Category object.

---

#### `PUT /categories/:id/toggle` — permission: `category.manage`

Toggle the `is_active` flag on a category. No request body.

Returns `200` with updated Category object.

---

#### `DELETE /categories/:id` — permission: `category.delete`

Delete a category. Returns `204`.

---

#### `PUT /categories/reorder` — permission: `category.manage`

Bulk-update `sort_order` for multiple categories in one request.

**Request body:** array of `{ id, sort_order }`:
```json
[
  { "id": "64f1a2b3c4d5e6f7a8b9c0d2", "sort_order": 1 },
  { "id": "64f1a2b3c4d5e6f7a8b9c0d3", "sort_order": 2 }
]
```

Returns `204`.

---

### Blog / Community Administration

#### `POST /blog/posts` — permission: `blog.create`

Create a new blog post (starts in `draft` status).

**Request body:**
```json
{
  "slug": "xu-huong-thu-dong-2025",
  "title": "Xu hướng Thu Đông 2025",
  "subtitle": "...",
  "body": "Nội dung bài viết...",
  "post_type": "XU HƯỚNG TIÊU ĐIỂM",
  "depth": "deep",
  "series_name": "MODAMI INSIGHT",
  "series_no": 12,
  "series_quarter": "Q4/2025",
  "cover_image": "https://...",
  "cover_caption": "...",
  "read_time_min": 8,
  "word_count": 1200,
  "author": {
    "name": "Nguyen Van A",
    "title": "Fashion Editor",
    "bio": "..."
  },
  "key_points": ["Màu sắc trung tính trở lại", "..."],
  "references": ["https://..."],
  "hashtags": ["thuong-hieu", "vintage"],
  "cta_link": "https://...",
  "is_featured": false
}
```

| Field | Required | Validation |
|---|---|---|
| `slug` | Yes | URL-safe, unique |
| `title` | Yes | |
| `depth` | No | `quick` / `deep` |

Returns `201` with created BlogPost object.

---

#### `PUT /blog/posts/:id` — permission: `blog.update`

Partial update of a blog post. All fields optional.

**Request body:** same fields as create, all as optional (`*string` / `*int` / etc.). Only send changed fields.

Returns `200` with updated BlogPost object.

---

#### `POST /blog/posts/:id/publish` — permission: `blog.publish`

Transition a blog post from `draft` → `published`. Sets `published_at` to now.

No request body. Returns `200` with updated BlogPost object.

---

#### `DELETE /blog/posts/:id` — permission: `blog.delete`

Delete a blog post. Returns `204`.

---

## Error Codes Reference

| HTTP | Scenario | FE handling |
|---|---|---|
| `400` | Missing required field / validation failure | Map `error.details[].field` to form error |
| `401` | Token missing or expired | Refresh token or redirect to login |
| `403` | Permission denied (e.g., editing another seller's product) | Show "không có quyền" message |
| `404` | Product / category / post not found | Show not-found screen |
| `409` | Slug already exists (categories, blog posts) | Prompt user to change slug |
| `422` | Invalid state transition (e.g., submit a non-draft) | Show current status + allowed actions |
| `500` | Internal server error | Show retry button |

---

## Screens → API Mapping

| Screen | APIs to call |
|---|---|
| Home | `GET /home-feeds` |
| Product feed | `GET /products/feed` |
| Product detail | `GET /products/slug/:slug` or `GET /products/:id`, `POST /products/:id/view` |
| Product search | `GET /search?q=...` |
| Category browse | `GET /categories`, `GET /categories/:slug/children`, `GET /products/search?category_id=...` |
| Hashtag page | `GET /hashtags/:tag/products` |
| Seller profile | `GET /sellers/:id`, `GET /sellers/:id/products`, `GET /sellers/:id/reviews`, `GET /sellers/:id/stats` |
| Community feed | `GET /community` |
| Blog post detail | `GET /blog/posts/:slug` |
| Trend reports | `GET /blog/reports` |
| My listings | `GET /products/me` |
| Create listing | `POST /products` → `POST /products/:id/submit` |
| Edit listing | `PUT /products/:id` |
| Resubmit rejected | `POST /products/:id/resubmit` |
| Archive listing | `POST /products/:id/archive` |
