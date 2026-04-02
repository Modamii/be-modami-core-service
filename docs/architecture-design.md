# ModaMi Core Service — Database & API Design v3 (Production)

---

## 1. Tổng quan thay đổi so với v2

| Thay đổi | Lý do |
|----------|-------|
| `listings` → `products` | Đây là sản phẩm thật (quần áo), không phải tin đăng. User nghĩ "sản phẩm", code nên reflect |
| Thêm `is_verified` | Badge "Đã xác thực" trên UI (Image 2) |
| Thêm `categories` collection | Master data cho danh mục, quản lý từ admin |
| Thêm `packages` + `subscriptions` | Gói thành viên Style/Elite (Image 1) |
| Thêm `saved_products` | User lưu sản phẩm để xem sau (khác favorite/yêu thích) |
| Thêm `daily_stats`, `seller_stats_snapshots` | Dashboard thống kê cho admin + seller |
| Thêm `reports` | Báo cáo vi phạm sản phẩm |
| Thêm `hashtags` | Master data cho hashtag/tag trending |

---

## 2. Collections (21 total)

### 2.1. `products` — Sản phẩm

```go
type Product struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    SellerID  primitive.ObjectID `bson:"seller_id"`
    Status    ProductStatus      `bson:"status"`
    Version   int64              `bson:"version"`

    // --- Thông tin sản phẩm ---
    Title       string `bson:"title"`
    Slug        string `bson:"slug"`
    Description string `bson:"description"`
    Price       int64  `bson:"price"` // VND

    // --- Phân loại ---
    CategoryID primitive.ObjectID `bson:"category_id"`
    Condition  string             `bson:"condition"` // new | like_new | good | fair
    Size       string             `bson:"size"`
    Brand      string             `bson:"brand,omitempty"`
    Color      string             `bson:"color,omitempty"`
    Material   string             `bson:"material,omitempty"`

    // --- Media (max 6) ---
    Images []ProductImage `bson:"images"`

    // --- Flags ---
    IsVerified bool `bson:"is_verified"` // "Đã xác thực" badge
    IsFeatured bool `bson:"is_featured"` // "Nổi bật" tag
    IsSelect   bool `bson:"is_select"`   // ModaMi Select program

    // --- Credit ---
    CreditCost int `bson:"credit_cost"` // cost to unlock contact info

    // --- Hashtags (max 10) ---
    Hashtags []string `bson:"hashtags,omitempty"`

    // --- Timestamps ---
    CreatedAt   time.Time  `bson:"created_at"`
    UpdatedAt   time.Time  `bson:"updated_at"`
    PublishedAt *time.Time `bson:"published_at,omitempty"`
    SoldAt      *time.Time `bson:"sold_at,omitempty"`
    DeletedAt   *time.Time `bson:"deleted_at,omitempty"`
}

type ProductStatus string
const (
    ProductStatusDraft    ProductStatus = "draft"
    ProductStatusPending  ProductStatus = "pending"
    ProductStatusActive   ProductStatus = "active"
    ProductStatusSold     ProductStatus = "sold"
    ProductStatusArchived ProductStatus = "archived"
)

type ProductImage struct {
    URL      string `bson:"url"`
    Position int    `bson:"position"` // 0 = cover
    Width    int    `bson:"width,omitempty"`
    Height   int    `bson:"height,omitempty"`
}
```

**Indexes:**
```javascript
db.products.createIndex({ seller_id: 1, status: 1, created_at: -1 })
db.products.createIndex({ status: 1, category_id: 1, published_at: -1 })
db.products.createIndex({ status: 1, is_featured: 1, published_at: -1 })
db.products.createIndex({ status: 1, is_select: 1, published_at: -1 })
db.products.createIndex({ status: 1, is_verified: 1, published_at: -1 })
db.products.createIndex({ slug: 1 }, { unique: true })
db.products.createIndex({ hashtags: 1, status: 1 })
db.products.createIndex({ price: 1, status: 1 })
db.products.createIndex({ brand: 1, status: 1 })
db.products.createIndex(
    { deleted_at: 1 },
    { partialFilterExpression: { deleted_at: { $ne: null } } }
)
```

---

### 2.2. `product_stats` — Counters (hot data, tách riêng)

