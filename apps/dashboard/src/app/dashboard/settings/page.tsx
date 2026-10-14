"use client";

import { useEffect, useState } from "react";
import DashboardNav from "../../../components/DashboardNav";
import { api, getToken } from "../../../lib/api";

type Me = {
  user: { email: string; role: string };
  tenant: { plan: string; task_credits: number; task_credits_used: number };
};

export default function Settings() {
  const [me, setMe] = useState<Me | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      window.location.href = "/login";
      return;
    }
    api<Me>("/api/v1/me")
      .then(setMe)
      .catch((e) => setError(e.message));
  }, []);

  return (
    <div className="min-h-screen bg-[#F5F0EB]">
      <DashboardNav />
      <main className="max-w-7xl mx-auto p-8">
        <h1 className="text-2xl font-semibold mb-8">Settings</h1>
        {error && <p className="mb-4 text-sm text-red-700">{error}</p>}
        <div className="max-w-xl space-y-6">
          <div className="p-6 bg-white border border-gray-200 rounded-lg">
            <h3 className="font-medium mb-4">Account</h3>
            <div className="flex justify-between items-center py-2 border-b border-gray-100">
              <span className="text-sm">{me?.user.email || "—"}</span>
              <span className="text-xs px-2 py-1 bg-gray-100 rounded">{me?.user.role || "—"}</span>
            </div>
          </div>
          <div className="p-6 bg-white border border-gray-200 rounded-lg">
            <h3 className="font-medium mb-4">Task Credits</h3>
            <p className="text-sm text-gray-600 mb-2">Plan: {me?.tenant.plan || "—"}</p>
            <p className="text-sm text-gray-600 mb-2">Current balance: {me?.tenant.task_credits ?? "—"}</p>
            <p className="text-sm text-gray-600">Used this month: {me?.tenant.task_credits_used ?? "—"}</p>
          </div>
        </div>
      </main>
    </div>
  );
}
