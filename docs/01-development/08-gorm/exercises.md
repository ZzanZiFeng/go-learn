# GORM 练习 (Exercises)

## 练习 1: 模型定义

### 任务
为电商系统设计以下模型：

1. **Product** 商品
   - ID, UUID, Name, Description, Price, Stock, CategoryID
   - 软删除支持
   - 自动时间戳

2. **Category** 分类
   - ID, Name, Slug, ParentID (自引用)
   - 唯一约束

3. **Review** 评论
   - ID, ProductID, UserID, Rating (1-5), Content
   - 复合索引

### 提示
```go
type Product struct {
    ID          uint           `gorm:"primaryKey"`
    UUID        string         `gorm:"type:uuid;uniqueIndex"`
    Name        string         `gorm:"size:255;not null;index"`
    Description string         `gorm:"type:text"`
    Price       float64        `gorm:"type:decimal(10,2);not null;check:price >= 0"`
    Stock       int            `gorm:"default:0;check:stock >= 0"`
    CategoryID  uint           `gorm:"index"`
    Category    *Category
    Reviews     []Review       `gorm:"foreignKey:ProductID"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// 实现 BeforeCreate 自动生成 UUID
```

---

## 练习 2: CRUD 服务

### 任务
实现 ProductService，包含以下方法：

1. `Create(product *Product) error`
2. `GetByID(id uint) (*Product, error)`
3. `GetByUUID(uuid string) (*Product, error)`
4. `Update(id uint, updates map[string]interface{}) error`
5. `Delete(id uint) error`
6. `List(page, pageSize int, categoryID *uint) ([]Product, int64, error)`

### 要求
- 正确处理 ErrRecordNotFound
- 返回业务错误
- 支持按分类过滤

### 参考结构
```go
type ProductService struct {
    db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
    return &ProductService{db: db}
}

func (s *ProductService) GetByID(id uint) (*Product, error) {
    var product Product
    err := s.db.First(&product, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrProductNotFound
        }
        return nil, err
    }
    return &product, nil
}
```

---

## 练习 3: 关联查询

### 任务
实现以下查询方法：

1. **GetProductWithCategory** - 获取商品及其分类
2. **GetProductWithReviews** - 获取商品及最新 10 条评论
3. **GetCategoryWithProducts** - 获取分类及其下所有商品
4. **GetProductDetails** - 获取商品完整信息（分类、评论、评论用户）

### 要求
- 使用 Preload 避免 N+1
- 评论需要关联用户信息
- 只加载必要字段

### 示例
```go
func (s *ProductService) GetProductDetails(id uint) (*Product, error) {
    var product Product
    err := s.db.
        Preload("Category").
        Preload("Reviews", func(db *gorm.DB) *gorm.DB {
            return db.Order("created_at DESC").Limit(10)
        }).
        Preload("Reviews.User", func(db *gorm.DB) *gorm.DB {
            return db.Select("id", "name", "avatar")
        }).
        First(&product, id).Error
    // ...
}
```

---

## 练习 4: 复杂查询

### 任务
实现商品搜索功能：

```go
type ProductSearchParams struct {
    Keyword    string
    CategoryID *uint
    MinPrice   *float64
    MaxPrice   *float64
    MinRating  *float64
    InStock    *bool
    SortBy     string   // name, price, rating, created_at
    SortOrder  string   // asc, desc
    Page       int
    PageSize   int
}

func (s *ProductService) Search(params ProductSearchParams) (*SearchResult, error)

type SearchResult struct {
    Products   []ProductDTO
    Total      int64
    Page       int
    PageSize   int
    TotalPages int
}
```

### 要求
- 支持关键词搜索（名称和描述）
- 支持价格范围过滤
- 支持平均评分过滤
- 支持库存状态过滤
- 动态排序（验证排序字段）
- 分页

### 提示
```go
// 使用 Scopes 组织查询条件
func FilterByPriceRange(min, max *float64) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if min != nil {
            db = db.Where("price >= ?", *min)
        }
        if max != nil {
            db = db.Where("price <= ?", *max)
        }
        return db
    }
}

