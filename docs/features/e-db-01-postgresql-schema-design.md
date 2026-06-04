---
title: "E-DB-01 PostgreSQL DB 스키마 설계 및 구축"
created: 2026-06-04
updated: '2026-06-04'
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "feature"
status: "planned"
tags: ["feature", "e_db", "e-db-01"]
history:
- "2026-06-04 danya.kim <danya.kim@thundersoft.com>: create feature document from XLSX feature code"
- '2026-06-04 danya.kim <danya.kim@thundersoft.com>: organize features as flat feature-code documents'
---

# E-DB-01 PostgreSQL DB 스키마 설계 및 구축

## Feature Metadata

| 항목 | 값 |
| --- | --- |
| 기능 ID | `E-DB-01` |
| 중분류 | E. DB/저장 시스템 |
| 기능명 | PostgreSQL DB 스키마 설계 및 구축 |
| 우선순위 | P0 |
| 난이도 | 중 |

## 현재 기반 우선 원칙

본 문서는 기능 구현 후보이며, 현재 코드와 `docs/stable/*` 문서를 대체하지 않는다.

현재 구현 기준과 충돌하면 stable 문서와 실제 코드 검증 결과를 우선한다.

## 관제팀 구현 범위

GORM model, AutoMigrate, repository, index, testcontainers 기반 저장소 테스트

## 원본 XLSX 상세 설명

정규화된 RDB 스키마 설계(로봇/임무/사용자/탐지결과), 인덱스 최적화(시공간 쿼리), 파티셔닝 전략(월별), Connection Pooling(PgBouncer), 마이그레이션 관리(Alembic/Flyway).

## 구현 메모

- 구현 전 stable 계약과 현재 코드 구조를 먼저 확인한다.
- 외부 로봇팀 계약에 영향을 주는 변경은 `docs/stable/robot-interface.md`를 함께 갱신한다.
- WebRTC, 센서, 녹화, 로봇 등록 흐름을 바꾸면 Python Mock Robot 기반 E2E 검증을 수행한다.

## 관련 문서

- `docs/stable/data-storage.md`
- `docs/stable/go-gorm-persistence.md`
