# Tasks: Go后端学习教程（四轨并行结构）

**Input**: Design documents from `/specs/001-go-backend-learning/`
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓

**Organization**: Tasks are organized by **四轨并行结构** (Four-Track Parallel Structure):
- **轨道一**: 开发教程 (Development Track) - Go语言和后端开发知识
- **轨道二**: 实战教程 (Practice Track) - 项目实战
- **轨道三**: 部署教程 (Deployment Track) - 部署与运维
- **轨道四**: 测试教程 (Testing Track) - 测试与调试

## Format: `[ID] [P?] [Track-Chapter] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Track-Chapter]**: Which track and chapter (e.g., DEV-00, PRAC-01, DEPLOY-02, TEST-03)
- Include exact file paths in descriptions

## Path Conventions

```
docs/
├── 00-introduction/          # 总体介绍
├── 01-development/           # 轨道一：开发教程
├── 02-practice/              # 轨道二：实战教程
├── 03-deployment/            # 轨道三：部署教程
└── 04-testing/               # 轨道四：测试教程

projects/                     # 实战项目源码
├── todo-cli/
├── todo-api/
├── auth-service/
└── fullstack-demo/

tests/                        # 测试示例代码
examples/                     # 独立代码示例
infra/                        # 基础设施配置
```

---

## Phase 1: Setup (Infrastructure & Project Structure)

**Purpose**: Project initialization, development environment, and shared infrastructure

- [ ] T001 Create project root structure with docs/, projects/, tests/, examples/, infra/ directories
- [ ] T002 [P] Create root README.md with project overview and four-track learning path at /README.md
- [ ] T003 [P] Create docs/00-introduction/README.md with tutorial roadmap, prerequisites, and track overview
- [ ] T004 [P] Create infra/docker-compose.yml with PostgreSQL, Redis, RabbitMQ services
- [ ] T005 [P] Create infra/postgres/init.sql with initial database setup scripts
- [ ] T006 [P] Create infra/redis/redis.conf with Redis configuration
- [ ] T007 [P] Create infra/monitoring/prometheus.yml with Prometheus scrape configuration
- [ ] T008 [P] Create infra/monitoring/grafana/ with Grafana dashboards and datasources provisioning
- [ ] T009 [P] Create root go.mod and go.work for workspace management
- [ ] T010 [P] Create examples/go.mod for shared example code

---

## Phase 2: Foundational (Base Project Structures)

**Purpose**: Create base structures for all four tracks and practical projects

- [ ] T011 Create projects/todo-cli/go.mod and basic project structure (cmd/, internal/)
- [ ] T012 [P] Create projects/todo-api/go.mod and layered structure (cmd/, internal/handlers,services,repositories,models/)
- [ ] T013 [P] Create projects/auth-service/go.mod and project structure
- [ ] T014 [P] Create projects/fullstack-demo/go.mod with backend/ and docker-compose.yml
- [ ] T015 [P] Create tests/go.mod and basic test utilities
- [ ] T016 [P] Create docs/01-development/ directory structure for all 12 chapters
- [ ] T017 [P] Create docs/02-practice/ directory structure for all 4 projects
- [ ] T018 [P] Create docs/03-deployment/ directory structure for all 5 chapters
- [ ] T019 [P] Create docs/04-testing/ directory structure for all 5 chapters

**Checkpoint**: Foundation ready - all tracks can be implemented in parallel

---

## Phase 3: 轨道一 - 开发教程 Chapter 00-01 (环境与语法 - P1)

### DEV-00: 环境搭建 (User Story 1)

**Goal**: Enable learners to set up Go development environment and run first program

- [ ] T020 [DEV-00] Create docs/01-development/00-environment/README.md with chapter overview and objectives
- [ ] T021 [P] [DEV-00] Write install-guide.md with Go installation for macOS/Windows/Linux
- [ ] T022 [P] [DEV-00] Write gopath-modules.md explaining GOPATH vs Go Modules
- [ ] T023 [P] [DEV-00] Write vscode-setup.md with VS Code Go plugin configuration
- [ ] T024 [DEV-00] Write first-program.md with Hello World example and explanation
- [ ] T025 [P] [DEV-00] Create examples/syntax/hello/main.go with Hello World example
- [ ] T026 [DEV-00] Add exercises and solutions in docs/01-development/00-environment/

### DEV-01: 语法基础 (User Story 2)

**Goal**: Enable learners to understand Go basics with JS/TS comparisons