// 使用子查询计算平均评分
subQuery := db.Model(&Review{}).
    Select("product_id, AVG(rating) as avg_rating").
    Group("product_id")
```

---

## 练习 5: 事务操作

### 任务
实现订单创建功能（需要事务）：

```go
type CreateOrderRequest struct {
    UserID    uint
    Items     []OrderItemRequest
    AddressID uint
}

type OrderItemRequest struct {
    ProductID uint
    Quantity  int
}

func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
```

### 要求
1. 验证商品存在且有库存
2. 计算订单总价
3. 扣减商品库存（乐观锁）
4. 创建订单和订单项
5. 任何失败都回滚

### 提示
```go
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    var order Order

    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. 锁定并验证商品
        for _, item := range req.Items {
            var product Product
            if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
                First(&product, item.ProductID).Error; err != nil {
                return fmt.Errorf("product %d not found", item.ProductID)
            }

            if product.Stock < item.Quantity {
                return fmt.Errorf("insufficient stock for product %d", item.ProductID)
            }
        }

        // 2. 创建订单
        // 3. 创建订单项
        // 4. 扣减库存

        return nil
    })

    return &order, err
}
```

---

## 练习 6: 钩子实现

### 任务
为模型实现以下钩子：

1. **Product.BeforeCreate** - 自动生成 UUID 和 Slug
2. **Product.AfterCreate** - 记录创建日志
3. **Order.BeforeCreate** - 自动生成订单号
4. **Order.AfterUpdate** - 状态变更通知
5. **Review.BeforeCreate** - 验证评分范围

### 示例
```go
func (p *Product) BeforeCreate(tx *gorm.DB) error {
    // 生成 UUID
    if p.UUID == "" {
        p.UUID = uuid.New().String()
    }

    // 生成 Slug
    if p.Slug == "" {
        p.Slug = generateSlug(p.Name)
    }

    return nil
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
    // 生成订单号: ORD-YYYYMMDD-XXXXX
    o.OrderNumber = fmt.Sprintf("ORD-%s-%05d",
        time.Now().Format("20060102"),
        rand.Intn(100000))
    return nil
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
    if r.Rating < 1 || r.Rating > 5 {
        return errors.New("rating must be between 1 and 5")
    }
    return nil
}
```

---

## 综合项目: 博客 API 数据层

### 项目描述
使用 GORM 实现博客系统的完整数据访问层。

### 模型设计

```go
// User 用户
type User struct {
    ID        uint           `gorm:"primaryKey"`
    UUID      string         `gorm:"type:uuid;uniqueIndex"`
    Username  string         `gorm:"size:50;uniqueIndex;not null"`
    Email     string         `gorm:"size:255;uniqueIndex;not null"`
    Password  string         `gorm:"size:255;not null"`
    Name      string         `gorm:"size:100"`
    Bio       string         `gorm:"type:text"`
    Avatar    string         `gorm:"size:500"`
    Role      string         `gorm:"size:20;default:'user'"`
    Posts     []Post         `gorm:"foreignKey:AuthorID"`
    Comments  []Comment      `gorm:"foreignKey:UserID"`
    Followers []User         `gorm:"many2many:user_follows;joinForeignKey:FollowingID;joinReferences:UserID"`
    Following []User         `gorm:"many2many:user_follows;joinForeignKey:UserID;joinReferences:FollowingID"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Post 文章
type Post struct {
    ID          uint           `gorm:"primaryKey"`
    UUID        string         `gorm:"type:uuid;uniqueIndex"`
    Title       string         `gorm:"size:255;not null"`
    Slug        string         `gorm:"size:255;uniqueIndex"`
    Content     string         `gorm:"type:text"`
    Excerpt     string         `gorm:"size:500"`
    Status      string         `gorm:"size:20;default:'draft';index"`
    AuthorID    uint           `gorm:"index;not null"`
    Author      *User
    Tags        []Tag          `gorm:"many2many:post_tags;"`
    Comments    []Comment      `gorm:"foreignKey:PostID"`
    ViewCount   int            `gorm:"default:0"`
    PublishedAt *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// Tag 标签
type Tag struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"size:50;uniqueIndex;not null"`
    Slug      string `gorm:"size:50;uniqueIndex;not null"`
    Posts     []Post `gorm:"many2many:post_tags;"`
    CreatedAt time.Time
}

