"use client";

import { useEffect, useState } from "react";
import DashboardNav from "../../../components/DashboardNav";
import { api, CostReport, getToken } from "../../../lib/api";

export default function Costs() {
  const [report, setReport] = useState<CostReport | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      window.location.href = "/login";
      return;
    }
    api<CostReport>("/api/v1/costs")
      .then(setReport)
      .catch((e) => setError(e.message));
  }, []);

  return (
    <div className="min-h-screen bg-[#F5F0EB]">
      <DashboardNav />
      <main className="max-w-7xl mx-auto p-8">
        <h1 className="text-2xl font-semibold mb-8">Cost Reports</h1>
        {error && <p className="mb-4 text-sm text-red-700">{error}</p>}
        <div className="grid grid-cols-3 gap-6 mb-8">
          <div className="p-6 bg-white border border-gray-200 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">Estimated monthly (nodes)</p>
            <p className="text-2xl font-semibold">${(report?.total ?? 0).toFixed(2)}</p>
          </div>
          <div className="p-6 bg-white border border-gray-200 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">After applied decisions</p>
            <p className="text-2xl font-semibold">${(report?.optimized ?? 0).toFixed(2)}</p>
          </div>
          <div className="p-6 bg-white border border-gray-200 rounded-lg">
            <p className="text-sm text-gray-500 mb-1">Predicted savings</p>
            <p className="text-2xl font-semibold text-green-600">${(report?.savings ?? 0).toFixed(2)}</p>
          </div>
        </div>
        <div className="p-6 bg-white border border-gray-200 rounded-lg">
          <h3 className="font-medium mb-4">Savings by decision type</h3>
          <div className="space-y-3">
            {Object.entries(report?.breakdown || {}).map(([k, v]) => (
              <div key={k} className="flex justify-between text-sm">
                <span>{k}</span>
                <span className="font-medium">${v.toFixed(2)}</span>
              </div>
            ))}
            {Object.keys(report?.breakdown || {}).length === 0 && (
              <p className="text-sm text-gray-500">No applied decisions yet. Numbers come from node cost labels and applied placements.</p>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