- [ ] T027 [DEV-01] Create docs/01-development/01-syntax/README.md with chapter overview
- [ ] T028 [P] [DEV-01] Write 01-variables.md covering var, :=, const with JS comparison
- [ ] T029 [P] [DEV-01] Write 02-basic-types.md covering int, string, bool, float64
- [ ] T030 [P] [DEV-01] Write 03-composite-types.md covering array, slice, map with JS comparison
- [ ] T031 [P] [DEV-01] Write 04-control-flow.md covering if, for, switch
- [ ] T032 [P] [DEV-01] Write 05-functions.md covering params, returns, multiple returns
- [ ] T033 [P] [DEV-01] Write 06-pointers.md covering pointer basics
- [ ] T034 [P] [DEV-01] Write 07-error-handling.md covering error, panic, recover
- [ ] T035 [P] [DEV-01] Write 08-packages.md covering import, go mod
- [ ] T036 [DEV-01] Write js-comparison.md summarizing all Go vs JS/TS differences
- [ ] T037 [P] [DEV-01] Create examples/syntax/variables/main.go with variable examples
- [ ] T038 [P] [DEV-01] Create examples/syntax/types/main.go with type examples
- [ ] T039 [P] [DEV-01] Create examples/syntax/slices/main.go with slice examples
- [ ] T040 [P] [DEV-01] Create examples/syntax/maps/main.go with map examples
- [ ] T041 [P] [DEV-01] Create examples/syntax/control/main.go with control flow examples
- [ ] T042 [P] [DEV-01] Create examples/syntax/functions/main.go with function examples
- [ ] T043 [P] [DEV-01] Create examples/syntax/pointers/main.go with pointer examples
- [ ] T044 [P] [DEV-01] Create examples/syntax/errors/main.go with error handling examples
- [ ] T045 [DEV-01] Add exercises and solutions in docs/01-development/01-syntax/

**Checkpoint**: DEV-00, DEV-01 complete - Learners can set up Go and write basic programs

---

## Phase 4: 轨道一 - 开发教程 Chapter 02-03 (结构体与并发 - P2)

### DEV-02: 结构体与接口 (User Story 3)

**Goal**: Enable learners to design code with structs and interfaces

- [ ] T046 [DEV-02] Create docs/01-development/02-struct-interface/README.md with chapter overview
- [ ] T047 [P] [DEV-02] Write 01-structs.md covering struct definition and initialization
- [ ] T048 [P] [DEV-02] Write 02-struct-tags.md covering json, db tags
- [ ] T049 [P] [DEV-02] Write 03-methods.md covering value vs pointer receivers
- [ ] T050 [P] [DEV-02] Write 04-interfaces.md covering interface definition and implicit implementation
- [ ] T051 [P] [DEV-02] Write 05-type-assertion.md covering type assertions and switches
- [ ] T052 [P] [DEV-02] Write 06-embedding.md covering struct embedding (composition)
- [ ] T053 [P] [DEV-02] Write 07-generics.md covering generics basics (Go 1.18+)
- [ ] T054 [P] [DEV-02] Create examples/syntax/structs/main.go with struct examples
- [ ] T055 [P] [DEV-02] Create examples/syntax/methods/main.go with method examples
- [ ] T056 [P] [DEV-02] Create examples/syntax/interfaces/main.go with interface examples
- [ ] T057 [P] [DEV-02] Create examples/syntax/embedding/main.go with composition examples
- [ ] T058 [DEV-02] Add exercises and solutions in docs/01-development/02-struct-interface/

### DEV-03: 并发编程 (User Story 4)

**Goal**: Enable learners to write concurrent programs with goroutines and channels

- [ ] T059 [DEV-03] Create docs/01-development/03-concurrency/README.md with chapter overview
- [ ] T060 [P] [DEV-03] Write 01-goroutines.md covering goroutine basics with async/await comparison
- [ ] T061 [P] [DEV-03] Write 02-channels.md covering channel creation, send, receive
- [ ] T062 [P] [DEV-03] Write 03-buffered-chan.md covering buffered channels
- [ ] T063 [P] [DEV-03] Write 04-select.md covering select for multiplexing
- [ ] T064 [P] [DEV-03] Write 05-sync-package.md covering WaitGroup, Mutex, RWMutex
- [ ] T065 [P] [DEV-03] Write 06-context.md covering context for cancellation and timeout
- [ ] T066 [P] [DEV-03] Write 07-patterns.md covering worker pool, fan-out/fan-in
- [ ] T067 [P] [DEV-03] Write 08-race-detection.md covering race detection and avoidance
- [ ] T068 [P] [DEV-03] Create examples/concurrency/goroutines/main.go with goroutine examples
- [ ] T069 [P] [DEV-03] Create examples/concurrency/channels/main.go with channel examples
- [ ] T070 [P] [DEV-03] Create examples/concurrency/select/main.go with select examples
- [ ] T071 [P] [DEV-03] Create examples/concurrency/sync/main.go with sync primitives examples
- [ ] T072 [P] [DEV-03] Create examples/concurrency/context/main.go with context examples
- [ ] T073 [P] [DEV-03] Create examples/concurrency/patterns/main.go with pattern examples
- [ ] T074 [DEV-03] Add exercises and solutions in docs/01-development/03-concurrency/

**Checkpoint**: DEV-02, DEV-03 complete - Learners understand Go's OOP and concurrency

---

## Phase 5: 轨道一 - 开发教程 Chapter 04-06 (架构与Web API - P2)

### DEV-04: 后端架构 (User Story 5)

**Goal**: Enable learners to structure Go projects professionally

