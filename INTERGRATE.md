# INTERGRATE - FE API Integration Guide

Tai lieu nay duoc sinh tu `docs/swagger.json` de FE co the tich hop API day du theo contract backend.

## 1) Global Contract

- **Swagger version**: `2.0`
- **Service**: `Modami Core Service API`
- **Version**: `1.0`
- **Schemes**: `http, https`
- **Host (default swagger)**: `localhost:8087`
- **Base path**: `/v1/core-services`

### Request Rules

- Header mac dinh: `Content-Type: application/json`
- Header auth: `Authorization: Bearer <access_token>` cho endpoint co `BearerAuth`
- Full URL mau: `<scheme>://<host>` + `basePath` + `path`

### Standard Envelope

- **Success**: `internal_adapter_handler.StandardSuccessEnvelope`
  - `success`: boolean
  - `data`: object (noi dung tra ve thuc te)
  - `meta`: object (phan trang, thong tin bo sung...)
- **Error**: `internal_adapter_handler.StandardErrorEnvelope`
  - `success`: boolean
  - `error`: object (message/code/details tuy endpoint)
  - `meta`: object

### HTTP Status Mapping cho FE

| Status | Y nghia FE nen xu ly |
|---|---|
| `200` | Thanh cong, render du lieu |
| `201` | Tao moi thanh cong, update state/local cache |
| `204` | Thanh cong khong body, chi can refresh list/state |
| `400` | Validation sai input, map loi vao form |
| `401` | Chua dang nhap/het han token, redirect login hoac refresh token |
| `403` | Khong du quyen, hien thong bao permission |
| `404` | Khong tim thay resource, hien empty/not-found screen |
| `500` | Loi he thong, hien retry + fallback UI |

## 2) DTO/Schema Validation Master

### `github_com_modami_core-service_internal_domain.CategoryOrder`

| field | type | required | validation |
|---|---|---|---|
| `id` | `string` | no | - |
| `sort_order` | `integer` | no | - |

### `github_com_modami_core-service_internal_dto.CreateProductRequest`

| field | type | required | validation |
|---|---|---|---|
| `brand` | `string` | no | - |
| `category_id` | `string` | yes | - |
| `color` | `string` | no | - |
| `condition` | `string` | yes | enum=new, like_new, good, fair |
| `credit_cost` | `integer` | no | - |
| `description` | `string` | yes | minLength=20; maxLength=5000 |
| `hashtags` | `array<string>` | no | maxItems=10 |
| `images` | `array<github_com_modami_core-service_internal_dto.ImageInput>` | yes | minItems=1; maxItems=6 |
| `material` | `string` | no | - |
| `price` | `integer` | yes | minimum=1000 |
| `size` | `string` | yes | - |
| `title` | `string` | yes | minLength=5; maxLength=200 |

### `github_com_modami_core-service_internal_dto.ImageInput`

| field | type | required | validation |
|---|---|---|---|
| `height` | `integer` | no | - |
| `position` | `integer` | no | - |
| `url` | `string` | yes | - |
| `width` | `integer` | no | - |

### `github_com_modami_core-service_internal_dto.ResubmitRequest`

| field | type | required | validation |
|---|---|---|---|
| `brand` | `string` | no | - |
| `category_id` | `string` | no | - |
| `color` | `string` | no | - |
| `condition` | `string` | no | enum=new, like_new, good, fair |
| `description` | `string` | no | minLength=20; maxLength=5000 |
| `hashtags` | `array<string>` | no | maxItems=10 |
| `images` | `array<github_com_modami_core-service_internal_dto.ImageInput>` | no | minItems=1; maxItems=6 |
| `material` | `string` | no | - |
| `note` | `string` | no | - |
| `price` | `integer` | no | minimum=1000 |
| `size` | `string` | no | - |
| `title` | `string` | no | minLength=5; maxLength=200 |

### `github_com_modami_core-service_internal_dto.UpdateProductRequest`