// Comment 评论
type Comment struct {
    ID        uint           `gorm:"primaryKey"`
    Content   string         `gorm:"type:text;not null"`
    PostID    uint           `gorm:"index;not null"`
    Post      *Post
    UserID    uint           `gorm:"index;not null"`
    User      *User
    ParentID  *uint          `gorm:"index"`
    Parent    *Comment
    Replies   []Comment      `gorm:"foreignKey:ParentID"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

### Repository 实现

```go
// UserRepository
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uint) (*User, error)
    FindByUsername(ctx context.Context, username string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uint) error
    Follow(ctx context.Context, userID, followingID uint) error
    Unfollow(ctx context.Context, userID, followingID uint) error
    GetFollowers(ctx context.Context, userID uint, page, pageSize int) ([]User, int64, error)
    GetFollowing(ctx context.Context, userID uint, page, pageSize int) ([]User, int64, error)
}

// PostRepository
type PostRepository interface {
    Create(ctx context.Context, post *Post) error
    FindByID(ctx context.Context, id uint) (*Post, error)
    FindBySlug(ctx context.Context, slug string) (*Post, error)
    Update(ctx context.Context, post *Post) error
    Delete(ctx context.Context, id uint) error
    Publish(ctx context.Context, id uint) error
    List(ctx context.Context, opts PostListOptions) ([]Post, int64, error)
    ListByAuthor(ctx context.Context, authorID uint, page, pageSize int) ([]Post, int64, error)
    ListByTag(ctx context.Context, tagSlug string, page, pageSize int) ([]Post, int64, error)
    Search(ctx context.Context, query string, page, pageSize int) ([]Post, int64, error)
    IncrementViewCount(ctx context.Context, id uint) error
}

// TagRepository
type TagRepository interface {
    Create(ctx context.Context, tag *Tag) error
    FindByID(ctx context.Context, id uint) (*Tag, error)
    FindBySlug(ctx context.Context, slug string) (*Tag, error)
    FindOrCreate(ctx context.Context, name string) (*Tag, error)
    List(ctx context.Context) ([]Tag, error)
    GetPopularTags(ctx context.Context, limit int) ([]TagWithCount, error)
}

// CommentRepository
type CommentRepository interface {
    Create(ctx context.Context, comment *Comment) error
    FindByID(ctx context.Context, id uint) (*Comment, error)
    Update(ctx context.Context, comment *Comment) error
    Delete(ctx context.Context, id uint) error
    ListByPost(ctx context.Context, postID uint, page, pageSize int) ([]Comment, int64, error)
}
```

### 验收标准

1. **模型正确**
   - 所有关联关系正确定义
   - 索引和约束合理
   - 软删除支持

2. **CRUD 完整**
   - 所有基本操作实现
   - 错误处理正确
   - 返回业务错误

3. **查询优化**
   - 使用 Preload 避免 N+1
   - 只查询需要的字段
   - 分页实现正确

4. **事务正确**
   - 关注/取消关注使用事务
   - 文章发布使用事务

5. **测试覆盖**
   - 每个 Repository 有单元测试
   - 使用 SQLite 内存数据库

### 提交要求
- 完整的代码实现
- 单元测试
- 示例使用代码