- [ ] T075 [DEV-04] Create docs/01-development/04-architecture/README.md with chapter overview
- [ ] T076 [P] [DEV-04] Write 01-project-layout.md covering standard Go project structure
- [ ] T077 [P] [DEV-04] Write 02-layered-arch.md covering Handler/Service/Repository pattern
- [ ] T078 [P] [DEV-04] Write 03-dependency-injection.md covering DI patterns
- [ ] T079 [P] [DEV-04] Write 04-config-management.md covering Viper configuration
- [ ] T080 [P] [DEV-04] Write 05-error-design.md covering error handling strategies
- [ ] T081 [P] [DEV-04] Write 06-microservices-intro.md covering microservices concepts and API gateway
- [ ] T082 [P] [DEV-04] Create examples/patterns/layered/ with layered architecture example project
- [ ] T083 [P] [DEV-04] Create examples/patterns/config/main.go with Viper example
- [ ] T084 [DEV-04] Add exercises and solutions in docs/01-development/04-architecture/

### DEV-05: Web API开发 (User Story 6)

**Goal**: Enable learners to build RESTful API services

- [ ] T085 [DEV-05] Create docs/01-development/05-web-api/README.md with chapter overview
- [ ] T086 [P] [DEV-05] Write 01-http-basics.md covering net/http package
- [ ] T087 [P] [DEV-05] Write 02-gin-intro.md covering Gin framework basics
- [ ] T088 [P] [DEV-05] Write 03-routing.md covering path params, query params, route groups
- [ ] T089 [P] [DEV-05] Write 04-request-binding.md covering JSON, Form, URI binding
- [ ] T090 [P] [DEV-05] Write 05-response.md covering JSON responses and error responses
- [ ] T091 [P] [DEV-05] Write 06-validation.md covering request validation with binding tags
- [ ] T092 [P] [DEV-05] Write 07-middleware.md covering logging, recovery, CORS middleware
- [ ] T093 [P] [DEV-05] Write 08-error-handling.md covering API error handling patterns
- [ ] T094 [P] [DEV-05] Write 09-file-upload.md covering file upload handling
- [ ] T095 [P] [DEV-05] Write 10-swagger.md covering Swagger/OpenAPI documentation
- [ ] T096 [P] [DEV-05] Create examples/web/basic-server/main.go with net/http example
- [ ] T097 [P] [DEV-05] Create examples/web/gin-basic/main.go with basic Gin example
- [ ] T098 [P] [DEV-05] Create examples/web/routing/main.go with routing examples
- [ ] T099 [P] [DEV-05] Create examples/web/middleware/main.go with middleware examples
- [ ] T100 [P] [DEV-05] Create examples/web/validation/main.go with validation examples
- [ ] T101 [DEV-05] Add exercises and solutions in docs/01-development/05-web-api/

### DEV-06: 认证授权 (User Story 7)

**Goal**: Enable learners to implement authentication and authorization

- [ ] T102 [DEV-06] Create docs/01-development/06-authentication/README.md with chapter overview
- [ ] T103 [P] [DEV-06] Write 01-auth-intro.md covering authentication vs authorization concepts
- [ ] T104 [P] [DEV-06] Write 02-password.md covering bcrypt password hashing
- [ ] T105 [P] [DEV-06] Write 03-jwt-basics.md covering JWT structure and principles
- [ ] T106 [P] [DEV-06] Write 04-jwt-impl.md covering JWT generation, validation, refresh
- [ ] T107 [P] [DEV-06] Write 05-jwt-middleware.md covering JWT authentication middleware
- [ ] T108 [P] [DEV-06] Write 06-session.md covering session-based authentication
- [ ] T109 [P] [DEV-06] Write 07-jwt-vs-session.md comparing JWT and session approaches
- [ ] T110 [P] [DEV-06] Write 08-oauth2-intro.md covering OAuth2 principles
- [ ] T111 [P] [DEV-06] Write 09-oauth2-github.md covering GitHub OAuth2 integration
- [ ] T112 [P] [DEV-06] Write 10-rbac.md covering role-based access control
- [ ] T113 [P] [DEV-06] Create examples/auth/jwt/main.go with JWT examples
- [ ] T114 [P] [DEV-06] Create examples/auth/middleware/main.go with auth middleware
- [ ] T115 [P] [DEV-06] Create examples/auth/session/main.go with session examples
- [ ] T116 [P] [DEV-06] Create examples/auth/oauth2/main.go with OAuth2 examples
- [ ] T117 [DEV-06] Add exercises and solutions in docs/01-development/06-authentication/

**Checkpoint**: DEV-04, DEV-05, DEV-06 complete - Learners can build secure web APIs

---

## Phase 6: 轨道一 - 开发教程 Chapter 07-08 (数据库 - P2)

### DEV-07: 数据库基础 (User Story 8)

**Goal**: Enable learners to work with PostgreSQL using database/sql

