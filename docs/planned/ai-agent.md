---
title: "ai-agent"
created: 2026-05-26
updated: 2026-05-26
author: "danya.kim <danya.kim@thundersoft.com>"
editors: ["danya.kim <danya.kim@thundersoft.com>"]
type: "design"
status: "planned"
tags: ["ai-agent", "eino", "llm", "sop", "control"]
history:
- "2026-05-26 danya.kim <danya.kim@thundersoft.com>: moved into docs/planned lifecycle structure"
---

# AI Agent

> Canonical document for Go/Eino based AI Agent requirements.

## 문서 목적

관제센터 AI Agent 기능을 Go 기반 Eino로 구현하기 위한 기능, 입력 데이터, 출력 포맷, 도구, 실행 흐름, 저장 정책을 정의한다.

본 문서는 PoC 기준이며, AI Agent는 SOP 근거를 바탕으로 제어 명령 초안을 생성할 수 있다. 단, 명령 실행은 반드시 관제요원 또는 지휘관 승인 후 서버가 수행한다.

## Eino 적용 기준

공식 Eino 문서 기준으로 Eino는 Go 기반 LLM 애플리케이션 프레임워크이며, ChatModel, Tool, Retriever, Chain/Graph/Workflow, Agent 패턴을 제공한다.

AI Web에서는 다음 기준으로 사용한다.

- 반복적이고 예측 가능한 기능: Eino Graph 또는 Chain
- 관제요원의 자유 질의응답: Eino Agent
- DB/API 조회: Eino Tool
- SOP/매뉴얼 검색: Retriever 또는 별도 SOP 검색 Tool
- 위험도 계산처럼 규칙이 있는 처리: LLM 단독 판단이 아니라 deterministic function 또는 Graph node

참고:

- Eino Overview: https://www.cloudwego.io/docs/eino/overview/
- Eino Chain/Graph/Workflow: https://www.cloudwego.io/docs/eino/core_modules/chain_and_graph_orchestration/
- Eino Agent with Tools: https://www.cloudwego.io/docs/eino/quick_start/agent_llm_with_tools/
- Agent or Graph: https://www.cloudwego.io/docs/eino/overview/graph_or_agent/

## AI Agent 목표

### PoC 목표

- 임무 현재 상황을 자연어로 요약한다.
- 탐지, 센서, 로봇 상태, 통신 상태를 종합해 위험도를 설명한다.
- 산악조난, 붕괴현장, 지하시설 시나리오에 맞는 SOP 후보를 제안한다.
- SOP 근거가 있는 경우 관제요원이 승인할 수 있는 제어 명령 초안을 생성한다.
- 관제요원 질문에 대해 DB에 저장된 근거 기반으로 답변한다.
- 이벤트 타임라인을 기반으로 임무 종료 보고서 초안을 생성한다.

### 비목표

- E-Stop, Return-to-Home, PTZ, Waypoint 명령을 Agent가 사람 승인 없이 자동 실행하지 않는다.
- SOP 근거가 없는 제어 명령 초안을 생성하지 않는다.
- 근거 없는 탐지 결과, 센서값, 위치를 생성하지 않는다.
- 소방 SOP 전문 지식을 임의로 만들어내지 않는다.
- 실시간 제어 루프에 Agent를 넣지 않는다.

## 사용자 기능

### 1. 상황 요약

관제 화면 우측 패널에 현재 임무 상황을 요약한다.

입력:

- mission
- latest telemetry
- latest sensor readings
- recent detection results
- recent events
- module status
- scenario type

출력:

- 현재 상황 한 줄 요약
- 주요 이벤트 3~5개
- 위험 신호
- 데이터 신선도
- 확인 필요 항목

### 2. 위험도 분석

센서, 탐지, 로봇 상태, 통신 상태를 종합해 위험도와 근거를 제공한다.

위험도:

- `normal`
- `caution`
- `warning`
- `critical`

주요 위험 근거:

- 가스 농도 상승
- 산소 농도 저하
- 고온 또는 온도 급상승
- 인명/열원 탐지
- 통신 품질 저하
- 배터리 부족
- 로봇 고착/전도/모듈 장애
- 위치 신뢰도 저하

### 3. SOP 매핑 및 대응 제안

시나리오와 이벤트 유형을 기준으로 SOP 후보를 제안한다.

