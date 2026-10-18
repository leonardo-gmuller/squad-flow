export type Kind = string
export type Status = 'backlog' | 'in_progress' | 'in_qa' | 'changes_requested' | 'done'
export type Priority = 'low' | 'medium' | 'high' | 'critical'

export interface Demand {
  id: string
  number: number
  title: string
  description: string
  kind: Kind
  kind_name?: string
  priority: Priority
  status: Status
  assignee: string
  reviewer: string
  assignee_id: string
  reviewer_id: string
  reference: string
  event_codes: string[]
  package_id: string
  release_id: string
  release_name: string
  product: string
  acceptance_criteria: string
  due_date: string | null
  qa_cycle: number
  return_count: number
  created_at: string
  updated_at: string
  completed_at: string | null
}
export interface TestCase {
  id: string
  demand_id: string
  cycle: number
  title: string
  steps: string
  expected: string
  actual: string
  result: 'passed' | 'failed'
  author: string
  author_id: string
  created_at: string
}
export interface Transition {
  id: string
  demand_id: string
  from_status: Status
  to_status: Status
  actor: string
  actor_id: string
  reason: string
  created_at: string
}
export interface Detail {
  demand: Demand
  tests: TestCase[]
  transitions: Transition[]
}
export interface Metrics {
  total: number
  active: number
  in_qa: number
  returned: number
  delivered: number
  first_pass_percent: number
  average_lead_hours: number
  by_kind: Record<string, number>
  by_status: Record<string, number>
  by_priority: Record<string, number>
  overdue: number
  unassigned: number
  total_returns: number
  rework_percent: number
  on_time_percent: number
  deliveries_with_due_date: number
  delivered_30_days: number
  median_lead_hours: number
  p85_lead_hours: number
  average_active_age_hours: number
  daily_deliveries: { date: string; count: number }[]
}

export interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'member'
  created_at: string
}
export interface ReportRequest {
  prompt: string
  demand_id?: string
  from?: string
  to?: string
  previous_report_id?: string
}
export interface GeneratedReport {
  id: string
  demand_id: string
  prompt: string
  created_at: string
  source_data?: Record<string, unknown>
  content: {
    html?: string
    title: string
    summary: string
    metrics: { label: string; value: number; unit: string }[]
    sections: { title: string; body: string }[]
    charts: { title: string; kind: 'bar' | 'line'; points: { label: string; value: number }[] }[]
    tables: { title: string; columns: string[]; rows: string[][] }[]
    recommendations: string[]
    missing_data: string[]
    source_ids: string[]
  }
}
export interface Workstream {
  id: string
  code: string
  name: string
  color: string
  active: boolean
  created_at: string
  updated_at: string
}
export interface WorkstreamInput {
  name: string
  color: string
  active: boolean
}
export interface Member {
  id: string
  name: string
  email: string
}

export interface CreateInput {
  title: string
  description: string
  kind: Kind
  priority: Priority
  assignee_id: string
  reviewer_id: string
  reference: string
  event_codes: string[]
  package_id: string
  release_id: string
  product: string
  acceptance_criteria: string
  due_date: string | null
}

// An OS is a container; only its items are counted in the workflow metrics.
export interface WorkPackage {
  id: string
  reference: string
  title: string
  description: string
  initiative: string
  kind: Kind
  created_at: string
  updated_at: string
  total: number
  done: number
  in_qa: number
}
export interface CreatePackage {
  reference: string
  title: string
  description: string
  initiative: string
  kind: Kind
  create_first_item: boolean
  release_id: string
  assignee_id: string
  reviewer_id: string
  product: string
}

export interface Release {
  id: string
  name: string
  description: string
  release_date: string | null
  archived_at: string | null
  created_at: string
  updated_at: string
  total: number
  done: number
  in_qa: number
}
export interface ReleaseInput {
  name: string
  description: string
  release_date: string | null
}
export interface BoardFilter {
  release_id?: string
  archive?: string
  kind?: string
  owner?: string
  q?: string
  status?: Status
  limit?: number
  offset?: number
}
export interface BoardColumn {
  status: Status
  total: number
  items: Demand[]
}
export interface Board {
  columns: BoardColumn[]
  owners: string[]
}

export interface UserNotification {
  id: string
  demand_id: string
  kind: 'assigned' | 'unassigned' | 'status_changed' | 'qa_passed'
  title: string
  message: string
  read_at: string | null
  created_at: string
}
export interface NotificationPage {
  items: UserNotification[]
  unread_count: number
  next_cursor: string
}

export interface ReportJobEvent {
  id: string
  job_id: string
  type: string
  agent?: string
  message?: string
  attempt: number
  created_at: string
}
export interface ReportJob {
  id: string
  request: ReportRequest
  status: 'queued' | 'processing' | 'completed' | 'failed'
  attempts: number
  max_attempts: number
  available_at: string
  progress_agent: string
  progress_message: string
  error_message: string
  report_id: string
  created_at: string
  updated_at: string
  completed_at: string | null
  events?: ReportJobEvent[]
}
