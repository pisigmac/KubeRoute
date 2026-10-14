"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import DashboardNav from "../../../components/DashboardNav";
import { api, Cluster, Decision, getToken } from "../../../lib/api";

export default function Decisions() {
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [clusterId, setClusterId] = useState("");
  const [decisions, setDecisions] = useState<Decision[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      window.location.href = "/login";
      return;
    }
    api<{ clusters: Cluster[] }>("/api/v1/clusters")
      .then((d) => {
        const list = d.clusters || [];
        setClusters(list);
        if (list[0]) setClusterId(list[0].id);
      })
      .catch((e) => setError(e.message));
  }, []);

  useEffect(() => {
    if (!clusterId) return;
    api<{ decisions: Decision[] }>(`/api/v1/clusters/${clusterId}/decisions`)
      .then((d) => setDecisions(d.decisions || []))
      .catch((e) => setError(e.message));
  }, [clusterId]);

  async function act(id: string, action: "apply" | "reject") {
    setBusy(id);
    try {
      await api(`/api/v1/clusters/${clusterId}/decisions/${id}/${action}`, { method: "POST" });
      const d = await api<{ decisions: Decision[] }>(`/api/v1/clusters/${clusterId}/decisions`);
      setDecisions(d.decisions || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed");
    } finally {
      setBusy(null);
    }
  }

  return (
    <div className="min-h-screen bg-[#F5F0EB]">
      <DashboardNav />
      <main className="max-w-7xl mx-auto p-8">
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-semibold">Routing Decisions</h1>
          <select className="border rounded-lg px-3 py-2 text-sm" value={clusterId} onChange={(e) => setClusterId(e.target.value)}>
            {clusters.map((c) => (
              <option key={c.id} value={c.id}>{c.name}</option>
            ))}
          </select>
        </div>
        {error && <p className="mb-4 text-sm text-red-700">{error}</p>}
        <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-200">
              <tr>
                <th className="px-4 py-3 text-left font-medium">ID</th>
                <th className="px-4 py-3 text-left font-medium">Type</th>
                <th className="px-4 py-3 text-left font-medium">Workload</th>
                <th className="px-4 py-3 text-left font-medium">Target</th>
                <th className="px-4 py-3 text-left font-medium">Confidence</th>
                <th className="px-4 py-3 text-left font-medium">Reason</th>
                <th className="px-4 py-3 text-left font-medium">Status</th>
                <th className="px-4 py-3 text-left font-medium">Actions</th>
              </tr>
            </thead>
            <tbody>
              {decisions.map((d) => (
                <tr key={d.id} className="border-b border-gray-100">
                  <td className="px-4 py-3 font-mono text-xs">
                    <Link className="underline" href={`/dashboard/decisions/${d.id}?cluster=${clusterId}`}>
                      {d.id.slice(0, 8)}
                    </Link>
                  </td>
                  <td className="px-4 py-3"><span className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs">{d.decision_type}</span></td>
                  <td className="px-4 py-3 text-xs">{d.workload_namespace}/{d.workload_name}</td>
                  <td className="px-4 py-3 text-xs">{d.target_node_name}</td>
                  <td className="px-4 py-3">{(d.confidence_score * 100).toFixed(0)}%</td>
                  <td className="px-4 py-3 text-gray-600">{d.reason}</td>
                  <td className="px-4 py-3 text-xs">{d.status}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2 items-center">
                      <Link className="px-3 py-1 border border-gray-300 text-xs rounded" href={`/dashboard/decisions/${d.id}?cluster=${clusterId}`}>Review</Link>
                      {d.status === "pending" && (
                        <>
                          <button disabled={busy === d.id} onClick={() => act(d.id, "apply")} className="px-3 py-1 bg-[#1A1A1A] text-white text-xs rounded">Apply</button>
                          <button disabled={busy === d.id} onClick={() => act(d.id, "reject")} className="px-3 py-1 border border-gray-300 text-xs rounded">Reject</button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {decisions.length === 0 && <p className="p-6 text-sm text-gray-500">No decisions for this cluster.</p>}
        </div>
      </main>
    </div>
  );
}
