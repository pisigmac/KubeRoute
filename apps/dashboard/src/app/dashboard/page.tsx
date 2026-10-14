"use client";

import { FormEvent, useEffect, useState } from "react";
import DashboardNav from "../../components/DashboardNav";
import { api, Cluster, getToken } from "../../lib/api";

export default function Dashboard() {
  const [clusters, setClusters] = useState<Cluster[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("kind-local");
  const [provider, setProvider] = useState("kind");
  const [region, setRegion] = useState("local");
  const [busy, setBusy] = useState(false);

  async function load() {
    if (!getToken()) {
      window.location.href = "/login";
      return;
    }
    try {
      const data = await api<{ clusters: Cluster[] }>("/api/v1/clusters");
      setClusters(data.clusters || []);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed to load");
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function register(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api("/api/v1/clusters", {
        method: "POST",
        body: JSON.stringify({ name, provider, region }),
      });
      setOpen(false);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "register failed");
    } finally {
      setBusy(false);
    }
  }

  const nodes = clusters.reduce((n, c) => n + (c.node_count || 0), 0);

  return (
    <div className="min-h-screen bg-[#F5F0EB] text-[#1A1A1A]">
      <DashboardNav />
      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">Cluster Topology</h1>
            <p className="text-sm text-gray-500">Clusters registered with the control plane (live API)</p>
          </div>
          <button
            onClick={() => setOpen(true)}
            className="px-4 py-2 bg-black text-white text-sm font-medium rounded-lg hover:bg-gray-800 transition"
          >
            + Register Cluster
          </button>
        </div>

        {error && <p className="mb-4 text-sm text-red-700">{error}</p>}

        <div className="grid sm:grid-cols-3 gap-5 mb-8">
          <div className="p-5 bg-white border border-gray-200/80 rounded-xl">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Active Clusters</p>
            <p className="text-3xl font-extrabold">{clusters.length}</p>
          </div>
          <div className="p-5 bg-white border border-gray-200/80 rounded-xl">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Managed Nodes</p>
            <p className="text-3xl font-extrabold">{nodes}</p>
          </div>
          <div className="p-5 bg-white border border-gray-200/80 rounded-xl">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Source</p>
            <p className="text-sm font-semibold mt-2">GET /api/v1/clusters</p>
          </div>
        </div>

        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-100">
            <h2 className="font-bold text-base">Registered Clusters</h2>
          </div>
          {clusters.length === 0 ? (
            <p className="p-6 text-sm text-gray-500">No clusters yet. Register one, then run the agent with that cluster id.</p>
          ) : (
            <div className="divide-y divide-gray-100">
              {clusters.map((c) => (
                <div key={c.id} className="p-6 flex items-center justify-between">
                  <div>
                    <h3 className="font-bold text-base">{c.name}</h3>
                    <p className="text-xs text-gray-500 mt-0.5 font-mono">{c.id}</p>
                    <p className="text-xs text-gray-500 mt-0.5">{c.provider} · {c.region}</p>
                  </div>
                  <div className="flex items-center gap-8 text-sm">
                    <div>
                      <p className="text-xs text-gray-400">Nodes</p>
                      <p className="font-semibold">{c.node_count}</p>
                    </div>
                    <span className="px-2.5 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                      {c.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>

      {open && (
        <div className="fixed inset-0 bg-black/40 flex items-center justify-center p-4 z-50">
          <form onSubmit={register} className="bg-white rounded-xl p-6 w-full max-w-md space-y-4">
            <h2 className="font-bold text-lg">Register cluster</h2>
            <label className="block text-sm">
              Name
              <input className="mt-1 w-full border rounded-lg px-3 py-2" value={name} onChange={(e) => setName(e.target.value)} required />
            </label>
            <label className="block text-sm">
              Provider
              <select className="mt-1 w-full border rounded-lg px-3 py-2" value={provider} onChange={(e) => setProvider(e.target.value)}>
                <option value="kind">kind</option>
                <option value="aws">aws</option>
                <option value="gcp">gcp</option>
                <option value="azure">azure</option>
                <option value="on-prem">on-prem</option>
              </select>
            </label>
            <label className="block text-sm">
              Region
              <input className="mt-1 w-full border rounded-lg px-3 py-2" value={region} onChange={(e) => setRegion(e.target.value)} />
            </label>
            <div className="flex justify-end gap-2">
              <button type="button" onClick={() => setOpen(false)} className="px-3 py-2 text-sm border rounded-lg">Cancel</button>
              <button type="submit" disabled={busy} className="px-3 py-2 text-sm bg-black text-white rounded-lg">{busy ? "Saving…" : "Register"}</button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