- [ ] T118 [DEV-07] Create docs/01-development/07-database/README.md with chapter overview
- [ ] T119 [P] [DEV-07] Write 01-sql-review.md reviewing SQL basics (CRUD, JOIN)
- [ ] T120 [P] [DEV-07] Write 02-postgres-setup.md covering PostgreSQL setup (local + Docker)
- [ ] T121 [P] [DEV-07] Write 03-database-sql.md covering database/sql standard library
- [ ] T122 [P] [DEV-07] Write 04-pgx-driver.md covering pgx driver usage
- [ ] T123 [P] [DEV-07] Write 05-connection-pool.md covering connection pool configuration
- [ ] T124 [P] [DEV-07] Write 06-prepared-stmt.md covering prepared statements (SQL injection prevention)
- [ ] T125 [P] [DEV-07] Write 07-transactions.md covering transaction handling
- [ ] T126 [P] [DEV-07] Write 08-migrations.md covering database migrations
- [ ] T127 [P] [DEV-07] Create examples/database/connect/main.go with connection example
- [ ] T128 [P] [DEV-07] Create examples/database/query/main.go with query examples
- [ ] T129 [P] [DEV-07] Create examples/database/transaction/main.go with transaction examples
- [ ] T130 [DEV-07] Add exercises and solutions in docs/01-development/07-database/

### DEV-08: GORM框架 (User Story 9)

**Goal**: Enable learners to use GORM ORM effectively

- [ ] T131 [DEV-08] Create docs/01-development/08-gorm/README.md with chapter overview
- [ ] T132 [P] [DEV-08] Write 01-gorm-intro.md covering GORM introduction and setup
- [ ] T133 [P] [DEV-08] Write 02-model-definition.md covering model definition with struct tags
- [ ] T134 [P] [DEV-08] Write 03-crud.md covering Create, Read, Update, Delete operations
- [ ] T135 [P] [DEV-08] Write 04-query-builder.md covering query building with Where, Order, Limit
- [ ] T136 [P] [DEV-08] Write 05-associations.md covering 1:1, 1:N, M:N relationships
- [ ] T137 [P] [DEV-08] Write 06-preload.md covering eager and lazy loading
- [ ] T138 [P] [DEV-08] Write 07-hooks.md covering GORM hooks (BeforeCreate, AfterUpdate, etc.)
- [ ] T139 [P] [DEV-08] Write 08-transactions.md covering GORM transactions
- [ ] T140 [P] [DEV-08] Write 09-raw-sql.md covering raw SQL with GORM
- [ ] T141 [P] [DEV-08] Write 10-best-practices.md covering GORM best practices
- [ ] T142 [P] [DEV-08] Create examples/gorm/models/models.go with User, Post, Comment, Tag models
- [ ] T143 [P] [DEV-08] Create examples/gorm/crud/main.go with CRUD examples
- [ ] T144 [P] [DEV-08] Create examples/gorm/associations/main.go with relationship examples
- [ ] T145 [P] [DEV-08] Create examples/gorm/queries/main.go with advanced query examples
- [ ] T146 [DEV-08] Add exercises and solutions in docs/01-development/08-gorm/

**Checkpoint**: DEV-07, DEV-08 complete - Learners can work with PostgreSQL and GORM

---

## Phase 7: 轨道一 - 开发教程 Chapter 09-11 (缓存/队列/可观测 - P2/P3)

### DEV-09: 缓存策略 (User Story 10)

**Goal**: Enable learners to implement caching strategies

- [ ] T147 [DEV-09] Create docs/01-development/09-cache/README.md with chapter overview
- [ ] T148 [P] [DEV-09] Write 01-cache-intro.md covering caching concepts
- [ ] T149 [P] [DEV-09] Write 02-redis-basics.md covering Redis installation and basic commands
- [ ] T150 [P] [DEV-09] Write 03-go-redis.md covering go-redis client usage
- [ ] T151 [P] [DEV-09] Write 04-cache-patterns.md covering Cache-Aside, Write-Through patterns
- [ ] T152 [P] [DEV-09] Write 05-cache-problems.md covering penetration, breakdown, avalanche
- [ ] T153 [P] [DEV-09] Write 06-local-cache.md covering go-cache for local caching
- [ ] T154 [P] [DEV-09] Write 07-multi-level.md covering multi-level cache architecture
- [ ] T155 [P] [DEV-09] Write 08-cache-consistency.md covering cache-database consistency
- [ ] T156 [P] [DEV-09] Create examples/cache/redis/main.go with Redis examples
- [ ] T157 [P] [DEV-09] Create examples/cache/local/main.go with local cache examples
- [ ] T158 [P] [DEV-09] Create examples/cache/patterns/main.go with cache pattern examples
- [ ] T159 [DEV-09] Add exercises and solutions in docs/01-development/09-cache/

### DEV-10: 消息队列 (User Story 11)

**Goal**: Enable learners to implement async task processing

