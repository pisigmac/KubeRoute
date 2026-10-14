"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import DashboardNav from "../../../../components/DashboardNav";
import { api, Decision, getToken } from "../../../../lib/api";

function clusterFromQuery(): string {
  if (typeof window === "undefined") return "";
  return new URLSearchParams(window.location.search).get("cluster") || "";
}

function fmtWhen(iso?: string | null) {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function affinityLines(patch: Decision["proposed_patch"], currentNode?: string, proposed = true): string {
  if (!proposed) {
    if (!currentNode) return "nodeAffinity: (none / unscheduled)";
    return `required: kubernetes.io/hostname In [${currentNode}]`;
  }
  if (!patch) return "(no patch)";
  if (patch.op === "unpin" || patch.nodeAffinity == null) return "nodeAffinity: null (unpin)";
  try {
    return JSON.stringify(patch.nodeAffinity, null, 2);
  } catch {
    return String(patch.nodeAffinity);
  }
}

export default function DecisionPR() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const [clusterId, setClusterId] = useState("");
  const [d, setD] = useState<Decision | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function load(cid: string) {
    const row = await api<Decision>(`/api/v1/clusters/${cid}/decisions/${id}`);
    setD(row);
  }

  useEffect(() => {
    if (!getToken()) {
      window.location.href = "/login";
      return;
    }
    const cid = clusterFromQuery();
    if (!cid) {
      setError("Missing cluster query param. Open this page from the decisions list.");
      return;
    }
    setClusterId(cid);
    load(cid).catch((e) => setError(e.message));
  }, [id]);

  async function act(action: "apply" | "reject" | "revert") {
    if (!clusterId) return;
    setBusy(true);
    setError(null);
    try {
      const res = await api<Decision | { status: string }>(
        `/api/v1/clusters/${clusterId}/decisions/${id}/${action}`,
        { method: "POST" }
      );
      if (action === "revert" && res && "id" in res) {
        window.location.href = `/dashboard/decisions/${res.id}?cluster=${clusterId}`;
        return;
      }
      await load(clusterId);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed");
    } finally {
      setBusy(false);
    }
  }

  const current = d?.params?.current_node || d?.solver_trace?.current || "";
  const target = d?.target_node_name || d?.solver_trace?.chosen || "";
  const ns = d?.workload_namespace || d?.params?.namespace || "default";
  const name = d?.workload_name || d?.params?.workload || "";
  const savingsHr = d?.solver_trace?.savings_per_hour ?? (d?.predicted_cost_savings ? d.predicted_cost_savings / 730 : 0);
  const candidates = d?.solver_trace?.candidates || [];
  const title = useMemo(() => {
    if (!d) return "Placement";
    if (d.params?.op === "unpin") return `Unpin ${ns}/${name}`;
    return `Move ${ns}/${name} from ${current || "unscheduled"} → ${target || "?"}`;
  }, [d, ns, name, current, target]);

  return (
    <div className="min-h-screen bg-[#F5F0EB]">
      <DashboardNav />
      <main className="max-w-5xl mx-auto p-8">
        <p className="text-sm mb-4">
          <Link href="/dashboard/decisions" className="underline text-gray-600">
            ← Decisions
          </Link>
        </p>
        {error && <p className="mb-4 text-sm text-red-700">{error}</p>}
        {!d && !error && <p className="text-sm text-gray-500">Loading…</p>}
        {d && (
          <>
            <div className="flex flex-wrap items-start justify-between gap-4 mb-6">
              <div>
                <p className="text-xs font-mono text-gray-500 mb-1">{d.id}</p>
                <h1 className="text-2xl font-semibold">{title}</h1>
                <p className="text-sm text-gray-600 mt-1">{d.reason}</p>
              </div>
              <span className="px-3 py-1 rounded-full text-xs bg-white border border-gray-200 uppercase tracking-wide">
                {d.status}
              </span>
            </div>

            <div className="grid sm:grid-cols-3 gap-4 mb-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <p className="text-xs text-gray-500 uppercase">$/hour impact</p>
                <p className="text-xl font-semibold mt-1">
                  {savingsHr >= 0 ? "+" : ""}
                  ${savingsHr.toFixed(4)}/hr
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  ~${(d.predicted_cost_savings || 0).toFixed(2)} / month
                </p>
              </div>
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <p className="text-xs text-gray-500 uppercase">Blast radius</p>
                <p className="text-sm mt-2">
                  {d.solver_trace?.blast_radius?.kind || d.workload_kind || "Deployment"}{" "}
                  {ns}/{name}
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  {d.solver_trace?.blast_radius?.workloads ?? 1} workload
                </p>
              </div>
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <p className="text-xs text-gray-500 uppercase">Checks</p>
                <ul className="mt-2 text-sm space-y-1">
                  {(d.solver_trace?.checks || ["fits CPU/mem on target"]).map((c) => (
                    <li key={c}>✓ {c}</li>
                  ))}
                </ul>
              </div>
            </div>

            <div className="grid md:grid-cols-2 gap-4 mb-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <p className="text-xs text-gray-500 uppercase mb-2">Current affinity</p>
                <pre className="text-xs overflow-auto whitespace-pre-wrap">{affinityLines(d.proposed_patch, current, false)}</pre>
              </div>
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <p className="text-xs text-gray-500 uppercase mb-2">Proposed patch</p>
                <pre className="text-xs overflow-auto whitespace-pre-wrap">{affinityLines(d.proposed_patch, target, true)}</pre>
              </div>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg overflow-hidden mb-6">
              <div className="px-4 py-3 border-b border-gray-100">
                <h2 className="font-medium">Solver trace</h2>
                <p className="text-xs text-gray-500">
                  {d.solver_trace?.objective || "minimize_hourly_cost"}
                  {d.solver_trace?.constraints?.length
                    ? ` · ${d.solver_trace.constraints.join(", ")}`
                    : ""}
                </p>
              </div>
              <table className="w-full text-sm">
                <thead className="bg-gray-50 border-b border-gray-200">
                  <tr>
                    <th className="px-4 py-2 text-left font-medium">Node</th>
                    <th className="px-4 py-2 text-left font-medium">$/hr</th>
                    <th className="px-4 py-2 text-left font-medium">Verdict</th>
                    <th className="px-4 py-2 text-left font-medium">Why</th>
                  </tr>
                </thead>
                <tbody>
                  {candidates.map((c) => (
                    <tr key={c.node} className="border-b border-gray-100">
                      <td className="px-4 py-2 font-mono text-xs">{c.node}</td>
                      <td className="px-4 py-2">{c.cost_per_hour ?? "—"}</td>
                      <td className="px-4 py-2">
                        <span
                          className={`px-2 py-0.5 rounded text-xs ${
                            c.verdict === "chosen"
                              ? "bg-green-50 text-green-800"
                              : c.verdict === "current"
                                ? "bg-amber-50 text-amber-800"
                                : "bg-gray-100 text-gray-700"
                          }`}
                        >
                          {c.verdict}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-gray-600">{c.why}</td>
                    </tr>
                  ))}
                  {candidates.length === 0 && (
                    <tr>
                      <td className="px-4 py-4 text-sm text-gray-500" colSpan={4}>
                        No candidate rows (revert or scaling).
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-4 mb-6">
              <h2 className="font-medium mb-3">Timeline</h2>
              <ol className="text-sm space-y-2">
                {(d.timeline || []).map((ev, i) => (
                  <li key={i} className="flex gap-3">
                    <span className="font-mono text-xs text-gray-500 w-40 shrink-0">{fmtWhen(ev.at)}</span>
                    <span>{ev.event}{ev.by ? ` · ${ev.by}` : ""}</span>
                  </li>
                ))}
              </ol>
            </div>

            <div className="flex gap-2">
              {d.status === "pending" && (
                <>
                  <button
                    disabled={busy}
                    onClick={() => act("apply")}
                    className="px-4 py-2 bg-[#1A1A1A] text-white text-sm rounded-lg"
                  >
                    Apply
                  </button>
                  <button
                    disabled={busy}
                    onClick={() => act("reject")}
                    className="px-4 py-2 border border-gray-300 text-sm rounded-lg"
                  >
                    Reject
                  </button>
                </>
              )}
              {(d.status === "applied" || d.status === "executed") && !d.revert_of && (
                <button
                  disabled={busy}
                  onClick={() => act("revert")}
                  className="px-4 py-2 border border-gray-300 text-sm rounded-lg"
                >
                  Revert
                </button>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  );
}