```go
type ProductStats struct {
    ProductID     primitive.ObjectID `bson:"_id"`
    ViewCount     int64              `bson:"view_count"`
    FavoriteCount int64              `bson:"favorite_count"`
    SaveCount     int64              `bson:"save_count"`
    UnlockCount   int64              `bson:"unlock_count"`
    ShareCount    int64              `bson:"share_count"`
    UpdatedAt     time.Time          `bson:"updated_at"`
}
```

---

### 2.3. `product_moderations` — Lịch sử xét duyệt

```go
type ProductModeration struct {
    ID          primitive.ObjectID  `bson:"_id,omitempty"`
    ProductID   primitive.ObjectID  `bson:"product_id"`
    Round       int                 `bson:"round"`
    Action      string              `bson:"action"` // submitted | approved | rejected
    RejectCode  string              `bson:"reject_code,omitempty"`
    Reason      string              `bson:"reason,omitempty"`
    Note        string              `bson:"note,omitempty"`
    Suggestion  string              `bson:"suggestion,omitempty"`
    ModeratorID *primitive.ObjectID `bson:"moderator_id,omitempty"`
    CreatedAt   time.Time           `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.product_moderations.createIndex({ product_id: 1, created_at: -1 })
db.product_moderations.createIndex({ action: 1, created_at: -1 })
db.product_moderations.createIndex({ moderator_id: 1, created_at: -1 })
```

---

### 2.4. `select_products` — ModaMi Select metadata

```go
type SelectProduct struct {
    ProductID     primitive.ObjectID `bson:"_id"`
    Campaign      string             `bson:"campaign"`
    Story         string             `bson:"story"`
    Provenance    string             `bson:"provenance"`
    Year          int                `bson:"year,omitempty"`
    CertificateID string             `bson:"certificate_id"`
    VerifiedBy    primitive.ObjectID `bson:"verified_by"`
    VerifiedAt    time.Time          `bson:"verified_at"`
    CreatedAt     time.Time          `bson:"created_at"`
}
```

---

### 2.5. `categories` — Danh mục sản phẩm (Master Data)

```go
type Category struct {
    ID           primitive.ObjectID  `bson:"_id,omitempty"`
    Name         string              `bson:"name"`        // "Outerwear"
    NameVI       string              `bson:"name_vi"`     // "Áo khoác"
    Slug         string              `bson:"slug"`        // "outerwear"
    Icon         string              `bson:"icon,omitempty"`
    ParentID     *primitive.ObjectID `bson:"parent_id,omitempty"` // subcategory
    SortOrder    int                 `bson:"sort_order"`
    IsActive     bool                `bson:"is_active"`
    ProductCount int64               `bson:"product_count"` // denormalized
    CreatedAt    time.Time           `bson:"created_at"`
    UpdatedAt    time.Time           `bson:"updated_at"`
}
```

**Indexes:**
```javascript
db.categories.createIndex({ slug: 1 }, { unique: true })
db.categories.createIndex({ parent_id: 1, sort_order: 1 })
db.categories.createIndex({ is_active: 1, sort_order: 1 })
```

**Seed Data:**
```javascript
[
    { name: "Tops",        name_vi: "Áo",       slug: "tops",        sort_order: 1 },
    { name: "Bottoms",     name_vi: "Quần",     slug: "bottoms",     sort_order: 2 },
    { name: "Dresses",     name_vi: "Váy/Đầm", slug: "dresses",     sort_order: 3 },
    { name: "Outerwear",   name_vi: "Áo khoác", slug: "outerwear",   sort_order: 4 },
    { name: "Shoes",       name_vi: "Giày dép", slug: "shoes",       sort_order: 5 },
    { name: "Bags",        name_vi: "Túi xách", slug: "bags",        sort_order: 6 },
    { name: "Accessories", name_vi: "Phụ kiện", slug: "accessories", sort_order: 7 },
]
```

---

### 2.6. `packages` — Gói thành viên (Master Data)

```go
type Package struct {
    ID   primitive.ObjectID `bson:"_id,omitempty"`
    Code string             `bson:"code"` // "curator" | "style" | "elite"
    Name string             `bson:"name"`
    Tier int                `bson:"tier"` // 0=free, 1=style, 2=elite

    // Pricing
    PriceMonthly int64  `bson:"price_monthly"`
    PriceYearly  int64  `bson:"price_yearly"`
    Currency     string `bson:"currency"` // "VND"

    // Benefits
    CreditsPerMonth int    `bson:"credits_per_month"`
    SearchBoost     bool   `bson:"search_boost"`
    SearchPriority  bool   `bson:"search_priority"`
    BadgeName       string `bson:"badge_name"`
    PrioritySupport bool   `bson:"priority_support"`
    FeaturedSlots   int    `bson:"featured_slots"`

    IsActive  bool      `bson:"is_active"`
    SortOrder int       `bson:"sort_order"`
    CreatedAt time.Time `bson:"created_at"`
    UpdatedAt time.Time `bson:"updated_at"`
}
```

**Indexes:**
```javascript
db.packages.createIndex({ code: 1 }, { unique: true })
db.packages.createIndex({ is_active: 1, sort_order: 1 })
```

**Seed Data:**
```javascript
[
    {
        code: "curator", name: "Curator", tier: 0,
        price_monthly: 0, price_yearly: 0,
        credits_per_month: 5, search_boost: false, search_priority: false,
        badge_name: "", priority_support: false, featured_slots: 0
    },
    {
        code: "style", name: "Style", tier: 1,
        price_monthly: 99000, price_yearly: 990000,
        credits_per_month: 50, search_boost: true, search_priority: false,
        badge_name: "Style Curator", priority_support: false, featured_slots: 3
    },
    {
        code: "elite", name: "Elite", tier: 2,
        price_monthly: 249000, price_yearly: 2490000,
        credits_per_month: 200, search_boost: true, search_priority: true,
        badge_name: "Elite", priority_support: true, featured_slots: 10
    }
]
```

---

### 2.7. `subscriptions` — Đăng ký gói

```go
type Subscription struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    UserID    primitive.ObjectID `bson:"user_id"`
    PackageID primitive.ObjectID `bson:"package_id"`

    BillingCycle     string `bson:"billing_cycle"` // monthly | yearly
    PricePaid        int64  `bson:"price_paid"`
    CreditsAllocated int    `bson:"credits_allocated"`
    CreditsUsed      int    `bson:"credits_used"`

    Status       string     `bson:"status"` // active | expired | cancelled
    AutoRenew    bool       `bson:"auto_renew"`
    StartDate    time.Time  `bson:"start_date"`
    EndDate      time.Time  `bson:"end_date"`
    CancelledAt  *time.Time `bson:"cancelled_at,omitempty"`
    CancelReason string     `bson:"cancel_reason,omitempty"`

    CreatedAt time.Time `bson:"created_at"`
    UpdatedAt time.Time `bson:"updated_at"`
}
```

**Indexes:**
```javascript
db.subscriptions.createIndex({ user_id: 1, status: 1 })
db.subscriptions.createIndex({ user_id: 1, end_date: -1 })
db.subscriptions.createIndex({ status: 1, end_date: 1 })        // cron: expiring
db.subscriptions.createIndex({ status: 1, auto_renew: 1, end_date: 1 }) // cron: auto-renew
```

---

### 2.8. `orders`

```go
type Order struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    OrderCode string             `bson:"order_code"`
    Version   int64              `bson:"version"`

    BuyerID   primitive.ObjectID `bson:"buyer_id"`
    SellerID  primitive.ObjectID `bson:"seller_id"`
    ProductID primitive.ObjectID `bson:"product_id"`

    Snapshot OrderSnapshot `bson:"snapshot"` // immutable

    ItemPrice        int64  `bson:"item_price"`
    ShippingFee      int64  `bson:"shipping_fee"`
    PlatformFee      int64  `bson:"platform_fee"`
    TotalPrice       int64  `bson:"total_price"`

    Shipping         ShippingInfo `bson:"shipping"`
    TrackingCode     string       `bson:"tracking_code,omitempty"`
    ShippingProvider string       `bson:"shipping_provider,omitempty"`

    Status       OrderStatus `bson:"status"`
    CancelReason string      `bson:"cancel_reason,omitempty"`
    CancelledBy  string      `bson:"cancelled_by,omitempty"`

    CreatedAt   time.Time  `bson:"created_at"`
    UpdatedAt   time.Time  `bson:"updated_at"`
    ConfirmedAt *time.Time `bson:"confirmed_at,omitempty"`
    ShippedAt   *time.Time `bson:"shipped_at,omitempty"`
    DeliveredAt *time.Time `bson:"delivered_at,omitempty"`
    CancelledAt *time.Time `bson:"cancelled_at,omitempty"`
}