- [ ] T160 [DEV-10] Create docs/01-development/10-message-queue/README.md with chapter overview
- [ ] T161 [P] [DEV-10] Write 01-mq-intro.md covering message queue concepts
- [ ] T162 [P] [DEV-10] Write 02-rabbitmq-setup.md covering RabbitMQ setup and management UI
- [ ] T163 [P] [DEV-10] Write 03-rabbitmq-go.md covering Go RabbitMQ client usage
- [ ] T164 [P] [DEV-10] Write 04-work-queues.md covering work queue pattern
- [ ] T165 [P] [DEV-10] Write 05-pubsub.md covering publish/subscribe pattern
- [ ] T166 [P] [DEV-10] Write 06-delayed-queue.md covering delayed message queues
- [ ] T167 [P] [DEV-10] Write 07-dead-letter.md covering dead letter queues
- [ ] T168 [P] [DEV-10] Write 08-reliability.md covering message acknowledgment and persistence
- [ ] T169 [P] [DEV-10] Create examples/mq/basic/main.go with basic publish/consume
- [ ] T170 [P] [DEV-10] Create examples/mq/worker/main.go with worker queue example
- [ ] T171 [P] [DEV-10] Create examples/mq/delayed/main.go with delayed queue example
- [ ] T172 [DEV-10] Add exercises and solutions in docs/01-development/10-message-queue/

### DEV-11: 可观测性 (User Story 12)

**Goal**: Enable learners to build observable services

- [ ] T173 [DEV-11] Create docs/01-development/11-observability/README.md with chapter overview
- [ ] T174 [P] [DEV-11] Write 01-observability-intro.md covering three pillars of observability
- [ ] T175 [P] [DEV-11] Write 02-structured-logging.md covering Zap structured logging
- [ ] T176 [P] [DEV-11] Write 03-log-levels.md covering log levels and context
- [ ] T177 [P] [DEV-11] Write 04-prometheus.md covering Prometheus metrics collection
- [ ] T178 [P] [DEV-11] Write 05-custom-metrics.md covering custom metrics (counters, gauges, histograms)
- [ ] T179 [P] [DEV-11] Write 06-grafana.md covering Grafana dashboard setup
- [ ] T180 [P] [DEV-11] Write 07-alerting.md covering alerting rules
- [ ] T181 [P] [DEV-11] Write 08-tracing-intro.md covering distributed tracing concepts
- [ ] T182 [P] [DEV-11] Write 09-jaeger.md covering Jaeger integration
- [ ] T183 [P] [DEV-11] Write 10-elk.md covering ELK log aggregation (optional)
- [ ] T184 [P] [DEV-11] Create examples/observability/logging/main.go with Zap examples
- [ ] T185 [P] [DEV-11] Create examples/observability/metrics/main.go with Prometheus examples
- [ ] T186 [P] [DEV-11] Create examples/observability/tracing/main.go with Jaeger examples
- [ ] T187 [DEV-11] Add exercises and solutions in docs/01-development/11-observability/

**Checkpoint**: 轨道一 complete - All 12 development tutorial chapters done

---

## Phase 8: 轨道二 - 实战教程 (Practice Track)

### PRAC-00: 项目概览

- [ ] T188 [PRAC-00] Create docs/02-practice/00-overview/README.md with project overview and difficulty guide

### PRAC-01: Todo CLI (入门项目)

**Goal**: Build a command-line Todo application

- [ ] T189 [PRAC-01] Create docs/02-practice/01-todo-cli/README.md with project introduction
- [ ] T190 [P] [PRAC-01] Write step-01-init.md covering project initialization
- [ ] T191 [P] [PRAC-01] Write step-02-crud.md covering add/list/complete/delete operations
- [ ] T192 [P] [PRAC-01] Write step-03-storage.md covering JSON file storage
- [ ] T193 [P] [PRAC-01] Write step-04-polish.md covering CLI polish and UX improvements
- [ ] T194 [PRAC-01] Implement projects/todo-cli/main.go with complete CLI application
- [ ] T195 [P] [PRAC-01] Implement projects/todo-cli/internal/todo/todo.go with Todo model and operations
- [ ] T196 [P] [PRAC-01] Implement projects/todo-cli/internal/storage/json.go with JSON file storage

### PRAC-02: Todo API (进阶项目)

**Goal**: Build a RESTful Todo API with database

- [ ] T197 [PRAC-02] Create docs/02-practice/02-todo-api/README.md with project introduction
- [ ] T198 [P] [PRAC-02] Write step-01-setup.md covering project setup with Gin
- [ ] T199 [P] [PRAC-02] Write step-02-routes.md covering CRUD routes implementation
- [ ] T200 [P] [PRAC-02] Write step-03-database.md covering PostgreSQL integration with GORM
- [ ] T201 [P] [PRAC-02] Write step-04-auth.md covering JWT authentication
- [ ] T202 [P] [PRAC-02] Write step-05-cache.md covering Redis caching layer
- [ ] T203 [PRAC-02] Implement projects/todo-api/cmd/api/main.go with server entry point
- [ ] T204 [P] [PRAC-02] Implement projects/todo-api/internal/models/todo.go with Todo model
- [ ] T205 [P] [PRAC-02] Implement projects/todo-api/internal/models/user.go with User model
- [ ] T206 [P] [PRAC-02] Implement projects/todo-api/internal/handlers/todo.go with Todo handlers
- [ ] T207 [P] [PRAC-02] Implement projects/todo-api/internal/handlers/auth.go with Auth handlers
- [ ] T208 [P] [PRAC-02] Implement projects/todo-api/internal/services/todo.go with Todo service
- [ ] T209 [P] [PRAC-02] Implement projects/todo-api/internal/repositories/todo.go with Todo repository
- [ ] T210 [P] [PRAC-02] Implement projects/todo-api/internal/middleware/auth.go with JWT middleware

