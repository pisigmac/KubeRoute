const TOKEN_KEY = "kuberoute_token";

export function apiBase(): string {
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const res = await fetch(`${apiBase()}${path}`, { ...init, headers });
  if (res.status === 401) {
    clearToken();
    if (typeof window !== "undefined") window.location.href = "/login";
    throw new Error("unauthorized");
  }
  if (res.status === 204) return undefined as T;
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(body.error || res.statusText);
  }
  return body as T;
}

export type Cluster = {
  id: string;
  name: string;
  provider: string;
  region: string;
  status: string;
  node_count: number;
  last_heartbeat?: string;
};

export type Decision = {
  id: string;
  cluster_id: string;
  decision_type: string;
  confidence_score: number;
  reason: string;
  status: string;
  applied: boolean;
  executed?: boolean;
  workload_name?: string;
  workload_namespace?: string;
  workload_kind?: string;
  target_node_name?: string;
  predicted_cost_savings?: number;
  params?: Record<string, string>;
  solver_trace?: SolverTrace;
  proposed_patch?: ProposedPatch;
  revert_of?: string;
  timeline?: TimelineEvent[];
  created_at?: string;
  applied_at?: string;
  executed_at?: string;
};

export type TimelineEvent = {
  at?: string;
  event: string;
  by?: string;
};

export type SolverCandidate = {
  node: string;
  cost_per_hour?: number;
  verdict: string;
  why: string;
};

export type SolverTrace = {
  objective?: string;
  chosen?: string;
  current?: string;
  candidates?: SolverCandidate[];
  constraints?: string[];
  checks?: string[];
  blast_radius?: { kind?: string; namespace?: string; name?: string; workloads?: number };
  savings_per_hour?: number;
};

export type ProposedPatch = {
  op?: string;
  kind?: string;
  namespace?: string;
  name?: string;
  nodeAffinity?: unknown;
};

export type CostReport = {
  total: number;
  optimized: number;
  savings: number;
  breakdown: Record<string, number>;
};