type OrderSnapshot struct {
    Title     string `bson:"title"`
    ImageURL  string `bson:"image_url"`
    Brand     string `bson:"brand"`
    Condition string `bson:"condition"`
    Size      string `bson:"size"`
    Category  string `bson:"category"`
}

type ShippingInfo struct {
    Name     string `bson:"name"`
    Phone    string `bson:"phone"`
    Address  string `bson:"address"`
    Province string `bson:"province"`
    District string `bson:"district"`
    Ward     string `bson:"ward"`
}
```

**Indexes:**
```javascript
db.orders.createIndex({ order_code: 1 }, { unique: true })
db.orders.createIndex({ buyer_id: 1, created_at: -1 })
db.orders.createIndex({ seller_id: 1, created_at: -1 })
db.orders.createIndex({ product_id: 1 })
db.orders.createIndex({ status: 1, created_at: -1 })
db.orders.createIndex({ status: 1, seller_id: 1, created_at: -1 })
```

---

### 2.9. `order_events` — Audit trail

```go
type OrderEvent struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    OrderID    primitive.ObjectID `bson:"order_id"`
    FromStatus string             `bson:"from_status"`
    ToStatus   string             `bson:"to_status"`
    ActorID    primitive.ObjectID `bson:"actor_id"`
    ActorType  string             `bson:"actor_type"`
    Note       string             `bson:"note,omitempty"`
    CreatedAt  time.Time          `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.order_events.createIndex({ order_id: 1, created_at: 1 })
```