PoC에서는 실제 소방 SOP 전체 RAG를 구축하지 않고, 시나리오별 seed SOP rule set으로 시작한다.

출력은 반드시 다음을 포함한다.

- SOP 후보명
- 적용 이유
- 관련 이벤트/센서 근거
- 추천 대응
- 사람이 승인해야 하는 항목
- confidence

### 4. SOP 기반 제어 명령 초안

AI Agent는 SOP 추천 결과에 근거해 제어 명령 초안을 생성할 수 있다.

지원 모드:

- `advisory`: Agent가 대응 권고만 생성한다.
- `approval_required`: Agent가 명령 초안을 생성하고, 사람이 승인하면 서버가 실행한다.

PoC 기준은 `approval_required`이다.

명령 초안 후보:

- `return_to_home`: 통신 저하, 배터리 부족, 가스 위험 등 복귀 권고 조건
- `ptz`: 인명/열원 후보 확대 관찰
- `waypoint`: 추가 탐색 후보 지점 제안
- `estop`: 고착, 전도, 충돌 위험 등 긴급 정지 권고

명령 초안에는 반드시 SOP 근거와 관련 이벤트/센서 ID를 포함한다.

### 5. 관제 질의응답

관제요원 또는 지휘관이 자연어로 질문하면 Agent가 관련 도구를 호출해 답변한다.

예시 질문:

- 현재 가장 위험한 구역은 어디야?
- 최근 5분간 가스 수치 변화 알려줘.
- 인명 탐지 후보가 몇 건이야?
- 통신 상태가 나빠진 시점이 언제야?
- 지금 구조대원 진입해도 되는 근거가 있어?
- 이 상황에 맞는 SOP 후보는 뭐야?

### 6. 보고서 초안

임무 종료 후 이벤트, 탐지, 센서 추세, 제어 명령 이력을 기반으로 보고서 초안을 생성한다.

PoC 출력:

- 임무 개요
- 주요 타임라인
- 탐지 결과 요약
- 위험 이벤트 요약
- 제어 명령 이력
- SOP 제안 이력
- 후속 조치 제안

## Eino 컴포넌트 설계

### Go 패키지 후보

```text
internal/agent
  service.go
  prompts.go
  graphs.go
  tools.go
  guardrails.go
  schemas.go
  repository.go
```

### AgentService

역할:

- 상황 요약 실행
- 위험도 분석 실행
- SOP 추천 실행
- 관제 Q&A 실행
- 보고서 초안 생성
- agent run, tool call, output 저장

### Eino 구성

```text
ChatModel
  + ToolsNode
  + Graph/Chain
  + Guardrail
  + Output Parser
```

권장 구조:

```text
Deterministic Graphs
  - MissionContextGraph
  - RiskAssessmentGraph
  - SopRecommendationGraph
  - ControlDraftGraph
  - ReportDraftGraph

ControlCenterAgent
  - ChatModel
  - Tools
  - 필요한 경우 위 Graph를 Tool로 호출
```

Eino 공식 문서의 권장처럼, 안정적 백엔드 기능은 Graph로 만들고 Agent가 필요할 때 Graph를 Tool처럼 호출하는 구조를 우선한다.

## Graph 명세

### MissionContextGraph

목적:

- 현재 임무 데이터를 Agent 입력에 적합한 구조로 압축한다.

입력:

```json
{
  "missionId": "uuid",
  "timeWindowMinutes": 10
}
```

처리:

1. mission 조회
2. latest telemetry 조회
3. latest sensor readings 조회
4. latest module status 조회
5. recent events 조회
6. recent detections 조회
7. 데이터 신선도 계산

출력:

```json
{
  "mission": {},
  "robotStatus": {},
  "sensorSummary": {},
  "moduleSummary": {},
  "recentEvents": [],
  "recentDetections": [],
  "dataFreshness": {}
}
```

### RiskAssessmentGraph

목적:

- 규칙 기반 위험도와 LLM 설명을 결합한다.

처리:

1. 센서 임계값 평가
2. 탐지 이벤트 평가
3. 통신/배터리/모듈 상태 평가
4. 시나리오별 가중치 적용
5. LLM으로 사람이 읽기 쉬운 설명 생성

출력:

```json
{
  "riskLevel": "warning",
  "score": 72,
  "reasons": [
    {
      "type": "gas_risk",
      "severity": "warning",
      "evidenceRef": "event:..."
    }
  ],
  "summary": "CO 농도 상승과 통신 저하가 동시에 발생해 접근 주의가 필요합니다."
}
```

### SopRecommendationGraph

목적:

- 시나리오와 위험 이벤트에 맞는 SOP 후보를 제안한다.

처리:

1. mission type 확인
2. event type 확인
3. seed SOP rule set 조회
4. 관련 SOP 후보 정렬
5. LLM으로 대응 제안 문장화

출력:

```json
{
  "recommendations": [
    {
      "sopId": "collapse-gas-001",
      "title": "붕괴 현장 유해가스 의심 대응",
      "confidence": 0.82,
      "reason": "CO 농도 상승과 산소 농도 저하가 동시에 관측됨",
      "recommendedActions": [
        "구조대원 진입 전 환기 가능 여부 확인",
        "로봇 추가 탐색 유지",
        "고위험 구역 접근 제한"
      ],
      "requiresHumanApproval": true,
      "evidenceRefs": [
        "sensor_reading:...",
        "event:..."
      ]
    }
  ]
}
```

### ControlDraftGraph

목적:

- SOP 추천과 현재 임무 상태를 기반으로 사람이 승인할 제어 명령 초안을 생성한다.

처리:

1. SOP 추천 결과 확인
2. 관련 이벤트/센서 근거 확인
3. 명령 가능 범위 확인
4. 금지 조건 확인
5. command draft 생성

출력:

```json
{
  "controlDrafts": [
    {
      "draftId": "draft-001",
      "commandType": "return_to_home",
      "priority": "high",
      "reason": "배터리 부족과 통신 저하가 동시에 발생해 SOP 기준 복귀 검토가 필요합니다.",
      "sopId": "common-return-001",
      "requiresApproval": true,
      "approvalRole": "operator",
      "commandPayload": {
        "mode": "safe_return"
      },
      "evidenceRefs": [
        "event:...",
        "telemetry_snapshot:..."
      ],
      "expiresAt": "2026-05-12T05:01:00Z"
    }
  ]
}
```

### ReportDraftGraph

목적:

- 임무 종료 후 보고서 초안을 생성한다.

처리:

1. mission 정보 조회
2. 이벤트 타임라인 조회
3. 탐지 결과 요약
4. 센서 위험 구간 요약
5. 제어 명령 이력 요약
6. SOP 추천 이력 요약

출력:

```json
{
  "title": "지하시설 탐색 임무 보고서 초안",
  "sections": [
    {
      "heading": "주요 타임라인",
      "content": "..."
    }
  ],
  "evidenceRefs": [
    "event:...",
    "control_command:..."
  ]
}
```

## Tool 명세

### Read-only Tools

Agent가 자유 질의응답에서 호출할 수 있는 조회 도구이다.

| Tool | 목적 |
| --- | --- |
| `get_mission_snapshot` | 임무 기본 정보, 현재 상태 조회 |
| `get_recent_events` | 최근 이벤트 조회 |
| `get_latest_robot_status` | 최신 telemetry 조회 |
| `get_sensor_trend` | 센서 추세 조회 |
| `get_detection_results` | 탐지 결과 조회 |
| `get_module_status` | PTZ, Jetson, 5G, 음성 모듈 상태 조회 |
| `get_storage_object_metadata` | 스냅샷, point cloud, map object metadata 조회 |
| `search_sop_candidates` | seed SOP 또는 RAG 후보 검색 |
| `get_control_command_history` | 제어 명령 이력 조회 |

### Write Tools

Agent가 직접 외부 세계를 제어하지 않도록 write tool은 승인 흐름을 전제로 제한한다.

| Tool | 목적 | 승인 |
| --- | --- | --- |
| `create_situation_summary` | 상황 요약 저장 | 자동 가능 |
| `create_sop_recommendation` | SOP 추천 저장 | 자동 가능 |
| `create_control_draft` | SOP 근거 기반 제어 명령 초안 저장 | 자동 가능 |
| `create_report_draft` | 보고서 초안 저장 | 자동 가능 |
| `acknowledge_agent_output` | 사용자가 Agent 결과 확인 | 사용자 액션 필요 |
| `approve_control_draft` | 사람이 명령 초안을 승인 | 사용자 액션 필요 |