### PRAC-03: Auth Service (进阶项目)

**Goal**: Build a complete authentication service

- [ ] T211 [PRAC-03] Create docs/02-practice/03-auth-service/README.md with project introduction
- [ ] T212 [P] [PRAC-03] Write step-01-jwt.md covering JWT implementation
- [ ] T213 [P] [PRAC-03] Write step-02-session.md covering session management
- [ ] T214 [P] [PRAC-03] Write step-03-oauth2.md covering OAuth2 integration
- [ ] T215 [P] [PRAC-03] Write step-04-rbac.md covering RBAC implementation
- [ ] T216 [PRAC-03] Implement projects/auth-service/cmd/api/main.go with server entry point
- [ ] T217 [P] [PRAC-03] Implement projects/auth-service/internal/models/ with User, Session, Role models
- [ ] T218 [P] [PRAC-03] Implement projects/auth-service/internal/handlers/ with auth handlers
- [ ] T219 [P] [PRAC-03] Implement projects/auth-service/internal/services/ with auth services

### PRAC-04: Fullstack Demo (高级项目)

**Goal**: Build a complete full-stack application

- [ ] T220 [PRAC-04] Create docs/02-practice/04-fullstack-demo/README.md with project introduction
- [ ] T221 [P] [PRAC-04] Write step-01-planning.md covering system design
- [ ] T222 [P] [PRAC-04] Write step-02-backend.md covering backend implementation
- [ ] T223 [P] [PRAC-04] Write step-03-frontend.md covering frontend integration
- [ ] T224 [P] [PRAC-04] Write step-04-integration.md covering full integration
- [ ] T225 [P] [PRAC-04] Write step-05-deploy.md covering Docker deployment
- [ ] T226 [PRAC-04] Implement projects/fullstack-demo/backend/ with complete backend
- [ ] T227 [P] [PRAC-04] Create projects/fullstack-demo/docker-compose.yml with all services
- [ ] T228 [P] [PRAC-04] Create projects/fullstack-demo/Dockerfile with multi-stage build

**Checkpoint**: 轨道二 complete - All 4 practical projects done

---

## Phase 9: 轨道三 - 部署教程 (Deployment Track - User Story 13)

### DEPLOY-00: 部署概览

- [ ] T229 [DEPLOY-00] Create docs/03-deployment/00-overview/README.md with deployment overview

### DEPLOY-01: 编译与打包

**Goal**: Enable learners to compile Go programs

- [ ] T230 [DEPLOY-01] Create docs/03-deployment/01-compilation/README.md with chapter overview
- [ ] T231 [P] [DEPLOY-01] Write basic-build.md covering go build basics
- [ ] T232 [P] [DEPLOY-01] Write cross-compile.md covering GOOS/GOARCH cross-compilation
- [ ] T233 [P] [DEPLOY-01] Write build-flags.md covering build flags and ldflags
- [ ] T234 [DEPLOY-01] Add exercises and solutions in docs/03-deployment/01-compilation/

### DEPLOY-02: Docker容器化

**Goal**: Enable learners to containerize Go applications

- [ ] T235 [DEPLOY-02] Create docs/03-deployment/02-docker/README.md with chapter overview
- [ ] T236 [P] [DEPLOY-02] Write dockerfile.md covering Dockerfile basics
- [ ] T237 [P] [DEPLOY-02] Write multi-stage.md covering multi-stage builds
- [ ] T238 [P] [DEPLOY-02] Write compose.md covering Docker Compose
- [ ] T239 [DEPLOY-02] Add exercises and solutions in docs/03-deployment/02-docker/

### DEPLOY-03: 配置管理

**Goal**: Enable learners to manage configuration

- [ ] T240 [DEPLOY-03] Create docs/03-deployment/03-config/README.md with chapter overview
- [ ] T241 [P] [DEPLOY-03] Write env-vars.md covering environment variables
- [ ] T242 [P] [DEPLOY-03] Write config-files.md covering config file management
- [ ] T243 [P] [DEPLOY-03] Write secrets.md covering secrets management
- [ ] T244 [DEPLOY-03] Add exercises and solutions in docs/03-deployment/03-config/

### DEPLOY-04: CI/CD流水线

**Goal**: Enable learners to set up CI/CD