| field | type | required | validation |
|---|---|---|---|
| `brand` | `string` | no | - |
| `category_id` | `string` | no | - |
| `color` | `string` | no | - |
| `condition` | `string` | no | enum=new, like_new, good, fair |
| `credit_cost` | `integer` | no | - |
| `description` | `string` | no | minLength=20; maxLength=5000 |
| `hashtags` | `array<string>` | no | maxItems=10 |
| `images` | `array<github_com_modami_core-service_internal_dto.ImageInput>` | no | minItems=1; maxItems=6 |
| `material` | `string` | no | - |
| `price` | `integer` | no | minimum=1000 |
| `size` | `string` | no | - |
| `title` | `string` | no | minLength=5; maxLength=200 |

### `internal_adapter_handler.CreateCategoryRequest`

| field | type | required | validation |
|---|---|---|---|
| `icon` | `string` | no | - |
| `name` | `string` | yes | - |
| `name_vi` | `string` | yes | - |
| `parent_id` | `string` | no | - |
| `slug` | `string` | yes | - |
| `sort_order` | `integer` | no | - |

### `internal_adapter_handler.StandardErrorEnvelope`

| field | type | required | validation |
|---|---|---|---|
| `error` | `object` | no | - |
| `meta` | `object` | no | - |
| `success` | `boolean` | no | - |

### `internal_adapter_handler.StandardSuccessEnvelope`

| field | type | required | validation |
|---|---|---|---|
| `data` | `object` | no | - |
| `meta` | `object` | no | - |
| `success` | `boolean` | no | - |

### `internal_adapter_handler.UpdateCategoryRequest`

| field | type | required | validation |
|---|---|---|---|
| `icon` | `string` | no | - |
| `name` | `string` | no | - |
| `name_vi` | `string` | no | - |
| `slug` | `string` | no | - |
| `sort_order` | `integer` | no | - |

## 3) Endpoint Catalog (Full)

Moi endpoint ben duoi gom: auth, input (path/query/body), output, error codes, va sample request cho FE.

## Admin APIs

### `POST /categories`

- **Summary**: Admin: create category
- **Auth required**: `yes`
- **Consumes**: `application/json`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `body` | `body` | yes | `internal_adapter_handler.CreateCategoryRequest` | Category |

**Output / Error**