Agent 직접 실행 금지:

- `send_estop`
- `send_ptz`
- `send_waypoint`
- `send_return_to_home`

AI Agent는 제어 명령을 직접 실행하지 않는다. 필요한 경우 SOP 근거가 포함된 제어 명령 초안을 생성하고, 사람이 승인하면 기존 Control Service가 명령을 실행한다.

## 제어 승인 흐름

```text
Agent creates control draft
  -> Go Server stores draft
  -> Browser shows draft with SOP evidence
  -> Operator/Commander approves or rejects
  -> Go Server checks RBAC and draft validity
  -> Control Service creates control_command
  -> command DataChannel sends to Robot Peer
  -> controlAck updates command status
```

승인 전 검증:

- draft가 만료되지 않았는지 확인
- 승인자 role이 충분한지 확인
- draft의 missionId, robotId가 현재 활성 임무와 일치하는지 확인
- 관련 SOP/evidenceRefs가 존재하는지 확인
- 동일 명령 중복 실행 여부 확인
- E-Stop은 별도 긴급 승인 UX 또는 즉시 사용자 버튼으로 실행할 수 있지만, Agent가 자동 실행하지 않는다.

## 출력 JSON 계약

### SituationSummary

```json
{
  "summaryId": "uuid",
  "missionId": "uuid",
  "riskLevel": "warning",
  "headline": "지하시설 내부에서 통신 저하와 CO 농도 상승이 관측되었습니다.",
  "keyFindings": [
    "최근 5분간 CO 농도가 상승했습니다.",
    "로봇 위치 신뢰도가 낮아졌습니다."
  ],
  "recommendedChecks": [
    "가스 센서 캘리브레이션 상태 확인",
    "Return-to-Home 가능 여부 확인"
  ],
  "evidenceRefs": [
    "sensor_reading:...",
    "event:..."
  ],
  "generatedAt": "2026-05-12T05:00:00Z"
}
```

### AgentAnswer

```json
{
  "answerId": "uuid",
  "missionId": "uuid",
  "question": "최근 5분간 가장 위험한 이벤트가 뭐야?",
  "answer": "최근 5분 기준 가장 위험한 이벤트는 CO 농도 상승입니다.",
  "riskLevel": "warning",
  "evidenceRefs": [
    "event:...",
    "sensor_reading:..."
  ],
  "limitations": [
    "Robot팀 센서 임계값이 최종 확정되지 않았습니다."
  ],
  "generatedAt": "2026-05-12T05:00:00Z"
}
```

## Guardrail 정책

### 근거 제한

Agent는 다음 데이터만 근거로 사용한다.

- PostgreSQL에 저장된 임무/센서/탐지/이벤트/제어/모듈 상태
- MinIO object metadata
- seed SOP rule set 또는 승인된 SOP 문서
- 사용자가 현재 대화에서 제공한 정보

### 금지 응답

- 확인되지 않은 생존자 수 단정
- 확인되지 않은 위치 단정
- 구조대원 진입 가능 여부를 최종 판단처럼 표현
- 로봇 제어 명령을 사람 승인 없이 자동 실행
- SOP 근거 없는 제어 명령 초안 생성
- 센서 임계값을 임의로 생성
- SOP 원문에 없는 절차를 실제 SOP처럼 표현

### 필수 표현

위험 권고에는 반드시 다음을 포함한다.

- 근거 이벤트 또는 센서 ID
- confidence
- 사람이 승인해야 하는 조치 또는 제어 명령 초안
- 데이터 신선도 또는 한계

## API 후보

### 상황 요약 생성

```http
POST /api/missions/{missionId}/agent/situation-summary
```

Request:

```json
{
  "timeWindowMinutes": 10,
  "trigger": "manual"
}
```

Response:

```json
{
  "agentRunId": "uuid",
  "summaryId": "uuid",
  "status": "completed"
}
```

### SOP 추천 생성

```http
POST /api/missions/{missionId}/agent/sop-recommendations
```

### 제어 명령 초안 생성

```http
POST /api/missions/{missionId}/agent/control-drafts
```

Request:

```json
{
  "sopRecommendationId": "uuid",
  "timeWindowMinutes": 10
}
```

### 제어 명령 초안 승인