---

### 2.10. `favorites` — Yêu thích (♡)

```go
type Favorite struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    UserID    primitive.ObjectID `bson:"user_id"`
    ProductID primitive.ObjectID `bson:"product_id"`
    CreatedAt time.Time          `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.favorites.createIndex({ user_id: 1, product_id: 1 }, { unique: true })
db.favorites.createIndex({ user_id: 1, created_at: -1 })
db.favorites.createIndex({ product_id: 1 })
```

---

### 2.11. `saved_products` — Lưu sản phẩm (bookmark)

```go
type SavedProduct struct {
    ID           primitive.ObjectID  `bson:"_id,omitempty"`
    UserID       primitive.ObjectID  `bson:"user_id"`
    ProductID    primitive.ObjectID  `bson:"product_id"`
    CollectionID *primitive.ObjectID `bson:"collection_id,omitempty"`
    Note         string              `bson:"note,omitempty"`
    CreatedAt    time.Time           `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.saved_products.createIndex({ user_id: 1, product_id: 1 }, { unique: true })
db.saved_products.createIndex({ user_id: 1, created_at: -1 })
db.saved_products.createIndex({ user_id: 1, collection_id: 1, created_at: -1 })
```

---

### 2.12. `saved_collections` — Bộ sưu tập bookmark

```go
type SavedCollection struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    UserID    primitive.ObjectID `bson:"user_id"`
    Name      string             `bson:"name"`
    ItemCount int                `bson:"item_count"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

**Indexes:**
```javascript
db.saved_collections.createIndex({ user_id: 1, created_at: -1 })
```

---

### 2.13. `follows`

```go
type Follow struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    FollowerID primitive.ObjectID `bson:"follower_id"`
    SellerID   primitive.ObjectID `bson:"seller_id"`
    CreatedAt  time.Time          `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.follows.createIndex({ follower_id: 1, seller_id: 1 }, { unique: true })
db.follows.createIndex({ seller_id: 1 })
db.follows.createIndex({ follower_id: 1, created_at: -1 })
```

---

### 2.14. `reviews`

```go
type Review struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    OrderID   primitive.ObjectID `bson:"order_id"`
    ProductID primitive.ObjectID `bson:"product_id"`
    BuyerID   primitive.ObjectID `bson:"buyer_id"`
    SellerID  primitive.ObjectID `bson:"seller_id"`
    Rating    int                `bson:"rating"`
    Comment   string             `bson:"comment,omitempty"`
    Images    []string           `bson:"images,omitempty"` // review ảnh
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

**Indexes:**
```javascript
db.reviews.createIndex({ order_id: 1 }, { unique: true })
db.reviews.createIndex({ seller_id: 1, created_at: -1 })
db.reviews.createIndex({ product_id: 1, created_at: -1 })
db.reviews.createIndex({ buyer_id: 1, created_at: -1 })
db.reviews.createIndex({ seller_id: 1, rating: 1 })
```

---

### 2.15. `credit_wallets`

```go
type CreditWallet struct {
    UserID      primitive.ObjectID `bson:"_id"`
    Balance     int                `bson:"balance"`
    TotalEarned int                `bson:"total_earned"`
    TotalSpent  int                `bson:"total_spent"`
    Version     int64              `bson:"version"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}
