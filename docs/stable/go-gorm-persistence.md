---
title: "go-gorm-persistence"
created: 2026-05-22
updated: '2026-05-26'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>", "dhkimxx <dhkimxx@naver.com>"]
type: "guide"
status: "stable"
tags: ["server", "go", "gorm", "persistence", "postgresql"]
source_url: "https://github.com/dhkimxx/ai-agent-skills/blob/main/skills/go-gorm-persistence/SKILL.md"
history:
- "2026-05-22 danya.kim <danya.kim@thundersoft.com>: initial entry"
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: added Go GORM persistence guide from skill reference'
- '2026-05-22 danya.kim <danya.kim@thundersoft.com>: clarified recorder chunk file format scope'
- '2026-05-22 dhkimxx <dhkimxx@naver.com>: documented staged GORM adoption and recording storage object metadata rule'
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/stable lifecycle structure"
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: flattened from harness directory into stable docs"
- '2026-05-26 danya.kim <danya.kim@thundersoft.com>: flattened persistence guide from harness directory into stable docs'
---
# Go GORM Persistence

이 문서는 Go 서버에서 GORM 기반 persistence 코드를 설계하거나 수정할 때 따를 기준을 정의한다.

이 문서는 robot-center 서버 persistence 작성 원칙이며, 구체적인 DB DDL, REST JSON, DataChannel payload, MinIO object key는 정의하지 않는다.

## 적용 범위

적용 대상:

- Go backend persistence code
- GORM entity/model 설계
- repository 구현
- service transaction 경계
- PostgreSQL migration과 테스트
- recorder-worker의 metadata write path
- MinIO file metadata cleanup과 DB row 정합성

비적용 대상:

- HTTP request/response schema
- WebRTC signaling schema
- media track format
- recorder chunk 상세 파일 포맷
- replay manifest schema

## 기본 흐름

1. Entity 중심으로 시작한다.
2. PK, FK, nullable 관계, cascade 정책, unique/index 제약을 먼저 확인한다.
3. 현재 코드베이스의 layer 구조를 따른다.
4. Handler는 입출력 변환, Service는 비즈니스 규칙과 transaction 경계, Repository는 DB query를 담당한다.
5. 기존 repository 생성 방식, `*gorm.DB` 주입 방식, 테스트 DB helper, migration 방식을 먼저 읽는다.
6. 관계 조회가 반복되면 흩어진 `Preload`, `Joins`, filter를 재사용 가능한 GORM scope로 모은다.
7. 새 테이블/컬럼을 추가하면 entity tag, migration, fixture, repository test를 함께 갱신한다.
8. PostgreSQL 전용 기능은 SQLite 대체 검증으로 충분한지 따로 판단한다.

현재 P0 전환 기준:

- DB connection lifecycle은 GORM이 소유한다.
- 기존 raw SQL은 `*gorm.DB.DB()`에서 얻은 `*sql.DB`로 단계적으로 유지할 수 있다.
- PostGIS, JSONB aggregate, row lock처럼 GORM 전환 리스크가 큰 query는 별도 repository test를 추가한 뒤 옮긴다.
- 최종 목표는 service별 repository가 `context.Context`와 `*gorm.DB` transaction을 직접 받는 구조다.

PostgreSQL에서 특히 주의할 기능:

- UUID 타입
- JSON/JSONB query
- row lock
- `FOR UPDATE`
- `SKIP LOCKED`
- regex/order expression
- timezone conversion

## Entity 설계

원칙:

- 공통 PK/timestamp model이 있으면 재사용한다.
- UUID PK는 생성 책임을 한 곳에 둔다.
- 필수 FK는 `not null`과 index를 명시한다.
- 선택 FK는 pointer type을 사용한다.
- 관계는 `foreignKey`, `references`를 명시한다.
- 삭제 정책이 중요한 관계는 `OnDelete:CASCADE` 또는 `OnDelete:SET NULL`을 tag에 드러낸다.
- 단순 JSON/list 저장은 GORM serializer나 `gorm.io/datatypes`를 먼저 검토한다.

주의:

- DB별 타입 제어, validation, query helper가 필요하면 `driver.Valuer`, `sql.Scanner`, `GormDataType`, `GormDBDataType` 구현을 검토한다.
- FK 값을 바꿀 때 이미 로드된 association struct가 있으면 stale association 저장을 피하도록 association field를 비운다.

## Repository 패턴

Repository는 DB query에 집중한다.

Repository에서 하지 않는 일:

- request parsing
- permission 판단
- 외부 API 호출
- queue publish
- object storage cleanup 실행
- HTTP response DTO 조립

규칙:

- Service가 의존할 repository interface와 GORM 구현체를 분리한다.
- GORM 구현체는 global DB getter를 직접 호출하지 않는다.
- `*gorm.DB`는 생성자로 주입한다.
- repository method는 `context.Context`를 받고 `WithContext(ctx)`로 전파한다.
- `Create`/`Save` 직후 관계 포함 응답이 필요하면 저장된 struct를 그대로 반환하지 말고 detail 조회 method로 재조회한다.
- 기본 조회와 관계 포함 조회를 분리한다.
- update method는 의도별로 나눈다.

조회 method 예:

```go
type TeamRepository interface {
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Team, error)
    FindDetailByID(ctx context.Context, id uuid.UUID) (*entity.Team, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TeamStatus) error
}
```

구현체 기준:

```go
type TeamRepositoryImpl struct {
    db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepositoryImpl {
    return &TeamRepositoryImpl{db: db}
}

func (r *TeamRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.Team, error) {
    var team entity.Team
    if err := r.db.WithContext(ctx).First(&team, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &team, nil
}
```

## GORM Scope

관계 로딩과 query composition은 scope로 모은다.

원칙:

- Scope는 `func(db *gorm.DB) *gorm.DB` 형태로 둔다.
- 여러 repository/domain에서 합성할 수 있게 작게 유지한다.
- 관계 로딩 scope 이름은 `WithXxxRelations`를 기본으로 한다.
- 조회 목적이 다르면 `WithXxxListRelations`, `WithXxxDetailRelations`, `WithXxxExportRelations`로 분리한다.
- nested preload는 callback 안에서 필요한 조건을 명시한다.

예:

```go
func WithTeamDetailRelations(db *gorm.DB) *gorm.DB {
    return db.
        Preload("Settings").
        Preload("Members", func(db *gorm.DB) *gorm.DB {
            return db.Where("status = ?", "ACTIVE").Order("created_at asc")
        })
}

func WithTeamListRelations(db *gorm.DB) *gorm.DB {
    return db.Preload("Settings")
}
```

주의:

- `Preload`는 상위 query의 filter/order를 자동 상속하지 않는다.
- join이 들어간 query는 ambiguous column을 피하기 위해 컬럼명을 테이블명까지 붙인다.
- many-side join으로 row가 중복될 수 있으면 primary key `Distinct` 또는 subquery 기준을 먼저 잡는다.

## Pagination

Pagination boilerplate는 공통 helper로 분리한다.

원칙:

- Repository는 filter가 적용된 base query를 만든다.
- 공통 helper가 `Count`, `Offset`, `Limit`을 담당한다.
- count query에는 `Order`, `Offset`, `Limit`, relation preload를 섞지 않는다.
- fetch query에만 relation scope와 정렬을 붙인다.
- page size 상한을 둔다.

흐름:

```text
Repository
-> build filtered base query
-> pagination helper count
-> pagination helper fetch with relation scopes
```

## Update 기준

원칙:

- 전체 entity 갱신과 단일 field/status 변경을 분리한다.
- 일부 컬럼만 바꿀 때는 `Model(...).Where(...).Update(...)` 또는 `Updates(map[string]any{...})`를 우선한다.
- `Save`는 zero value까지 저장한다는 점을 명시적으로 감수할 때만 사용한다.
- 상태 전이는 조건부 update와 `RowsAffected` 확인을 검토한다.

예:

```go
func (r *TeamRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.TeamStatus) error {
    return r.db.WithContext(ctx).
        Model(&entity.Team{}).
        Where("id = ?", id).
        Update("status", status).
        Error
}
```

## Transaction 경계

여러 aggregate를 함께 바꾸거나 검증 후 쓰기가 필요한 작업은 Service가 transaction 경계를 소유한다.

원칙:

- Service에서 `db.Transaction(func(tx *gorm.DB) error { ... })`를 사용한다.
- Transaction 내부 repository 호출은 `tx`로 새 repository 구현체를 만든다.
- Transaction 내부에서 원래 repository의 `r.db`를 사용하지 않는다.
- 단일 repository 안에서 관련 row 정리가 완결되는 삭제 작업은 repository가 transaction을 소유할 수 있다.
- 여러 repository와 도메인 검증이 엮이면 Service가 transaction을 소유한다.

예:

```go
err := db.Transaction(func(tx *gorm.DB) error {
    teamRepository := NewTeamRepository(tx)
    return teamRepository.UpdateStatus(ctx, teamID, entity.TeamStatusActive)
})
```

경쟁 소비나 resource claim 로직은 row lock을 검토한다.

검토 대상:

- `FOR UPDATE`
- `SKIP LOCKED`
- `WHERE status IN (...)`
- 좁은 범위의 deadlock/lock wait retry

## 외부 Side Effect

DB transaction은 rollback 가능한 DB row 변경에 집중한다.

원칙:

- object storage, queue, 외부 API 같은 side effect는 commit 이후 수행한다.
- 외부 cleanup 대상은 DB row 변경 전에 식별한다.
- 실제 cleanup은 commit 이후 처리한다.
- soft hide, hard delete, cascade delete 정책을 먼저 확인한다.
- 조회 조건과 외부 cleanup 범위를 같은 정책에 맞춘다.

robot-center 기준:

- MinIO object 삭제는 DB transaction 안에서 수행하지 않는다.
- recording metadata 변경과 object cleanup은 실패 보상 전략을 분리한다.
- DB row는 MinIO object 본문 자체의 source of truth가 아니라 object metadata의 source of truth다.
- 업로드 완료된 recording file은 `storage_objects`에 idempotent하게 기록한다.
- `recording_chunks.manifest_object_id`는 manifest object row를 참조한다.
- `recording_chunks.metadata`는 기존 UI와 manifest 생성을 위한 요약 캐시로만 사용한다.

## Migration

원칙:

- AutoMigrate는 새 테이블/컬럼 추가에만 제한적으로 사용한다.
- drop, rename, data transform, FK 재작성은 명시 migration을 둔다.
- 새 entity를 추가하면 migration 순서를 확인한다.
- 참조 대상 table이 먼저 생성되어야 한다.
- 운영 데이터가 있는 rename/type 변경은 SQL 또는 Go migration으로 분리한다.
- Go migration은 transaction 안에서 실행한다.
- 기존 schema를 읽어야 하면 현재 entity struct 대신 local struct 또는 `tx.Table("...")`를 사용한다.
- migration은 재실행 가능성을 고려한다.

재실행성 확인 대상:

- duplicate column
- duplicate index
- FK 존재 여부
- 이미 migrate된 row 처리

## 테스트 전략

Repository test는 운영 DB와 같은 PostgreSQL 계열 testcontainers를 우선한다.

원칙:

- 테스트 helper가 container 시작, 임시 DB 생성, migration 적용, cleanup을 캡슐화한다.
- SQLite in-memory는 GORM tag/hook 같은 좁은 테스트에만 제한적으로 사용한다.
- lock, JSON, UUID, regex, FK 동작은 PostgreSQL 기반 테스트를 우선한다.
- 상태 전이 repository는 `RowsAffected == 0`, `ErrRecordNotFound`, 동시성 경로를 함께 검증한다.

## 자주 빠지는 함정

- `Find`는 row가 없어도 error를 내지 않는다.
- 필수 단건 조회는 `First` 또는 `Take`를 쓴다.
- `Preload`는 상위 query 조건을 자동 상속하지 않는다.
- `Save` 전에 로드된 association struct가 있으면 FK 변경 의도와 다른 저장이 발생할 수 있다.
- `Count` query에 preload/order/join이 섞이면 느리거나 잘못된 SQL이 나올 수 있다.
- Transaction 안에서 외부 side effect를 수행하면 rollback 불가능한 불일치가 생긴다.
- Raw SQL을 쓸 때는 PostgreSQL 방언과 identifier quoting을 현재 프로젝트 기준에 맞춘다.