```http
POST /api/missions/{missionId}/agent/control-drafts/{draftId}/approve
```

Response:

```json
{
  "controlCommandId": "uuid",
  "status": "sent"
}
```

### 제어 명령 초안 반려

```http
POST /api/missions/{missionId}/agent/control-drafts/{draftId}/reject
```

### 관제 Q&A

```http
POST /api/missions/{missionId}/agent/chat
```

Request:

```json
{
  "message": "현재 가장 위험한 이벤트가 뭐야?",
  "timeWindowMinutes": 10
}
```

### 보고서 초안 생성

```http
POST /api/missions/{missionId}/agent/report-draft
```

### Agent run 조회

```http
GET /api/agent/runs/{agentRunId}
```

## DB 확장 후보

AI Agent 기능 구현 시 다음 테이블을 추가한다.

### agent_runs

- id
- mission_id
- user_id
- run_type: situation_summary, risk_assessment, sop_recommendation, control_draft, chat, report_draft
- status: running, completed, failed
- input_payload
- output_payload
- model_name
- prompt_version
- started_at
- completed_at
- error_message

### agent_tool_calls

- id
- agent_run_id
- tool_name
- input_payload
- output_payload
- started_at
- completed_at
- error_message

### situation_summaries

- id
- mission_id
- agent_run_id
- risk_level
- headline
- key_findings
- recommended_checks
- evidence_refs
- generated_at
- acknowledged_by
- acknowledged_at

### sop_recommendations

- id
- mission_id
- agent_run_id
- sop_id
- title
- confidence
- reason
- recommended_actions
- requires_human_approval
- evidence_refs
- generated_at
- acknowledged_by
- acknowledged_at

### control_drafts

- id
- mission_id
- agent_run_id
- sop_recommendation_id
- robot_id
- command_type
- priority
- reason
- command_payload
- evidence_refs
- status: draft, approved, rejected, expired, executed
- requires_approval
- approval_role
- expires_at
- approved_by
- approved_at
- rejected_by
- rejected_at
- control_command_id
- generated_at

### agent_messages

- id
- mission_id
- agent_run_id
- user_id
- role: user, assistant, system, tool
- content
- evidence_refs
- created_at

### sop_documents

- id
- scenario_type
- title
- version
- content
- metadata
- created_at
- updated_at

PoC에서는 `sop_documents`를 seed data로 시작하고, RAG가 필요해지면 chunk/embedding 테이블 또는 외부 vector store를 추가한다.

## UI 반영

### Live Dashboard

추가 패널:

- Situation Summary
- Risk Assessment
- SOP Recommendation
- Control Draft Approval
- Ask Agent

표시 원칙:

- Agent 결과는 탐지/센서 원본과 구분해서 표시한다.
- Agent 권고와 명령 초안은 구분해서 표시한다.
- 명령 초안은 SOP 근거, 관련 이벤트, 만료 시간, 승인 버튼을 함께 표시한다.
- 모든 권고에는 근거 링크가 있어야 한다.
- 데이터가 오래되면 stale 상태를 표시한다.

### Event Detail

추가 항목:

- 이 이벤트가 Agent 요약에 사용되었는지
- 관련 SOP 추천
- 관련 제어 명령 초안
- 관련 Agent run

### Mission Records

추가 항목:

- 상황 요약 이력
- SOP 추천 이력
- 제어 명령 초안 승인/반려 이력
- Agent Q&A 이력
- 보고서 초안

## PoC 우선순위

### P0

- 상황 요약 생성
- 위험도 분석
- read-only tools
- agent_runs/tool_calls 저장
- evidenceRefs 포함

### P1

- SOP 추천
- SOP 기반 제어 명령 초안
- 제어 명령 초안 승인/반려 UX
- 관제 Q&A
- Live Dashboard Agent 패널
- situation_summaries/sop_recommendations/control_drafts 저장

### P2

- 보고서 초안
- RAG 기반 SOP 검색
- streaming response
- multi-agent 구조
- 평가/피드백 루프

## 미확정 항목

- 사용할 LLM provider
- 모델명과 timeout
- SOP seed data 범위
- 위험도 threshold
- 센서별 임계값
- Agent 응답 streaming 여부
- RAG vector store 사용 여부
- Agent 결과 승인/반려 UX