| status | description | schema |
|---|---|---|
| `201` | Created | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/categories' \
  -H 'Authorization: Bearer {TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{"TODO":"fill_body_based_on_schema"}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `PUT /categories/reorder`

- **Summary**: Admin: reorder categories
- **Auth required**: `yes`
- **Consumes**: `application/json`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `body` | `body` | yes | `array<github_com_modami_core-service_internal_domain.CategoryOrder>` | Ordered ids |

**Output / Error**

| status | description | schema |
|---|---|---|
| `204` | No Content | `-` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X PUT '{SCHEME}://{HOST}/v1/core-services/categories/reorder' \
  -H 'Authorization: Bearer {TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{"TODO":"fill_body_based_on_schema"}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `PUT /categories/{id}`

- **Summary**: Admin: update category
- **Auth required**: `yes`
- **Consumes**: `application/json`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Category ID |
| `body` | `body` | yes | `internal_adapter_handler.UpdateCategoryRequest` | Fields |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X PUT '{SCHEME}://{HOST}/v1/core-services/categories/{id}' \
  -H 'Authorization: Bearer {TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{"TODO":"fill_body_based_on_schema"}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `DELETE /categories/{id}`

- **Summary**: Admin: delete category
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Category ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `204` | No Content | `-` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X DELETE '{SCHEME}://{HOST}/v1/core-services/categories/{id}' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `PUT /categories/{id}/toggle`

- **Summary**: Admin: toggle category active
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Category ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X PUT '{SCHEME}://{HOST}/v1/core-services/categories/{id}/toggle' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## Categories APIs

### `GET /categories`

- **Summary**: List categories
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**: none

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/categories'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /categories/{slug}`

- **Summary**: Get category by slug
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `slug` | `path` | yes | `string` | Category slug |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/categories/{slug}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /categories/{slug}/children`

- **Summary**: List child categories
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `slug` | `path` | yes | `string` | Parent category slug |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/categories/{slug}/children'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## Hashtags APIs

### `GET /hashtags/suggest`

- **Summary**: Suggest hashtags (autocomplete)
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `q` | `query` | yes | `string` | Prefix |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/hashtags/suggest'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /hashtags/trending`

- **Summary**: Trending hashtags
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `limit` | `query` | no | `integer` | Max tags |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/hashtags/trending'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## Products APIs

### `GET /hashtags/{tag}/products`

- **Summary**: Products by hashtag
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `tag` | `path` | yes | `string` | Hashtag (without #) |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/hashtags/{tag}/products'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `POST /products`

- **Summary**: Create product
- **Auth required**: `yes`
- **Consumes**: `application/json`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `body` | `body` | yes | `github_com_modami_core-service_internal_dto.CreateProductRequest` | Product payload |

**Output / Error**

| status | description | schema |
|---|---|---|
| `201` | Created | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/products' \
  -H 'Authorization: Bearer {TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{"TODO":"fill_body_based_on_schema"}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/featured`

- **Summary**: Featured products
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/featured'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/feed`

- **Summary**: Product feed (cursor)
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/feed'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/me`

- **Summary**: List my products (seller)
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `status` | `query` | no | `string` | Filter by status |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/me' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/select`

- **Summary**: Curated select products
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/select'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/slug/{slug}`

- **Summary**: Get product by slug
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `slug` | `path` | yes | `string` | URL slug |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/slug/{slug}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/{id}`

- **Summary**: Get product by ID
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/{id}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `PUT /products/{id}`

- **Summary**: Update product
- **Auth required**: `yes`
- **Consumes**: `application/json`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |
| `body` | `body` | yes | `github_com_modami_core-service_internal_dto.UpdateProductRequest` | Fields to update |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X PUT '{SCHEME}://{HOST}/v1/core-services/products/{id}' \
  -H 'Authorization: Bearer {TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{"TODO":"fill_body_based_on_schema"}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `DELETE /products/{id}`

- **Summary**: Delete product
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `204` | No Content | `-` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X DELETE '{SCHEME}://{HOST}/v1/core-services/products/{id}' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `POST /products/{id}/archive`

- **Summary**: Archive product
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/products/{id}/archive' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/{id}/moderation`

- **Summary**: List moderation history for product
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/{id}/moderation'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `POST /products/{id}/resubmit`

- **Summary**: Resubmit rejected product
- **Auth required**: `yes`
- **Consumes**: `application/json`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |
| `body` | `body` | yes | `github_com_modami_core-service_internal_dto.ResubmitRequest` | Updates |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/products/{id}/resubmit' \
  -H 'Authorization: Bearer {TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{"TODO":"fill_body_based_on_schema"}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /products/{id}/similar`

- **Summary**: Similar products
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |
| `limit` | `query` | no | `integer` | Max items (default 10, max 50) |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/{id}/similar'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `POST /products/{id}/submit`

- **Summary**: Submit product for review
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/products/{id}/submit' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `POST /products/{id}/unarchive`

- **Summary**: Unarchive product
- **Auth required**: `yes`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `401` | Unauthorized | `internal_adapter_handler.StandardErrorEnvelope` |
| `403` | Forbidden | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/products/{id}/unarchive' \
  -H 'Authorization: Bearer {TOKEN}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `POST /products/{id}/view`

- **Summary**: Track product view
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `-`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Product ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `204` | No Content | `-` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X POST '{SCHEME}://{HOST}/v1/core-services/products/{id}/view'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /sellers/{seller_id}/products`

- **Summary**: List products by seller
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `seller_id` | `path` | yes | `string` | Seller user ID |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/sellers/{seller_id}/products'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## Search APIs

### `GET /products/search`

- **Summary**: Search products (catalog)
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `q` | `query` | no | `string` | Search query |
| `category_id` | `query` | no | `string` | Category filter |
| `condition` | `query` | no | `string` | Condition |
| `brand` | `query` | no | `string` | Brand |
| `min_price` | `query` | no | `integer` | Min price |
| `max_price` | `query` | no | `integer` | Max price |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/products/search'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /search`

- **Summary**: Search products (alias of GET /products/search)
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `q` | `query` | no | `string` | Search query |
| `category_id` | `query` | no | `string` | Category filter |
| `condition` | `query` | no | `string` | Condition |
| `brand` | `query` | no | `string` | Brand |
| `min_price` | `query` | no | `integer` | Min price |
| `max_price` | `query` | no | `integer` | Max price |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/search'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /search/suggest`

- **Summary**: Hashtag suggestions (alias of GET /hashtags/suggest)
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `q` | `query` | yes | `string` | Prefix |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `400` | Bad Request | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/search/suggest'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /search/trending`

- **Summary**: Trending hashtags (alias of GET /hashtags/trending)
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `limit` | `query` | no | `integer` | Max tags |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/search/trending'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## Sellers APIs

### `GET /sellers/{id}`

- **Summary**: Public seller profile
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Seller user ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/sellers/{id}'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /sellers/{id}/products`

- **Summary**: Seller's listed products
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Seller user ID |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/sellers/{id}/products'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /sellers/{id}/reviews`

- **Summary**: Seller reviews
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Seller user ID |
| `cursor` | `query` | no | `string` | Pagination cursor |
| `limit` | `query` | no | `integer` | Page size |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/sellers/{id}/reviews'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

### `GET /sellers/{id}/stats`

- **Summary**: Seller public stats
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**

| name | in | required | type/schema | note |
|---|---|---|---|---|
| `id` | `path` | yes | `string` | Seller user ID |

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | OK | `internal_adapter_handler.StandardSuccessEnvelope` |
| `404` | Not Found | `internal_adapter_handler.StandardErrorEnvelope` |
| `500` | Internal Server Error | `internal_adapter_handler.StandardErrorEnvelope` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/sellers/{id}/stats'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## System APIs

### `GET /health`

- **Summary**: Health check
- **Auth required**: `no`
- **Consumes**: `-`
- **Produces**: `application/json`

**Input**: none

**Output / Error**

| status | description | schema |
|---|---|---|
| `200` | Example: {\"status\":\"ok\"} | `object` |

**cURL sample**

```bash
curl -X GET '{SCHEME}://{HOST}/v1/core-services/health'
```

**FE integration notes**

- Validate input theo bang schema o muc `2)` truoc khi goi API.
- Luon parse theo envelope `success/data/meta` hoac `success/error/meta`.
- Xu ly status code theo bang mapping muc `1)`; khong hard-code message server.

## 4) FE Flow Mapping (Goi y)

- **Home feed**: `GET /products/feed`, `GET /products/featured`, `GET /products/select`
- **Search**: `GET /search`, `GET /search/suggest`, `GET /search/trending`, `GET /products/search`
- **Product detail**: `GET /products/{id}` hoac `GET /products/slug/{slug}`, sau do `POST /products/{id}/view`
- **Seller profile**: `GET /sellers/{id}`, `GET /sellers/{id}/products`, `GET /sellers/{id}/stats`, `GET /sellers/{id}/reviews`
- **My products**: `GET /products/me`, `POST /products`, `PUT /products/{id}`, `POST /products/{id}/submit`, `POST /products/{id}/resubmit`, `POST /products/{id}/archive`, `POST /products/{id}/unarchive`
- **Admin category**: `/categories*` nhom endpoint CRUD/reorder/toggle

## 5) TypeScript Starter Types (FE copy nhanh)

```ts
export type ApiSuccess<T = unknown, M = Record<string, unknown>> = {
  success: true;
  data: T;
  meta?: M;
};

export type ApiError<E = Record<string, unknown>, M = Record<string, unknown>> = {
  success: false;
  error: E;
  meta?: M;
};

export type ApiEnvelope<T = unknown, E = Record<string, unknown>, M = Record<string, unknown>> =
  | ApiSuccess<T, M>
  | ApiError<E, M>;
```

## 6) QA Checklist cho FE Integration

- Verify du path param/query param/body cho tung endpoint truoc merge.
- Verify required field + enum + min/max length + min/max items.
- Verify render dung khi `data` rong (empty list, null details...).
- Verify luong loi `400/401/403/404/500` tren tung man hinh chinh.
- Verify loading state + retry state + timeout/network-fail state.
- Verify auth header chi attach vao endpoint can auth.

---

Nguon su that contract: `docs/swagger.json` / `docs/swagger.yaml`. Neu backend doi contract, can regenerate/cap nhat file nay cung luc.