```

---

### 2.16. `credit_transactions`

```go
type CreditTransaction struct {
    ID           primitive.ObjectID  `bson:"_id,omitempty"`
    UserID       primitive.ObjectID  `bson:"user_id"`
    Amount       int                 `bson:"amount"`
    Type         string              `bson:"type"` // purchase | unlock | refund | reward | subscription_alloc | expire
    RefType      string              `bson:"ref_type,omitempty"` // product | order | subscription
    RefID        *primitive.ObjectID `bson:"ref_id,omitempty"`
    BalanceAfter int                 `bson:"balance_after"`
    Description  string              `bson:"description"`
    CreatedAt    time.Time           `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.credit_transactions.createIndex({ user_id: 1, created_at: -1 })
db.credit_transactions.createIndex({ ref_type: 1, ref_id: 1 })
db.credit_transactions.createIndex({ type: 1, created_at: -1 })
```

---

### 2.17. `contact_unlocks`

```go
type ContactUnlock struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    BuyerID    primitive.ObjectID `bson:"buyer_id"`
    ProductID  primitive.ObjectID `bson:"product_id"`
    SellerID   primitive.ObjectID `bson:"seller_id"`
    CreditTxID primitive.ObjectID `bson:"credit_tx_id"`
    CreatedAt  time.Time          `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.contact_unlocks.createIndex({ buyer_id: 1, product_id: 1 }, { unique: true })
db.contact_unlocks.createIndex({ seller_id: 1, created_at: -1 })
db.contact_unlocks.createIndex({ buyer_id: 1, created_at: -1 })
```

---

### 2.18. `reports` — Báo cáo vi phạm

```go
type Report struct {
    ID         primitive.ObjectID  `bson:"_id,omitempty"`
    ReporterID primitive.ObjectID  `bson:"reporter_id"`
    TargetType string              `bson:"target_type"` // product | user | review
    TargetID   primitive.ObjectID  `bson:"target_id"`
    Reason     string              `bson:"reason"` // scam | fake | inappropriate | spam | other
    Detail     string              `bson:"detail,omitempty"`
    Status     string              `bson:"status"` // pending | reviewed | resolved | dismissed
    ResolvedBy *primitive.ObjectID `bson:"resolved_by,omitempty"`
    ResolvedAt *time.Time          `bson:"resolved_at,omitempty"`
    Resolution string              `bson:"resolution,omitempty"`
    CreatedAt  time.Time           `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.reports.createIndex({ status: 1, created_at: -1 })
db.reports.createIndex({ target_type: 1, target_id: 1 })
db.reports.createIndex({ reporter_id: 1, created_at: -1 })
```

---

### 2.19. `daily_stats` — Admin Dashboard

```go
type DailyStat struct {
    ID   primitive.ObjectID `bson:"_id,omitempty"`
    Date string             `bson:"date"` // "2026-03-29"

    // Products
    NewProducts         int `bson:"new_products"`
    ApprovedProducts    int `bson:"approved_products"`
    RejectedProducts    int `bson:"rejected_products"`
    SoldProducts        int `bson:"sold_products"`
    TotalActiveProducts int `bson:"total_active_products"`

    // Orders
    NewOrders       int   `bson:"new_orders"`
    CompletedOrders int   `bson:"completed_orders"`
    CancelledOrders int   `bson:"cancelled_orders"`
    TotalGMV        int64 `bson:"total_gmv"`

    // Users
    NewUsers         int `bson:"new_users"`
    ActiveUsers      int `bson:"active_users"`
    NewSubscriptions int `bson:"new_subscriptions"`

    // Credits
    CreditsPurchased int `bson:"credits_purchased"`
    CreditsSpent     int `bson:"credits_spent"`
    UnlockCount      int `bson:"unlock_count"`

    // Revenue
    SubscriptionRevenue int64 `bson:"subscription_revenue"`
    PlatformFeeRevenue  int64 `bson:"platform_fee_revenue"`

    CreatedAt time.Time `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.daily_stats.createIndex({ date: -1 }, { unique: true })
```

---

### 2.20. `seller_stats_snapshots` — Seller performance (monthly)

```go
type SellerStatsSnapshot struct {
    ID              primitive.ObjectID `bson:"_id,omitempty"`
    SellerID        primitive.ObjectID `bson:"seller_id"`
    Period          string             `bson:"period"` // "2026-03"
    ProductsListed  int                `bson:"products_listed"`
    ProductsSold    int                `bson:"products_sold"`
    TotalRevenue    int64              `bson:"total_revenue"`
    OrdersCompleted int                `bson:"orders_completed"`
    OrdersCancelled int                `bson:"orders_cancelled"`
    AvgRating       float64            `bson:"avg_rating"`
    NewReviews      int                `bson:"new_reviews"`
    NewFollowers    int                `bson:"new_followers"`
    ProfileViews    int                `bson:"profile_views"`
    UnlocksReceived int                `bson:"unlocks_received"`
    CreatedAt       time.Time          `bson:"created_at"`
}
```

**Indexes:**
```javascript
db.seller_stats_snapshots.createIndex({ seller_id: 1, period: -1 }, { unique: true })
db.seller_stats_snapshots.createIndex({ period: -1, total_revenue: -1 })
```

---

### 2.21. `hashtags` — Trending tags

```go
type Hashtag struct {
    Tag        string    `bson:"_id"` // "vintage"
    UsageCount int64     `bson:"usage_count"`
    UpdatedAt  time.Time `bson:"updated_at"`
}
```

**Indexes:**
```javascript
db.hashtags.createIndex({ usage_count: -1 })
```

---

## 3. Collection Map

```
┌─────────────────────────────────────────────────────────────────────┐
│                          CORE SERVICE DB                            │
│                                                                     │
│  ┌─── PRODUCT DOMAIN ─────────┐  ┌─── ORDER DOMAIN ──────────┐    │
│  │                             │  │                            │    │
│  │  products                   │  │  orders                    │    │
│  │  product_stats              │  │  order_events              │    │
│  │  product_moderations        │  │                            │    │
│  │  select_products            │  └────────────────────────────┘    │
│  │                             │                                    │
│  └─────────────────────────────┘  ┌─── CREDIT DOMAIN ─────────┐    │
│                                    │                            │    │
│  ┌─── MASTER DATA ────────────┐   │  credit_wallets            │    │
│  │                             │   │  credit_transactions       │    │
│  │  categories                 │   │  contact_unlocks           │    │
│  │  packages                   │   │                            │    │
│  │  hashtags                   │   └────────────────────────────┘    │
│  │                             │                                    │
│  └─────────────────────────────┘  ┌─── SUBSCRIPTION DOMAIN ───┐    │
│                                    │                            │    │
│  ┌─── ENGAGEMENT DOMAIN ──────┐   │  subscriptions             │    │
│  │                             │   │                            │    │
│  │  favorites                  │   └────────────────────────────┘    │
│  │  saved_products             │                                    │
│  │  saved_collections          │  ┌─── ANALYTICS DOMAIN ──────┐    │
│  │  follows                    │  │                            │    │
│  │  reviews                    │  │  daily_stats               │    │
│  │                             │  │  seller_stats_snapshots    │    │
│  └─────────────────────────────┘  │                            │    │
│                                    └────────────────────────────┘    │
│  ┌─── TRUST & SAFETY ─────────┐                                    │
│  │                             │                                    │
│  │  reports                    │                                    │
│  │                             │                                    │
│  └─────────────────────────────┘                                    │
│                                                                     │
│  Total: 21 collections                                              │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 4. Full API List (113 endpoints)

### 4.1. Product APIs (16 public)

```
# CRUD
POST   /api/v1/products                         # Tạo sản phẩm (draft)
GET    /api/v1/products/:id                      # Chi tiết sản phẩm (by ID)
GET    /api/v1/products/slug/:slug               # Chi tiết sản phẩm (by slug)
PUT    /api/v1/products/:id                      # Cập nhật sản phẩm
DELETE /api/v1/products/:id                      # Soft delete

# Seller actions
POST   /api/v1/products/:id/submit              # Gửi duyệt
POST   /api/v1/products/:id/resubmit            # Sửa & gửi lại
POST   /api/v1/products/:id/archive             # Ẩn
POST   /api/v1/products/:id/unarchive           # Hiện lại
GET    /api/v1/products/:id/moderation           # Trạng thái duyệt

# Queries
GET    /api/v1/products/me                       # Sản phẩm của tôi
GET    /api/v1/products/seller/:seller_id        # Sản phẩm của seller (public)
GET    /api/v1/products/feed                     # Feed trang chủ
GET    /api/v1/products/featured                 # Nổi bật
GET    /api/v1/products/select                   # ModaMi Select
GET    /api/v1/products/similar/:id              # Tương tự

# Tracking
POST   /api/v1/products/:id/view                # Track view
```

### 4.2. Search APIs (4)

```
GET    /api/v1/search                            # Full-text search + filters
GET    /api/v1/search/suggest?q=                 # Autocomplete
GET    /api/v1/search/trending                   # Trending terms
GET    /api/v1/search/hashtag/:tag               # Theo hashtag
```

### 4.3. Order APIs (9 public)

```
POST   /api/v1/orders                            # Tạo đơn hàng
GET    /api/v1/orders/:id                        # Chi tiết đơn
GET    /api/v1/orders/my-purchases               # Lịch sử mua
GET    /api/v1/orders/my-sales                   # Đơn bán của tôi
PUT    /api/v1/orders/:id/confirm                # Seller xác nhận
PUT    /api/v1/orders/:id/ship                   # Cập nhật giao hàng
PUT    /api/v1/orders/:id/receive                # Buyer xác nhận nhận
PUT    /api/v1/orders/:id/cancel                 # Huỷ đơn
GET    /api/v1/orders/:id/events                 # Timeline trạng thái
```

### 4.4. Category APIs (3 public)

```
GET    /api/v1/categories                        # Tất cả (tree)
GET    /api/v1/categories/:slug                  # Chi tiết
GET    /api/v1/categories/:slug/children         # Subcategories
```

### 4.5. Favorite APIs (4)

```
POST   /api/v1/favorites/:product_id             # Thêm yêu thích
DELETE /api/v1/favorites/:product_id             # Bỏ yêu thích
GET    /api/v1/favorites                         # Danh sách
GET    /api/v1/favorites/check/:product_id       # Kiểm tra
```

### 4.6. Saved Product APIs (8)

```
POST   /api/v1/saved                             # Lưu sản phẩm
DELETE /api/v1/saved/:product_id                 # Bỏ lưu
GET    /api/v1/saved                             # Danh sách đã lưu
GET    /api/v1/saved/check/:product_id           # Kiểm tra

# Collections
POST   /api/v1/saved/collections                 # Tạo collection
GET    /api/v1/saved/collections                 # List collections
PUT    /api/v1/saved/collections/:id             # Rename
DELETE /api/v1/saved/collections/:id             # Xóa
PUT    /api/v1/saved/:product_id/move/:collection_id
```

### 4.7. Follow APIs (5)

```
POST   /api/v1/follows/:seller_id                # Theo dõi
DELETE /api/v1/follows/:seller_id                # Bỏ theo dõi
GET    /api/v1/follows/following                 # Đang theo dõi
GET    /api/v1/follows/followers                 # Người theo dõi tôi
GET    /api/v1/follows/check/:seller_id          # Kiểm tra
```

### 4.8. Review APIs (4)

```
POST   /api/v1/reviews                           # Tạo đánh giá
GET    /api/v1/reviews/seller/:seller_id         # Theo seller
GET    /api/v1/reviews/product/:product_id       # Theo sản phẩm
GET    /api/v1/reviews/my-reviews                # Tôi đã viết
```

### 4.9. Credit APIs (6)

```
GET    /api/v1/credits/balance                   # Số dư
GET    /api/v1/credits/transactions              # Lịch sử
POST   /api/v1/credits/purchase                  # Mua credit
POST   /api/v1/products/:id/unlock              # Mở khóa liên hệ
GET    /api/v1/unlocks                           # Đã unlock
GET    /api/v1/unlocks/check/:product_id         # Kiểm tra
```

### 4.10. Subscription APIs (7)

```
GET    /api/v1/packages                          # Danh sách gói
GET    /api/v1/packages/:code                    # Chi tiết gói
POST   /api/v1/subscriptions                     # Đăng ký
GET    /api/v1/subscriptions/current             # Gói hiện tại
PUT    /api/v1/subscriptions/current/cancel      # Huỷ
PUT    /api/v1/subscriptions/current/renew       # Bật auto-renew
POST   /api/v1/subscriptions/upgrade             # Nâng cấp
GET    /api/v1/subscriptions/history             # Lịch sử
```

### 4.11. Report APIs (2 public)

```
POST   /api/v1/reports                           # Báo cáo vi phạm
GET    /api/v1/reports/my-reports                 # Báo cáo đã gửi
```

### 4.12. Seller Profile APIs (4)

```
GET    /api/v1/sellers/:id                       # Profile
GET    /api/v1/sellers/:id/products              # Sản phẩm
GET    /api/v1/sellers/:id/reviews               # Đánh giá
GET    /api/v1/sellers/:id/stats                 # Stats công khai
```

### 4.13. Hashtag APIs (3)

```
GET    /api/v1/hashtags/trending                 # Trending
GET    /api/v1/hashtags/:tag/products            # Sản phẩm theo tag
GET    /api/v1/hashtags/suggest?q=               # Autocomplete
```

---

### 4.14. Admin — Product Moderation (9)

```
GET    /api/v1/products/pending            # Queue duyệt
GET    /api/v1/products/:id                # Chi tiết (admin view)
POST   /api/v1/products/:id/approve        # Duyệt
POST   /api/v1/products/:id/reject         # Từ chối
PUT    /api/v1/products/:id/feature        # Đánh dấu nổi bật
PUT    /api/v1/products/:id/unfeature      # Bỏ nổi bật
PUT    /api/v1/products/:id/verify         # Xác thực
POST   /api/v1/products/:id/select         # Đưa vào Select
DELETE /api/v1/products/:id                # Hard delete
```

### 4.15. Admin — Orders (4)

```
GET    /api/v1/orders                      # Tất cả đơn
GET    /api/v1/orders/:id                  # Chi tiết
PUT    /api/v1/orders/:id/force-cancel     # Force huỷ
PUT    /api/v1/orders/:id/force-complete   # Force hoàn thành
```

### 4.16. Admin — Categories (5)

```
POST   /api/v1/categories                  # Tạo
PUT    /api/v1/categories/:id              # Update
PUT    /api/v1/categories/:id/toggle       # Bật/tắt
DELETE /api/v1/categories/:id              # Xóa
PUT    /api/v1/categories/reorder          # Sắp xếp
```

### 4.17. Admin — Packages (3)

```
POST   /api/v1/packages                    # Tạo
PUT    /api/v1/packages/:id                # Update
PUT    /api/v1/packages/:id/toggle         # Bật/tắt
```

### 4.18. Admin — Reports (4)

```
GET    /api/v1/reports                     # Danh sách
GET    /api/v1/reports/:id                 # Chi tiết
PUT    /api/v1/reports/:id/resolve         # Xử lý
PUT    /api/v1/reports/:id/dismiss         # Bỏ qua
```

### 4.19. Admin — Sellers (5)

```
GET    /api/v1/sellers                     # Danh sách
GET    /api/v1/sellers/:id                 # Chi tiết
PUT    /api/v1/sellers/:id/ban             # Ban
PUT    /api/v1/sellers/:id/unban           # Unban
GET    /api/v1/subscriptions               # Tất cả subscriptions
GET    /api/v1/credits/transactions        # Tất cả credit tx
```

### 4.20. Admin — Dashboard (8)

```
GET    /api/v1/dashboard/overview          # Tổng quan hôm nay
GET    /api/v1/dashboard/stats             # Stats theo date range
GET    /api/v1/dashboard/revenue           # Revenue breakdown
GET    /api/v1/dashboard/products          # Product stats
GET    /api/v1/dashboard/categories        # Top categories
GET    /api/v1/dashboard/top-sellers       # Top sellers
GET    /api/v1/dashboard/subscriptions     # Subscription metrics
GET    /api/v1/dashboard/moderation        # Moderation stats
```

---

## 5. API Summary

| Domain | Public | Admin | Total |
|--------|--------|-------|-------|
| Product | 17 | 9 | 26 |
| Search | 4 | 0 | 4 |
| Order | 9 | 4 | 13 |
| Category | 3 | 5 | 8 |
| Favorite | 4 | 0 | 4 |
| Saved | 9 | 0 | 9 |
| Follow | 5 | 0 | 5 |
| Review | 4 | 0 | 4 |
| Credit & Unlock | 6 | 1 | 7 |
| Subscription | 8 | 3 | 11 |
| Report | 2 | 4 | 6 |
| Seller Profile | 4 | 5 | 9 |
| Hashtag | 3 | 0 | 3 |
| Dashboard | 0 | 8 | 8 |
| **TOTAL** | **78** | **39** | **117** |