- [ ] T245 [DEPLOY-04] Create docs/03-deployment/04-cicd/README.md with chapter overview
- [ ] T246 [P] [DEPLOY-04] Write github-actions.md covering GitHub Actions setup
- [ ] T247 [P] [DEPLOY-04] Write gitlab-ci.md covering GitLab CI setup
- [ ] T248 [P] [DEPLOY-04] Create infra/ci/.github/workflows/test.yml with test workflow
- [ ] T249 [P] [DEPLOY-04] Create infra/ci/.github/workflows/deploy.yml with deploy workflow
- [ ] T250 [DEPLOY-04] Add exercises and solutions in docs/03-deployment/04-cicd/

### DEPLOY-05: 运维监控

**Goal**: Enable learners to operate and monitor services

- [ ] T251 [DEPLOY-05] Create docs/03-deployment/05-operations/README.md with chapter overview
- [ ] T252 [P] [DEPLOY-05] Write health-check.md covering health check endpoints
- [ ] T253 [P] [DEPLOY-05] Write log-management.md covering production log management
- [ ] T254 [P] [DEPLOY-05] Write alerting.md covering alerting setup
- [ ] T255 [DEPLOY-05] Add exercises and solutions in docs/03-deployment/05-operations/

**Checkpoint**: 轨道三 complete - All 5 deployment chapters done

---

## Phase 10: 轨道四 - 测试教程 (Testing Track - User Story 14)

### TEST-00: 测试概览

- [ ] T256 [TEST-00] Create docs/04-testing/00-overview/README.md with testing overview

### TEST-01: 单元测试

**Goal**: Enable learners to write unit tests

- [ ] T257 [TEST-01] Create docs/04-testing/01-unit-testing/README.md with chapter overview
- [ ] T258 [P] [TEST-01] Write basics.md covering go test basics
- [ ] T259 [P] [TEST-01] Write table-driven.md covering table-driven tests
- [ ] T260 [P] [TEST-01] Write mocking.md covering testify mocking
- [ ] T261 [P] [TEST-01] Write coverage.md covering test coverage
- [ ] T262 [P] [TEST-01] Create tests/unit/calculator_test.go with calculator test example
- [ ] T263 [P] [TEST-01] Create tests/unit/user_service_test.go with service test example
- [ ] T264 [P] [TEST-01] Create tests/unit/mock_examples_test.go with mock examples
- [ ] T265 [TEST-01] Add exercises and solutions in docs/04-testing/01-unit-testing/

### TEST-02: 集成测试

**Goal**: Enable learners to write integration tests

- [ ] T266 [TEST-02] Create docs/04-testing/02-integration/README.md with chapter overview
- [ ] T267 [P] [TEST-02] Write database.md covering database integration tests
- [ ] T268 [P] [TEST-02] Write api.md covering API integration tests
- [ ] T269 [P] [TEST-02] Write testcontainers.md covering testcontainers-go
- [ ] T270 [P] [TEST-02] Create tests/integration/database_test.go with DB test example
- [ ] T271 [P] [TEST-02] Create tests/integration/api_test.go with API test example
- [ ] T272 [TEST-02] Add exercises and solutions in docs/04-testing/02-integration/

### TEST-03: E2E测试

**Goal**: Enable learners to write end-to-end tests

- [ ] T273 [TEST-03] Create docs/04-testing/03-e2e/README.md with chapter overview
- [ ] T274 [P] [TEST-03] Write setup.md covering E2E test setup
- [ ] T275 [P] [TEST-03] Write scenarios.md covering test scenarios
- [ ] T276 [P] [TEST-03] Create tests/e2e/user_flow_test.go with user flow test
- [ ] T277 [TEST-03] Add exercises and solutions in docs/04-testing/03-e2e/

### TEST-04: 性能测试

**Goal**: Enable learners to write performance tests

- [ ] T278 [TEST-04] Create docs/04-testing/04-performance/README.md with chapter overview
- [ ] T279 [P] [TEST-04] Write benchmarks.md covering Go benchmarks
- [ ] T280 [P] [TEST-04] Write profiling.md covering pprof profiling
- [ ] T281 [P] [TEST-04] Create tests/benchmark/json_bench_test.go with JSON benchmark
- [ ] T282 [P] [TEST-04] Create tests/benchmark/concurrent_bench_test.go with concurrent benchmark
- [ ] T283 [TEST-04] Add exercises and solutions in docs/04-testing/04-performance/

### TEST-05: 调试技巧

**Goal**: Enable learners to debug Go programs

- [ ] T284 [TEST-05] Create docs/04-testing/05-debugging/README.md with chapter overview
- [ ] T285 [P] [TEST-05] Write compile-errors.md covering compile error debugging
- [ ] T286 [P] [TEST-05] Write runtime-errors.md covering runtime error debugging
- [ ] T287 [P] [TEST-05] Write delve.md covering delve debugger
- [ ] T288 [P] [TEST-05] Write pprof.md covering pprof for performance debugging
- [ ] T289 [P] [TEST-05] Create tests/debugging/bug-01-nil-pointer/ with nil pointer bug scenario
- [ ] T290 [P] [TEST-05] Create tests/debugging/bug-02-race-condition/ with race condition scenario
- [ ] T291 [P] [TEST-05] Create tests/debugging/bug-03-deadlock/ with deadlock scenario
- [ ] T292 [P] [TEST-05] Create tests/debugging/bug-04-memory-leak/ with memory leak scenario
- [ ] T293 [TEST-05] Add exercises and solutions in docs/04-testing/05-debugging/

