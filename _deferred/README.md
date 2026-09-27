# Deferred code

이 디렉터리는 2권 single-window core에서 제외한 1권 및 3–4권 POC를 보관한다. `_deferred`로 시작하므로 Go의 `./...` 대상에 포함되지 않는다.

이 source는 옮길 당시의 API를 보존한 archive다. 2권 model과 API를 의도적으로 단순화했으므로, 이 디렉터리만 따로 빌드하는 것은 지원하지 않는다. 재개할 때는 아래 대상 권과 archive source를 먼저 읽고, 현재 2권 core의 ownership 경계를 우회하지 않도록 API와 test를 다시 연결한다.

## 1권 — GUI Application API

| 주제 | 보존된 구현 | 검증 | 재개 시 할 일 |
|---|---|---|---|
| Accessibility traversal | `internal/client/accessibility.go` | archive test | semantic node contract, screen-reader event, keyboard navigation을 추가 |
| App navigation | `internal/app/chat/chat_test.go` | archive test | generic navigation component와 route/state restoration을 추가 |

## 3권 — Advanced Mobile Operating System Features

| 주제 | 보존된 구현 | 검증 | 재개 시 할 일 |
|---|---|---|---|
| Split / multi-display | `internal/server/transition.go`, 기존 simulator/runtime test | archive test | 가변 divider, pane 교체, display 간 window 이동을 다시 설계 |
| History / task | `internal/server/task.go`, 기존 simulator | archive test | multi-task 선택·종료·state restore UI를 추가 |
| Package trust / permission | `internal/server/package.go`, `permission.go` | archive test | installer UI와 capability별 concrete adapter를 추가 |
| Capture / recording | `internal/server/recording.go` | archive test | recording indicator와 capture policy를 추가 |

## 4권 — Mobile Operating System Performance & Optimization

| 주제 | 보존된 구현 | 검증 | 재개 시 할 일 |
|---|---|---|---|
| Resource accounting | `internal/server/quota.go` | archive test | allocation/resource profiling과 pressure signal을 연결 |
| Fault baseline | `internal/server/watchdog.go`, `fault_test.go` | archive test | renderer recovery, visual/performance regression, inspector, hot reload을 추가 |

## 2권에서 유지할 경계

2권은 single-window core의 public contract만 가진다. archive의 feature는 현재 core API에 섞지 않으며, 필요해질 때 선택적으로 되살린다.