**Checkpoint**: 轨道四 complete - All 5 testing chapters done

---

## Phase 11: Polish & Final Integration

**Purpose**: Final integration, review, and project completion

- [ ] T294 [P] Complete docs/00-introduction/README.md with full track links and learning paths
- [ ] T295 [P] Update README.md with complete project documentation
- [ ] T296 [P] Create CONTRIBUTING.md with contribution guidelines
- [ ] T297 [P] Review all code examples for consistency and runnability
- [ ] T298 [P] Review all JS/TS comparison tables for accuracy
- [ ] T299 [P] Create infra/monitoring/grafana/dashboards/go-app.json with application dashboard
- [ ] T300 [P] Validate all Docker Compose configurations work together
- [ ] T301 Final review of learning path flow and prerequisites
- [ ] T302 Create comprehensive index of all examples in examples/README.md

---

## Dependencies & Execution Order

### Track Dependencies

```
轨道一 (开发教程) ────────────────────────────────────────────────►
    │
    │ DEV-00,01 完成后
    ▼
轨道二 (实战教程) ────────────────────────────────────────────────►
    │                         │
    │ PRAC-02完成后           │ DEV-02完成后
    ▼                         ▼
轨道三 (部署教程)          轨道四 (测试教程)
```

### Chapter Dependencies Within Track 1

```
DEV-00 (环境) → DEV-01 (语法) → DEV-02 (结构体/接口) → DEV-03 (并发)
                                                            │
                                                            ▼
DEV-04 (架构) → DEV-05 (Web API) → DEV-06 (认证) → DEV-07 (数据库)
                                                            │
                                                            ▼
DEV-08 (GORM) → DEV-09 (缓存) → DEV-10 (消息队列) → DEV-11 (可观测)
```

### Parallel Opportunities

- **Setup Phase**: T002-T010 can all run in parallel
- **Foundational Phase**: T011-T019 can run in parallel
- **Within each chapter**: Files marked [P] can be created in parallel
- **Different tracks**: Once foundations are complete, different tracks can be worked on in parallel

---

## Task Summary

| Phase | Track | Tasks | Parallelizable |
|-------|-------|-------|----------------|
| Phase 1 | Setup | T001-T010 (10) | 8 |
| Phase 2 | Foundational | T011-T019 (9) | 8 |
| Phase 3 | DEV-00, DEV-01 | T020-T045 (26) | 21 |
| Phase 4 | DEV-02, DEV-03 | T046-T074 (29) | 25 |
| Phase 5 | DEV-04, DEV-05, DEV-06 | T075-T117 (43) | 38 |
| Phase 6 | DEV-07, DEV-08 | T118-T146 (29) | 26 |
| Phase 7 | DEV-09, DEV-10, DEV-11 | T147-T187 (41) | 37 |
| Phase 8 | PRAC-01 to PRAC-04 | T188-T228 (41) | 32 |
| Phase 9 | DEPLOY-01 to DEPLOY-05 | T229-T255 (27) | 20 |
| Phase 10 | TEST-01 to TEST-05 | T256-T293 (38) | 30 |
| Phase 11 | Polish | T294-T302 (9) | 7 |

**Total Tasks**: 302
**Parallelizable Tasks**: 252 (83%)

---

## Implementation Strategy

### MVP First (Track 1 Chapters 00-01)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: DEV-00 + DEV-01
4. **STOP and VALIDATE**: Learners can set up Go and write basic programs

### Incremental Delivery

1. **Release 1**: DEV-00 + DEV-01 → Learners can get started with Go
2. **Release 2**: + DEV-02 + DEV-03 → Learners understand Go's unique features
3. **Release 3**: + DEV-04 to DEV-08 + PRAC-01,02 → Learners can build APIs with databases
4. **Release 4**: + DEV-09 to DEV-11 + PRAC-03,04 → Learners can build production-ready services
5. **Release 5**: + 轨道三 + 轨道四 → Complete learning path

### Parallel Track Strategy

With multiple contributors:
1. Team completes Setup + Foundational together
2. Once Phase 3 (DEV-00,01) is complete:
   - Contributor A: Continue Track 1 (DEV-02+)
   - Contributor B: Start Track 2 (PRAC-01)
   - Contributor C: Start Track 4 (TEST-01)
3. After Track 2 PRAC-02:
   - Contributor D: Start Track 3 (DEPLOY-01)

---

## Notes

- **[P]** tasks = different files, no dependencies
- **[Track-Chapter]** label maps task to specific track and chapter
- Each chapter should be independently completable
- All code examples must include expected output (per FR-003)
- Every chapter must include JS/TS comparisons where applicable (per FR-004)
- Commit after each task or logical group
- Stop at any checkpoint to validate independently
